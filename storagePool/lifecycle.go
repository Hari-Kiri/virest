package storagePool

import (
	"github.com/Hari-Kiri/virest/utilities"
)

func marshalPool(storagePool utilities.StoragePool) (string, error) {
	xmlConfig, err := storagePool.Marshal()
	if err != nil {
		return "", wrap("marshal pool config", err)
	}
	return xmlConfig, nil
}

// Define creates a new inactive storage pool from the given model.
// On success it returns the UUID of the defined pool.
// Equivalent to virsh pool-define.
func (c *Connection) Define(storagePool utilities.StoragePool, flags utilities.DefineFlags) (uuid string, err error) {
	xmlConfig, err := marshalPool(storagePool)
	if err != nil {
		return "", err
	}

	defined, err := c.hv.StoragePoolDefineXML(xmlConfig, flags.Libvirt())
	if err != nil {
		return "", wrap("define storage pool", err)
	}
	defer finishFree(defined, &err)

	uuid, err = defined.GetUUIDString()
	if err != nil {
		return "", wrap("get defined pool uuid", err)
	}
	return uuid, nil
}

// DefineAs defines an inactive pool from virsh-style -as parameters.
// On success it returns the UUID of the defined pool.
// Equivalent to virsh pool-define-as.
func (c *Connection) DefineAs(params utilities.DefineAsParams, flags utilities.DefineFlags) (string, error) {
	model, err := params.BuildModel()
	if err != nil {
		return "", wrap("build pool model", err)
	}
	return c.Define(model, flags)
}

// CreateTransient creates and starts a transient (non-persistent) pool from the model.
// On success it returns the UUID of the created pool.
// Equivalent to virsh pool-create.
func (c *Connection) CreateTransient(storagePool utilities.StoragePool, flags utilities.CreateFlags) (uuid string, err error) {
	config, err := marshalPool(storagePool)
	if err != nil {
		return "", err
	}

	created, err := c.hv.CreateTransient(config, flags.Libvirt())
	if err != nil {
		return "", wrap("create transient storage pool", err)
	}
	defer finishFree(created, &err)

	uuid, err = created.GetUUIDString()
	if err != nil {
		return "", wrap("get transient pool uuid", err)
	}
	return uuid, nil
}

// CreateAs creates and starts a transient pool from virsh-style -as parameters.
// On success it returns the UUID of the created pool.
// Equivalent to virsh pool-create-as.
func (c *Connection) CreateAs(params utilities.CreateAsParams, flags utilities.CreateFlags) (string, error) {
	model, err := params.BuildModel()
	if err != nil {
		return "", wrap("build pool model", err)
	}
	return c.CreateTransient(model, flags)
}

// Build builds the underlying storage for a defined pool.
// ref is a pool name or UUID (virsh-style).
func (c *Connection) Build(ref string, flags utilities.BuildFlags) (err error) {
	pool, err := c.lookupPool(ref)
	if err != nil {
		return err
	}
	defer finishFree(pool, &err)
	return wrap("build storage pool", pool.Build(flags.Libvirt()))
}

// Start starts an inactive storage pool.
// Equivalent to virsh pool-start.
// ref is a pool name or UUID.
func (c *Connection) Start(ref string, flags utilities.CreateFlags) (err error) {
	pool, err := c.lookupPool(ref)
	if err != nil {
		return err
	}
	defer finishFree(pool, &err)
	return wrap("start storage pool", pool.Create(flags.Libvirt()))
}

// Destroy stops an active storage pool.
// ref is a pool name or UUID.
func (c *Connection) Destroy(ref string) (err error) {
	pool, err := c.lookupPool(ref)
	if err != nil {
		return err
	}
	defer finishFree(pool, &err)
	return wrap("destroy storage pool", pool.Destroy())
}

// Delete deletes the underlying storage resources of a pool.
// ref is a pool name or UUID.
func (c *Connection) Delete(ref string, flags utilities.DeleteFlags) (err error) {
	pool, err := c.lookupPool(ref)
	if err != nil {
		return err
	}
	defer finishFree(pool, &err)
	return wrap("delete storage pool", pool.Delete(flags.Libvirt()))
}

// Undefine removes a storage pool definition.
// ref is a pool name or UUID.
func (c *Connection) Undefine(ref string) (err error) {
	pool, err := c.lookupPool(ref)
	if err != nil {
		return err
	}
	defer finishFree(pool, &err)
	return wrap("undefine storage pool", pool.Undefine())
}

// Refresh refreshes the list of volumes in the storage pool.
// ref is a pool name or UUID.
func (c *Connection) Refresh(ref string) (err error) {
	pool, err := c.lookupPool(ref)
	if err != nil {
		return err
	}
	defer finishFree(pool, &err)
	return wrap("refresh storage pool", pool.Refresh(0))
}

// Autostart configures whether the pool starts with the host.
// ref is a pool name or UUID.
func (c *Connection) Autostart(ref string, autostart bool) (err error) {
	pool, err := c.lookupPool(ref)
	if err != nil {
		return err
	}
	defer finishFree(pool, &err)
	return wrap("set storage pool autostart", pool.SetAutostart(autostart))
}
