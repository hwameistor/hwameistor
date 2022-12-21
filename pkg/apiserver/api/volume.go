package api

import (
	"time"
)

// LocalPool is storage pool struct
type LocalPool struct {
	// Supported pool name: HDD_POOL, SSD_POOL, NVMe_POOL 存储池
	Name string `json:"name,omitempty"`
}

// Volume
type Volume struct {
	// local volume name 名称
	Name string `json:"name,omitempty"`

	// local volume state 状态
	State State `json:"state,omitempty"`

	// replica number 副本数
	ReplicaNumber int64 `json:"replicaNumber,omitempty"`

	// VolumeGroup is the group name of the local volumes. It is designed for the scheduling and allocating. 磁盘组
	VolumeGroup string `json:"volumeGroup,omitempty"`

	// size 容量
	RequiredCapacityBytes int64 `json:"requiredCapacityBytes,omitempty"`

	// PersistentVolumeClaimNamespace is the namespace of the associated PVC 命名空间
	PersistentVolumeClaimNamespace string `json:"pvcNamespace,omitempty"`

	// PersistentVolumeClaimName is the name of the associated PVC 绑定PVC
	PersistentVolumeClaimName string `json:"pvcName,omitempty"`

	// Convertible 转换高可用模式
	Convertible bool `json:"convertible,omitempty"`

	// createTime 创建时间
	CreateTime time.Time `json:"createTime,omitempty"`
}

type VolumeItemsList struct {
	// volumes
	Volumes []*Volume `json:"volumes,omitempty"`
}

// VolumeList
type VolumeList struct {
	// volumes
	VolumeItemsList VolumeItemsList `json:"items,omitempty"`
	// page 信息
	Page *Pagination `json:"pagination,omitempty"`
}

// VolumeReplica
type VolumeReplica struct {
	// replica name
	Name string `json:"name,omitempty"`

	// replica state
	State State `json:"state,omitempty"`

	// Synced is the sync state of the volume replica, which is important in HA volume 同步状态
	Synced bool `json:"synced,omitempty"`

	// NodeName is the assigned node where the volume replica is located 节点
	NodeName string `json:"nodeName,omitempty"`

	// RequiredCapacityBytes 容量
	RequiredCapacityBytes int64 `json:"requiredCapacityBytes,omitempty"`

	// StoragePath is a real path of the volume replica, like /dev/sdg.
	StoragePath string `json:"storagePath,omitempty"`

	// DevicePath is a link path of the StoragePath of the volume replica,
	// e.g. /dev/LocalStorage_PoolHDD/pvc-fbf3ffc3-66db-4dae-9032-bda3c61b8f85
	DevicePath string `json:"devicePath,omitempty"`
}

// VolumeReplicaList
type VolumeReplicaList struct {
	// volume name
	VolumeName string `json:"volumeName,omitempty"`
	// VolumeReplicas
	VolumeReplicas []*VolumeReplica `json:"volumeReplicas,omitempty"`
}

// VolumeOperationList
type VolumeOperationListByNode struct {
	// node name
	NodeName string `json:"nodeName,omitempty"`
	//// VolumeOperations
	//VolumeMigrateOperations []*VolumeMigrateOperation `json:"items,omitempty"`
	// VolumeMigrateOperationItemsList
	VolumeMigrateOperationItemsList VolumeMigrateOperationItemsList `json:"items,omitempty"`
	// page 信息
	Page *Pagination `json:"pagination,omitempty"`
}

// VolumeMigrateOperationItemsList
type VolumeMigrateOperationItemsList struct {
	// VolumeMigrateOperations
	VolumeMigrateOperations []*VolumeMigrateOperation `json:"volumeMigrateOperations,omitempty"`
}

// VolumeOperationByVolume
type VolumeOperationByVolume struct {
	// VolumeName
	VolumeName string `json:"volumeName,omitempty"`
	// VolumeMigrateOperationItemsList
	VolumeMigrateOperationItemsList VolumeMigrateOperationItemsList `json:"items,omitempty"`
	//// VolumeMigrateOperations
	//VolumeMigrateOperations []*VolumeMigrateOperation `json:"items,omitempty"`
}

// VolumeOperationByMigrate
type VolumeOperationByMigrate struct {
	// VolumeMigrateName name
	VolumeMigrateName string `json:"volumeMigrateName,omitempty"`
	// VolumeOperation
	VolumeMigrateOperation *VolumeMigrateOperation `json:"volumeMigrateOperation,omitempty"`
}

// VolumeMigrateOperation
type VolumeMigrateOperation struct {
	// VolumeMigrateName 迁移CRD名称
	Name string `json:"name"`

	// State 迁移状态
	State State `json:"state,omitempty"`

	// VolumeName 迁移卷名称
	VolumeName string `json:"volumeName"`

	// SourceNode 迁移源节点
	SourceNode string `json:"sourceNode"`

	// TargetNode 迁移目的节点
	TargetNode string `json:"targetNode"`

	// StartTime 迁移开始时间
	StartTime time.Time `json:"startTime,omitempty"`

	// EndTime 迁移结束时间
	EndTime time.Time `json:"endTime,omitempty"`
}

// VolumeConvertOperation
type VolumeConvertOperation struct {
	// VolumeConvert Name 转换CRD名称
	Name string `json:"name"`

	// State 转换状态
	State State `json:"state,omitempty"`

	// VolumeName 转换卷名称
	VolumeName string `json:"volumeName"`

	// ReplicaNumber 副本数
	ReplicaNumber string `json:"replicaNumber"`

	// StartTime 转换开始时间
	StartTime time.Time `json:"startTime,omitempty"`

	// EndTime 转换结束时间
	EndTime time.Time `json:"endTime,omitempty"`
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

// VolumeGroupVolumeInfo defines the observed volume state of VolumeGroup
type VolumeGroupVolumeInfo struct {
	// VolumeName
	VolumeName string `json:"volumeName,omitempty"`
	// NodeNames
	NodeNames []string `json:"nodeNames,omitempty"`
	// State
	State State `json:"state,omitempty"`
}

// VolumeGroup defines the observed state of VolumeGroup
type VolumeGroup struct {
	// Name
	Name string `json:"name"`
	// NodeNames
	NodeNames []string `json:"nodeNames,omitempty"`
	// VolumeGroupVolumeInfo
	VolumeGroupVolumeInfos []VolumeGroupVolumeInfo `json:"volumeGroupVolumeInfos,omitempty"`
}

type VolumeMigrateRspBody struct {
	VolumeMigrateInfo *VolumeMigrateInfo `json:"data,omitempty"`
}

type VolumeMigrateInfo struct {
	VolumeName   string `json:"volumeName,omitempty"`
	SrcNode      string `json:"srcNode,omitempty"`
	SelectedNode string `json:"selectedNode,omitempty"`
}

//type VolumeMigrateInfo struct {
//	VolumeName   string `form:"volumeName" json:"volumeName" binding:"required"`
//	SrcNode      string `form:"srcNode" json:"srcNode" binding:"required"`
//	SelectedNode string `form:"selectedNode" json:"selectedNode" binding:"required"`
//}

type VolumeConvertRspBody struct {
	VolumeConvertInfo *VolumeConvertInfo `json:"data,omitempty"`
}

type VolumeConvertInfo struct {
	VolumeName string `json:"volumeName"`
	ReplicaNum int64  `json:"replicaNum"`
}
