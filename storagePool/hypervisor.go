package storagePool

import (
	"libvirt.org/go/libvirt"
)

// hypervisor is the connection-level surface used by Connection methods.
// Tests inject fakes; production uses a libvirt-backed adapter.
type hypervisor interface {
	Close() (int, error)
	ListAllStoragePools(flags libvirt.ConnectListAllStoragePoolsFlags) ([]poolHandle, error)
	LookupStoragePoolByUUIDString(uuid string) (poolHandle, error)
	LookupStoragePoolByName(name string) (poolHandle, error)
	StoragePoolDefineXML(xmlConfig string, flags libvirt.StoragePoolDefineFlags) (poolHandle, error)
	CreateTransient(config string, flags libvirt.StoragePoolCreateFlags) (poolHandle, error)
	GetStoragePoolCapabilities(flags uint32) (string, error)
	FindStoragePoolSources(poolType, srcSpec string, flags uint32) (string, error)
	StoragePoolEventLifecycleRegister(pool poolHandle, callback libvirt.StoragePoolEventLifecycleCallback) (int, error)
	StoragePoolEventRefreshRegister(pool poolHandle, callback libvirt.StoragePoolEventGenericCallback) (int, error)
	StoragePoolEventDeregister(callbackID int) error
}

// poolHandle is the storage-pool object surface used by Connection methods.
type poolHandle interface {
	Free() error
	Ref() error
	GetUUIDString() (string, error)
	GetName() (string, error)
	GetInfo() (*libvirt.StoragePoolInfo, error)
	GetAutostart() (bool, error)
	IsPersistent() (bool, error)
	GetXMLDesc(flags libvirt.StorageXMLFlags) (string, error)
	Build(flags libvirt.StoragePoolBuildFlags) error
	Create(flags libvirt.StoragePoolCreateFlags) error
	Destroy() error
	Delete(flags libvirt.StoragePoolDeleteFlags) error
	Undefine() error
	Refresh(flags uint32) error
	SetAutostart(autostart bool) error
	// underlying returns the native libvirt pool when available (events need it).
	underlying() *libvirt.StoragePool
}
