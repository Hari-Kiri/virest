package storagePool

import (
	"encoding/xml"

	"github.com/Hari-Kiri/virest/utilities"
)

// PoolState is the storage pool state, abstracted so callers never
// import libvirt.org/go/libvirt.
type PoolState uint

const (
	// PoolStateInactive is an inactive pool.
	PoolStateInactive PoolState = iota
	// PoolStateActive is an active pool.
	PoolStateActive
	// PoolStatePersistent is a persistent pool.
	PoolStatePersistent
	// PoolStateRunning is a running pool.
	PoolStateRunning
	// PoolStateDeleted is a deleted pool.
	PoolStateDeleted
	// PoolStateInactiveBuild is an inactive pool being built.
	PoolStateInactiveBuild
)

// String returns a human-readable state name.
func (s PoolState) String() string {
	switch s {
	case PoolStateInactive:
		return "inactive"
	case PoolStateActive:
		return "active"
	case PoolStatePersistent:
		return "persistent"
	case PoolStateRunning:
		return "running"
	case PoolStateDeleted:
		return "deleted"
	case PoolStateInactiveBuild:
		return "inactive-build"
	default:
		return "unknown"
	}
}

// Info is volatile information about a storage pool.
type Info struct {
	Name       string   `json:"name"`
	Uuid       string   `json:"uuid"`
	State      PoolState `json:"state"`
	Persistent bool      `json:"persistent"`
	Autostart  bool      `json:"autostart"`
	Capacity   uint64    `json:"capacity"`
	Allocation uint64    `json:"allocation"`
	Available  uint64    `json:"available"`
}

// Detail is the full storage pool description (typed model, not raw XML).
type Detail struct {
	utilities.StoragePool
}

// PoolSummary is a condensed view of a storage pool used by List.
type PoolSummary struct {
	Uuid       string                      `json:"uuid"`
	Name       string                      `json:"name"`
	State      PoolState                   `json:"state"`
	Autostart  bool                        `json:"autostart"`
	Persistent bool                        `json:"persistent"`
	Capacity   utilities.StoragePoolSize   `json:"capacity"`
	Allocation utilities.StoragePoolSize   `json:"allocation"`
	Available  utilities.StoragePoolSize   `json:"available"`
}

// EventKind selects which storage-pool event to wait for or stream.
type EventKind uint

const (
	// EventLifecycle waits for storage pool lifecycle events.
	EventLifecycle EventKind = 0
	// EventRefresh waits for storage pool refresh events.
	EventRefresh EventKind = 1
)

// PoolEventLifecycle is a storage pool event lifecycle value,
// abstracted so callers never import libvirt.org/go/libvirt.
type PoolEventLifecycle uint

const (
	// PoolEventLifecycleUndefined is an undefined event.
	PoolEventLifecycleUndefined PoolEventLifecycle = iota
	// PoolEventLifecycleStarted is a pool started event.
	PoolEventLifecycleStarted
	// PoolEventLifecycleStopped is a pool stopped event.
	PoolEventLifecycleStopped
	// PoolEventLifecycleAborted is a pool aborted event.
	PoolEventLifecycleAborted
)

// Event is a storage pool lifecycle or refresh notification.
type Event struct {
	EventLifecycle PoolEventLifecycle `json:"eventLifecycle"`
	EventRefresh   int                `json:"eventRefresh"`
	Timestamp      int64              `json:"timestamp"`
	TimestampNano  int64              `json:"timestampNano"`
	EventId        int                `json:"eventId"`
}

// SourceSpec is an optional storage-pool source element used by FindSources.
type SourceSpec struct {
	XMLName   xml.Name  `xml:"source" json:"-"`
	Host      Host      `xml:"host" json:"host"`
	Initiator Initiator `xml:"initiator" json:"initiator"`
}

// Host is a source host attribute pair.
type Host struct {
	Name string `xml:"name,attr" json:"name"`
	Port int    `xml:"port,attr" json:"port"`
}

// Initiator holds an iSCSI IQN.
type Initiator struct {
	Iqn Iqn `xml:"iqn" json:"iqn"`
}

// Iqn is an iSCSI qualified name.
type Iqn struct {
	Name string `xml:"name,attr" json:"name"`
}

// Sources is the discovered set of storage pool sources.
type Sources struct {
	XMLName xml.Name                      `xml:"sources" json:"-"`
	Source  []utilities.StoragePoolSource `xml:"source" json:"source"`
}

// Capabilities describes supported storage pool types and formats.
type Capabilities struct {
	XMLName xml.Name         `xml:"storagepoolCapabilities" json:"-"`
	Pool    []CapabilityPool `xml:"pool" json:"pool"`
}

// CapabilityPool is one pool type entry in Capabilities.
type CapabilityPool struct {
	Type        string      `xml:"type,attr" json:"type"`
	Supported   string      `xml:"supported,attr" json:"supported"`
	VolOptions  VolOptions  `xml:"volOptions" json:"volOptions"`
	PoolOptions PoolOptions `xml:"poolOptions" json:"poolOptions"`
}

// VolOptions describes volume format options for a pool type.
type VolOptions struct {
	DefaultFormat DefaultFormat `xml:"defaultFormat" json:"defaultFormat"`
	Enum          Enum          `xml:"enum" json:"enum"`
}

// PoolOptions describes pool format options for a pool type.
type PoolOptions struct {
	DefaultFormat DefaultFormat `xml:"defaultFormat" json:"defaultFormat"`
	Enum          Enum          `xml:"enum" json:"enum"`
}

// DefaultFormat is a default format type attribute.
type DefaultFormat struct {
	Type string `xml:"type,attr" json:"type"`
}

// Enum lists allowed format values.
type Enum struct {
	Name  string   `xml:"name,attr" json:"name"`
	Value []string `xml:"value" json:"value"`
}
