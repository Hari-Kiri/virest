package utilities

import "testing"

func TestStoragePoolMarshalRejectsUnknownType(t *testing.T) {
	pool := StoragePool{Name: "x", Type: PoolType("not-a-protocol")}
	if _, err := pool.Marshal(); err == nil {
		t.Fatal("expected unsupported type error")
	}
}

func TestStoragePoolMarshalKnownType(t *testing.T) {
	pool := StoragePool{
		Name:   "My-Pool-dir",
		Type:   PoolTypeDir,
		Target: &StoragePoolTarget{Path: "/var/lib/libvirt/images"},
	}
	doc, err := pool.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out StoragePool
	if err := out.Unmarshal(doc); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Type != PoolTypeDir || out.Name != "My-Pool-dir" {
		t.Fatalf("round-trip: %+v", out)
	}
}

func TestStoragePoolUnmarshalSetsPoolType(t *testing.T) {
	const doc = `<pool type='iscsi'><name>My-Pool-iscsi</name></pool>`
	var pool StoragePool
	if err := pool.Unmarshal(doc); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if pool.Type != PoolTypeISCSI {
		t.Fatalf("Type=%q", pool.Type)
	}
}
