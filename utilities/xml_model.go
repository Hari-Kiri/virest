package utilities

import (
	"encoding/xml"
	"fmt"

	"libvirt.org/go/libvirtxml"
)

// Brand-neutral storage-pool models.
// Callers import utilities — not libvirt.org/go/libvirtxml.
// Nested source/target types remain libvirtxml aliases; StoragePool.Type is PoolType.

type (
	// StoragePoolTarget is the pool target element (path, permissions, …).
	StoragePoolTarget = libvirtxml.StoragePoolTarget
	// StoragePoolSize is a capacity/allocation/available size with unit.
	StoragePoolSize = libvirtxml.StoragePoolSize
	// StoragePoolSource is the pool source element.
	StoragePoolSource = libvirtxml.StoragePoolSource
	// StoragePoolSourceHost is a source host (name/port).
	StoragePoolSourceHost = libvirtxml.StoragePoolSourceHost
	// StoragePoolSourceDevice is a source device path.
	StoragePoolSourceDevice = libvirtxml.StoragePoolSourceDevice
	// StoragePoolSourceDir is a source directory path.
	StoragePoolSourceDir = libvirtxml.StoragePoolSourceDir
	// StoragePoolSourceFormat is a source format type attribute.
	StoragePoolSourceFormat = libvirtxml.StoragePoolSourceFormat
	// StoragePoolSourceInitiator is an iSCSI initiator element.
	StoragePoolSourceInitiator = libvirtxml.StoragePoolSourceInitiator
	// StoragePoolSourceInitiatorIQN is an initiator IQN.
	StoragePoolSourceInitiatorIQN = libvirtxml.StoragePoolSourceInitiatorIQN
	// StoragePoolSourceAuth is source authentication (chap/ceph).
	StoragePoolSourceAuth = libvirtxml.StoragePoolSourceAuth
	// StoragePoolSourceAuthSecret is the secret reference for auth.
	StoragePoolSourceAuthSecret = libvirtxml.StoragePoolSourceAuthSecret
	// StoragePoolSourceAdapter is a SCSI/FC adapter element.
	StoragePoolSourceAdapter = libvirtxml.StoragePoolSourceAdapter
	// StoragePoolSourceProtocol is a source protocol (e.g. NFS version).
	StoragePoolSourceProtocol = libvirtxml.StoragePoolSourceProtocol
	// StoragePoolFeatures is optional pool features.
	StoragePoolFeatures = libvirtxml.StoragePoolFeatures
	// StoragePoolRefresh is optional refresh configuration.
	StoragePoolRefresh = libvirtxml.StoragePoolRefresh
	// StoragePoolFSCommandline is FS backend mount options.
	StoragePoolFSCommandline = libvirtxml.StoragePoolFSCommandline
	// StoragePoolRBDCommandline is RBD backend config options.
	StoragePoolRBDCommandline = libvirtxml.StoragePoolRBDCommandline
)

// StoragePool is the typed storage-pool configuration model.
// Type is PoolType (supported protocol constants), not a free string.
type StoragePool struct {
	XMLName        xml.Name                 `xml:"pool" json:"-"`
	Type           PoolType                 `xml:"type,attr" json:"type"`
	Name           string                   `xml:"name,omitempty" json:"name,omitempty"`
	UUID           string                   `xml:"uuid,omitempty" json:"uuid,omitempty"`
	Allocation     *StoragePoolSize         `xml:"allocation" json:"allocation,omitempty"`
	Capacity       *StoragePoolSize         `xml:"capacity" json:"capacity,omitempty"`
	Available      *StoragePoolSize         `xml:"available" json:"available,omitempty"`
	Features       *StoragePoolFeatures     `xml:"features" json:"features,omitempty"`
	Target         *StoragePoolTarget       `xml:"target" json:"target,omitempty"`
	Source         *StoragePoolSource       `xml:"source" json:"source,omitempty"`
	Refresh        *StoragePoolRefresh      `xml:"refresh" json:"refresh,omitempty"`
	FSCommandline  *StoragePoolFSCommandline
	RBDCommandline *StoragePoolRBDCommandline
}

// Marshal encodes the pool as libvirt storage-pool XML.
// Type must be empty or a Known PoolType.
func (s StoragePool) Marshal() (string, error) {
	if s.Type != "" && !s.Type.Known() {
		return "", fmt.Errorf("%w: %s", errPoolAsUnsupported, s.Type)
	}
	raw := s.toLibvirt()
	return raw.Marshal()
}

// Unmarshal decodes libvirt storage-pool XML into s.
func (s *StoragePool) Unmarshal(doc string) error {
	var raw libvirtxml.StoragePool
	err := raw.Unmarshal(doc)
	if err != nil {
		return err
	}
	*s = fromLibvirt(raw)
	return nil
}

func (s StoragePool) toLibvirt() libvirtxml.StoragePool {
	return libvirtxml.StoragePool{
		Type:           s.Type.String(),
		Name:           s.Name,
		UUID:           s.UUID,
		Allocation:     s.Allocation,
		Capacity:       s.Capacity,
		Available:      s.Available,
		Features:       s.Features,
		Target:         s.Target,
		Source:         s.Source,
		Refresh:        s.Refresh,
		FSCommandline:  s.FSCommandline,
		RBDCommandline: s.RBDCommandline,
	}
}

func fromLibvirt(raw libvirtxml.StoragePool) StoragePool {
	return StoragePool{
		Type:           PoolType(raw.Type),
		Name:           raw.Name,
		UUID:           raw.UUID,
		Allocation:     raw.Allocation,
		Capacity:       raw.Capacity,
		Available:      raw.Available,
		Features:       raw.Features,
		Target:         raw.Target,
		Source:         raw.Source,
		Refresh:        raw.Refresh,
		FSCommandline:  raw.FSCommandline,
		RBDCommandline: raw.RBDCommandline,
	}
}
