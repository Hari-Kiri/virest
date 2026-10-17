package utilities

import (
	"testing"

	"libvirt.org/go/libvirt"
)

func TestFlagsDefineCreateBuildDelete(t *testing.T) {
	if Flags.Define.None != 0 {
		t.Fatalf("Define.None=%v", Flags.Define.None)
	}
	if Flags.Define.Validate.Libvirt() != libvirt.STORAGE_POOL_DEFINE_VALIDATE {
		t.Fatalf("Define.Validate mismatch")
	}
	if Flags.Create.Normal.Libvirt() != libvirt.STORAGE_POOL_CREATE_NORMAL {
		t.Fatalf("Create.Normal mismatch")
	}
	if Flags.Create.WithBuild.Libvirt() != libvirt.STORAGE_POOL_CREATE_WITH_BUILD {
		t.Fatalf("Create.WithBuild mismatch")
	}
	if Flags.Create.WithBuildOverwrite.Libvirt() != libvirt.STORAGE_POOL_CREATE_WITH_BUILD_OVERWRITE {
		t.Fatalf("Create.WithBuildOverwrite mismatch")
	}
	if Flags.Create.WithBuildNoOverwrite.Libvirt() != libvirt.STORAGE_POOL_CREATE_WITH_BUILD_NO_OVERWRITE {
		t.Fatalf("Create.WithBuildNoOverwrite mismatch")
	}
	if Flags.Build.New.Libvirt() != libvirt.STORAGE_POOL_BUILD_NEW {
		t.Fatalf("Build.New mismatch")
	}
	if Flags.Build.Repair.Libvirt() != libvirt.STORAGE_POOL_BUILD_REPAIR {
		t.Fatalf("Build.Repair mismatch")
	}
	if Flags.Build.Resize.Libvirt() != libvirt.STORAGE_POOL_BUILD_RESIZE {
		t.Fatalf("Build.Resize mismatch")
	}
	if Flags.Build.NoOverwrite.Libvirt() != libvirt.STORAGE_POOL_BUILD_NO_OVERWRITE {
		t.Fatalf("Build.NoOverwrite mismatch")
	}
	if Flags.Build.Overwrite.Libvirt() != libvirt.STORAGE_POOL_BUILD_OVERWRITE {
		t.Fatalf("Build.Overwrite mismatch")
	}
	if Flags.Delete.Normal.Libvirt() != libvirt.STORAGE_POOL_DELETE_NORMAL {
		t.Fatalf("Delete.Normal mismatch")
	}
	if Flags.Delete.Zeroed.Libvirt() != libvirt.STORAGE_POOL_DELETE_ZEROED {
		t.Fatalf("Delete.Zeroed mismatch")
	}
}

func TestFlagsListAndXML(t *testing.T) {
	cases := []struct {
		name string
		got  libvirt.ConnectListAllStoragePoolsFlags
		want libvirt.ConnectListAllStoragePoolsFlags
	}{
		{"Inactive", Flags.List.Inactive.Libvirt(), libvirt.CONNECT_LIST_STORAGE_POOLS_INACTIVE},
		{"Active", Flags.List.Active.Libvirt(), libvirt.CONNECT_LIST_STORAGE_POOLS_ACTIVE},
		{"Persistent", Flags.List.Persistent.Libvirt(), libvirt.CONNECT_LIST_STORAGE_POOLS_PERSISTENT},
		{"Transient", Flags.List.Transient.Libvirt(), libvirt.CONNECT_LIST_STORAGE_POOLS_TRANSIENT},
		{"Autostart", Flags.List.Autostart.Libvirt(), libvirt.CONNECT_LIST_STORAGE_POOLS_AUTOSTART},
		{"NoAutostart", Flags.List.NoAutostart.Libvirt(), libvirt.CONNECT_LIST_STORAGE_POOLS_NO_AUTOSTART},
		{"Dir", Flags.List.Dir.Libvirt(), libvirt.CONNECT_LIST_STORAGE_POOLS_DIR},
		{"FS", Flags.List.FS.Libvirt(), libvirt.CONNECT_LIST_STORAGE_POOLS_FS},
		{"NetFS", Flags.List.NetFS.Libvirt(), libvirt.CONNECT_LIST_STORAGE_POOLS_NETFS},
		{"Logical", Flags.List.Logical.Libvirt(), libvirt.CONNECT_LIST_STORAGE_POOLS_LOGICAL},
		{"Disk", Flags.List.Disk.Libvirt(), libvirt.CONNECT_LIST_STORAGE_POOLS_DISK},
		{"ISCSI", Flags.List.ISCSI.Libvirt(), libvirt.CONNECT_LIST_STORAGE_POOLS_ISCSI},
		{"SCSI", Flags.List.SCSI.Libvirt(), libvirt.CONNECT_LIST_STORAGE_POOLS_SCSI},
		{"Mpath", Flags.List.Mpath.Libvirt(), libvirt.CONNECT_LIST_STORAGE_POOLS_MPATH},
		{"RBD", Flags.List.RBD.Libvirt(), libvirt.CONNECT_LIST_STORAGE_POOLS_RBD},
		{"Sheepdog", Flags.List.Sheepdog.Libvirt(), libvirt.CONNECT_LIST_STORAGE_POOLS_SHEEPDOG},
		{"Gluster", Flags.List.Gluster.Libvirt(), libvirt.CONNECT_LIST_STORAGE_POOLS_GLUSTER},
		{"ZFS", Flags.List.ZFS.Libvirt(), libvirt.CONNECT_LIST_STORAGE_POOLS_ZFS},
		{"VStorage", Flags.List.VStorage.Libvirt(), libvirt.CONNECT_LIST_STORAGE_POOLS_VSTORAGE},
		{"ISCSIDirect", Flags.List.ISCSIDirect.Libvirt(), libvirt.CONNECT_LIST_STORAGE_POOLS_ISCSI_DIRECT},
	}
	n := len(cases)
	for i := 0; i < n; i++ {
		c := cases[i]
		if c.got != c.want {
			t.Fatalf("%s: got %v want %v", c.name, c.got, c.want)
		}
	}
	if Flags.XML.None != 0 {
		t.Fatalf("XML.None=%v", Flags.XML.None)
	}
	if Flags.XML.Inactive.Libvirt() != libvirt.STORAGE_XML_INACTIVE {
		t.Fatalf("XML.Inactive mismatch")
	}
	combined := Flags.List.Active | Flags.List.Dir
	if combined.Libvirt() != libvirt.CONNECT_LIST_STORAGE_POOLS_ACTIVE|libvirt.CONNECT_LIST_STORAGE_POOLS_DIR {
		t.Fatalf("OR combine failed: %v", combined)
	}
}

func TestFlagsNegativeDistinct(t *testing.T) {
	if Flags.Create.WithBuild == Flags.Create.WithBuildOverwrite {
		t.Fatal("WithBuild must differ from WithBuildOverwrite")
	}
	if Flags.Build.Overwrite == Flags.Build.NoOverwrite {
		t.Fatal("Overwrite must differ from NoOverwrite")
	}
	if Flags.Delete.Normal == Flags.Delete.Zeroed {
		t.Fatal("Delete Normal must differ from Zeroed")
	}
	if Flags.Define.Validate == Flags.Define.None {
		t.Fatal("Validate must differ from None")
	}
	if Flags.List.Active == Flags.List.Inactive {
		t.Fatal("List Active must differ from Inactive")
	}
	if Flags.XML.Inactive == Flags.XML.None {
		t.Fatal("XML Inactive must differ from None")
	}
}

func BenchmarkFlagsLibvirtConvert(b *testing.B) {
	f := Flags.Create.WithBuild
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = f.Libvirt()
	}
}

func BenchmarkFlagsListLibvirtConvert(b *testing.B) {
	f := Flags.List.Active | Flags.List.Dir
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = f.Libvirt()
	}
}
