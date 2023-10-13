package api

import (
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

// LocalPool is storage pool struct
type LocalPool struct {
	// Supported pool name: HDD_POOL, SSD_POOL, NVMe_POOL 存储池
	Name string `json:"name,omitempty"`
}

type Volume struct {
	apisv1alpha1.LocalVolume
}

type VolumeItemsList struct {
	// volumes
	Volumes []*Volume `json:"items,omitempty"`
}

type VolumeList struct {
	Volumes []*Volume `json:"items"`
	// page 信息
	Page *Pagination `json:"pagination,omitempty"`
}

type VolumeReplica struct {
	apisv1alpha1.LocalVolumeReplica

	//// replica state todo
	//State State `json:"state,omitempty"`
}

type VolumeReplicaList struct {
	// volume name
	VolumeName string `json:"volumeName,omitempty"`
	// VolumeReplicas
	VolumeReplicas []*VolumeReplica `json:"volumeReplicas,omitempty"`
}

type VolumeOperationListByNode struct {
	// node name
	NodeName string `json:"nodeName,omitempty"`
	// VolumeOperations
	VolumeMigrateOperations []*VolumeMigrateOperation `json:"items,omitempty"`
	// page 信息
	Page *Pagination `json:"pagination,omitempty"`
}

type VolumeOperationByVolume struct {
	// VolumeName
	VolumeName string `json:"volumeName,omitempty"`
	//// OperationList
	//OperationList []Operation `json:"items"`
	// VolumeMigrateOperations
	VolumeMigrateOperations []*VolumeMigrateOperation `json:"volumeMigrateOperations,omitempty"`
	// VolumeConvertOperations
	VolumeConvertOperations []*VolumeConvertOperation `json:"VolumeConvertOperations,omitempty"`
	// VolumeExpandOperations
	VolumeExpandOperations []*VolumeExpandOperation `json:"VolumeExpandOperations,omitempty"`
}

type VolumeOperationByMigrate struct {
	// VolumeMigrateName name
	VolumeMigrateName string `json:"volumeMigrateName,omitempty"`
	// VolumeOperation
	VolumeMigrateOperation *VolumeMigrateOperation `json:"volumeMigrateOperation,omitempty"`
}

type VolumeMigrateOperation struct {
	apisv1alpha1.LocalVolumeMigrate
}

type VolumeConvertOperation struct {
	apisv1alpha1.LocalVolumeConvert
}

type VolumeExpandOperation struct {
	apisv1alpha1.LocalVolumeExpand
}

type VolumeSnapshotOperation struct {
	apisv1alpha1.LocalVolumeSnapshot
}

// LocalVolumeMigrateSpec defines the desired state of LocalVolumeMigrate
type LocalVolumeMigrateSpec struct {
	// volumeName
	VolumeName string `json:"volumeName"`
	// sourceNode
	SourceNode string `json:"sourceNode"`
	// targetNodesSuggested
	TargetNodesSuggested []string `json:"targetNodesSuggested"`
	// migrateAllVols
	MigrateAllVols bool `json:"migrateAllVols,omitempty"`
	// abort
	Abort bool `json:"abort,omitempty"`
}

// LocalVolumeMigrateStatus defines the observed state of LocalVolumeMigrate
type LocalVolumeMigrateStatus struct {
	// record the volume's replica number, it will be set internally
	OriginalReplicaNumber int64 `json:"originalReplicaNumber,omitempty"`
	// record the node where the specified replica is migrated to
	TargetNode string `json:"targetNode,omitempty"`

	// State of the operation, e.g. submitted, started, completed, abort, ...
	State State `json:"state,omitempty"`
	// error message to describe some states
	Message string `json:"message,omitempty"`
}

type LocalVolumeMigrate struct {
	Kind     string `yaml:"kind"`
	Metadata struct {
		Name      string `yaml:"name"`
		Namespace string `yaml:"namespace"`
	}

	Spec   LocalVolumeMigrateSpec   `json:"spec,omitempty"`
	Status LocalVolumeMigrateStatus `json:"status,omitempty"`
}

type LocalVolumeReplica struct {
	Kind     string `yaml:"kind"`
	Metadata struct {
		Name      string `yaml:"name"`
		Namespace string `yaml:"namespace"`
	}

	Spec   LocalVolumeReplicaSpec   `json:"spec,omitempty"`
	Status LocalVolumeReplicaStatus `json:"status,omitempty"`
}

// LocalVolumeReplicaSpec defines the desired state of LocalVolumeReplica
type LocalVolumeReplicaSpec struct {
	// VolumeName is the name of the volume, e.g. pvc-fbf3ffc3-66db-4dae-9032-bda3c61b8f85
	VolumeName string `json:"volumeName,omitempty"`

	// PoolName is the name of the storage pool, e.g. LocalStorage_PoolHDD, LocalStorage_PoolSSD, etc..
	PoolName string `json:"poolName,omitempty"`

	// NodeName is the assigned node where the volume replica is located
	NodeName string `json:"nodeName,omitempty"`

	RequiredCapacityBytes int64 `json:"requiredCapacityBytes,omitempty"`

	Delete bool `json:"delete,omitempty"`
}

// LocalVolumeReplicaStatus defines the observed state of LocalVolumeReplica
type LocalVolumeReplicaStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster

	// StoragePath is a real path of the volume replica, like /dev/sdg.
	StoragePath string `json:"storagePath,omitempty"`

	// DevicePath is a link path of the StoragePath of the volume replica,
	// e.g. /dev/LocalStorage_PoolHDD/pvc-fbf3ffc3-66db-4dae-9032-bda3c61b8f85
	DevicePath string `json:"devPath,omitempty"`

	// Disks is a list of physical disks where the volume replica is spread cross, especially for striped LVM volume replica
	Disks []string `json:"disks,omitempty"`

	// AllocatedCapacityBytes is the real allocated capacity in bytes
	AllocatedCapacityBytes int64 `json:"allocatedCapacityBytes,omitempty"`

	// State is the phase of volume replica, e.g. Creating, Ready, NotReady, ToBeDeleted, Deleted
	State State `json:"state,omitempty"`

	// Synced is the sync state of the volume replica, which is important in HA volume
	Synced bool `json:"synced,omitempty"`

	// HAState is state for ha replica, replica.Status.State == Ready only when HAState is Consistent of nil
	HAState *HAState `json:"haState,omitempty"`

	// InUse is one of volume replica's states, which indicates the replica is used by a Pod or not
	InUse bool `json:"inUse,omitempty"`
}

// HAState is state for ha replica
type HAState struct {
	// Consistent, Inconsistent, replica is ready only when consistent
	State State `json:"state"`
	// Reason is why this state happened
	Reason string `json:"reason,omitempty"`
}

// VolumeGroup defines the observed state of VolumeGroup
type VolumeGroup struct {
	apisv1alpha1.LocalVolumeGroup
	// Volumes
	Volumes []apisv1alpha1.LocalVolume `json:"items,omitempty"`
}

type VolumeGroupList struct {
	// VolumeGroups
	VolumeGroups []VolumeGroup `json:"items"`
}

type VolumeMigrateInfo struct {
	VolumeName   string `json:"volumeName,omitempty"`
	SrcNode      string `json:"srcNode,omitempty"`
	SelectedNode string `json:"selectedNode,omitempty"`
}

type VolumeMigrateRspBody struct {
	VolumeMigrateInfo *VolumeMigrateInfo `json:"data,omitempty"`
}

type VolumeMigrateReqBody struct {
	SrcNode      string `json:"srcNode,omitempty"`
	SelectedNode string `json:"selectedNode,omitempty"`
	Abort        bool   `json:"abort,omitempty default:false"`
}

type VolumeConvertReqBody struct {
	Abort bool `json:"abort,omitempty default:false"`
}

type VolumeConvertRspBody struct {
	VolumeConvertInfo *VolumeConvertInfo `json:"data,omitempty"`
}

type VolumeConvertInfo struct {
	VolumeName string `json:"volumeName,omitempty"`
	ReplicaNum int64  `json:"replicaNum"`
}

type VolumeExpandReqBody struct {
	//VolumeName     string `json:"volumeName,omitempty"`
	TargetCapacity string `json:"targetCapacity"`
	Abort          bool   `json:"abort"`
}

type VolumeExpandRspBody struct {
	VolumeExpandInfo *VolumeExpandInfo `json:"data,omitempty"`
}

type VolumeExpandInfo struct {
	VolumeName          string `json:"volumeName"`
	TargetCapacityBytes int64  `json:"targetCapacityBytes"`
}

type VolumeSnapshotRepBody struct {
	VolumeName string `json:"volumeName,omitempty"`
	Capacity   string `json:"capacity"`
	Abort      bool   `json:"abort"`
}

type VolumeSnapshotRspBody struct {
	VolumeSnapshotInfo *VolumeSnapshotInfo `json:"data,omitempty"`
}

type VolumeSnapshotInfo struct {
	Name          string `json:"Name"`
	CapacityBytes int64  `json:"capacityBytes"`
	SourceVolume  string `json:"sourceVolume"`
}
