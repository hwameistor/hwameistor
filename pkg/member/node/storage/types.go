package storage

import (
	"errors"
	"os"
	"syscall"

	apisv1alpha1 "github.com/hwameistor/local-storage/pkg/apis/hwameistor/v1alpha1"
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
////go:generate mockgen -source=types.go -destination=../../../member/node/storage/pools_mock.go  -package=storage
type LocalPoolManager interface {
	ExtendPools(localDisks []*apisv1alpha1.LocalDisk) error

	ExtendPoolsInfo(localDisks map[string]*apisv1alpha1.LocalDisk) (map[string]*apisv1alpha1.LocalPool, error)

	GetReplicas() (map[string]*apisv1alpha1.LocalVolumeReplica, error)
}

// LocalVolumeReplicaManager interface
//go:generate mockgen -source=types.go -destination=../../../member/node/storage/replica_mock.go  -package=storage
type LocalVolumeReplicaManager interface {
	CreateVolumeReplica(replica *apisv1alpha1.LocalVolumeReplica) (*apisv1alpha1.LocalVolumeReplica, error)
	DeleteVolumeReplica(replica *apisv1alpha1.LocalVolumeReplica) error
	ExpandVolumeReplica(replica *apisv1alpha1.LocalVolumeReplica, newCapacityBytes int64) (*apisv1alpha1.LocalVolumeReplica, error)
	GetVolumeReplica(replica *apisv1alpha1.LocalVolumeReplica) (*apisv1alpha1.LocalVolumeReplica, error)
	TestVolumeReplica(replica *apisv1alpha1.LocalVolumeReplica) (*apisv1alpha1.LocalVolumeReplica, error)

	// consistencyCheck on all volume replicas by comparing VolumeReplica and underlying volumes
	// will log all the check results for alerting or other purpose, but not block anything
	ConsistencyCheck()
}

// LocalRegistry interface
////go:generate mockgen -source=types.go -destination=../../../member/node/storage/registry_mock.go  -package=storage
type LocalRegistry interface {
	Init()

	Disks() map[string]*apisv1alpha1.LocalDisk
	Pools() map[string]*apisv1alpha1.LocalPool
	VolumeReplicas() map[string]*apisv1alpha1.LocalVolumeReplica
	HasVolumeReplica(replica *apisv1alpha1.LocalVolumeReplica) bool
	UpdateNodeForVolumeReplica(replica *apisv1alpha1.LocalVolumeReplica)
	SyncResourcesToNodeCRD(localDisks map[string]*apisv1alpha1.LocalDisk) error
}

// DeviceInfo struct
type DeviceInfo struct {
	OSFileInfo   os.FileInfo
	SysTStat     *syscall.Stat_t
	Path         string
	Name         string
	Major        uint32
	Minor        uint32
	MajMinString string
}

// LocalVolumeExecutor interface
////go:generate mockgen -source=types.go -destination=../../../member/node/storage/replica_executor_mock.go  -package=storage
type LocalVolumeExecutor interface {
	CreateVolumeReplica(replica *apisv1alpha1.LocalVolumeReplica) (*apisv1alpha1.LocalVolumeReplica, error)
	DeleteVolumeReplica(replica *apisv1alpha1.LocalVolumeReplica) error
	ExpandVolumeReplica(replica *apisv1alpha1.LocalVolumeReplica, newCapacityBytes int64) (*apisv1alpha1.LocalVolumeReplica, error)
	TestVolumeReplica(replica *apisv1alpha1.LocalVolumeReplica) (*apisv1alpha1.LocalVolumeReplica, error)
	// GetReplicas return all replicas
	GetReplicas() (map[string]*apisv1alpha1.LocalVolumeReplica, error)

	// consistencyCheck on all volume replicas by comparing VolumeReplica and underlying volumes
	// will log all the check results for alerting or other purpose, but not block anything
	ConsistencyCheck(crdReplicas map[string]*apisv1alpha1.LocalVolumeReplica)
}

// LocalPoolExecutor interface
////go:generate mockgen -source=types.go -destination=../../../member/node/storage/pools_executor_mock.go  -package=storage
type LocalPoolExecutor interface {
	ExtendPools(localDisks []*apisv1alpha1.LocalDisk) error
	ExtendPoolsInfo(localDisks map[string]*apisv1alpha1.LocalDisk) (map[string]*apisv1alpha1.LocalPool, error)
	GetReplicas() (map[string]*apisv1alpha1.LocalVolumeReplica, error)
}
