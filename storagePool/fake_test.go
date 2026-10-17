package storagePool

import (
	"sync"

	"github.com/Hari-Kiri/virest/utilities"
	"libvirt.org/go/libvirt"
)

func newConnectionForTest(hv hypervisor) *Connection {
	return &Connection{hv: hv}
}

type fakeHypervisor struct {
	mu sync.Mutex

	closeErr      error
	closeCalls    int
	listPools     []poolHandle
	listErr       error
	listFlagsSeen libvirt.ConnectListAllStoragePoolsFlags

	poolsByUUID map[string]poolHandle
	poolsByName map[string]poolHandle
	lookupErr   error
	lookupByNameErr error

	definePool   poolHandle
	defineErr    error
	defineXML    string
	defineFlags  libvirt.StoragePoolDefineFlags

	createTransientPool   poolHandle
	createTransientErr    error
	createTransientConfig string
	createTransientFlags  libvirt.StoragePoolCreateFlags

	capabilitiesXML string
	capabilitiesErr error

	findSourcesXML string
	findSourcesErr error
	findPoolType   string
	findSrcSpec    string

	lifecycleRegisterErr error
	refreshRegisterErr   error
	deregisterErr        error
	deregisterCalls      []int

	lifecycleCallback libvirt.StoragePoolEventLifecycleCallback
	refreshCallback   libvirt.StoragePoolEventGenericCallback
	nextCallbackID    int
}

func newFakeHypervisor() *fakeHypervisor {
	return &fakeHypervisor{
		poolsByUUID:    map[string]poolHandle{},
		poolsByName:    map[string]poolHandle{},
		nextCallbackID: 1,
	}
}

// registerPool stores pool under UUID or name lookup based on ref shape.
func (f *fakeHypervisor) registerPool(ref string, p poolHandle) {
	if utilities.LooksLikeUUID(ref) {
		f.poolsByUUID[ref] = p
		return
	}
	f.poolsByName[ref] = p
}

func (f *fakeHypervisor) Close() (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closeCalls++
	return 0, f.closeErr
}

func (f *fakeHypervisor) ListAllStoragePools(flags libvirt.ConnectListAllStoragePoolsFlags) ([]poolHandle, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.listFlagsSeen = flags
	if f.listErr != nil {
		return nil, f.listErr
	}
	out := make([]poolHandle, len(f.listPools))
	copy(out, f.listPools)
	return out, nil
}

func (f *fakeHypervisor) LookupStoragePoolByUUIDString(uuid string) (poolHandle, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.lookupErr != nil {
		return nil, f.lookupErr
	}
	pool, ok := f.poolsByUUID[uuid]
	if !ok {
		return nil, libvirt.Error{Code: libvirt.ERR_NO_STORAGE_POOL, Message: "not found"}
	}
	return pool, nil
}

func (f *fakeHypervisor) LookupStoragePoolByName(name string) (poolHandle, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.lookupByNameErr != nil {
		return nil, f.lookupByNameErr
	}
	if f.lookupErr != nil {
		return nil, f.lookupErr
	}
	pool, ok := f.poolsByName[name]
	if !ok {
		return nil, libvirt.Error{Code: libvirt.ERR_NO_STORAGE_POOL, Message: "not found"}
	}
	return pool, nil
}

func (f *fakeHypervisor) StoragePoolDefineXML(xmlConfig string, flags libvirt.StoragePoolDefineFlags) (poolHandle, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.defineXML = xmlConfig
	f.defineFlags = flags
	if f.defineErr != nil {
		return nil, f.defineErr
	}
	return f.definePool, nil
}

func (f *fakeHypervisor) CreateTransient(config string, flags libvirt.StoragePoolCreateFlags) (poolHandle, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.createTransientConfig = config
	f.createTransientFlags = flags
	if f.createTransientErr != nil {
		return nil, f.createTransientErr
	}
	return f.createTransientPool, nil
}

func (f *fakeHypervisor) GetStoragePoolCapabilities(flags uint32) (string, error) {
	if f.capabilitiesErr != nil {
		return "", f.capabilitiesErr
	}
	return f.capabilitiesXML, nil
}

func (f *fakeHypervisor) FindStoragePoolSources(poolType, srcSpec string, flags uint32) (string, error) {
	f.findPoolType = poolType
	f.findSrcSpec = srcSpec
	if f.findSourcesErr != nil {
		return "", f.findSourcesErr
	}
	return f.findSourcesXML, nil
}

func (f *fakeHypervisor) StoragePoolEventLifecycleRegister(pool poolHandle, callback libvirt.StoragePoolEventLifecycleCallback) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.lifecycleRegisterErr != nil {
		return 0, f.lifecycleRegisterErr
	}
	f.lifecycleCallback = callback
	id := f.nextCallbackID
	f.nextCallbackID++
	return id, nil
}

func (f *fakeHypervisor) StoragePoolEventRefreshRegister(pool poolHandle, callback libvirt.StoragePoolEventGenericCallback) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.refreshRegisterErr != nil {
		return 0, f.refreshRegisterErr
	}
	f.refreshCallback = callback
	id := f.nextCallbackID
	f.nextCallbackID++
	return id, nil
}

func (f *fakeHypervisor) StoragePoolEventDeregister(callbackID int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.deregisterCalls = append(f.deregisterCalls, callbackID)
	return f.deregisterErr
}

func (f *fakeHypervisor) fireLifecycle(event libvirt.StoragePoolEventLifecycle) {
	f.mu.Lock()
	cb := f.lifecycleCallback
	f.mu.Unlock()
	if cb != nil {
		cb(nil, nil, &event)
	}
}

func (f *fakeHypervisor) fireRefresh() {
	f.mu.Lock()
	cb := f.refreshCallback
	f.mu.Unlock()
	if cb != nil {
		cb(nil, nil)
	}
}

type fakePool struct {
	mu sync.Mutex

	uuid       string
	name       string
	info       *libvirt.StoragePoolInfo
	autostart  bool
	persistent bool
	xmlDesc    string

	freeCalls int
	refCalls  int

	getUUIDErr       error
	getNameErr       error
	getInfoErr       error
	getAutostartErr  error
	isPersistentErr  error
	getXMLErr        error
	buildErr         error
	createErr        error
	destroyErr       error
	deleteErr        error
	undefineErr      error
	refreshErr       error
	setAutostartErr  error
	refErr           error
	freeErr          error

	buildFlags         libvirt.StoragePoolBuildFlags
	createFlags        libvirt.StoragePoolCreateFlags
	deleteFlags        libvirt.StoragePoolDeleteFlags
	refreshFlags       uint32
	setAutostartValue  *bool
	getXMLFlags        libvirt.StorageXMLFlags
}

func (p *fakePool) Free() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.freeCalls++
	return p.freeErr
}

func (p *fakePool) Ref() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.refCalls++
	return p.refErr
}

func (p *fakePool) GetUUIDString() (string, error) {
	if p.getUUIDErr != nil {
		return "", p.getUUIDErr
	}
	return p.uuid, nil
}

func (p *fakePool) GetName() (string, error) {
	if p.getNameErr != nil {
		return "", p.getNameErr
	}
	return p.name, nil
}

func (p *fakePool) GetInfo() (*libvirt.StoragePoolInfo, error) {
	if p.getInfoErr != nil {
		return nil, p.getInfoErr
	}
	return p.info, nil
}

func (p *fakePool) GetAutostart() (bool, error) {
	if p.getAutostartErr != nil {
		return false, p.getAutostartErr
	}
	return p.autostart, nil
}

func (p *fakePool) IsPersistent() (bool, error) {
	if p.isPersistentErr != nil {
		return false, p.isPersistentErr
	}
	return p.persistent, nil
}

func (p *fakePool) GetXMLDesc(flags libvirt.StorageXMLFlags) (string, error) {
	p.getXMLFlags = flags
	if p.getXMLErr != nil {
		return "", p.getXMLErr
	}
	return p.xmlDesc, nil
}

func (p *fakePool) Build(flags libvirt.StoragePoolBuildFlags) error {
	p.buildFlags = flags
	return p.buildErr
}

func (p *fakePool) Create(flags libvirt.StoragePoolCreateFlags) error {
	p.createFlags = flags
	return p.createErr
}

func (p *fakePool) Destroy() error { return p.destroyErr }

func (p *fakePool) Delete(flags libvirt.StoragePoolDeleteFlags) error {
	p.deleteFlags = flags
	return p.deleteErr
}

func (p *fakePool) Undefine() error { return p.undefineErr }

func (p *fakePool) Refresh(flags uint32) error {
	p.refreshFlags = flags
	return p.refreshErr
}

func (p *fakePool) SetAutostart(autostart bool) error {
	p.setAutostartValue = &autostart
	return p.setAutostartErr
}

func (p *fakePool) underlying() *libvirt.StoragePool { return nil }
