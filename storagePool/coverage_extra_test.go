package storagePool

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Hari-Kiri/virest/utilities"
	"libvirt.org/go/libvirt"
)

func TestConnectPositiveNegative(t *testing.T) {
	orig := connectWithAuth
	t.Cleanup(func() { connectWithAuth = orig })

	connectWithAuth = func(uri string, auth *libvirt.ConnectAuth, flags libvirt.ConnectFlags) (utilities.Connection, error) {
		return utilities.Connection{}, errors.New("dial fail")
	}
	if _, err := Connect("qemu:///system"); err == nil {
		t.Fatal("expected dial error")
	}

	connectWithAuth = func(uri string, auth *libvirt.ConnectAuth, flags libvirt.ConnectFlags) (utilities.Connection, error) {
		return utilities.Connection{Connect: nil}, nil
	}
	if _, err := Connect("qemu:///system"); !errors.Is(err, errNilLibvirtConnect) {
		t.Fatalf("want nil-conn sentinel, got %v", err)
	}
}

func TestLookupClosedConnection(t *testing.T) {
	conn := &Connection{}
	if _, err := conn.lookupPool("x"); !errors.Is(err, errConnectionClosed) {
		t.Fatalf("got %v", err)
	}
	var nilConn *Connection
	if _, err := nilConn.lookupPool("x"); !errors.Is(err, errConnectionClosed) {
		t.Fatalf("got %v", err)
	}
}

func TestFreePoolsFromAndFinishHelpers(t *testing.T) {
	ok := &fakePool{}
	bad := &fakePool{freeErr: errors.New("free fail")}
	err := freePoolsFrom([]poolHandle{ok, bad}, 0)
	if err == nil {
		t.Fatal("expected free error")
	}

	var primary error
	finishFree(&fakePool{freeErr: errors.New("x")}, &primary)
	if primary == nil {
		t.Fatal("expected finishFree to set err")
	}
	primary = errors.New("keep")
	finishFree(&fakePool{freeErr: errors.New("x")}, &primary)
	if primary.Error() != "keep" {
		t.Fatalf("should keep primary: %v", primary)
	}

	hv := newFakeHypervisor()
	hv.deregisterErr = errors.New("dereg")
	var derr error
	finishDeregister(hv, 7, &derr)
	if derr == nil {
		t.Fatal("expected deregister error")
	}
}

func TestListMidFailureFreesRemainder(t *testing.T) {
	good := &fakePool{
		name: "a",
		uuid: "a",
		xmlDesc: `<pool type='dir'><name>a</name><uuid>a</uuid><capacity unit='bytes'>1</capacity></pool>`,
		info: &libvirt.StoragePoolInfo{State: libvirt.STORAGE_POOL_RUNNING, Capacity: 1},
	}
	bad := &fakePool{getInfoErr: errors.New("info fail"), xmlDesc: `<pool type='dir'><name>b</name><uuid>b</uuid></pool>`}
	rest := &fakePool{name: "c"}
	hv := newFakeHypervisor()
	hv.listPools = []poolHandle{good, bad, rest}
	conn := newConnectionForTest(hv)
	if _, err := conn.List(0, 0); err == nil {
		t.Fatal("expected error")
	}
	if rest.freeCalls < 1 {
		t.Fatalf("expected remainder free, freeCalls=%d", rest.freeCalls)
	}
}

func TestListFreeError(t *testing.T) {
	pool := &fakePool{
		name: "p",
		uuid: "u",
		xmlDesc: `<pool type='dir'><name>p</name><uuid>u</uuid></pool>`,
		info:    &libvirt.StoragePoolInfo{},
		freeErr: errors.New("free boom"),
	}
	// first Free is after summarize (withPoolRef frees refs); last Free on list item fails
	hv := newFakeHypervisor()
	hv.listPools = []poolHandle{pool}
	conn := newConnectionForTest(hv)
	// summarize uses Ref/Free; then List calls Free again — set freeErr after summarize by using freeErr always
	if _, err := conn.List(0, 0); err == nil {
		// may fail in withPoolRef free or list free — either is fine as long as error surfaces
		t.Fatal("expected free-related error")
	}
}

func TestInfoRefFailure(t *testing.T) {
	pool := &fakePool{refErr: errors.New("ref fail")}
	hv := newFakeHypervisor()
	hv.registerPool("u", pool)
	conn := newConnectionForTest(hv)
	if _, err := conn.Info("u"); err == nil {
		t.Fatal("expected ref error")
	}
}

func TestDumpXMLErrors(t *testing.T) {
	pool := &fakePool{getXMLErr: errors.New("xml fail")}
	hv := newFakeHypervisor()
	hv.registerPool("u", pool)
	conn := newConnectionForTest(hv)
	if _, err := conn.Dump("u", 0); err == nil {
		t.Fatal("expected xml error")
	}

	pool.getXMLErr = nil
	pool.xmlDesc = "not-xml"
	if _, err := conn.Dump("u", 0); err == nil {
		t.Fatal("expected unmarshal error")
	}
}

func TestFindSourcesUnmarshalError(t *testing.T) {
	hv := newFakeHypervisor()
	hv.findSourcesXML = "bad"
	conn := newConnectionForTest(hv)
	if _, err := conn.FindSources(utilities.PoolTypeDir, SourceSpec{}); err == nil {
		t.Fatal("expected unmarshal error")
	}
}

func TestWaitRefreshEvent(t *testing.T) {
	pool := &fakePool{}
	hv := newFakeHypervisor()
	hv.registerPool("u", pool)
	conn := newConnectionForTest(hv)

	done := make(chan Event, 1)
	go func() {
		ev, err := conn.WaitEvent("u", EventRefresh)
		if err != nil {
			t.Errorf("wait refresh: %v", err)
			return
		}
		done <- ev
	}()

	deadline := time.After(2 * time.Second)
	for {
		hv.mu.Lock()
		ready := hv.refreshCallback != nil
		hv.mu.Unlock()
		if ready {
			break
		}
		select {
		case <-deadline:
			t.Fatal("refresh callback not registered")
		case <-time.After(5 * time.Millisecond):
		}
	}
	hv.fireRefresh()
	select {
	case ev := <-done:
		if ev.EventRefresh != 1 {
			t.Fatalf("refresh=%d", ev.EventRefresh)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}
}

func TestWaitEventRegisterErrors(t *testing.T) {
	pool := &fakePool{}
	hv := newFakeHypervisor()
	hv.registerPool("u", pool)
	hv.lifecycleRegisterErr = errors.New("reg fail")
	conn := newConnectionForTest(hv)
	if _, err := conn.WaitEvent("u", EventLifecycle); err == nil {
		t.Fatal("expected register error")
	}

	hv.lifecycleRegisterErr = nil
	hv.refreshRegisterErr = errors.New("reg fail")
	if _, err := conn.WaitEvent("u", EventRefresh); err == nil {
		t.Fatal("expected refresh register error")
	}
}

func TestStreamEventsEmitErrorAndLifecycle(t *testing.T) {
	pool := &fakePool{}
	hv := newFakeHypervisor()
	hv.registerPool("u", pool)
	conn := newConnectionForTest(hv)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errCh := make(chan error, 1)
	go func() {
		errCh <- conn.StreamEvents(ctx, "u", EventLifecycle, func(ev Event) error {
			return errors.New("emit fail")
		})
	}()

	deadline := time.After(2 * time.Second)
	for {
		hv.mu.Lock()
		ready := hv.lifecycleCallback != nil
		hv.mu.Unlock()
		if ready {
			break
		}
		select {
		case <-deadline:
			t.Fatal("callback not registered")
		case <-time.After(5 * time.Millisecond):
		}
	}
	hv.fireLifecycle(libvirt.StoragePoolEventLifecycle{Event: libvirt.STORAGE_POOL_EVENT_STARTED})
	select {
	case err := <-errCh:
		if err == nil {
			t.Fatal("expected emit error")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}

	if err := conn.StreamEvents(ctx, "u", 99, func(Event) error { return nil }); err == nil {
		t.Fatal("expected unsupported kind")
	}
}

func TestDefineMarshalEdge(t *testing.T) {
	// empty name/type still marshals; force define path with free error on defined pool
	defined := &fakePool{uuid: "u", freeErr: errors.New("free defined")}
	hv := newFakeHypervisor()
	hv.definePool = defined
	conn := newConnectionForTest(hv)
	if _, err := conn.Define(utilities.StoragePool{Name: "p", Type: utilities.PoolTypeDir}, 0); err == nil {
		t.Fatal("expected free error after define")
	}
}

func TestCreateTransientMarshalEdge(t *testing.T) {
	created := &fakePool{uuid: "u", freeErr: errors.New("free created")}
	hv := newFakeHypervisor()
	hv.createTransientPool = created
	conn := newConnectionForTest(hv)
	if _, err := conn.CreateTransient(utilities.StoragePool{Name: "p", Type: utilities.PoolTypeDir}, 0); err == nil {
		t.Fatal("expected free error after create transient")
	}
}

func TestRunParallelEmpty(t *testing.T) {
	if err := runParallel(); err != nil {
		t.Fatalf("empty parallel: %v", err)
	}
}
