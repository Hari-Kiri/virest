package storagePool

import (
	"encoding/xml"

	"github.com/Hari-Kiri/virest/utilities"
	"libvirt.org/go/libvirt"
)

// Dump returns the full pool description as a typed model.
// ref is a pool name or UUID. Equivalent to virsh pool-dumpxml.
func (c *Connection) Dump(ref string, xmlFlags utilities.XMLFlags) (detail Detail, err error) {
	pool, err := c.lookupPool(ref)
	if err != nil {
		return Detail{}, err
	}
	defer finishFree(pool, &err)

	model, err := poolXML(pool, xmlFlags.Libvirt())
	if err != nil {
		return Detail{}, err
	}
	return Detail{StoragePool: model}, nil
}

func poolXML(pool poolHandle, flags libvirt.StorageXMLFlags) (utilities.StoragePool, error) {
	doc, err := pool.GetXMLDesc(flags)
	if err != nil {
		return utilities.StoragePool{}, wrap("get pool xml", err)
	}
	var model utilities.StoragePool
	err = model.Unmarshal(doc)
	if err != nil {
		return utilities.StoragePool{}, wrap("unmarshal pool xml", err)
	}
	return model, nil
}

// Capabilities returns supported storage pool types and formats.
func (c *Connection) Capabilities() (Capabilities, error) {
	doc, err := c.hv.GetStoragePoolCapabilities(0)
	if err != nil {
		return Capabilities{}, wrap("get storage pool capabilities", err)
	}
	var caps Capabilities
	err = xml.Unmarshal([]byte(doc), &caps)
	if err != nil {
		return Capabilities{}, wrap("unmarshal storage pool capabilities", err)
	}
	return caps, nil
}

// FindSources discovers available storage pool sources for the given pool type.
func (c *Connection) FindSources(poolType utilities.PoolType, src SourceSpec) (Sources, error) {
	srcXML, err := xml.MarshalIndent(src, "", "  ")
	if err != nil {
		return Sources{}, wrap("marshal source spec", err)
	}
	doc, err := c.hv.FindStoragePoolSources(poolType.String(), string(srcXML), 0)
	if err != nil {
		return Sources{}, wrap("find storage pool sources", err)
	}
	var sources Sources
	err = xml.Unmarshal([]byte(doc), &sources)
	if err != nil {
		return Sources{}, wrap("unmarshal storage pool sources", err)
	}
	return sources, nil
}
