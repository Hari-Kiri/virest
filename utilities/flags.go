package utilities

import "libvirt.org/go/libvirt"

// Brand-neutral storage-pool flag types (wrap libvirt flag enums).
// Prefer Flags.* constants and these types on Connection APIs — not libvirt imports.

// DefineFlags wraps libvirt.StoragePoolDefineFlags (Define / DefineAs).
type DefineFlags uint

// CreateFlags wraps libvirt.StoragePoolCreateFlags (Start / CreateTransient / CreateAs).
type CreateFlags uint

// BuildFlags wraps libvirt.StoragePoolBuildFlags (Build).
type BuildFlags uint

// DeleteFlags wraps libvirt.StoragePoolDeleteFlags (Delete).
type DeleteFlags uint

// ListFlags wraps libvirt.ConnectListAllStoragePoolsFlags (List / List*).
type ListFlags uint

// XMLFlags wraps libvirt.StorageXMLFlags (List xmlFlags, Dump).
type XMLFlags uint

// Libvirt returns the underlying libvirt define flags.
func (f DefineFlags) Libvirt() libvirt.StoragePoolDefineFlags {
	return libvirt.StoragePoolDefineFlags(f)
}

// Libvirt returns the underlying libvirt create/start flags.
func (f CreateFlags) Libvirt() libvirt.StoragePoolCreateFlags {
	return libvirt.StoragePoolCreateFlags(f)
}

// Libvirt returns the underlying libvirt build flags.
func (f BuildFlags) Libvirt() libvirt.StoragePoolBuildFlags {
	return libvirt.StoragePoolBuildFlags(f)
}

// Libvirt returns the underlying libvirt delete flags.
func (f DeleteFlags) Libvirt() libvirt.StoragePoolDeleteFlags {
	return libvirt.StoragePoolDeleteFlags(f)
}

// Libvirt returns the underlying libvirt list-all-storage-pools flags.
func (f ListFlags) Libvirt() libvirt.ConnectListAllStoragePoolsFlags {
	return libvirt.ConnectListAllStoragePoolsFlags(f)
}

// Libvirt returns the underlying libvirt storage XML flags.
func (f XMLFlags) Libvirt() libvirt.StorageXMLFlags {
	return libvirt.StorageXMLFlags(f)
}

type defineFlagConsts struct {
	None     DefineFlags
	Validate DefineFlags
}

type createFlagConsts struct {
	Normal               CreateFlags
	WithBuild            CreateFlags
	WithBuildOverwrite   CreateFlags
	WithBuildNoOverwrite CreateFlags
}

type buildFlagConsts struct {
	New         BuildFlags
	Repair      BuildFlags
	Resize      BuildFlags
	NoOverwrite BuildFlags
	Overwrite   BuildFlags
}

type deleteFlagConsts struct {
	Normal DeleteFlags
	Zeroed DeleteFlags
}

type listFlagConsts struct {
	Inactive    ListFlags
	Active      ListFlags
	Persistent  ListFlags
	Transient   ListFlags
	Autostart   ListFlags
	NoAutostart ListFlags
	Dir         ListFlags
	FS          ListFlags
	NetFS       ListFlags
	Logical     ListFlags
	Disk        ListFlags
	ISCSI       ListFlags
	SCSI        ListFlags
	Mpath       ListFlags
	RBD         ListFlags
	Sheepdog    ListFlags
	Gluster     ListFlags
	ZFS         ListFlags
	VStorage    ListFlags
	ISCSIDirect ListFlags
}

type xmlFlagConsts struct {
	None     XMLFlags
	Inactive XMLFlags
}

type flagsRoot struct {
	Define defineFlagConsts
	Create createFlagConsts
	Build  buildFlagConsts
	Delete deleteFlagConsts
	List   listFlagConsts
	XML    xmlFlagConsts
}

// Flags groups all storage-pool libvirt flag constants for callers.
var Flags = flagsRoot{
	Define: defineFlagConsts{
		None:     0,
		Validate: DefineFlags(libvirt.STORAGE_POOL_DEFINE_VALIDATE),
	},
	Create: createFlagConsts{
		Normal:               CreateFlags(libvirt.STORAGE_POOL_CREATE_NORMAL),
		WithBuild:            CreateFlags(libvirt.STORAGE_POOL_CREATE_WITH_BUILD),
		WithBuildOverwrite:   CreateFlags(libvirt.STORAGE_POOL_CREATE_WITH_BUILD_OVERWRITE),
		WithBuildNoOverwrite: CreateFlags(libvirt.STORAGE_POOL_CREATE_WITH_BUILD_NO_OVERWRITE),
	},
	Build: buildFlagConsts{
		New:         BuildFlags(libvirt.STORAGE_POOL_BUILD_NEW),
		Repair:      BuildFlags(libvirt.STORAGE_POOL_BUILD_REPAIR),
		Resize:      BuildFlags(libvirt.STORAGE_POOL_BUILD_RESIZE),
		NoOverwrite: BuildFlags(libvirt.STORAGE_POOL_BUILD_NO_OVERWRITE),
		Overwrite:   BuildFlags(libvirt.STORAGE_POOL_BUILD_OVERWRITE),
	},
	Delete: deleteFlagConsts{
		Normal: DeleteFlags(libvirt.STORAGE_POOL_DELETE_NORMAL),
		Zeroed: DeleteFlags(libvirt.STORAGE_POOL_DELETE_ZEROED),
	},
	List: listFlagConsts{
		Inactive:    ListFlags(libvirt.CONNECT_LIST_STORAGE_POOLS_INACTIVE),
		Active:      ListFlags(libvirt.CONNECT_LIST_STORAGE_POOLS_ACTIVE),
		Persistent:  ListFlags(libvirt.CONNECT_LIST_STORAGE_POOLS_PERSISTENT),
		Transient:   ListFlags(libvirt.CONNECT_LIST_STORAGE_POOLS_TRANSIENT),
		Autostart:   ListFlags(libvirt.CONNECT_LIST_STORAGE_POOLS_AUTOSTART),
		NoAutostart: ListFlags(libvirt.CONNECT_LIST_STORAGE_POOLS_NO_AUTOSTART),
		Dir:         ListFlags(libvirt.CONNECT_LIST_STORAGE_POOLS_DIR),
		FS:          ListFlags(libvirt.CONNECT_LIST_STORAGE_POOLS_FS),
		NetFS:       ListFlags(libvirt.CONNECT_LIST_STORAGE_POOLS_NETFS),
		Logical:     ListFlags(libvirt.CONNECT_LIST_STORAGE_POOLS_LOGICAL),
		Disk:        ListFlags(libvirt.CONNECT_LIST_STORAGE_POOLS_DISK),
		ISCSI:       ListFlags(libvirt.CONNECT_LIST_STORAGE_POOLS_ISCSI),
		SCSI:        ListFlags(libvirt.CONNECT_LIST_STORAGE_POOLS_SCSI),
		Mpath:       ListFlags(libvirt.CONNECT_LIST_STORAGE_POOLS_MPATH),
		RBD:         ListFlags(libvirt.CONNECT_LIST_STORAGE_POOLS_RBD),
		Sheepdog:    ListFlags(libvirt.CONNECT_LIST_STORAGE_POOLS_SHEEPDOG),
		Gluster:     ListFlags(libvirt.CONNECT_LIST_STORAGE_POOLS_GLUSTER),
		ZFS:         ListFlags(libvirt.CONNECT_LIST_STORAGE_POOLS_ZFS),
		VStorage:    ListFlags(libvirt.CONNECT_LIST_STORAGE_POOLS_VSTORAGE),
		ISCSIDirect: ListFlags(libvirt.CONNECT_LIST_STORAGE_POOLS_ISCSI_DIRECT),
	},
	XML: xmlFlagConsts{
		None:     0,
		Inactive: XMLFlags(libvirt.STORAGE_XML_INACTIVE),
	},
}
