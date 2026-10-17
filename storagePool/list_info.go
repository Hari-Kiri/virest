package storagePool

import (
	"sync"

	"github.com/Hari-Kiri/virest/utilities"
	"libvirt.org/go/libvirt"
)

// Info returns volatile information about a storage pool (usage, state, flags).
// ref is a pool name or UUID (virsh-style).
func (c *Connection) Info(ref string) (info Info, err error) {
	pool, err := c.lookupPool(ref)
	if err != nil {
		return Info{}, err
	}
	defer finishFree(pool, &err)

	uuid := ref
	if !utilities.LooksLikeUUID(ref) {
		uuid, err = pool.GetUUIDString()
		if err != nil {
			return Info{}, wrap("get pool uuid", err)
		}
	}

	name, poolInfo, autostart, persistent, err := loadInfoAttrs(pool)
	if err != nil {
		return Info{}, err
	}

	info = Info{
		Name:       name,
		Uuid:       uuid,
		Autostart:  autostart,
		Persistent: persistent,
	}
	if poolInfo != nil {
		info.State = PoolState(poolInfo.State)
		info.Capacity = poolInfo.Capacity
		info.Allocation = poolInfo.Allocation
		info.Available = poolInfo.Available
	}
	return info, nil
}

// List returns storage pools matching listFlags, with XML size fields using xmlFlags.
// Primary API equivalent to virsh pool-list.
func (c *Connection) List(listFlags utilities.ListFlags, xmlFlags utilities.XMLFlags) ([]PoolSummary, error) {
	pools, err := c.hv.ListAllStoragePools(listFlags.Libvirt())
	if err != nil {
		return nil, wrap("list storage pools", err)
	}

	n := len(pools)
	result := make([]PoolSummary, n)
	for i := 0; i < n; i++ {
		summary, sumErr := summarizePool(pools[i], xmlFlags.Libvirt())
		freeErr := pools[i].Free()
		if sumErr != nil {
			_ = freePoolsFrom(pools, i+1)
			return nil, sumErr
		}
		if freeErr != nil {
			_ = freePoolsFrom(pools, i+1)
			return nil, wrap("free pool", freeErr)
		}
		result[i] = summary
	}
	return result, nil
}

// ListActive lists active pools (virsh pool-list default).
func (c *Connection) ListActive(xmlFlags utilities.XMLFlags) ([]PoolSummary, error) {
	return c.List(utilities.Flags.List.Active, xmlFlags)
}

// ListInactive lists inactive pools (virsh pool-list --inactive).
func (c *Connection) ListInactive(xmlFlags utilities.XMLFlags) ([]PoolSummary, error) {
	return c.List(utilities.Flags.List.Inactive, xmlFlags)
}

// ListAll lists active and inactive pools (virsh pool-list --all).
func (c *Connection) ListAll(xmlFlags utilities.XMLFlags) ([]PoolSummary, error) {
	return c.List(utilities.Flags.List.Active|utilities.Flags.List.Inactive, xmlFlags)
}

func summarizePool(pool poolHandle, xmlFlags libvirt.StorageXMLFlags) (PoolSummary, error) {
	model, poolInfo, autostart, persistent, err := loadSummaryAttrs(pool, xmlFlags)
	if err != nil {
		return PoolSummary{}, err
	}

	summary := PoolSummary{
		Uuid:       model.UUID,
		Name:       model.Name,
		Autostart:  autostart,
		Persistent: persistent,
	}
	if poolInfo != nil {
		summary.State = PoolState(poolInfo.State)
	}
	if model.Capacity != nil {
		summary.Capacity = *model.Capacity
	}
	if model.Allocation != nil {
		summary.Allocation = *model.Allocation
	}
	if model.Available != nil {
		summary.Available = *model.Available
	}
	return summary, nil
}

func loadInfoAttrs(pool poolHandle) (name string, info *libvirt.StoragePoolInfo, autostart, persistent bool, err error) {
	err = runParallel(
		func() error {
			return withPoolRef(pool, "ref pool for name", func() error {
				var getErr error
				name, getErr = pool.GetName()
				return wrap("get pool name", getErr)
			})
		},
		func() error {
			return withPoolRef(pool, "ref pool for info", func() error {
				var getErr error
				info, getErr = pool.GetInfo()
				return wrap("get pool info", getErr)
			})
		},
		func() error {
			return withPoolRef(pool, "ref pool for autostart", func() error {
				var getErr error
				autostart, getErr = pool.GetAutostart()
				return wrap("get pool autostart", getErr)
			})
		},
		func() error {
			return withPoolRef(pool, "ref pool for persistent", func() error {
				var getErr error
				persistent, getErr = pool.IsPersistent()
				return wrap("get pool persistent", getErr)
			})
		},
	)
	return name, info, autostart, persistent, err
}

func loadSummaryAttrs(
	pool poolHandle,
	xmlFlags libvirt.StorageXMLFlags,
) (model utilities.StoragePool, info *libvirt.StoragePoolInfo, autostart, persistent bool, err error) {
	// Sequential on List hot path: avoid N×4 goroutines under HA load.
	err = withPoolRef(pool, "ref pool for detail", func() error {
		var xmlErr error
		model, xmlErr = poolXML(pool, xmlFlags)
		return xmlErr
	})
	if err != nil {
		return model, info, autostart, persistent, err
	}
	err = withPoolRef(pool, "ref pool for info", func() error {
		var getErr error
		info, getErr = pool.GetInfo()
		return wrap("get pool info", getErr)
	})
	if err != nil {
		return model, info, autostart, persistent, err
	}
	err = withPoolRef(pool, "ref pool for autostart", func() error {
		var getErr error
		autostart, getErr = pool.GetAutostart()
		return wrap("get pool autostart", getErr)
	})
	if err != nil {
		return model, info, autostart, persistent, err
	}
	err = withPoolRef(pool, "ref pool for persistent", func() error {
		var getErr error
		persistent, getErr = pool.IsPersistent()
		return wrap("get pool persistent", getErr)
	})
	return model, info, autostart, persistent, err
}

// runParallel runs tasks concurrently and returns the first error.
func runParallel(tasks ...func() error) error {
	n := len(tasks)
	if n == 0 {
		return nil
	}
	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		firstErr error
	)
	wg.Add(n)
	for i := 0; i < n; i++ {
		task := tasks[i]
		go func() {
			defer wg.Done()
			taskErr := task()
			if taskErr == nil {
				return
			}
			mu.Lock()
			if firstErr == nil {
				firstErr = taskErr
			}
			mu.Unlock()
		}()
	}
	wg.Wait()
	return firstErr
}

// withPoolRef Refs pool, runs fn, then Frees. Prefer fn's error over free errors.
func withPoolRef(pool poolHandle, refOp string, fn func() error) error {
	refErr := pool.Ref()
	if refErr != nil {
		return wrap(refOp, refErr)
	}
	fnErr := fn()
	freeErr := pool.Free()
	if fnErr != nil {
		return fnErr
	}
	return wrap("free pool", freeErr)
}
