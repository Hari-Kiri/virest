package utilities

// PoolType is a libvirt storage-pool type (backend / protocol).
// Prefer these constants over free strings for pool Type fields and FindSources.
type PoolType string

// Supported storage-pool types (libvirt pool type attribute).
const (
	PoolTypeDir         PoolType = "dir"
	PoolTypeFS          PoolType = "fs"
	PoolTypeNetFS       PoolType = "netfs"
	PoolTypeLogical     PoolType = "logical"
	PoolTypeDisk        PoolType = "disk"
	PoolTypeISCSI       PoolType = "iscsi"
	PoolTypeISCSIDirect PoolType = "iscsi-direct"
	PoolTypeSCSI        PoolType = "scsi"
	PoolTypeMpath       PoolType = "mpath"
	PoolTypeRBD         PoolType = "rbd"
	PoolTypeSheepdog    PoolType = "sheepdog"
	PoolTypeGluster     PoolType = "gluster"
	PoolTypeZFS         PoolType = "zfs"
	PoolTypeVStorage    PoolType = "vstorage"
)

// String returns the libvirt type attribute value.
func (t PoolType) String() string {
	return string(t)
}

// Known reports whether t is a DefineAs/CreateAs-supported pool type.
func (t PoolType) Known() bool {
	switch t {
	case PoolTypeDir,
		PoolTypeFS,
		PoolTypeNetFS,
		PoolTypeLogical,
		PoolTypeDisk,
		PoolTypeISCSI,
		PoolTypeISCSIDirect,
		PoolTypeSCSI,
		PoolTypeMpath,
		PoolTypeRBD,
		PoolTypeSheepdog,
		PoolTypeGluster,
		PoolTypeZFS,
		PoolTypeVStorage:
		return true
	}
	return false
}
