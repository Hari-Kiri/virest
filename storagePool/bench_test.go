package storagePool

import (
	"testing"

	"github.com/Hari-Kiri/virest/utilities"
	"libvirt.org/go/libvirt"
)

func BenchmarkInfo(b *testing.B) {
	pool := &fakePool{
		name:       "p",
		uuid:       "u",
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
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := conn.Info("u"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkList(b *testing.B) {
	const n = 32
	pools := make([]poolHandle, n)
	for i := 0; i < n; i++ {
		pools[i] = &fakePool{
			name: "p",
			uuid: "u",
			xmlDesc: `<pool type='dir'><name>p</name><uuid>u</uuid>` +
				`<capacity unit='bytes'>10</capacity>` +
				`<allocation unit='bytes'>1</allocation>` +
				`<available unit='bytes'>9</available></pool>`,
			info: &libvirt.StoragePoolInfo{State: libvirt.STORAGE_POOL_RUNNING, Capacity: 10},
		}
	}
	hv := newFakeHypervisor()
	hv.listPools = pools
	conn := newConnectionForTest(hv)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hv.listPools = pools
		if _, err := conn.List(0, 0); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLifecycle(b *testing.B) {
	pool := &fakePool{}
	hv := newFakeHypervisor()
	hv.registerPool("u", pool)
	conn := newConnectionForTest(hv)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := conn.Build("u", 0); err != nil {
			b.Fatal(err)
		}
		if err := conn.Start("u", 0); err != nil {
			b.Fatal(err)
		}
		if err := conn.Destroy("u"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDefine(b *testing.B) {
	hv := newFakeHypervisor()
	hv.definePool = &fakePool{uuid: "u"}
	conn := newConnectionForTest(hv)
	model := utilities.StoragePool{Name: "p", Type: utilities.PoolTypeDir}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := conn.Define(model, 0); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCreateTransient(b *testing.B) {
	hv := newFakeHypervisor()
	hv.createTransientPool = &fakePool{uuid: "u"}
	conn := newConnectionForTest(hv)
	model := utilities.StoragePool{Name: "p", Type: utilities.PoolTypeDir}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := conn.CreateTransient(model, 0); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDefineAs(b *testing.B) {
	hv := newFakeHypervisor()
	hv.definePool = &fakePool{uuid: "u"}
	conn := newConnectionForTest(hv)
	params := utilities.DefineAsParams{Name: "images", Type: utilities.PoolTypeDir, TargetPath: "/var/lib/libvirt/images"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := conn.DefineAs(params, 0); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCreateAs(b *testing.B) {
	hv := newFakeHypervisor()
	hv.createTransientPool = &fakePool{uuid: "u"}
	conn := newConnectionForTest(hv)
	params := utilities.CreateAsParams{Name: "images", Type: utilities.PoolTypeDir, TargetPath: "/var/lib/libvirt/images"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := conn.CreateAs(params, 0); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLookupByName(b *testing.B) {
	pool := &fakePool{name: "p", uuid: "u"}
	hv := newFakeHypervisor()
	hv.registerPool("p", pool)
	conn := newConnectionForTest(hv)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := conn.Destroy("p"); err != nil {
			b.Fatal(err)
		}
	}
}
