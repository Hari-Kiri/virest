package utilities

import "testing"

func TestPoolTypeStringAndKnown(t *testing.T) {
	cases := []struct {
		t    PoolType
		want string
	}{
		{PoolTypeDir, "dir"},
		{PoolTypeFS, "fs"},
		{PoolTypeNetFS, "netfs"},
		{PoolTypeLogical, "logical"},
		{PoolTypeDisk, "disk"},
		{PoolTypeISCSI, "iscsi"},
		{PoolTypeISCSIDirect, "iscsi-direct"},
		{PoolTypeSCSI, "scsi"},
		{PoolTypeMpath, "mpath"},
		{PoolTypeRBD, "rbd"},
		{PoolTypeSheepdog, "sheepdog"},
		{PoolTypeGluster, "gluster"},
		{PoolTypeZFS, "zfs"},
		{PoolTypeVStorage, "vstorage"},
	}
	n := len(cases)
	for i := 0; i < n; i++ {
		c := cases[i]
		if c.t.String() != c.want {
			t.Fatalf("%v.String()=%q want %q", c.t, c.t.String(), c.want)
		}
		if !c.t.Known() {
			t.Fatalf("%v should be Known", c.t)
		}
	}
	if PoolType("unknown").Known() {
		t.Fatal("unknown must not be Known")
	}
	if PoolType("").Known() {
		t.Fatal("empty must not be Known")
	}
}

func TestPoolTypeNegativeDistinct(t *testing.T) {
	if PoolTypeDir == PoolTypeFS {
		t.Fatal("Dir must differ from FS")
	}
	if PoolTypeISCSI == PoolTypeISCSIDirect {
		t.Fatal("ISCSI must differ from ISCSIDirect")
	}
}

func BenchmarkPoolTypeKnown(b *testing.B) {
	t := PoolTypeDir
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = t.Known()
	}
}
