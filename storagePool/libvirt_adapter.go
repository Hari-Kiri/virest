package storagePool

import (
	"libvirt.org/go/libvirt"
)

type libvirtHypervisor struct {
	conn *libvirt.Connect
}

type libvirtPool struct {
	pool *libvirt.StoragePool
}

func (h *libvirtHypervisor) Close() (int, error) {
	return h.conn.Close()
}

func (h *libvirtHypervisor) ListAllStoragePools(flags libvirt.ConnectListAllStoragePoolsFlags) ([]poolHandle, error) {
	pools, err := h.conn.ListAllStoragePools(flags)
	if err != nil {
		return nil, err
	}
	n := len(pools)
	out := make([]poolHandle, n)
	for i := 0; i < n; i++ {
		p := pools[i]
		out[i] = &libvirtPool{pool: &p}
	}
	return out, nil
}

func (h *libvirtHypervisor) LookupStoragePoolByUUIDString(uuid string) (poolHandle, error) {
	pool, err := h.conn.LookupStoragePoolByUUIDString(uuid)
	if err != nil {
		return nil, err
	}
	return &libvirtPool{pool: pool}, nil
}

func (h *libvirtHypervisor) LookupStoragePoolByName(name string) (poolHandle, error) {
	pool, err := h.conn.LookupStoragePoolByName(name)
	if err != nil {
		return nil, err
	}
	return &libvirtPool{pool: pool}, nil
}

func (h *libvirtHypervisor) StoragePoolDefineXML(xmlConfig string, flags libvirt.StoragePoolDefineFlags) (poolHandle, error) {
	pool, err := h.conn.StoragePoolDefineXML(xmlConfig, flags)
	if err != nil {
		return nil, err
	}
	return &libvirtPool{pool: pool}, nil
}

func (h *libvirtHypervisor) CreateTransient(config string, flags libvirt.StoragePoolCreateFlags) (poolHandle, error) {
	pool, err := h.conn.StoragePoolCreateXML(config, flags)
	if err != nil {
		return nil, err
	}
	return &libvirtPool{pool: pool}, nil
}

func (h *libvirtHypervisor) GetStoragePoolCapabilities(flags uint32) (string, error) {
	return h.conn.GetStoragePoolCapabilities(flags)
}

func (h *libvirtHypervisor) FindStoragePoolSources(poolType, srcSpec string, flags uint32) (string, error) {
	return h.conn.FindStoragePoolSources(poolType, srcSpec, flags)
}

func (h *libvirtHypervisor) StoragePoolEventLifecycleRegister(pool poolHandle, callback libvirt.StoragePoolEventLifecycleCallback) (int, error) {
	native := pool.underlying()
	if native == nil {
		return 0, errNoNativePool
	}
	return h.conn.StoragePoolEventLifecycleRegister(native, callback)
}

func (h *libvirtHypervisor) StoragePoolEventRefreshRegister(pool poolHandle, callback libvirt.StoragePoolEventGenericCallback) (int, error) {
	native := pool.underlying()
	if native == nil {
		return 0, errNoNativePool
	}
	return h.conn.StoragePoolEventRefreshRegister(native, callback)
}

func (h *libvirtHypervisor) StoragePoolEventDeregister(callbackID int) error {
	return h.conn.StoragePoolEventDeregister(callbackID)
}

func (p *libvirtPool) Free() error {
	return p.pool.Free()
}

func (p *libvirtPool) Ref() error {
	return p.pool.Ref()
}

func (p *libvirtPool) GetUUIDString() (string, error) {
	return p.pool.GetUUIDString()
}

func (p *libvirtPool) GetName() (string, error) {
	return p.pool.GetName()
}

func (p *libvirtPool) GetInfo() (*libvirt.StoragePoolInfo, error) {
	return p.pool.GetInfo()
}

func (p *libvirtPool) GetAutostart() (bool, error) {
	return p.pool.GetAutostart()
}

func (p *libvirtPool) IsPersistent() (bool, error) {
	return p.pool.IsPersistent()
}

func (p *libvirtPool) GetXMLDesc(flags libvirt.StorageXMLFlags) (string, error) {
	return p.pool.GetXMLDesc(flags)
}

func (p *libvirtPool) Build(flags libvirt.StoragePoolBuildFlags) error {
	return p.pool.Build(flags)
}

func (p *libvirtPool) Create(flags libvirt.StoragePoolCreateFlags) error {
	return p.pool.Create(flags)
}

func (p *libvirtPool) Destroy() error {
	return p.pool.Destroy()
}

func (p *libvirtPool) Delete(flags libvirt.StoragePoolDeleteFlags) error {
	return p.pool.Delete(flags)
}

func (p *libvirtPool) Undefine() error {
	return p.pool.Undefine()
}

func (p *libvirtPool) Refresh(flags uint32) error {
	return p.pool.Refresh(flags)
}

func (p *libvirtPool) SetAutostart(autostart bool) error {
	return p.pool.SetAutostart(autostart)
}

func (p *libvirtPool) underlying() *libvirt.StoragePool {
	return p.pool
}
