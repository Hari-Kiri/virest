package storagePool

import (
	"github.com/Hari-Kiri/virest/utilities"
)

// Name returns the pool name for ref (name or UUID).
// Equivalent to virsh pool-name.
func (c *Connection) Name(ref string) (name string, err error) {
	pool, err := c.lookupPool(ref)
	if err != nil {
		return "", err
	}
	defer finishFree(pool, &err)

	name, err = pool.GetName()
	if err != nil {
		return "", wrap("get pool name", err)
	}
	return name, nil
}

// UUIDByName returns the pool UUID for the given name.
// Equivalent to virsh pool-uuid.
func (c *Connection) UUIDByName(name string) (uuid string, err error) {
	pool, err := c.lookupPool(name)
	if err != nil {
		return "", err
	}
	defer finishFree(pool, &err)

	uuid, err = pool.GetUUIDString()
	if err != nil {
		return "", wrap("get pool uuid", err)
	}
	return uuid, nil
}

// FindSourcesAs discovers sources using virsh find-storage-pool-sources-as style params.
func (c *Connection) FindSourcesAs(poolType utilities.PoolType, params utilities.FindSourcesAsParams) (Sources, error) {
	src := SourceSpec{
		Host: Host{
			Name: params.Host,
			Port: params.Port,
		},
		Initiator: Initiator{
			Iqn: Iqn{Name: params.Initiator},
		},
	}
	return c.FindSources(poolType, src)
}
