package storage

import (
	"errors"
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

// variables
var (
	ErrorPoolNotFound                   = errors.New("not found pool")
	ErrorReplicaNotFound                = errors.New("not found replica")
	ErrorSnapshotNotFound               = errors.New("not found snapshot")
	ErrorReplicaExists                  = errors.New("already exists replica")
	ErrorLocalVolumeExistsInVolumeGroup = errors.New("already exists in volume group")
	ErrorInsufficientRequestResources   = errors.New("insufficient request resources")
	ErrorOverLimitedRequestResource     = errors.New("over limited request resources")
)

/* A set of interface for Hwameistor Local Object */

// LocalPoolManager is an interface to manage local storage pools
//go:generate mockgen -source=types.go -destination=../../../member/node/storage/pools_mock.go  -package=storage
type LocalPoolManager interface {
	ExtendPools(localDisks []*apisv1alpha1.LocalDevice) (bool, error)

	GetPools() (map[string]*apisv1alpha1.LocalPool, error)

	GetReplicas() (map[string]*apisv1alpha1.LocalVolumeReplica, error)

	ResizePhysicalVolumes(localDisks map[string]*apisv1alpha1.LocalDevice) error
}

// LocalVolumeReplicaManager interface
//go:generate mockgen -source=types.go -destination=../../../member/node/storage/replica_mock.go  -package=storage
type LocalVolumeReplicaManager interface {
	CreateVolumeReplica(replica *apisv1alpha1.LocalVolumeReplica) (*apisv1alpha1.LocalVolumeReplica, error)
	DeleteVolumeReplica(replica *apisv1alpha1.LocalVolumeReplica) error
	ExpandVolumeReplica(replica *apisv1alpha1.LocalVolumeReplica, newCapacityBytes int64) (*apisv1alpha1.LocalVolumeReplica, error)
	GetVolumeReplica(replica *apisv1alpha1.LocalVolumeReplica) (*apisv1alpha1.LocalVolumeReplica, error)
	TestVolumeReplica(replica *apisv1alpha1.LocalVolumeReplica) (*apisv1alpha1.LocalVolumeReplica, error)

	// ConsistencyCheck  on all volume replicas by comparing VolumeReplica and underlying volumes
	// will log all the check results for alerting or other purpose, but not block anything
	ConsistencyCheck()
}

// LocalVolumeReplicaSnapshotManager interface
//go:generate mockgen -source=types.go -destination=../../../member/node/storage/replica_mock.go  -package=storage
type LocalVolumeReplicaSnapshotManager interface {
	CreateVolumeReplicaSnapshot(replicaSnapshot *apisv1alpha1.LocalVolumeReplicaSnapshot) error
	DeleteVolumeReplicaSnapshot(replicaSnapshot *apisv1alpha1.LocalVolumeReplicaSnapshot) error
	UpdateVolumeReplicaSnapshot(replicaSnapshot *apisv1alpha1.LocalVolumeReplicaSnapshot) (*apisv1alpha1.LocalVolumeReplicaSnapshotStatus, error)
	GetVolumeReplicaSnapshot(replicaSnapshot *apisv1alpha1.LocalVolumeReplicaSnapshot) (*apisv1alpha1.LocalVolumeReplicaSnapshotStatus, error)

	LocalVolumeReplicaSnapshotRestoreManager
}

//go:generate mockgen -source=types.go -destination=../../../member/node/storage/replica_mock.go  -package=storage
type LocalVolumeReplicaSnapshotRestoreManager interface {
	RollbackVolumeReplicaSnapshot(snapshotRestore *apisv1alpha1.LocalVolumeReplicaSnapshotRestore) error
	RestoreVolumeReplicaSnapshot(snapshotRestore *apisv1alpha1.LocalVolumeReplicaSnapshotRestore) error
}

// LocalRegistry interface
//go:generate mockgen -source=types.go -destination=../../../member/node/storage/registry_mock.go  -package=storage
type LocalRegistry interface {
	Init()

	Disks() map[string]*apisv1alpha1.LocalDevice
	Pools() map[string]*apisv1alpha1.LocalPool
	VolumeReplicas() map[string]*apisv1alpha1.LocalVolumeReplica
	HasVolumeReplica(replica *apisv1alpha1.LocalVolumeReplica) bool
	UpdateNodeForVolumeReplica(replica *apisv1alpha1.LocalVolumeReplica)
	SyncNodeResources() error
	UpdateCondition(condition apisv1alpha1.StorageNodeCondition) error
	UpdatePoolExtendRecord(pool string, record apisv1alpha1.LocalDiskClaimSpec) error
}

/* A set of interface for executor to implement the above Hwameistor Local Object interface */

// LocalVolumeExecutor interface
//go:generate mockgen -source=types.go -destination=../../../member/node/storage/replica_executor_mock.go  -package=storage
type LocalVolumeExecutor interface {
	CreateVolumeReplica(replica *apisv1alpha1.LocalVolumeReplica) (*apisv1alpha1.LocalVolumeReplica, error)
	DeleteVolumeReplica(replica *apisv1alpha1.LocalVolumeReplica) error
	ExpandVolumeReplica(replica *apisv1alpha1.LocalVolumeReplica, newCapacityBytes int64) (*apisv1alpha1.LocalVolumeReplica, error)
	TestVolumeReplica(replica *apisv1alpha1.LocalVolumeReplica) (*apisv1alpha1.LocalVolumeReplica, error)
	// GetReplicas return all replicas
	GetReplicas() (map[string]*apisv1alpha1.LocalVolumeReplica, error)

	// ConsistencyCheck on all volume replicas by comparing VolumeReplica and underlying volumes
	// will log all the check results for alerting or other purpose, but not block anything
	ConsistencyCheck(crdReplicas map[string]*apisv1alpha1.LocalVolumeReplica)
}

type LocalVolumeReplicaSnapshotExecutor interface {
	CreateVolumeReplicaSnapshot(replicaSnapshot *apisv1alpha1.LocalVolumeReplicaSnapshot) error
	DeleteVolumeReplicaSnapshot(replicaSnapshot *apisv1alpha1.LocalVolumeReplicaSnapshot) error
	UpdateVolumeReplicaSnapshot(replicaSnapshot *apisv1alpha1.LocalVolumeReplicaSnapshot) (*apisv1alpha1.LocalVolumeReplicaSnapshotStatus, error)
	GetVolumeReplicaSnapshot(replicaSnapshot *apisv1alpha1.LocalVolumeReplicaSnapshot) (*apisv1alpha1.LocalVolumeReplicaSnapshotStatus, error)

	LocalVolumeReplicaSnapshotRestoreManager
}

// LocalPoolExecutor interface
//go:generate mockgen -source=types.go -destination=../../../member/node/storage/pools_executor_mock.go  -package=storage
type LocalPoolExecutor interface {
	ExtendPools(localDisks []*apisv1alpha1.LocalDevice) (bool, error)
	GetPools() (map[string]*apisv1alpha1.LocalPool, error)
	GetReplicas() (map[string]*apisv1alpha1.LocalVolumeReplica, error)
	ResizePhysicalVolumes(localDisks map[string]*apisv1alpha1.LocalDevice) error
}
