package utilities

import (
	"errors"
	"fmt"
	"strconv"
)

var (
	errPoolAsNameRequired       = errors.New("pool name is required")
	errPoolAsTypeRequired       = errors.New("pool type is required")
	errPoolAsTargetRequired     = errors.New("target path is required")
	errPoolAsHostRequired       = errors.New("source host is required")
	errPoolAsSourceNameRequired = errors.New("source name is required")
	errPoolAsSourceDevRequired  = errors.New("source device is required")
	errPoolAsInitiatorRequired  = errors.New("source initiator is required")
	errPoolAsAdapterRequired    = errors.New("adapter name or wwnn/wwpn is required")
	errPoolAsAuthIncomplete     = errors.New("auth type and username are both required when either is set")
	errPoolAsSecretConflict     = errors.New("provide only one of secret usage or secret uuid")
	errPoolAsUnsupported        = errors.New("unsupported pool type")
)

// poolAsParams is the unexported shared field set for virsh-style pool -as
// arguments. Callers use DefineAsParams or CreateAsParams only.
type poolAsParams struct {
	Name         string
	Type         PoolType
	TargetPath   string
	SourceHost   string
	SourcePort   int // optional; mapped to source host port attribute when > 0
	SourcePath   string
	SourceDev    string
	SourceName   string
	SourceFormat string

	// Auth (iscsi chap / rbd ceph). SecretUsage and SecretUUID are mutually exclusive.
	AuthType     string
	AuthUsername string
	SecretUsage  string
	SecretUUID   string

	// SourceInitiator is the initiator IQN (iscsi-direct; also allowed on iscsi).
	SourceInitiator string
	// SourceProtocolVer is NFS protocol version (netfs nfsvers=n).
	SourceProtocolVer string

	// SCSI / FC adapter (scsi pool).
	AdapterName            string
	AdapterWWNN            string
	AdapterWWPN            string
	AdapterParent          string
	AdapterParentWWNN      string
	AdapterParentWWPN      string
	AdapterParentFabricWWN string
}

// DefineAsParams is the caller-facing input for Connection.DefineAs
// (virsh pool-define-as).
type DefineAsParams poolAsParams

// CreateAsParams is the caller-facing input for Connection.CreateAs
// (virsh pool-create-as).
type CreateAsParams poolAsParams

// BuildModel builds a StoragePool from DefineAsParams.
func (p DefineAsParams) BuildModel() (StoragePool, error) {
	return buildPoolModel(poolAsParams(p))
}

// BuildModel builds a StoragePool from CreateAsParams.
func (p CreateAsParams) BuildModel() (StoragePool, error) {
	return buildPoolModel(poolAsParams(p))
}

// FindSourcesAsParams mirrors virsh find-storage-pool-sources-as typed args.
type FindSourcesAsParams struct {
	Host      string
	Port      int
	Initiator string
}

// buildPoolModel builds a StoragePool from virsh-style -as parameters.
// Supported types: dir, fs, netfs, logical, disk, iscsi, iscsi-direct, scsi,
// mpath, rbd, sheepdog, gluster, zfs, vstorage.
func buildPoolModel(p poolAsParams) (StoragePool, error) {
	if p.Name == "" {
		return StoragePool{}, errPoolAsNameRequired
	}
	if p.Type == "" {
		return StoragePool{}, errPoolAsTypeRequired
	}
	if err := validateAuth(p); err != nil {
		return StoragePool{}, err
	}

	switch p.Type {
	case PoolTypeDir:
		return buildDirPool(p)
	case PoolTypeFS:
		return buildFsPool(p)
	case PoolTypeNetFS:
		return buildNetfsPool(p)
	case PoolTypeLogical:
		return buildLogicalPool(p)
	case PoolTypeDisk:
		return buildDiskPool(p)
	case PoolTypeISCSI:
		return buildIscsiPool(p)
	case PoolTypeISCSIDirect:
		return buildIscsiDirectPool(p)
	case PoolTypeSCSI:
		return buildScsiPool(p)
	case PoolTypeMpath:
		return buildMpathPool(p)
	case PoolTypeRBD:
		return buildRbdPool(p)
	case PoolTypeSheepdog:
		return buildSheepdogPool(p)
	case PoolTypeGluster:
		return buildGlusterPool(p)
	case PoolTypeZFS:
		return buildZfsPool(p)
	case PoolTypeVStorage:
		return buildVstoragePool(p)
	}
	return StoragePool{}, fmt.Errorf("%w: %s", errPoolAsUnsupported, p.Type)
}

func validateAuth(p poolAsParams) error {
	if p.AuthType == "" && p.AuthUsername == "" && p.SecretUsage == "" && p.SecretUUID == "" {
		return nil
	}
	if p.AuthType == "" || p.AuthUsername == "" {
		return errPoolAsAuthIncomplete
	}
	if p.SecretUsage != "" && p.SecretUUID != "" {
		return errPoolAsSecretConflict
	}
	return nil
}

func buildDirPool(p poolAsParams) (StoragePool, error) {
	if p.TargetPath == "" {
		return StoragePool{}, errPoolAsTargetRequired
	}
	return StoragePool{
		Name:   p.Name,
		Type:   PoolTypeDir,
		Target: &StoragePoolTarget{Path: p.TargetPath},
	}, nil
}

func buildFsPool(p poolAsParams) (StoragePool, error) {
	if p.TargetPath == "" {
		return StoragePool{}, errPoolAsTargetRequired
	}
	return poolWithSource(p, PoolTypeFS, true), nil
}

func buildNetfsPool(p poolAsParams) (StoragePool, error) {
	if p.TargetPath == "" {
		return StoragePool{}, errPoolAsTargetRequired
	}
	if p.SourceHost == "" {
		return StoragePool{}, errPoolAsHostRequired
	}
	return poolWithSource(p, PoolTypeNetFS, true), nil
}

func buildLogicalPool(p poolAsParams) (StoragePool, error) {
	// virsh: target optional; source-name defaults to pool name when omitted.
	srcName := p.SourceName
	if srcName == "" {
		srcName = p.Name
	}
	cp := p
	cp.SourceName = srcName
	pool := poolWithSource(cp, PoolTypeLogical, false)
	if p.TargetPath != "" {
		pool.Target = &StoragePoolTarget{Path: p.TargetPath}
	}
	return pool, nil
}

func buildDiskPool(p poolAsParams) (StoragePool, error) {
	if p.TargetPath == "" {
		return StoragePool{}, errPoolAsTargetRequired
	}
	if p.SourceDev == "" {
		return StoragePool{}, errPoolAsSourceDevRequired
	}
	return poolWithSource(p, PoolTypeDisk, true), nil
}

func buildIscsiPool(p poolAsParams) (StoragePool, error) {
	if p.SourceHost == "" {
		return StoragePool{}, errPoolAsHostRequired
	}
	if p.SourceDev == "" {
		return StoragePool{}, errPoolAsSourceDevRequired
	}
	if p.TargetPath == "" {
		return StoragePool{}, errPoolAsTargetRequired
	}
	return poolWithSource(p, PoolTypeISCSI, true), nil
}

func buildIscsiDirectPool(p poolAsParams) (StoragePool, error) {
	if p.SourceHost == "" {
		return StoragePool{}, errPoolAsHostRequired
	}
	if p.SourceDev == "" {
		return StoragePool{}, errPoolAsSourceDevRequired
	}
	if p.SourceInitiator == "" {
		return StoragePool{}, errPoolAsInitiatorRequired
	}
	return poolWithSource(p, PoolTypeISCSIDirect, p.TargetPath != ""), nil
}

func buildScsiPool(p poolAsParams) (StoragePool, error) {
	if p.TargetPath == "" {
		return StoragePool{}, errPoolAsTargetRequired
	}
	if p.AdapterName == "" && (p.AdapterWWNN == "" || p.AdapterWWPN == "") {
		return StoragePool{}, errPoolAsAdapterRequired
	}
	return poolWithSource(p, PoolTypeSCSI, true), nil
}

func buildMpathPool(p poolAsParams) (StoragePool, error) {
	if p.TargetPath == "" {
		return StoragePool{}, errPoolAsTargetRequired
	}
	return poolWithSource(p, PoolTypeMpath, true), nil
}

func buildRbdPool(p poolAsParams) (StoragePool, error) {
	if p.SourceHost == "" {
		return StoragePool{}, errPoolAsHostRequired
	}
	if p.SourceName == "" {
		return StoragePool{}, errPoolAsSourceNameRequired
	}
	return poolWithSource(p, PoolTypeRBD, p.TargetPath != ""), nil
}

func buildSheepdogPool(p poolAsParams) (StoragePool, error) {
	if p.SourceHost == "" {
		return StoragePool{}, errPoolAsHostRequired
	}
	if p.SourceName == "" {
		return StoragePool{}, errPoolAsSourceNameRequired
	}
	return poolWithSource(p, PoolTypeSheepdog, p.TargetPath != ""), nil
}

func buildGlusterPool(p poolAsParams) (StoragePool, error) {
	if p.SourceHost == "" {
		return StoragePool{}, errPoolAsHostRequired
	}
	if p.SourceName == "" {
		return StoragePool{}, errPoolAsSourceNameRequired
	}
	return poolWithSource(p, PoolTypeGluster, p.TargetPath != ""), nil
}

func buildZfsPool(p poolAsParams) (StoragePool, error) {
	if p.SourceName == "" {
		return StoragePool{}, errPoolAsSourceNameRequired
	}
	return poolWithSource(p, PoolTypeZFS, p.TargetPath != ""), nil
}

func buildVstoragePool(p poolAsParams) (StoragePool, error) {
	if p.SourceName == "" {
		return StoragePool{}, errPoolAsSourceNameRequired
	}
	if p.TargetPath == "" {
		return StoragePool{}, errPoolAsTargetRequired
	}
	return poolWithSource(p, PoolTypeVStorage, true), nil
}

func poolWithSource(p poolAsParams, poolType PoolType, withTarget bool) StoragePool {
	pool := StoragePool{
		Name:   p.Name,
		Type:   poolType,
		Source: buildSource(p),
	}
	if withTarget && p.TargetPath != "" {
		pool.Target = &StoragePoolTarget{Path: p.TargetPath}
	}
	return pool
}

func buildSource(p poolAsParams) *StoragePoolSource {
	src := &StoragePoolSource{}
	filled := false

	if p.SourceName != "" {
		src.Name = p.SourceName
		filled = true
	}
	if p.SourceHost != "" {
		host := StoragePoolSourceHost{Name: p.SourceHost}
		if p.SourcePort > 0 {
			host.Port = strconv.Itoa(p.SourcePort)
		}
		src.Host = []StoragePoolSourceHost{host}
		filled = true
	}
	if p.SourcePath != "" {
		src.Dir = &StoragePoolSourceDir{Path: p.SourcePath}
		filled = true
	}
	if p.SourceDev != "" {
		src.Device = []StoragePoolSourceDevice{{Path: p.SourceDev}}
		filled = true
	}
	if p.SourceFormat != "" {
		src.Format = &StoragePoolSourceFormat{Type: p.SourceFormat}
		filled = true
	}
	if p.SourceProtocolVer != "" {
		src.Protocol = &StoragePoolSourceProtocol{Version: p.SourceProtocolVer}
		filled = true
	}
	if p.SourceInitiator != "" {
		src.Initiator = &StoragePoolSourceInitiator{
			IQN: StoragePoolSourceInitiatorIQN{Name: p.SourceInitiator},
		}
		filled = true
	}
	if auth := buildAuth(p); auth != nil {
		src.Auth = auth
		filled = true
	}
	if adapter := buildAdapter(p); adapter != nil {
		src.Adapter = adapter
		filled = true
	}
	if !filled {
		return nil
	}
	return src
}

func buildAuth(p poolAsParams) *StoragePoolSourceAuth {
	if p.AuthType == "" {
		return nil
	}
	auth := &StoragePoolSourceAuth{
		Type:     p.AuthType,
		Username: p.AuthUsername,
	}
	if p.SecretUsage != "" || p.SecretUUID != "" {
		auth.Secret = &StoragePoolSourceAuthSecret{
			Usage: p.SecretUsage,
			UUID:  p.SecretUUID,
		}
	}
	return auth
}

func buildAdapter(p poolAsParams) *StoragePoolSourceAdapter {
	if p.AdapterName == "" && p.AdapterWWNN == "" && p.AdapterWWPN == "" {
		return nil
	}
	adapter := &StoragePoolSourceAdapter{
		Parent: p.AdapterParent,
		WWNN:   p.AdapterWWNN,
		WWPN:   p.AdapterWWPN,
	}
	if p.AdapterName != "" {
		adapter.Type = "scsi_host"
		adapter.Name = p.AdapterName
		return adapter
	}
	adapter.Type = "fc_host"
	// Parent fabric / parent wwnn+wwpn: libvirtxml exposes Parent string and
	// ParentAddr; store fabric/wwnn hints in Parent when only those are set.
	if p.AdapterParent == "" {
		if p.AdapterParentFabricWWN != "" {
			adapter.Parent = p.AdapterParentFabricWWN
		} else if p.AdapterParentWWNN != "" && p.AdapterParentWWPN != "" {
			adapter.Parent = p.AdapterParentWWNN + "," + p.AdapterParentWWPN
		}
	}
	return adapter
}
