package storage

import (
	"errors"
	udsv1alpha1 "github.com/HwameiStor/local-storage/pkg/apis/uds/v1alpha1"
)

// variables
var (
	ErrorPoolNotFound                 = errors.New("not found pool")
	ErrorReplicaNotFound              = errors.New("not found replica")
	ErrorReplicaExists                = errors.New("already exists replica")
	ErrorInsufficientRequestResources = errors.New("insufficient request resources")
	ErrorOverLimitedRequestResource   = errors.New("over limited request resources")
)

// LocalPoolManager is an interface to manage local storage pools
type LocalPoolManager interface {
	ExtendPools(localDisks []*udsv1alpha1.LocalDisk) error

	ExtendPoolsInfo(localDisks map[string]*udsv1alpha1.LocalDisk) (map[string]*udsv1alpha1.LocalPool, error)

	GetReplicas() (map[string]*udsv1alpha1.LocalVolumeReplica, error)
}

// LocalVolumeReplicaManager interface
type LocalVolumeReplicaManager interface {
	CreateVolumeReplica(replica *udsv1alpha1.LocalVolumeReplica) (*udsv1alpha1.LocalVolumeReplica, error)
	DeleteVolumeReplica(replica *udsv1alpha1.LocalVolumeReplica) error
	ExpandVolumeReplica(replica *udsv1alpha1.LocalVolumeReplica, newCapacityBytes int64) (*udsv1alpha1.LocalVolumeReplica, error)
	GetVolumeReplica(replica *udsv1alpha1.LocalVolumeReplica) (*udsv1alpha1.LocalVolumeReplica, error)
	TestVolumeReplica(replica *udsv1alpha1.LocalVolumeReplica) (*udsv1alpha1.LocalVolumeReplica, error)

	// consistencyCheck on all volume replicas by comparing VolumeReplica and underlying volumes
	// will log all the check results for alerting or other purpose, but not block anything
	ConsistencyCheck()
}

// LocalDiskManager is an interface to manage local disks
type LocalDiskManager interface {
	// Discover all disks including HDD, SSD, NVMe, etc..
	DiscoverAvailableDisks() ([]*udsv1alpha1.LocalDisk, error)
	GetLocalDisks() (map[string]*udsv1alpha1.LocalDisk, error)
}

// LocalRegistry interface
type LocalRegistry interface {
	Init()

	Disks() map[string]*udsv1alpha1.LocalDisk
	Pools() map[string]*udsv1alpha1.LocalPool
	VolumeReplicas() map[string]*udsv1alpha1.LocalVolumeReplica
	HasVolumeReplica(replica *udsv1alpha1.LocalVolumeReplica) bool
	UpdateNodeForVolumeReplica(replica *udsv1alpha1.LocalVolumeReplica)
	SyncResourcesToNodeCRD(localDisks map[string]*udsv1alpha1.LocalDisk) error
}

// LocalDeviceListInterface interface
type LocalDeviceListInterface interface {
	GetDevicesInfo(string, map[string]struct{}) map[string]*DeviceInfo
}

// LocalVolumeExecutor interface
type LocalVolumeExecutor interface {
	CreateVolumeReplica(replica *udsv1alpha1.LocalVolumeReplica) (*udsv1alpha1.LocalVolumeReplica, error)
	DeleteVolumeReplica(replica *udsv1alpha1.LocalVolumeReplica) error
	ExpandVolumeReplica(replica *udsv1alpha1.LocalVolumeReplica, newCapacityBytes int64) (*udsv1alpha1.LocalVolumeReplica, error)
	TestVolumeReplica(replica *udsv1alpha1.LocalVolumeReplica) (*udsv1alpha1.LocalVolumeReplica, error)
	// GetReplicas return all replicas
	GetReplicas() (map[string]*udsv1alpha1.LocalVolumeReplica, error)

	// consistencyCheck on all volume replicas by comparing VolumeReplica and underlying volumes
	// will log all the check results for alerting or other purpose, but not block anything
	ConsistencyCheck(crdReplicas map[string]*udsv1alpha1.LocalVolumeReplica)
}

// LocalPoolExecutor interface
type LocalPoolExecutor interface {
	ExtendPools(localDisks []*udsv1alpha1.LocalDisk) error
	ExtendPoolsInfo(localDisks map[string]*udsv1alpha1.LocalDisk) (map[string]*udsv1alpha1.LocalPool, error)
	GetReplicas() (map[string]*udsv1alpha1.LocalVolumeReplica, error)
}
