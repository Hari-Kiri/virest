# storagePool — Implementation wiki

Go library mirroring the `virsh` **Storage Pool** command set (except interactive `pool-edit`). External callers use **typed Go structs** from `utilities` and `storagePool` — not raw XML and not `libvirtxml` imports. XML is marshaled only inside the package adapter.

**Import:** `github.com/Hari-Kiri/virest/storagePool`  
**Models / helpers:** `github.com/Hari-Kiri/virest/utilities` (`StoragePool`, `DefineAsParams`, `CreateAsParams`, `Flags`, …)

---

## Table of contents

1. [Quick start](#quick-start)
2. [Public I/O contract](#public-io-contract)
3. [Architecture](#architecture)
4. [Pool refs (name or UUID)](#pool-refs-name-or-uuid)
5. [virsh → Go cheat sheet](#virsh-→-go-cheat-sheet)
6. [Method reference](#method-reference)
   - [Connect](#connect) — `Connect`, `Close`
   - [Lifecycle](#lifecycle) — `Define`, `DefineAs`, `CreateTransient`, `CreateAs`, `Build`, `Start`, `Destroy`, `Delete`, `Undefine`, `Refresh`, `Autostart`
   - [Query / list](#query--list) — `List`, `ListActive`, `ListInactive`, `ListAll`, `Info`, `Dump`, `Name`, `UUIDByName`
   - [Discovery](#discovery) — `Capabilities`, `FindSources`, `FindSourcesAs`
   - [Events](#events) — `WaitEvent`, `StreamEvents`
7. [Persistent vs transient flows](#persistent-vs-transient-flows)
   - [Persistent](#persistent-survives-host-reboot-if-defined--autostart)
   - [Transient](#transient-active-until-destroy--reboot)
8. [Errors and Free](#errors-and-free)
9. [File map](#file-map)
10. [Related](#related)

---

## Quick start

```go
package main

import (
	"fmt"
	"log"

	"github.com/Hari-Kiri/virest/storagePool"
	"github.com/Hari-Kiri/virest/utilities"
)

func main() {
	conn, err := storagePool.Connect("qemu:///system")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// virsh pool-list --all
	pools, err := conn.ListAll(0)
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < len(pools); i++ {
		fmt.Printf("%s %s state=%v\n", pools[i].Name, pools[i].Uuid, pools[i].State)
	}

	// Define a dir pool (virsh pool-define-as), then start it
	uuid, err := conn.DefineAs(utilities.DefineAsParams{
		Name:       "My-Pool-dir",
		Type:       utilities.PoolTypeDir,
		TargetPath: "/var/lib/libvirt/images",
	}, 0)
	if err != nil {
		log.Fatal(err)
	}
	if err := conn.Build(uuid, utilities.Flags.Build.New); err != nil {
		// some pool types (e.g. dir) may soft-fail build — handle as needed
		log.Printf("build: %v", err)
	}
	if err := conn.Start(uuid, 0); err != nil {
		log.Fatal(err)
	}
}
```

---

## Public I/O contract

| Boundary | Format |
|----------|--------|
| Public `Connection` methods | Go structs from `utilities` (`StoragePool`, `DefineAsParams`, `CreateAsParams`, `Flags`, …) plus `Info`, `Detail` (return type), `PoolSummary`, string refs, `utilities` flags |
| Internal adapter → hypervisor | XML strings only inside `libvirt_adapter.go` |

- No public method takes or returns raw XML.
- Callers import `utilities` for pool models, **`utilities.PoolType*`** / `StoragePool.Type` (`PoolType`, not free strings), and **`utilities.Flags`** — **not** `libvirt.org/go/libvirtxml` (and prefer not raw `libvirt.STORAGE_POOL_*` on Connection APIs).
- Method names avoid `XML` (e.g. `CreateTransient`, not `CreateXML`). Prefer `Start` for pool-start (no library `Create` start method). Prefer `Dump` for pool-dumpxml (no library `Detail` method; return type is still `Detail`).
- HTTP testserver may JSON-encode the same structs; that is not XML on the wire for library callers.

---

## Architecture

```text
Caller / testserver / examples
        ↓
  Connection methods (lifecycle, list_info, detail, query_short, event)
        ↓
  lookupPool(ref)  — exactly one libvirt lookup (UUID or name)
        ↓
  hypervisor / poolHandle interfaces
        ↓
  libvirt_adapter → libvirt daemon
```

Shared non-domain logic lives in `utilities/` (`LooksLikeUUID`, `PoolType`, `Flags`, `DefineAsParams`/`CreateAsParams` builders, brand-neutral `StoragePool` aliases). Hypervisor orchestration stays in `storagePool/`.

---

## Pool refs (name or UUID)

Most mutating/query methods take `ref string`: a pool **name** or **UUID**.

| Input shape | Lookup used | Round-trips |
|-------------|-------------|-------------|
| `8-4-4-4-12` hex UUID (case-insensitive) | `LookupStoragePoolByUUIDString` | 1 |
| Any other string (e.g. `default`, `my-pool`) | `LookupStoragePoolByName` | 1 |

Rules:

- UUID-shaped strings **never** fall back to name lookup (same idea as virsh).
- Do **not** resolve names by scanning `List` (that would be N+1).
- Helper: `utilities.LooksLikeUUID(ref)`.

```go
info, err := conn.Info("default")                                    // by name
info, err = conn.Info("a1b2c3d4-e5f6-7890-abcd-ef1234567890")         // by UUID
```

---

## virsh → Go cheat sheet

| virsh | Library |
|-------|---------|
| `find-storage-pool-sources` | `FindSources(type, SourceSpec)` |
| `find-storage-pool-sources-as` | `FindSourcesAs(type, FindSourcesAsParams)` |
| `pool-autostart` | `Autostart(ref, bool)` |
| `pool-build` | `Build(ref, flags)` |
| `pool-create` | `CreateTransient(model, flags)` |
| `pool-create-as` | `CreateAs(params, flags)` |
| `pool-define` | `Define(model, flags)` |
| `pool-define-as` | `DefineAs(params, flags)` |
| `pool-delete` | `Delete(ref, flags)` |
| `pool-destroy` | `Destroy(ref)` |
| `pool-dumpxml` | `Dump(ref, xmlFlags)` |
| `pool-edit` | **Not implemented** — `Dump` → edit struct → `Define` |
| `pool-info` | `Info(ref)` |
| `pool-list` | `List` / `ListActive` / `ListInactive` / `ListAll` |
| `pool-name` | `Name(ref)` |
| `pool-refresh` | `Refresh(ref)` |
| `pool-start` | `Start(ref, flags)` |
| `pool-undefine` | `Undefine(ref)` |
| `pool-uuid` | `UUIDByName(name)` |
| `pool-event` | `WaitEvent` / `StreamEvents` |
| `pool-capabilities` | `Capabilities()` |

---

## Method reference

### Connect

#### `Connect(uri string) (*Connection, error)`

Opens a connection to the hypervisor. See [libvirt URI](https://libvirt.org/uri.html).

```go
conn, err := storagePool.Connect("qemu:///system")
if err != nil {
	return err
}
defer conn.Close()
```

#### `Close() error`

Releases the underlying libvirt connection. Safe on `nil` or already-closed `Connection` (returns `nil`). Always `defer conn.Close()` after a successful `Connect`.

---

### Lifecycle

#### `Define(model utilities.StoragePool, flags) (uuid string, error)`

Creates a **persistent, inactive** pool from a typed model (`virsh pool-define`). Returns the new pool UUID. Pool must be inactive if re-defining an existing definition (hypervisor rule).

`Define` accepts any `utilities.StoragePool` the hypervisor supports. Examples for common backends (same set as `DefineAs`) follow below.

Flags are `utilities.DefineFlags` (passed to `virStoragePoolDefineXML`):

| Flag | Value | Meaning |
|------|-------|---------|
| `0` (no flags) | `0` | Define the persistent pool from the marshaled model **without** extra schema validation. Usual default. |
| `utilities.Flags.Define.Validate` | `1 << 0` | Validate the generated XML against the storage-pool XML schema before defining. Fails early if the document is invalid for the hypervisor/libvirt schema. |

```go
model := utilities.StoragePool{
	Name: "My-Pool-dir",
	Type: utilities.PoolTypeDir,
	Target: &utilities.StoragePoolTarget{
		Path: "/var/lib/libvirt/images",
	},
}

// default: define without schema validation
uuid, err := conn.Define(model, 0)

// validate XML against schema, then define
uuid, err = conn.Define(model, utilities.Flags.Define.Validate)
```

Examples for common backends (same set as `DefineAs`):

```go
// dir
uuid, err := conn.Define(utilities.StoragePool{
	Name: "My-Pool-dir",
	Type: utilities.PoolTypeDir,
	Target: &utilities.StoragePoolTarget{
		Path: "/var/lib/libvirt/images",
	},
}, 0)

// fs
uuid, err = conn.Define(utilities.StoragePool{
	Name: "My-Pool-fs",
	Type: utilities.PoolTypeFS,
	Target: &utilities.StoragePoolTarget{
		Path: "/mnt/data",
	},
	Source: &utilities.StoragePoolSource{
		Device: []utilities.StoragePoolSourceDevice{{
			Path: "/dev/sdb1",
		}},
		Format: &utilities.StoragePoolSourceFormat{
			Type: "ext4",
		},
	},
}, 0)

// netfs (NFS)
uuid, err = conn.Define(utilities.StoragePool{
	Name: "My-Pool-netfs",
	Type: utilities.PoolTypeNetFS,
	Target: &utilities.StoragePoolTarget{
		Path: "/mnt/nfs",
	},
	Source: &utilities.StoragePoolSource{
		Host: []utilities.StoragePoolSourceHost{{
			Name: "nfs.example",
		}},
		Dir: &utilities.StoragePoolSourceDir{
			Path: "/export/images",
		},
		Format: &utilities.StoragePoolSourceFormat{
			Type: "nfs",
		},
		Protocol: &utilities.StoragePoolSourceProtocol{
			Version: "4",
		},
	},
}, 0)

// logical (LVM)
uuid, err = conn.Define(utilities.StoragePool{
	Name: "My-Pool-logical",
	Type: utilities.PoolTypeLogical,
	Source: &utilities.StoragePoolSource{
		Name: "vg0",
	},
}, 0)

// disk
uuid, err = conn.Define(utilities.StoragePool{
	Name: "My-Pool-disk",
	Type: utilities.PoolTypeDisk,
	Target: &utilities.StoragePoolTarget{
		Path: "/dev",
	},
	Source: &utilities.StoragePoolSource{
		Device: []utilities.StoragePoolSourceDevice{{
			Path: "/dev/sdb",
		}},
		Format: &utilities.StoragePoolSourceFormat{
			Type: "dos",
		},
	},
}, 0)

// iscsi
uuid, err = conn.Define(utilities.StoragePool{
	Name: "My-Pool-iscsi",
	Type: utilities.PoolTypeISCSI,
	Target: &utilities.StoragePoolTarget{
		Path: "/dev/disk/by-path",
	},
	Source: &utilities.StoragePoolSource{
		Host: []utilities.StoragePoolSourceHost{{
			Name: "storage.example",
			Port: "3260",
		}},
		Device: []utilities.StoragePoolSourceDevice{{
			Path: "iqn.2010-05.com.example:target",
		}},
		Auth: &utilities.StoragePoolSourceAuth{
			Type: "chap",
			Username: "user",
			Secret: &utilities.StoragePoolSourceAuthSecret{
				Usage: "libvirtiscsi",
			},
		},
	},
}, 0)

// iscsi-direct
uuid, err = conn.Define(utilities.StoragePool{
	Name: "My-Pool-iscsi-direct",
	Type: utilities.PoolTypeISCSIDirect,
	Source: &utilities.StoragePoolSource{
		Host: []utilities.StoragePoolSourceHost{{
			Name: "storage.example",
			Port: "3260",
		}},
		Device: []utilities.StoragePoolSourceDevice{{
			Path: "iqn.2010-05.com.example:target",
		}},
		Initiator: &utilities.StoragePoolSourceInitiator{
			IQN: utilities.StoragePoolSourceInitiatorIQN{
				Name: "iqn.2010-05.com.example:initiator",
			},
		},
	},
}, 0)

// scsi (scsi_host adapter)
uuid, err = conn.Define(utilities.StoragePool{
	Name: "My-Pool-scsi",
	Type: utilities.PoolTypeSCSI,
	Target: &utilities.StoragePoolTarget{
		Path: "/dev/disk/by-path",
	},
	Source: &utilities.StoragePoolSource{
		Adapter: &utilities.StoragePoolSourceAdapter{
			Type: "scsi_host",
			Name: "scsi_host5",
		},
	},
}, 0)

// scsi (fc_host adapter)
uuid, err = conn.Define(utilities.StoragePool{
	Name: "My-Pool-scsi",
	Type: utilities.PoolTypeSCSI,
	Target: &utilities.StoragePoolTarget{
		Path: "/dev/disk/by-path",
	},
	Source: &utilities.StoragePoolSource{
		Adapter: &utilities.StoragePoolSourceAdapter{
			Type: "fc_host",
			WWNN: "20000000c9848140",
			WWPN: "10000000c9848140",
		},
	},
}, 0)

// mpath
uuid, err = conn.Define(utilities.StoragePool{
	Name: "My-Pool-mpath",
	Type: utilities.PoolTypeMpath,
	Target: &utilities.StoragePoolTarget{
		Path: "/dev/mapper",
	},
}, 0)

// rbd (Ceph)
uuid, err = conn.Define(utilities.StoragePool{
	Name: "My-Pool-rbd",
	Type: utilities.PoolTypeRBD,
	Source: &utilities.StoragePoolSource{
		Name: "libvirt-pool",
		Host: []utilities.StoragePoolSourceHost{{
			Name: "mon.example",
		}},
		Auth: &utilities.StoragePoolSourceAuth{
			Type: "ceph",
			Username: "libvirt",
			Secret: &utilities.StoragePoolSourceAuthSecret{
				UUID: "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
			},
		},
	},
}, 0)

// sheepdog
uuid, err = conn.Define(utilities.StoragePool{
	Name: "My-Pool-sheepdog",
	Type: utilities.PoolTypeSheepdog,
	Source: &utilities.StoragePoolSource{
		Name: "vd",
		Host: []utilities.StoragePoolSourceHost{{
			Name: "localhost",
		}},
	},
}, 0)

// gluster
uuid, err = conn.Define(utilities.StoragePool{
	Name: "My-Pool-gluster",
	Type: utilities.PoolTypeGluster,
	Source: &utilities.StoragePoolSource{
		Name: "vol1",
		Host: []utilities.StoragePoolSourceHost{{
			Name: "111.222.111.222",
		}},
		Dir: &utilities.StoragePoolSourceDir{
			Path: "/",
		},
	},
}, 0)

// zfs
uuid, err = conn.Define(utilities.StoragePool{
	Name: "My-Pool-zfs",
	Type: utilities.PoolTypeZFS,
	Source: &utilities.StoragePoolSource{
		Name: "tank",
	},
}, 0)

// vstorage
uuid, err = conn.Define(utilities.StoragePool{
	Name: "My-Pool-vstorage",
	Type: utilities.PoolTypeVStorage,
	Target: &utilities.StoragePoolTarget{
		Path: "/mnt/vstorage",
	},
	Source: &utilities.StoragePoolSource{
		Name: "cluster1",
	},
}, 0)
```

#### `DefineAs(params utilities.DefineAsParams, flags) (uuid string, error)`

Same as `Define`, but builds the model from virsh-style `-as` fields via `utilities.DefineAsParams.BuildModel` (`virsh pool-define-as`).

Supported `Type` values (virsh `-as` backends): `dir`, `fs`, `netfs`, `logical`, `disk`, `iscsi`, `iscsi-direct`, `scsi`, `mpath`, `rbd`, `sheepdog`, `gluster`, `zfs`, `vstorage`. Unsupported types return a clear error. Hypervisor must still support the type at define time.

| Field | Typical use |
|-------|-------------|
| `Name` / `Type` | Required |
| `TargetPath` | Required for dir/fs/netfs/disk/iscsi/scsi/mpath/vstorage; optional elsewhere |
| `SourceHost` / `SourcePort` | netfs, iscsi*, rbd, sheepdog, gluster |
| `SourcePath` | netfs export / gluster dir |
| `SourceDev` / `SourceFormat` | fs, disk, iscsi target IQN |
| `SourceName` | logical VG, rbd/gluster/zfs/sheepdog/vstorage name |
| `AuthType` / `AuthUsername` / `SecretUsage` or `SecretUUID` | iscsi chap / rbd ceph (usage and uuid mutually exclusive) |
| `SourceInitiator` | iscsi-direct (required); optional on iscsi |
| `SourceProtocolVer` | netfs NFS version (`nfsvers`) |
| `AdapterName` or `AdapterWWNN`+`AdapterWWPN` | scsi / FC host |

Flags are the same `utilities.DefineFlags` as [`Define`](#definemodel-utilitiesstoragepool-flags-uuid-string-error) (passed through to `Define`):

| Flag | Value | Meaning |
|------|-------|---------|
| `0` (no flags) | `0` | Build the model from `-as` params and define **without** schema validation. Usual default. |
| `utilities.Flags.Define.Validate` | `1 << 0` | After `BuildModel`, validate the XML against the storage-pool schema before defining. Catches invalid combinations early. |

```go
params := utilities.DefineAsParams{
	Name: "My-Pool-dir",
	Type: utilities.PoolTypeDir,
	TargetPath: "/var/lib/libvirt/images",
}

// default: define-as without schema validation
uuid, err := conn.DefineAs(params, 0)

// validate XML against schema, then define
uuid, err = conn.DefineAs(params, utilities.Flags.Define.Validate)
```

Examples for common backends:

```go
// dir
uuid, err := conn.DefineAs(utilities.DefineAsParams{
	Name: "My-Pool-dir",
	Type: utilities.PoolTypeDir,
	TargetPath: "/var/lib/libvirt/images",
}, 0)

// fs
uuid, err = conn.DefineAs(utilities.DefineAsParams{
	Name: "My-Pool-fs",
	Type: utilities.PoolTypeFS,
	TargetPath: "/mnt/data",
	SourceDev: "/dev/sdb1",
	SourceFormat: "ext4",
}, 0)

// netfs (NFS)
uuid, err = conn.DefineAs(utilities.DefineAsParams{
	Name: "My-Pool-netfs",
	Type: utilities.PoolTypeNetFS,
	TargetPath: "/mnt/nfs",
	SourceHost: "nfs.example",
	SourcePath: "/export/images",
	SourceFormat: "nfs",
	SourceProtocolVer: "4",
}, 0)

// logical (LVM; SourceName defaults to Name when omitted)
uuid, err = conn.DefineAs(utilities.DefineAsParams{
	Name: "My-Pool-logical",
	Type: utilities.PoolTypeLogical,
	SourceName: "vg0",
}, 0)

// disk
uuid, err = conn.DefineAs(utilities.DefineAsParams{
	Name: "My-Pool-disk",
	Type: utilities.PoolTypeDisk,
	TargetPath: "/dev",
	SourceDev: "/dev/sdb",
	SourceFormat: "dos",
}, 0)

// iscsi
uuid, err = conn.DefineAs(utilities.DefineAsParams{
	Name: "My-Pool-iscsi",
	Type: utilities.PoolTypeISCSI,
	SourceHost: "storage.example",
	SourcePort: 3260,
	SourceDev: "iqn.2010-05.com.example:target",
	TargetPath: "/dev/disk/by-path",
	AuthType: "chap",
	AuthUsername: "user",
	SecretUsage: "libvirtiscsi",
}, 0)

// iscsi-direct
uuid, err = conn.DefineAs(utilities.DefineAsParams{
	Name: "My-Pool-iscsi-direct",
	Type: utilities.PoolTypeISCSIDirect,
	SourceHost: "storage.example",
	SourcePort: 3260,
	SourceDev: "iqn.2010-05.com.example:target",
	SourceInitiator: "iqn.2010-05.com.example:initiator",
}, 0)

// scsi (scsi_host adapter)
uuid, err = conn.DefineAs(utilities.DefineAsParams{
	Name: "My-Pool-scsi",
	Type: utilities.PoolTypeSCSI,
	TargetPath: "/dev/disk/by-path",
	AdapterName: "scsi_host5",
}, 0)

// scsi (fc_host adapter)
uuid, err = conn.DefineAs(utilities.DefineAsParams{
	Name: "My-Pool-scsi",
	Type: utilities.PoolTypeSCSI,
	TargetPath: "/dev/disk/by-path",
	AdapterWWNN: "20000000c9848140",
	AdapterWWPN: "10000000c9848140",
}, 0)

// mpath
uuid, err = conn.DefineAs(utilities.DefineAsParams{
	Name: "My-Pool-mpath",
	Type: utilities.PoolTypeMpath,
	TargetPath: "/dev/mapper",
}, 0)

// rbd (Ceph)
uuid, err = conn.DefineAs(utilities.DefineAsParams{
	Name: "My-Pool-rbd",
	Type: utilities.PoolTypeRBD,
	SourceHost: "mon.example",
	SourceName: "libvirt-pool",
	AuthType: "ceph",
	AuthUsername: "libvirt",
	SecretUUID: "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
}, 0)

// sheepdog
uuid, err = conn.DefineAs(utilities.DefineAsParams{
	Name: "My-Pool-sheepdog",
	Type: utilities.PoolTypeSheepdog,
	SourceHost: "localhost",
	SourceName: "vd",
}, 0)

// gluster
uuid, err = conn.DefineAs(utilities.DefineAsParams{
	Name: "My-Pool-gluster",
	Type: utilities.PoolTypeGluster,
	SourceHost: "111.222.111.222",
	SourceName: "vol1",
	SourcePath: "/",
}, 0)

// zfs
uuid, err = conn.DefineAs(utilities.DefineAsParams{
	Name: "My-Pool-zfs",
	Type: utilities.PoolTypeZFS,
	SourceName: "tank",
}, 0)

// vstorage
uuid, err = conn.DefineAs(utilities.DefineAsParams{
	Name: "My-Pool-vstorage",
	Type: utilities.PoolTypeVStorage,
	SourceName: "cluster1",
	TargetPath: "/mnt/vstorage",
}, 0)
```

#### `CreateTransient(model, flags) (uuid string, error)`

Creates and **starts** a **transient** (non-persistent) pool in one step (`virsh pool-create`). Survives until destroy or host reboot; no `Undefine`. Teardown: `Destroy` only.

Flags are `utilities.CreateFlags` (same set as [`Start`](#startref-flags-error); passed to `virStoragePoolCreateXML`). They control whether create also runs an implicit **build**:

| Flag | Value | Meaning |
|------|-------|---------|
| `utilities.Flags.Create.Normal` | `0` | Create and start the transient pool **without** building. Equivalent to passing `0`. |
| `utilities.Flags.Create.WithBuild` | `1 << 0` | Create/start and perform `Build` with default build flags (`utilities.Flags.Build.New`). |
| `utilities.Flags.Create.WithBuildOverwrite` | `1 << 1` | Create/start and build using `utilities.Flags.Build.Overwrite` (wipe existing data). Mutually exclusive with `WithBuildNoOverwrite`. |
| `utilities.Flags.Create.WithBuildNoOverwrite` | `1 << 2` | Create/start and build using `utilities.Flags.Build.NoOverwrite` (refuse if data would be overwritten). Mutually exclusive with `WithBuildOverwrite`. |

```go
model := utilities.StoragePool{
	Name: "My-Pool-dir",
	Type: utilities.PoolTypeDir,
	Target: &utilities.StoragePoolTarget{
		Path: "/tmp/virest-pool",
	},
}

// create + start only (no build)
uuid, err := conn.CreateTransient(model, utilities.Flags.Create.Normal)
// equivalent:
uuid, err = conn.CreateTransient(model, 0)
defer conn.Destroy(uuid)

// create + start + build (new)
uuid, err = conn.CreateTransient(model, utilities.Flags.Create.WithBuild)

// create + start + destructive build
uuid, err = conn.CreateTransient(model, utilities.Flags.Create.WithBuildOverwrite)

// create + start + safe build (no overwrite)
uuid, err = conn.CreateTransient(model, utilities.Flags.Create.WithBuildNoOverwrite)
```

Examples for common backends (same set as `Define`):

```go
// dir
uuid, err := conn.CreateTransient(utilities.StoragePool{
	Name: "My-Pool-dir",
	Type: utilities.PoolTypeDir,
	Target: &utilities.StoragePoolTarget{
		Path: "/tmp/virest-pool",
	},
}, 0)
defer conn.Destroy(uuid)

// fs
uuid, err = conn.CreateTransient(utilities.StoragePool{
	Name: "My-Pool-fs",
	Type: utilities.PoolTypeFS,
	Target: &utilities.StoragePoolTarget{
		Path: "/mnt/tmp-data",
	},
	Source: &utilities.StoragePoolSource{
		Device: []utilities.StoragePoolSourceDevice{{
			Path: "/dev/sdb1",
		}},
		Format: &utilities.StoragePoolSourceFormat{
			Type: "ext4",
		},
	},
}, utilities.Flags.Create.WithBuild)

// netfs (NFS)
uuid, err = conn.CreateTransient(utilities.StoragePool{
	Name: "My-Pool-netfs",
	Type: utilities.PoolTypeNetFS,
	Target: &utilities.StoragePoolTarget{
		Path: "/mnt/tmp-nfs",
	},
	Source: &utilities.StoragePoolSource{
		Host: []utilities.StoragePoolSourceHost{{
			Name: "nfs.example",
		}},
		Dir: &utilities.StoragePoolSourceDir{
			Path: "/export/images",
		},
		Format: &utilities.StoragePoolSourceFormat{
			Type: "nfs",
		},
		Protocol: &utilities.StoragePoolSourceProtocol{
			Version: "4",
		},
	},
}, 0)

// logical (LVM)
uuid, err = conn.CreateTransient(utilities.StoragePool{
	Name: "My-Pool-logical",
	Type: utilities.PoolTypeLogical,
	Source: &utilities.StoragePoolSource{
		Name: "vg0",
	},
}, 0)

// disk
uuid, err = conn.CreateTransient(utilities.StoragePool{
	Name: "My-Pool-disk",
	Type: utilities.PoolTypeDisk,
	Target: &utilities.StoragePoolTarget{
		Path: "/dev",
	},
	Source: &utilities.StoragePoolSource{
		Device: []utilities.StoragePoolSourceDevice{{
			Path: "/dev/sdb",
		}},
		Format: &utilities.StoragePoolSourceFormat{
			Type: "dos",
		},
	},
}, 0)

// iscsi
uuid, err = conn.CreateTransient(utilities.StoragePool{
	Name: "My-Pool-iscsi",
	Type: utilities.PoolTypeISCSI,
	Target: &utilities.StoragePoolTarget{
		Path: "/dev/disk/by-path",
	},
	Source: &utilities.StoragePoolSource{
		Host: []utilities.StoragePoolSourceHost{{
			Name: "storage.example",
			Port: "3260",
		}},
		Device: []utilities.StoragePoolSourceDevice{{
			Path: "iqn.2010-05.com.example:target",
		}},
		Auth: &utilities.StoragePoolSourceAuth{
			Type: "chap",
			Username: "user",
			Secret: &utilities.StoragePoolSourceAuthSecret{
				Usage: "libvirtiscsi",
			},
		},
	},
}, 0)

// iscsi-direct
uuid, err = conn.CreateTransient(utilities.StoragePool{
	Name: "My-Pool-iscsi-direct",
	Type: utilities.PoolTypeISCSIDirect,
	Source: &utilities.StoragePoolSource{
		Host: []utilities.StoragePoolSourceHost{{
			Name: "storage.example",
			Port: "3260",
		}},
		Device: []utilities.StoragePoolSourceDevice{{
			Path: "iqn.2010-05.com.example:target",
		}},
		Initiator: &utilities.StoragePoolSourceInitiator{
			IQN: utilities.StoragePoolSourceInitiatorIQN{
				Name: "iqn.2010-05.com.example:initiator",
			},
		},
	},
}, 0)

// scsi (scsi_host)
uuid, err = conn.CreateTransient(utilities.StoragePool{
	Name: "My-Pool-scsi",
	Type: utilities.PoolTypeSCSI,
	Target: &utilities.StoragePoolTarget{
		Path: "/dev/disk/by-path",
	},
	Source: &utilities.StoragePoolSource{
		Adapter: &utilities.StoragePoolSourceAdapter{
			Type: "scsi_host",
			Name: "scsi_host5",
		},
	},
}, 0)

// scsi (fc_host)
uuid, err = conn.CreateTransient(utilities.StoragePool{
	Name: "My-Pool-scsi",
	Type: utilities.PoolTypeSCSI,
	Target: &utilities.StoragePoolTarget{
		Path: "/dev/disk/by-path",
	},
	Source: &utilities.StoragePoolSource{
		Adapter: &utilities.StoragePoolSourceAdapter{
			Type: "fc_host",
			WWNN: "20000000c9848140",
			WWPN: "10000000c9848140",
		},
	},
}, 0)

// mpath
uuid, err = conn.CreateTransient(utilities.StoragePool{
	Name: "My-Pool-mpath",
	Type: utilities.PoolTypeMpath,
	Target: &utilities.StoragePoolTarget{
		Path: "/dev/mapper",
	},
}, 0)

// rbd (Ceph)
uuid, err = conn.CreateTransient(utilities.StoragePool{
	Name: "My-Pool-rbd",
	Type: utilities.PoolTypeRBD,
	Source: &utilities.StoragePoolSource{
		Name: "libvirt-pool",
		Host: []utilities.StoragePoolSourceHost{{
			Name: "mon.example",
		}},
		Auth: &utilities.StoragePoolSourceAuth{
			Type: "ceph",
			Username: "libvirt",
			Secret: &utilities.StoragePoolSourceAuthSecret{
				UUID: "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
			},
		},
	},
}, 0)

// sheepdog
uuid, err = conn.CreateTransient(utilities.StoragePool{
	Name: "My-Pool-sheepdog",
	Type: utilities.PoolTypeSheepdog,
	Source: &utilities.StoragePoolSource{
		Name: "vd",
		Host: []utilities.StoragePoolSourceHost{{
			Name: "localhost",
		}},
	},
}, 0)

// gluster
uuid, err = conn.CreateTransient(utilities.StoragePool{
	Name: "My-Pool-gluster",
	Type: utilities.PoolTypeGluster,
	Source: &utilities.StoragePoolSource{
		Name: "vol1",
		Host: []utilities.StoragePoolSourceHost{{
			Name: "111.222.111.222",
		}},
		Dir: &utilities.StoragePoolSourceDir{
			Path: "/",
		},
	},
}, 0)

// zfs
uuid, err = conn.CreateTransient(utilities.StoragePool{
	Name: "My-Pool-zfs",
	Type: utilities.PoolTypeZFS,
	Source: &utilities.StoragePoolSource{
		Name: "tank",
	},
}, 0)

// vstorage
uuid, err = conn.CreateTransient(utilities.StoragePool{
	Name: "My-Pool-vstorage",
	Type: utilities.PoolTypeVStorage,
	Target: &utilities.StoragePoolTarget{
		Path: "/mnt/tmp-vstorage",
	},
	Source: &utilities.StoragePoolSource{
		Name: "cluster1",
	},
}, 0)
```

#### `CreateAs(params utilities.CreateAsParams, flags) (uuid string, error)`

`CreateAsParams.BuildModel` then `CreateTransient` (`virsh pool-create-as`). Same fields as `DefineAsParams`. Teardown: `Destroy` only.

Flags are the same `utilities.CreateFlags` as [`CreateTransient`](#createtransientmodel-flags-uuid-string-error) / [`Start`](#startref-flags-error):

| Flag | Value | Meaning |
|------|-------|---------|
| `utilities.Flags.Create.Normal` | `0` | Create and start the transient pool **without** building. Equivalent to passing `0`. |
| `utilities.Flags.Create.WithBuild` | `1 << 0` | Create/start and perform `Build` with default build flags (`utilities.Flags.Build.New`). |
| `utilities.Flags.Create.WithBuildOverwrite` | `1 << 1` | Create/start and build using `utilities.Flags.Build.Overwrite` (wipe existing data). Mutually exclusive with `WithBuildNoOverwrite`. |
| `utilities.Flags.Create.WithBuildNoOverwrite` | `1 << 2` | Create/start and build using `utilities.Flags.Build.NoOverwrite` (refuse if data would be overwritten). Mutually exclusive with `WithBuildOverwrite`. |

```go
params := utilities.CreateAsParams{
	Name: "My-Pool-dir",
	Type: utilities.PoolTypeDir,
	TargetPath: "/tmp/virest-pool",
}

// create-as + start only (no build)
uuid, err := conn.CreateAs(params, utilities.Flags.Create.Normal)
// equivalent:
uuid, err = conn.CreateAs(params, 0)
defer conn.Destroy(uuid)

// create-as + start + build (new)
uuid, err = conn.CreateAs(params, utilities.Flags.Create.WithBuild)

// create-as + start + destructive build
uuid, err = conn.CreateAs(params, utilities.Flags.Create.WithBuildOverwrite)

// create-as + start + safe build (no overwrite)
uuid, err = conn.CreateAs(params, utilities.Flags.Create.WithBuildNoOverwrite)
```

Examples for common backends:

```go
// dir
uuid, err := conn.CreateAs(utilities.CreateAsParams{
	Name: "My-Pool-dir",
	Type: utilities.PoolTypeDir,
	TargetPath: "/tmp/virest-pool",
}, 0)
defer conn.Destroy(uuid)

// fs
uuid, err = conn.CreateAs(utilities.CreateAsParams{
	Name: "My-Pool-fs",
	Type: utilities.PoolTypeFS,
	TargetPath: "/mnt/tmp-data",
	SourceDev: "/dev/sdb1",
	SourceFormat: "ext4",
}, utilities.Flags.Create.WithBuild)

// netfs (NFS)
uuid, err = conn.CreateAs(utilities.CreateAsParams{
	Name: "My-Pool-netfs",
	Type: utilities.PoolTypeNetFS,
	TargetPath: "/mnt/tmp-nfs",
	SourceHost: "nfs.example",
	SourcePath: "/export/images",
	SourceFormat: "nfs",
	SourceProtocolVer: "4",
}, 0)

// logical (LVM; SourceName defaults to Name when omitted)
uuid, err = conn.CreateAs(utilities.CreateAsParams{
	Name: "My-Pool-logical",
	Type: utilities.PoolTypeLogical,
	SourceName: "vg0",
}, 0)

// disk
uuid, err = conn.CreateAs(utilities.CreateAsParams{
	Name: "My-Pool-disk",
	Type: utilities.PoolTypeDisk,
	TargetPath: "/dev",
	SourceDev: "/dev/sdb",
	SourceFormat: "dos",
}, 0)

// iscsi
uuid, err = conn.CreateAs(utilities.CreateAsParams{
	Name: "My-Pool-iscsi",
	Type: utilities.PoolTypeISCSI,
	SourceHost: "storage.example",
	SourcePort: 3260,
	SourceDev: "iqn.2010-05.com.example:target",
	TargetPath: "/dev/disk/by-path",
	AuthType: "chap",
	AuthUsername: "user",
	SecretUsage: "libvirtiscsi",
}, 0)

// iscsi-direct
uuid, err = conn.CreateAs(utilities.CreateAsParams{
	Name: "My-Pool-iscsi-direct",
	Type: utilities.PoolTypeISCSIDirect,
	SourceHost: "storage.example",
	SourcePort: 3260,
	SourceDev: "iqn.2010-05.com.example:target",
	SourceInitiator: "iqn.2010-05.com.example:initiator",
}, 0)

// scsi (scsi_host)
uuid, err = conn.CreateAs(utilities.CreateAsParams{
	Name: "My-Pool-scsi",
	Type: utilities.PoolTypeSCSI,
	TargetPath: "/dev/disk/by-path",
	AdapterName: "scsi_host5",
}, 0)

// scsi (fc_host)
uuid, err = conn.CreateAs(utilities.CreateAsParams{
	Name: "My-Pool-scsi",
	Type: utilities.PoolTypeSCSI,
	TargetPath: "/dev/disk/by-path",
	AdapterWWNN: "20000000c9848140",
	AdapterWWPN: "10000000c9848140",
}, 0)

// mpath
uuid, err = conn.CreateAs(utilities.CreateAsParams{
	Name: "My-Pool-mpath",
	Type: utilities.PoolTypeMpath,
	TargetPath: "/dev/mapper",
}, 0)

// rbd (Ceph)
uuid, err = conn.CreateAs(utilities.CreateAsParams{
	Name: "My-Pool-rbd",
	Type: utilities.PoolTypeRBD,
	SourceHost: "mon.example",
	SourceName: "libvirt-pool",
	AuthType: "ceph",
	AuthUsername: "libvirt",
	SecretUUID: "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
}, 0)

// sheepdog
uuid, err = conn.CreateAs(utilities.CreateAsParams{
	Name: "My-Pool-sheepdog",
	Type: utilities.PoolTypeSheepdog,
	SourceHost: "localhost",
	SourceName: "vd",
}, 0)

// gluster
uuid, err = conn.CreateAs(utilities.CreateAsParams{
	Name: "My-Pool-gluster",
	Type: utilities.PoolTypeGluster,
	SourceHost: "111.222.111.222",
	SourceName: "vol1",
	SourcePath: "/",
}, 0)

// zfs
uuid, err = conn.CreateAs(utilities.CreateAsParams{
	Name: "My-Pool-zfs",
	Type: utilities.PoolTypeZFS,
	SourceName: "tank",
}, 0)

// vstorage
uuid, err = conn.CreateAs(utilities.CreateAsParams{
	Name: "My-Pool-vstorage",
	Type: utilities.PoolTypeVStorage,
	SourceName: "cluster1",
	TargetPath: "/mnt/tmp-vstorage",
}, 0)
```

#### `Build(ref string, flags) error`

Builds underlying storage for a **defined**, typically **inactive** pool (`virsh pool-build`). `ref` is name or UUID. Some types (e.g. `dir`) may fail harmlessly depending on host state — callers decide whether to continue.

Flags are `utilities.BuildFlags` (bitwise-OR where allowed). Libvirt notes that **`OVERWRITE` / `NO_OVERWRITE` currently apply mainly to filesystem pools**.

| Flag | Value | Meaning |
|------|-------|---------|
| `utilities.Flags.Build.New` | `0` | Regular build from scratch (default). Initialize the pool’s backing store as a new pool. |
| `utilities.Flags.Build.Repair` | `1 << 0` | Repair / reinitialize an existing pool’s metadata or structure when the backend supports it. |
| `utilities.Flags.Build.Resize` | `1 << 1` | Extend / resize an existing pool (backend-dependent). |
| `utilities.Flags.Build.NoOverwrite` | `1 << 2` | Build only if it would **not** destroy existing data (refuse overwrite). Prefer when the target may already contain a filesystem or pool. |
| `utilities.Flags.Build.Overwrite` | `1 << 3` | Build and **overwrite** existing data on the target. Destructive — use only when intentional wipe is required. |

`NO_OVERWRITE` and `OVERWRITE` are mutually exclusive intents; do not OR them together.

```go
// default: new pool from scratch
err := conn.Build("My-Pool-dir", utilities.Flags.Build.New)
// equivalent:
err = conn.Build("My-Pool-dir", 0)

// repair / reinitialize
err = conn.Build("My-Pool-dir", utilities.Flags.Build.Repair)

// extend existing pool (if supported)
err = conn.Build("My-Pool-dir", utilities.Flags.Build.Resize)

// safe build: fail rather than clobber existing data (esp. fs pools)
err = conn.Build("My-Pool-fs", utilities.Flags.Build.NoOverwrite)

// destructive build: wipe and rebuild (esp. fs pools)
err = conn.Build("My-Pool-fs", utilities.Flags.Build.Overwrite)
```

#### `Start(ref, flags) error`

Starts an **already defined, inactive** pool (`virsh pool-start`). Do **not** confuse with `CreateTransient` (which creates + starts a non-persistent pool from a model).

Flags are `utilities.CreateFlags`. They control whether start also runs an implicit **build**:

| Flag | Value | Meaning |
|------|-------|---------|
| `utilities.Flags.Create.Normal` | `0` | Start the pool **without** building. Use when the pool was already built (or needs no build). |
| `utilities.Flags.Create.WithBuild` | `1 << 0` | Start and perform `Build` with default build flags (`utilities.Flags.Build.New`). Handy for first-time start of pools that need initialization. |
| `utilities.Flags.Create.WithBuildOverwrite` | `1 << 1` | Start and build using `utilities.Flags.Build.Overwrite` (wipe existing data). Mutually exclusive with `WithBuildNoOverwrite`. |
| `utilities.Flags.Create.WithBuildNoOverwrite` | `1 << 2` | Start and build using `utilities.Flags.Build.NoOverwrite` (refuse if data would be overwritten). Mutually exclusive with `WithBuildOverwrite`. |

```go
// start only (most common after Define + optional Build)
err := conn.Start("My-Pool-dir", utilities.Flags.Create.Normal)
err = conn.Start("My-Pool-dir", 0)

// start + build (new)
err = conn.Start("My-Pool-dir", utilities.Flags.Create.WithBuild)

// start + destructive build
err = conn.Start("My-Pool-fs", utilities.Flags.Create.WithBuildOverwrite)

// start + safe build (no overwrite)
err = conn.Start("My-Pool-fs", utilities.Flags.Create.WithBuildNoOverwrite)
```

#### `Destroy(ref) error`

Stops an active pool (`virsh pool-destroy`). Does not remove the definition (persistent pools remain defined).

```go
err := conn.Destroy("My-Pool-dir")
```

#### `Delete(ref, flags) error`

Deletes **backing storage resources** of a pool (`virsh pool-delete`). Destructive; pool is usually inactive. Does **not** remove the persistent XML definition — follow with `Undefine` when retiring a persistent pool. Use carefully on HA nodes.

Flags are `utilities.DeleteFlags`:

| Flag | Value | Meaning |
|------|-------|---------|
| `utilities.Flags.Delete.Normal` | `0` | Delete pool **metadata** / remove the pool’s managed resources in the normal (fast) way. Does not necessarily zero every block of underlying media. |
| `utilities.Flags.Delete.Zeroed` | `1 << 0` | Overwrite / clear underlying data to **zeros** before delete. Much slower; use when residual data must not remain on the media. |

```go
// fast / normal resource delete
err := conn.Delete("My-Pool-dir", utilities.Flags.Delete.Normal)
// equivalent:
err = conn.Delete("My-Pool-dir", 0)

// wipe data then delete (slow)
err = conn.Delete("My-Pool-dir", utilities.Flags.Delete.Zeroed)

// typical persistent teardown after Destroy:
_ = conn.Destroy("My-Pool-dir")
_ = conn.Delete("My-Pool-dir", utilities.Flags.Delete.Normal)
_ = conn.Undefine("My-Pool-dir")
```

#### `Undefine(ref) error`

Removes a **persistent** pool definition (`virsh pool-undefine`). Pool should be inactive. Not used for transient pools.

```go
err := conn.Undefine("My-Pool-dir")
```

#### `Refresh(ref) error`

Re-scans volumes in the pool (`virsh pool-refresh`). Flags fixed to `0` in this library.

```go
err := conn.Refresh("My-Pool-dir")
```

#### `Autostart(ref, autostart bool) error`

Enables or disables start-with-host (`virsh pool-autostart` / `virsh pool-autostart --disable`). Applies to **persistent** pools.

```go
// enable: pool starts automatically when libvirtd / the host comes up
err := conn.Autostart("My-Pool-dir", true)

// disable: pool remains defined but will not autostart
err = conn.Autostart("My-Pool-dir", false)
```

---

### Query / list

#### `List(listFlags ListFlags, xmlFlags XMLFlags) ([]PoolSummary, error)`

Primary `virsh pool-list` API. Pass `utilities.ListFlags` as `listFlags` (bitwise-OR combinations via `utilities.Flags.List.*`). `xmlFlags` are `utilities.XMLFlags` (`Flags.XML.*`) used when reading each pool’s XML for size fields. Returns typed summaries (not a text table). Attribute reads per pool are **sequential** (HA-safe; avoids N×goroutines). Prefer `utilities.Flags` — not raw `libvirt.CONNECT_LIST_*` / `STORAGE_XML_*` on the Connection API.

**`listFlags` (`utilities.ListFlags` / `utilities.Flags.List`):**

| Flag | Meaning |
|------|---------|
| `utilities.Flags.List.Inactive` | Inactive pools |
| `utilities.Flags.List.Active` | Active pools |
| `utilities.Flags.List.Persistent` | Persistent pools |
| `utilities.Flags.List.Transient` | Transient pools |
| `utilities.Flags.List.Autostart` | Autostart enabled |
| `utilities.Flags.List.NoAutostart` | Autostart disabled |
| `utilities.Flags.List.Dir` | Type `dir` |
| `utilities.Flags.List.FS` | Type `fs` |
| `utilities.Flags.List.NetFS` | Type `netfs` |
| `utilities.Flags.List.Logical` | Type `logical` |
| `utilities.Flags.List.Disk` | Type `disk` |
| `utilities.Flags.List.ISCSI` | Type `iscsi` |
| `utilities.Flags.List.SCSI` | Type `scsi` |
| `utilities.Flags.List.Mpath` | Type `mpath` |
| `utilities.Flags.List.RBD` | Type `rbd` |
| `utilities.Flags.List.Sheepdog` | Type `sheepdog` |
| `utilities.Flags.List.Gluster` | Type `gluster` |
| `utilities.Flags.List.ZFS` | Type `zfs` |
| `utilities.Flags.List.VStorage` | Type `vstorage` |
| `utilities.Flags.List.ISCSIDirect` | Type `iscsi-direct` |

**`xmlFlags` (`utilities.XMLFlags` / `utilities.Flags.XML`):**

| Flag | Meaning |
|------|---------|
| `utilities.Flags.XML.None` (`0`) | Active / current XML description (default) |
| `utilities.Flags.XML.Inactive` | Inactive (persistent config) XML when dumping/summarizing |

```go
import "github.com/Hari-Kiri/virest/utilities"

// state filters
pools, err := conn.List(utilities.Flags.List.Active, 0)
pools, err = conn.List(utilities.Flags.List.Inactive, 0)
pools, err = conn.List(
	utilities.Flags.List.Active|utilities.Flags.List.Inactive,
	0,
)

// persistence / autostart filters
pools, err = conn.List(utilities.Flags.List.Persistent, 0)
pools, err = conn.List(utilities.Flags.List.Transient, 0)
pools, err = conn.List(utilities.Flags.List.Autostart, 0)
pools, err = conn.List(utilities.Flags.List.NoAutostart, 0)

// type filters
pools, err = conn.List(utilities.Flags.List.Dir, 0)
pools, err = conn.List(utilities.Flags.List.FS, 0)
pools, err = conn.List(utilities.Flags.List.NetFS, 0)
pools, err = conn.List(utilities.Flags.List.Logical, 0)
pools, err = conn.List(utilities.Flags.List.Disk, 0)
pools, err = conn.List(utilities.Flags.List.ISCSI, 0)
pools, err = conn.List(utilities.Flags.List.SCSI, 0)
pools, err = conn.List(utilities.Flags.List.Mpath, 0)
pools, err = conn.List(utilities.Flags.List.RBD, 0)
pools, err = conn.List(utilities.Flags.List.Sheepdog, 0)
pools, err = conn.List(utilities.Flags.List.Gluster, 0)
pools, err = conn.List(utilities.Flags.List.ZFS, 0)
pools, err = conn.List(utilities.Flags.List.VStorage, 0)
pools, err = conn.List(utilities.Flags.List.ISCSIDirect, 0)

// combine filters (example: active dir pools) + inactive XML for size fields
pools, err = conn.List(
	utilities.Flags.List.Active|utilities.Flags.List.Dir,
	utilities.Flags.XML.Inactive,
)
```

`PoolSummary` fields: `Uuid`, `Name`, `State`, `Autostart`, `Persistent`, `Capacity`, `Allocation`, `Available`.

#### `ListActive(xmlFlags XMLFlags) ([]PoolSummary, error)`

≈ default `virsh pool-list` (active pools). Calls `List` with `utilities.Flags.List.Active`.

```go
import "github.com/Hari-Kiri/virest/utilities"

pools, err := conn.ListActive(0)
pools, err = conn.ListActive(utilities.Flags.XML.Inactive)
```

#### `ListInactive(xmlFlags XMLFlags) ([]PoolSummary, error)`

≈ `virsh pool-list --inactive`. Calls `List` with `utilities.Flags.List.Inactive`.

```go
import "github.com/Hari-Kiri/virest/utilities"

pools, err := conn.ListInactive(0)
pools, err = conn.ListInactive(utilities.Flags.XML.Inactive)
```

#### `ListAll(xmlFlags XMLFlags) ([]PoolSummary, error)`

≈ `virsh pool-list --all` (active | inactive).

```go
import "github.com/Hari-Kiri/virest/utilities"

pools, err := conn.ListAll(0)
pools, err = conn.ListAll(utilities.Flags.XML.Inactive)
```

#### `Info(ref) (Info, error)`

Volatile info: name, UUID, state, persistent, autostart, capacity/allocation/available (`virsh pool-info`). When `ref` is a name, UUID is resolved from the pool handle (one extra get after lookup).

```go
info, err := conn.Info("My-Pool-dir")
fmt.Printf("state=%v capacity=%d available=%d\n", info.State, info.Capacity, info.Available)
```

#### `Dump(ref, xmlFlags XMLFlags) (Detail, error)`

Full pool description as `utilities.StoragePool` embedded in return type `Detail` (`virsh pool-dumpxml`). `xmlFlags` are `utilities.XMLFlags` (`Flags.XML.None` or `Flags.XML.Inactive`).

```go
import "github.com/Hari-Kiri/virest/utilities"

detail, err := conn.Dump("My-Pool-dir", 0)
detail, err = conn.Dump("My-Pool-dir", utilities.Flags.XML.Inactive)
fmt.Printf("type=%s target=%s\n", detail.Type, detail.Target.Path)

// pool-edit substitute: dump → mutate → define (pool must be inactive)
detail.Target.Path = "/new/path"
_, err = conn.Define(detail.StoragePool, 0)
```

#### `Name(ref) (string, error)`

Returns pool name for a UUID (or name) ref (`virsh pool-name`).

```go
name, err := conn.Name("a1b2c3d4-e5f6-7890-abcd-ef1234567890")
```

#### `UUIDByName(name) (string, error)`

Returns UUID for a pool name (`virsh pool-uuid`).

```go
uuid, err := conn.UUIDByName("images")
```

---

### Discovery

#### `Capabilities() (Capabilities, error)`

Supported pool types and format options (`virsh pool-capabilities`).

```go
caps, err := conn.Capabilities()
for i := 0; i < len(caps.Pool); i++ {
	fmt.Println(caps.Pool[i].Type, caps.Pool[i].Supported)
}
```

#### `FindSources(poolType utilities.PoolType, src SourceSpec) (Sources, error)`

Discovers storage sources (`virsh find-storage-pool-sources`).

```go
sources, err := conn.FindSources(utilities.PoolTypeISCSI, storagePool.SourceSpec{
	Host: storagePool.Host{Name: "storage.example", Port: 3260},
	Initiator: storagePool.Initiator{
		Iqn: storagePool.Iqn{Name: "iqn.1993-08.org.debian:01:abc"},
	},
})
```

#### `FindSourcesAs(poolType utilities.PoolType, params utilities.FindSourcesAsParams) (Sources, error)`

Same as `FindSources` with virsh `-as` style fields (`virsh find-storage-pool-sources-as`).

```go
sources, err := conn.FindSourcesAs(utilities.PoolTypeISCSI, utilities.FindSourcesAsParams{
	Host:      "storage.example",
	Port:      3260,
	Initiator: "iqn.1993-08.org.debian:01:abc",
})
```

---

### Events

Prerequisite: the process must register and run the libvirt event loop:

```go
storagePool.RegisterEventLoop()
```

Without that loop, `WaitEvent` / `StreamEvents` will hang or fail.

Callers do not import `libvirt.org/go/libvirt`. Pool state uses `PoolState` (`Info.State`, `PoolSummary.State`) and event lifecycle uses `PoolEventLifecycle` (`Event.EventLifecycle`).

#### `WaitEvent(ref, kind EventKind) (Event, error)`

Blocks until one event (`virsh pool-event` one-shot style).

| `EventKind` | Meaning |
|-------------|---------|
| `EventLifecycle` (0) | Lifecycle events |
| `EventRefresh` (1) | Refresh events |

```go
ev, err := conn.WaitEvent("images", storagePool.EventLifecycle)
```

#### `StreamEvents(ctx, ref, kind, emit) error`

Registers for events and calls `emit` until `ctx` is cancelled or `emit` returns an error. Buffer size 8; when full, events may be **dropped**. Cancel / deadline → returns `nil`.

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

err := conn.StreamEvents(ctx, "images", storagePool.EventLifecycle, func(ev storagePool.Event) error {
	fmt.Printf("event id=%d\n", ev.EventId)
	return nil
})
```

---

## Persistent vs transient flows

### Persistent (survives host reboot if defined + autostart)

```text
Define / DefineAs
    → Build (optional; type-dependent)
    → Start / Create
    → Autostart(true) optional
    → Refresh / Info / Dump as needed
    → Destroy
    → Delete (optional, removes backing store)
    → Undefine
```

```go
uuid, err := conn.DefineAs(utilities.DefineAsParams{
	Name: "My-Pool-dir", Type: utilities.PoolTypeDir, TargetPath: "/var/lib/libvirt/images",
}, 0)
if err != nil {
	return err
}
_ = conn.Build(uuid, 0)
if err := conn.Start(uuid, 0); err != nil {
	return err
}
_ = conn.Autostart(uuid, true)

// later teardown
_ = conn.Destroy(uuid)
_ = conn.Undefine(uuid)
```

### Transient (active until destroy / reboot)

```text
CreateTransient / CreateAs → use → Destroy only
```

```go
uuid, err := conn.CreateAs(utilities.CreateAsParams{
	Name: "My-Pool-dir", Type: utilities.PoolTypeDir, TargetPath: "/tmp/virest-pool",
}, 0)
if err != nil {
	return err
}
defer conn.Destroy(uuid)
```

---

## Errors and Free

- Errors are wrapped with an operation name (`utilities.Wrap`), e.g. `lookup storage pool: …`.
- Inspect libvirt codes with `utilities.Code(err)` or `utilities.AsLibvirtError(err)`.
- Callers **must** `Close` the connection.
- The library **Free**s pool handles it acquires; callers never manage `poolHandle`.
- Methods on a closed/nil connection return `connection is closed`.

```go
err := conn.Destroy("missing-pool")
if code, ok := utilities.Code(err); ok {
	fmt.Println("libvirt code:", code)
}
```

---

## File map

| File | Role |
|------|------|
| `connect.go` | `Connect`, `Close`, `lookupPool` |
| `lifecycle.go` | Define / CreateTransient / `-As`, Build, Start, Destroy, Delete, Undefine, Refresh, Autostart |
| `list_info.go` | Info, List + Active/Inactive/All |
| `detail.go` | Dump, Capabilities, FindSources |
| `query_short.go` | Name, UUIDByName, FindSourcesAs |
| `event.go` | WaitEvent, StreamEvents |
| `hypervisor.go` | Internal `hypervisor` / `poolHandle` ports |
| `libvirt_adapter.go` | Production libvirt wiring |
| `types.go` | `Info`, `Detail`, `PoolSummary`, `Event`, `SourceSpec`, `Capabilities`, … |
| `errors.go` | `finishFree`, wrap helpers |
| `utilities/poolref.go` | `LooksLikeUUID` |
| `utilities/pool_type.go` | `PoolType` constants (`PoolTypeDir`, `PoolTypeISCSI`, …) |
| `utilities/pool_model.go` | `DefineAsParams`, `CreateAsParams` (unexported shared `poolAsParams`), `FindSourcesAsParams` |
| `utilities/flags.go` | `Flags`, `DefineFlags`, `CreateFlags`, `BuildFlags`, `DeleteFlags` |
| `utilities/xml_model.go` | Brand-neutral `StoragePool` / target / source / size aliases |

---

## Related

- Root [README.md](../README.md)
- Runnable demo: [examples/basic](../examples/basic)
- HTTP lab harness: `cmd/testserver` (Bearer + `Hypervisor-Uri`)
- Shared helpers: [utilities](../utilities) (`StoragePool` models, `DefineAsParams` / `CreateAsParams`, `LooksLikeUUID`, …)
