package utilities

import (
	"errors"
	"testing"
)

func TestBuildPoolModelDirSuccess(t *testing.T) {
	model, err := buildPoolModel(poolAsParams{
		Name:       "images",
		Type:       PoolTypeDir,
		TargetPath: "/var/lib/libvirt/images",
	})
	if err != nil {
		t.Fatalf("buildPoolModel: %v", err)
	}
	if model.Name != "images" || model.Type != PoolTypeDir {
		t.Fatalf("unexpected model: %+v", model)
	}
	if model.Target == nil || model.Target.Path != "/var/lib/libvirt/images" {
		t.Fatalf("unexpected target: %+v", model.Target)
	}
}

func TestBuildPoolModelFailures(t *testing.T) {
	cases := []struct {
		name string
		p    poolAsParams
		want error
	}{
		{"no name", poolAsParams{Type: PoolTypeDir, TargetPath: "/t"}, errPoolAsNameRequired},
		{"no type", poolAsParams{Name: "n", TargetPath: "/t"}, errPoolAsTypeRequired},
		{"dir no target", poolAsParams{Name: "n", Type: PoolTypeDir}, errPoolAsTargetRequired},
		{"unsupported", poolAsParams{Name: "n", Type: PoolType("unknown-backend"), TargetPath: "/t"}, errPoolAsUnsupported},
		{"iscsi no host", poolAsParams{Name: "n", Type: PoolTypeISCSI, SourceDev: "iqn.x", TargetPath: "/dev/disk/by-path"}, errPoolAsHostRequired},
		{"rbd no name", poolAsParams{Name: "n", Type: PoolTypeRBD, SourceHost: "mon"}, errPoolAsSourceNameRequired},
		{"auth incomplete", poolAsParams{Name: "n", Type: PoolTypeDir, TargetPath: "/t", AuthType: "chap"}, errPoolAsAuthIncomplete},
		{"secret conflict", poolAsParams{
			Name: "n", Type: PoolTypeRBD, SourceHost: "mon", SourceName: "pool",
			AuthType: "ceph", AuthUsername: "admin", SecretUsage: "u", SecretUUID: "uuid",
		}, errPoolAsSecretConflict},
	}
	n := len(cases)
	for i := 0; i < n; i++ {
		c := cases[i]
		_, err := buildPoolModel(c.p)
		if err == nil {
			t.Fatalf("%s: expected error", c.name)
		}
		if !errors.Is(err, c.want) {
			t.Fatalf("%s: err=%v want %v", c.name, err, c.want)
		}
	}
}

func TestBuildPoolModelIscsiRbdGluster(t *testing.T) {
	iscsi, err := buildPoolModel(poolAsParams{
		Name: "iscsi-pool", Type: PoolTypeISCSI,
		SourceHost: "storage.example", SourcePort: 3260,
		SourceDev:  "iqn.2010-05.com.example:target",
		TargetPath: "/dev/disk/by-path",
		AuthType:   "chap", AuthUsername: "user", SecretUsage: "libvirtiscsi",
	})
	if err != nil {
		t.Fatalf("iscsi: %v", err)
	}
	if iscsi.Source == nil || len(iscsi.Source.Host) != 1 || iscsi.Source.Host[0].Port != "3260" {
		t.Fatalf("iscsi host: %+v", iscsi.Source)
	}
	if iscsi.Source.Auth == nil || iscsi.Source.Auth.Type != "chap" {
		t.Fatalf("iscsi auth: %+v", iscsi.Source.Auth)
	}
	if len(iscsi.Source.Device) != 1 {
		t.Fatalf("iscsi device: %+v", iscsi.Source.Device)
	}

	rbd, err := buildPoolModel(poolAsParams{
		Name: "rbd-pool", Type: PoolTypeRBD,
		SourceHost: "mon.example", SourceName: "libvirt-pool",
		AuthType: "ceph", AuthUsername: "libvirt", SecretUUID: "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
	})
	if err != nil {
		t.Fatalf("rbd: %v", err)
	}
	if rbd.Source == nil || rbd.Source.Name != "libvirt-pool" || rbd.Source.Auth == nil {
		t.Fatalf("rbd source: %+v", rbd.Source)
	}

	gluster, err := buildPoolModel(poolAsParams{
		Name: "g", Type: PoolTypeGluster,
		SourceHost: "111.222.111.222", SourceName: "vol1", SourcePath: "/",
	})
	if err != nil {
		t.Fatalf("gluster: %v", err)
	}
	if gluster.Source == nil || gluster.Source.Dir == nil || gluster.Source.Dir.Path != "/" {
		t.Fatalf("gluster: %+v", gluster.Source)
	}
}

func TestBuildPoolModelScsiIscsiDirectZfs(t *testing.T) {
	scsi, err := buildPoolModel(poolAsParams{
		Name: "scsi", Type: PoolTypeSCSI, TargetPath: "/dev/disk/by-path",
		AdapterName: "scsi_host5",
	})
	if err != nil {
		t.Fatalf("scsi: %v", err)
	}
	if scsi.Source == nil || scsi.Source.Adapter == nil || scsi.Source.Adapter.Type != "scsi_host" {
		t.Fatalf("scsi adapter: %+v", scsi.Source)
	}

	fc, err := buildPoolModel(poolAsParams{
		Name: "fc", Type: PoolTypeSCSI, TargetPath: "/dev/disk/by-path",
		AdapterWWNN: "20000000c9848140", AdapterWWPN: "10000000c9848140",
	})
	if err != nil {
		t.Fatalf("fc: %v", err)
	}
	if fc.Source.Adapter.Type != "fc_host" {
		t.Fatalf("fc type=%s", fc.Source.Adapter.Type)
	}

	direct, err := buildPoolModel(poolAsParams{
		Name: "d", Type: PoolTypeISCSIDirect,
		SourceHost: "h", SourceDev: "iqn.x", SourceInitiator: "iqn.init",
	})
	if err != nil {
		t.Fatalf("iscsi-direct: %v", err)
	}
	if direct.Source.Initiator == nil || direct.Source.Initiator.IQN.Name != "iqn.init" {
		t.Fatalf("initiator: %+v", direct.Source.Initiator)
	}

	zfs, err := buildPoolModel(poolAsParams{Name: "z", Type: PoolTypeZFS, SourceName: "tank"})
	if err != nil {
		t.Fatalf("zfs: %v", err)
	}
	if zfs.Source == nil || zfs.Source.Name != "tank" {
		t.Fatalf("zfs: %+v", zfs.Source)
	}
}

func TestBuildPoolModelNetfsFsLogicalDiskMpathVstorageSheepdog(t *testing.T) {
	netfs, err := buildPoolModel(poolAsParams{
		Name: "nfs", Type: PoolTypeNetFS, TargetPath: "/mnt/nfs",
		SourceHost: "nfs.example", SourcePath: "/export", SourceFormat: "nfs",
		SourceProtocolVer: "4",
	})
	if err != nil {
		t.Fatalf("netfs: %v", err)
	}
	if netfs.Source.Protocol == nil || netfs.Source.Protocol.Version != "4" {
		t.Fatalf("protocol: %+v", netfs.Source.Protocol)
	}

	fs, err := buildPoolModel(poolAsParams{
		Name: "fs", Type: PoolTypeFS, TargetPath: "/mnt", SourceDev: "/dev/sdb1", SourceFormat: "ext4",
	})
	if err != nil || fs.Source == nil {
		t.Fatalf("fs: %v %+v", err, fs.Source)
	}

	logical, err := buildPoolModel(poolAsParams{Name: "vg", Type: PoolTypeLogical})
	if err != nil {
		t.Fatalf("logical: %v", err)
	}
	if logical.Source == nil || logical.Source.Name != "vg" {
		t.Fatalf("logical default source name: %+v", logical.Source)
	}

	if _, err := buildPoolModel(poolAsParams{
		Name: "disk", Type: PoolTypeDisk, TargetPath: "/dev", SourceDev: "/dev/sdb", SourceFormat: "dos",
	}); err != nil {
		t.Fatalf("disk: %v", err)
	}
	if _, err := buildPoolModel(poolAsParams{Name: "m", Type: PoolTypeMpath, TargetPath: "/dev/mapper"}); err != nil {
		t.Fatalf("mpath: %v", err)
	}
	if _, err := buildPoolModel(poolAsParams{
		Name: "v", Type: PoolTypeVStorage, SourceName: "cluster", TargetPath: "/mnt/vstorage",
	}); err != nil {
		t.Fatalf("vstorage: %v", err)
	}
	if _, err := buildPoolModel(poolAsParams{
		Name: "s", Type: PoolTypeSheepdog, SourceHost: "localhost", SourceName: "vd",
	}); err != nil {
		t.Fatalf("sheepdog: %v", err)
	}
}

func TestDefineAsCreateAsParamsWrappers(t *testing.T) {
	d := DefineAsParams{Name: "images", Type: PoolTypeDir, TargetPath: "/t"}
	c := CreateAsParams{Name: "tmp", Type: PoolTypeDir, TargetPath: "/tmp"}
	model, err := d.BuildModel()
	if err != nil {
		t.Fatalf("DefineAsParams.BuildModel: %v", err)
	}
	if model.Name != "images" {
		t.Fatalf("name=%s", model.Name)
	}
	created, err := c.BuildModel()
	if err != nil {
		t.Fatalf("CreateAsParams.BuildModel: %v", err)
	}
	if created.Name != "tmp" {
		t.Fatalf("create name=%s", created.Name)
	}
}

func TestStoragePoolMarshalRoundTrip(t *testing.T) {
	model := StoragePool{
		Name:   "images",
		Type: PoolTypeDir,
		Target: &StoragePoolTarget{Path: "/var/lib/libvirt/images"},
	}
	doc, err := model.Marshal()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out StoragePool
	if err := out.Unmarshal(doc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.Name != "images" || out.Target == nil || out.Target.Path != "/var/lib/libvirt/images" {
		t.Fatalf("round-trip: %+v", out)
	}
}

func BenchmarkBuildPoolModel(b *testing.B) {
	p := poolAsParams{Name: "images", Type: PoolTypeDir, TargetPath: "/var/lib/libvirt/images"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := buildPoolModel(p)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBuildPoolModelIscsi(b *testing.B) {
	p := poolAsParams{
		Name: "iscsi", Type: PoolTypeISCSI, SourceHost: "h", SourceDev: "iqn.x",
		TargetPath: "/dev/disk/by-path",
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := buildPoolModel(p)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBuildPoolModelMarshal(b *testing.B) {
	p := poolAsParams{Name: "images", Type: PoolTypeDir, TargetPath: "/var/lib/libvirt/images"}
	model, err := buildPoolModel(p)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := model.Marshal()
		if err != nil {
			b.Fatal(err)
		}
	}
}
