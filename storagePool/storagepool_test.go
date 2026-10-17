package storagePool

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Hari-Kiri/virest/utilities"
	"libvirt.org/go/libvirt"
)

func TestCloseNilAndClosed(t *testing.T) {
	var c *Connection
	if err := c.Close(); err != nil {
		t.Fatalf("nil close: %v", err)
	}

	hv := newFakeHypervisor()
	conn := newConnectionForTest(hv)
	if err := conn.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	if hv.closeCalls != 1 {
		t.Fatalf("closeCalls=%d", hv.closeCalls)
	}
	if err := conn.Close(); err != nil {
		t.Fatalf("second close: %v", err)
	}
}

func TestCloseError(t *testing.T) {
	hv := newFakeHypervisor()
	hv.closeErr = errors.New("boom")
	conn := newConnectionForTest(hv)
	if err := conn.Close(); err == nil {
		t.Fatal("expected error")
	}
}

func TestDefineSuccessAndFailures(t *testing.T) {
	defined := &fakePool{uuid: "u-1"}
	hv := newFakeHypervisor()
	hv.definePool = defined
	conn := newConnectionForTest(hv)

	uuid, err := conn.Define(utilities.StoragePool{Name: "p", Type: utilities.PoolTypeDir}, 0)
	if err != nil {
		t.Fatalf("define: %v", err)
	}
	if uuid != "u-1" {
		t.Fatalf("uuid=%s", uuid)
	}
	if defined.freeCalls != 1 {
		t.Fatalf("defined freeCalls=%d", defined.freeCalls)
	}
	if hv.defineXML == "" {
		t.Fatal("expected define xml")
	}

	hv.defineErr = errors.New("define failed")
	if _, err := conn.Define(utilities.StoragePool{Name: "p", Type: utilities.PoolTypeDir}, 1); err == nil {
		t.Fatal("expected define error")
	}

	hv.defineErr = nil
	hv.definePool = &fakePool{getUUIDErr: errors.New("uuid fail")}
	if _, err := conn.Define(utilities.StoragePool{Name: "p", Type: utilities.PoolTypeDir}, 0); err == nil {
		t.Fatal("expected uuid error")
	}
}

func TestCreateTransientSuccessAndFailures(t *testing.T) {
	created := &fakePool{uuid: "u-2"}
	hv := newFakeHypervisor()
	hv.createTransientPool = created
	conn := newConnectionForTest(hv)

	uuid, err := conn.CreateTransient(utilities.StoragePool{Name: "p", Type: utilities.PoolTypeDir}, 0)
	if err != nil {
		t.Fatalf("create transient: %v", err)
	}
	if uuid != "u-2" {
		t.Fatalf("uuid=%s", uuid)
	}
	if created.freeCalls != 1 {
		t.Fatalf("created freeCalls=%d", created.freeCalls)
	}
	if hv.createTransientConfig == "" {
		t.Fatal("expected create transient config")
	}

	hv.createTransientErr = errors.New("create transient failed")
	if _, err := conn.CreateTransient(utilities.StoragePool{Name: "p", Type: utilities.PoolTypeDir}, 1); err == nil {
		t.Fatal("expected create transient error")
	}

	hv.createTransientErr = nil
	hv.createTransientPool = &fakePool{getUUIDErr: errors.New("uuid fail")}
	if _, err := conn.CreateTransient(utilities.StoragePool{Name: "p", Type: utilities.PoolTypeDir}, 0); err == nil {
		t.Fatal("expected uuid error")
	}
}

func TestLifecycleOps(t *testing.T) {
	pool := &fakePool{}
	hv := newFakeHypervisor()
	hv.registerPool("abc", pool)
	conn := newConnectionForTest(hv)

	tests := []struct {
		name string
		call func() error
		check func(t *testing.T)
	}{
		{
			name: "build",
			call: func() error { return conn.Build("abc", utilities.Flags.Build.New) },
			check: func(t *testing.T) {
				if pool.buildFlags != libvirt.STORAGE_POOL_BUILD_NEW {
					t.Fatalf("flags=%v", pool.buildFlags)
				}
			},
		},
		{
			name: "start",
			call: func() error { return conn.Start("abc", utilities.Flags.Create.Normal) },
			check: func(t *testing.T) {
				if pool.createFlags != libvirt.STORAGE_POOL_CREATE_NORMAL {
					t.Fatalf("flags=%v", pool.createFlags)
				}
			},
		},
		{
			name: "destroy",
			call: func() error { return conn.Destroy("abc") },
		},
		{
			name: "delete",
			call: func() error { return conn.Delete("abc", utilities.Flags.Delete.Normal) },
			check: func(t *testing.T) {
				if pool.deleteFlags != libvirt.STORAGE_POOL_DELETE_NORMAL {
					t.Fatalf("flags=%v", pool.deleteFlags)
				}
			},
		},
		{
			name: "undefine",
			call: func() error { return conn.Undefine("abc") },
		},
		{
			name: "refresh",
			call: func() error { return conn.Refresh("abc") },
			check: func(t *testing.T) {
				if pool.refreshFlags != 0 {
					t.Fatalf("flags=%v", pool.refreshFlags)
				}
			},
		},
		{
			name: "autostart",
			call: func() error { return conn.Autostart("abc", true) },
			check: func(t *testing.T) {
				if pool.setAutostartValue == nil || !*pool.setAutostartValue {
					t.Fatal("expected autostart true")
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pool.freeCalls = 0
			if err := tc.call(); err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
			if pool.freeCalls != 1 {
				t.Fatalf("freeCalls=%d", pool.freeCalls)
			}
			if tc.check != nil {
				tc.check(t)
			}
		})
	}

	hv.lookupErr = errors.New("missing")
	if err := conn.Build("abc", 0); err == nil {
		t.Fatal("expected lookup error")
	}

	hv.lookupErr = nil
	pool.buildErr = errors.New("build fail")
	if err := conn.Build("abc", 0); err == nil {
		t.Fatal("expected build error")
	}
}

func TestDumpSuccessAndFailures(t *testing.T) {
	pool := &fakePool{
		xmlDesc: `<pool type='dir'><name>demo</name><uuid>u-1</uuid></pool>`,
	}
	hv := newFakeHypervisor()
	hv.registerPool("u-1", pool)
	conn := newConnectionForTest(hv)

	detail, err := conn.Dump("u-1", 0)
	if err != nil {
		t.Fatalf("dump: %v", err)
	}
	if detail.Name != "demo" {
		t.Fatalf("name=%s", detail.Name)
	}

	pool.getXMLErr = errors.New("xml fail")
	if _, err := conn.Dump("u-1", 0); err == nil {
		t.Fatal("expected xml error")
	}

	pool.getXMLErr = nil
	pool.xmlDesc = "not-xml"
	if _, err := conn.Dump("u-1", 0); err == nil {
		t.Fatal("expected unmarshal error")
	}
}

func TestInfoSuccessAndFieldFailure(t *testing.T) {
	pool := &fakePool{
		name:       "n1",
		autostart:  true,
		persistent: true,
		info: &libvirt.StoragePoolInfo{
			State:      libvirt.STORAGE_POOL_RUNNING,
			Capacity:   100,
			Allocation: 40,
			Available:  60,
		},
	}
	hv := newFakeHypervisor()
	hv.registerPool("u", pool)
	conn := newConnectionForTest(hv)

	info, err := conn.Info("u")
	if err != nil {
		t.Fatalf("info: %v", err)
	}
	if info.Name != "n1" || info.Capacity != 100 || !info.Autostart || !info.Persistent {
		t.Fatalf("unexpected info: %+v", info)
	}
	if pool.freeCalls < 1 {
		t.Fatal("expected original pool free")
	}

	pool.getNameErr = errors.New("name fail")
	if _, err := conn.Info("u"); err == nil {
		t.Fatal("expected name error")
	}
}

func TestListEmptySingleAndFailure(t *testing.T) {
	hv := newFakeHypervisor()
	conn := newConnectionForTest(hv)

	list, err := conn.List(0, 0)
	if err != nil {
		t.Fatalf("empty list: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("len=%d", len(list))
	}

	pool := &fakePool{
		xmlDesc: `<pool type='dir'><name>p</name><uuid>u</uuid><capacity unit='bytes'>10</capacity><allocation unit='bytes'>4</allocation><available unit='bytes'>6</available></pool>`,
		info:    &libvirt.StoragePoolInfo{State: libvirt.STORAGE_POOL_RUNNING},
	}
	hv.listPools = []poolHandle{pool}
	list, err = conn.List(3, 1)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if hv.listFlagsSeen != 3 {
		t.Fatalf("flags=%v", hv.listFlagsSeen)
	}
	if len(list) != 1 || list[0].Name != "p" || list[0].Capacity.Value != 10 {
		t.Fatalf("unexpected: %+v", list)
	}
	if pool.freeCalls < 1 {
		t.Fatal("expected pool free")
	}

	hv.listErr = errors.New("list fail")
	if _, err := conn.List(0, 0); err == nil {
		t.Fatal("expected list error")
	}

	hv.listErr = nil
	bad := &fakePool{getInfoErr: errors.New("info fail"), xmlDesc: `<pool type='dir'><name>p</name><uuid>u</uuid></pool>`}
	hv.listPools = []poolHandle{bad}
	if _, err := conn.List(0, 0); err == nil {
		t.Fatal("expected per-pool error")
	}
}

func TestCapabilitiesAndFindSources(t *testing.T) {
	hv := newFakeHypervisor()
	hv.capabilitiesXML = `<storagepoolCapabilities><pool type='dir' supported='yes'></pool></storagepoolCapabilities>`
	conn := newConnectionForTest(hv)

	caps, err := conn.Capabilities()
	if err != nil {
		t.Fatalf("caps: %v", err)
	}
	if len(caps.Pool) != 1 || caps.Pool[0].Type != utilities.PoolTypeDir.String() {
		t.Fatalf("caps=%+v", caps)
	}

	hv.capabilitiesErr = errors.New("caps fail")
	if _, err := conn.Capabilities(); err == nil {
		t.Fatal("expected caps error")
	}

	hv.capabilitiesErr = nil
	hv.capabilitiesXML = "bad"
	if _, err := conn.Capabilities(); err == nil {
		t.Fatal("expected unmarshal error")
	}

	hv.findSourcesXML = `<sources></sources>`
	sources, err := conn.FindSources(utilities.PoolTypeNetFS, SourceSpec{Host: Host{Name: "h", Port: 2049}})
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if hv.findPoolType != utilities.PoolTypeNetFS.String() || sources.Source == nil && len(sources.Source) != 0 {
		// empty sources is fine
	}
	if hv.findSrcSpec == "" {
		t.Fatal("expected src spec xml")
	}

	hv.findSourcesErr = errors.New("find fail")
	if _, err := conn.FindSources(utilities.PoolTypeNetFS, SourceSpec{}); err == nil {
		t.Fatal("expected find error")
	}
}

func TestWaitEventAndStreamEvents(t *testing.T) {
	pool := &fakePool{}
	hv := newFakeHypervisor()
	hv.registerPool("u", pool)
	conn := newConnectionForTest(hv)

	if _, err := conn.WaitEvent("u", 99); err == nil {
		t.Fatal("expected invalid kind")
	}

	done := make(chan Event, 1)
	go func() {
		ev, err := conn.WaitEvent("u", EventLifecycle)
		if err != nil {
			t.Errorf("wait: %v", err)
			return
		}
		done <- ev
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
	hv.fireLifecycle(libvirt.StoragePoolEventLifecycle{Event: libvirt.STORAGE_POOL_EVENT_STARTED, Detail: 1})

	select {
	case ev := <-done:
		if ev.EventLifecycle != PoolEventLifecycleStarted {
			t.Fatalf("event=%v", ev.EventLifecycle)
		}
		if len(hv.deregisterCalls) == 0 {
			t.Fatal("expected deregister")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for event")
	}

	ctx, cancel := context.WithCancel(context.Background())
	got := make(chan Event, 1)
	errCh := make(chan error, 1)
	go func() {
		errCh <- conn.StreamEvents(ctx, "u", EventRefresh, func(ev Event) error {
			got <- ev
			return nil
		})
	}()

	deadline = time.After(2 * time.Second)
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
	case ev := <-got:
		if ev.EventRefresh != 1 {
			t.Fatalf("refresh=%d", ev.EventRefresh)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout stream event")
	}
	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("stream err: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("stream did not exit")
	}

	if err := conn.StreamEvents(context.Background(), "u", EventLifecycle, nil); err == nil {
		t.Fatal("expected nil emit error")
	}
}

func TestErrorsAsLibvirt(t *testing.T) {
	hv := newFakeHypervisor()
	hv.lookupErr = libvirt.Error{Code: libvirt.ERR_NO_STORAGE_POOL, Message: "gone"}
	conn := newConnectionForTest(hv)
	err := conn.Destroy("x")
	if err == nil {
		t.Fatal("expected error")
	}
	lv, ok := utilities.AsLibvirtError(err)
	if !ok {
		t.Fatalf("AsLibvirtError failed: %v", err)
	}
	if lv.Code != libvirt.ERR_NO_STORAGE_POOL {
		t.Fatalf("code=%v", lv.Code)
	}
	if code, ok := utilities.Code(err); !ok || code != libvirt.ERR_NO_STORAGE_POOL {
		t.Fatalf("Code=%v ok=%v", code, ok)
	}

	plain := wrap("op", errors.New("plain"))
	if _, ok := utilities.AsLibvirtError(plain); ok {
		t.Fatal("plain error should not be libvirt.Error")
	}
}

func TestLookupByNameAndUUID(t *testing.T) {
	const uuid = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
	pool := &fakePool{uuid: uuid, name: "demo"}
	hv := newFakeHypervisor()
	hv.registerPool(uuid, pool)
	hv.registerPool("demo", pool)
	conn := newConnectionForTest(hv)

	if err := conn.Destroy("demo"); err != nil {
		t.Fatalf("destroy by name: %v", err)
	}
	if err := conn.Destroy(uuid); err != nil {
		t.Fatalf("destroy by uuid: %v", err)
	}
	if err := conn.Destroy("missing-pool"); err == nil {
		t.Fatal("expected missing name error")
	}
}

func TestStartDumpNameUUIDByName(t *testing.T) {
	pool := &fakePool{
		uuid:    "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		name:    "demo",
		xmlDesc: `<pool type='dir'><name>demo</name><uuid>a1b2c3d4-e5f6-7890-abcd-ef1234567890</uuid></pool>`,
	}
	hv := newFakeHypervisor()
	hv.registerPool("demo", pool)
	hv.registerPool(pool.uuid, pool)
	conn := newConnectionForTest(hv)

	if err := conn.Start("demo", 0); err != nil {
		t.Fatalf("start: %v", err)
	}
	detail, err := conn.Dump("demo", 0)
	if err != nil {
		t.Fatalf("dump: %v", err)
	}
	if detail.Name != "demo" {
		t.Fatalf("name=%s", detail.Name)
	}
	name, err := conn.Name(pool.uuid)
	if err != nil {
		t.Fatalf("name: %v", err)
	}
	if name != "demo" {
		t.Fatalf("got name %s", name)
	}
	uuid, err := conn.UUIDByName("demo")
	if err != nil {
		t.Fatalf("uuid by name: %v", err)
	}
	if uuid != pool.uuid {
		t.Fatalf("uuid=%s", uuid)
	}
	if _, err := conn.Name("nope"); err == nil {
		t.Fatal("expected name lookup error")
	}
}

func TestListPresets(t *testing.T) {
	hv := newFakeHypervisor()
	conn := newConnectionForTest(hv)
	if _, err := conn.ListActive(0); err != nil {
		t.Fatalf("ListActive: %v", err)
	}
	if hv.listFlagsSeen != libvirt.CONNECT_LIST_STORAGE_POOLS_ACTIVE {
		t.Fatalf("active flags=%v", hv.listFlagsSeen)
	}
	if _, err := conn.ListInactive(0); err != nil {
		t.Fatalf("ListInactive: %v", err)
	}
	if hv.listFlagsSeen != libvirt.CONNECT_LIST_STORAGE_POOLS_INACTIVE {
		t.Fatalf("inactive flags=%v", hv.listFlagsSeen)
	}
	if _, err := conn.ListAll(0); err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	want := libvirt.CONNECT_LIST_STORAGE_POOLS_ACTIVE | libvirt.CONNECT_LIST_STORAGE_POOLS_INACTIVE
	if hv.listFlagsSeen != want {
		t.Fatalf("all flags=%v want %v", hv.listFlagsSeen, want)
	}
}

func TestDefineAsCreateAsFindSourcesAs(t *testing.T) {
	defined := &fakePool{uuid: "u-as"}
	hv := newFakeHypervisor()
	hv.definePool = defined
	hv.createTransientPool = &fakePool{uuid: "u-cas"}
	hv.findSourcesXML = `<sources></sources>`
	conn := newConnectionForTest(hv)

	params := utilities.DefineAsParams{Name: "images", Type: utilities.PoolTypeDir, TargetPath: "/var/lib/libvirt/images"}
	uuid, err := conn.DefineAs(params, 0)
	if err != nil {
		t.Fatalf("DefineAs: %v", err)
	}
	if uuid != "u-as" {
		t.Fatalf("uuid=%s", uuid)
	}
	if hv.defineXML == "" {
		t.Fatal("expected define xml")
	}

	uuid, err = conn.CreateAs(utilities.CreateAsParams(params), 0)
	if err != nil {
		t.Fatalf("CreateAs: %v", err)
	}
	if uuid != "u-cas" {
		t.Fatalf("uuid=%s", uuid)
	}

	if _, err := conn.DefineAs(utilities.DefineAsParams{Type: utilities.PoolTypeDir}, 0); err == nil {
		t.Fatal("expected DefineAs validation error")
	}

	_, err = conn.FindSourcesAs(utilities.PoolTypeISCSI, utilities.FindSourcesAsParams{
		Host: "1.2.3.4", Port: 3260, Initiator: "iqn.1993-08.org.debian:01:abc",
	})
	if err != nil {
		t.Fatalf("FindSourcesAs: %v", err)
	}
	if hv.findPoolType != utilities.PoolTypeISCSI.String() {
		t.Fatalf("poolType=%s", hv.findPoolType)
	}
}
