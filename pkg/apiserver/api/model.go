package api

import (
	"strings"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

type State string

const (
	// purpose of the following CRDs is for operational job
	OperationStateSubmitted           State = "Submitted"
	OperationStateMigrateAddReplica   State = "AddReplica"
	OperationStateMigrateSyncReplica  State = "SyncReplica"
	OperationStateMigratePruneReplica State = "PruneReplica"
	OperationStateInProgress          State = "InProgress"
	OperationStateCompleted           State = "Completed"
	OperationStateToBeAborted         State = "ToBeAborted"
	OperationStateAborting            State = "Cancelled"
	OperationStateAborted             State = "Aborted"
	OperationStateFailed              State = "Failed"

	VolumeStateToBeUnmount State = "ToBeMounted"
	VolumeStateEmpty       State = ""
	VolumeStateCreated     State = "Created"
	VolumeStateCreating    State = "Creating"
	VolumeStateReady       State = "Ready"
	VolumeStateNotReady    State = "NotReady"
	VolumeStateToBeDeleted State = "ToBeDeleted"
	VolumeStateDeleted     State = "Deleted"

	VolumeStateUnknown State = "Unknown"

	VolumeReplicaStateInvalid     State = "Invalid"
	VolumeReplicaStateCreating    State = "Creating"
	VolumeReplicaStateReady       State = "Ready"
	VolumeReplicaStateNotReady    State = "NotReady"
	VolumeReplicaStateToBeDeleted State = "ToBeDeleted"
	VolumeReplicaStateDeleted     State = "Deleted"

	NodeStateReady            State = "Ready"
	NodeStateMaintain         State = "Maintain"
	NodeStateOffline          State = "Offline"
	NodeStateEmpty            State = ""
	NodeStateUnknown          State = "Unknown"
	NodeStateReadyAndNotReady State = "ReadyAndNotReady"

	NodeStateHealthy  State = "Healthy"
	NodeStateNotReady State = "NotReady"

	DriverStateEmpty    State = ""
	DriverStateReady    State = "Ready"
	DriverStateMaintain State = "Maintain"
	DriverStateOffline  State = "Offline"
	DriverStateUnknown  State = "Unknown"

	// LocalDiskUnclaimed represents that the disk is not bound to any LDC,
	// and is available for claiming.
	LocalDiskUnclaimed State = "Unclaimed"
	// LocalDiskReleased represents that the disk is released from the LDC,
	LocalDiskReleased State = "Released"
	// LocalDiskClaimed represents that the disk is bound to a LDC
	LocalDiskClaimed State = "Claimed"
	// LocalDiskInUse represents that the disk is in use but not claimed by a LDC
	LocalDiskInUse State = "Inuse"
	// LocalDiskReserved represents that the disk will be used in the feature
	LocalDiskReserved            State = "Reserved"
	LocalDiskReleaseReserved     State = "ReleaseReserved"
	LocalDiskEmpty               State = ""
	LocalDiskClaimedAndUnclaimed State = "ClaimedAndUnclaimed"
	LocalDiskPending             State = "Pending"
	LocalDiskBound               State = "Bound"
	LocalDiskAvailable           State = "Available"

	// LocalDiskActive is the state for the disk that is connected
	LocalDiskActive State = "Active"
	// LocalDiskInactive is the state for the disk that is disconnected
	LocalDiskInactive State = "Inactive"
	// LocalDiskUnknown is the state for the disk that cannot be determined
	// at this time(whether attached or detached)
	LocalDiskUnknown State = "Unknown"

	ModuleStatusRunning  State = "Running"
	ModuleStatusNotReady State = "NotReady"

	DrbdModuleStatusEnabled  State = "Enabled"
	DrbdModuleStatusDisabled State = "Disabled"

	ReservedSucceedState State = "Succeed"
	ReservedFailedState  State = "Failed"
)

type Pagination struct {
	// 总共有多少条目，请求时可以不用传递
	Total uint32 `json:"total,omitempty"`
	// 当前页索引，从 1 开始，为 0 时，会自动重置为默认值 constants.DefaultPage
	Page int32 `json:"page,omitempty"`
	// 总页数
	Pages int32 `json:"pages,omitempty"`
	// 每页数据量，为 -1 时表示查询全部，为 0 时会重置为默认值
	// constants.DefaultPageSize
	PageSize int32 `json:"pageSize,omitempty"`
	//// 排序规则，支持字符串和数字类型的字段进行排序
	//Sort string `json:"sort,omitempty"`
	//// 搜索关键字，支持模糊搜索,精准匹配和高级搜索.
	//Search string `protobuf:"bytes,5,opt,name=search,proto3" json:"search,omitempty"`
}

// disk path
const (
	DEV = "/dev/"
)

// disk class
const (
	DiskClassNameHDD  = "HDD"
	DiskClassNameSSD  = "SSD"
	DiskClassNameNVMe = "NVMe"
)

// consts
const (
	PoolNamePrefix  = "LocalStorage_Pool"
	PoolNameForHDD  = PoolNamePrefix + DiskClassNameHDD
	PoolNameForSSD  = PoolNamePrefix + DiskClassNameSSD
	PoolNameForNVMe = PoolNamePrefix + DiskClassNameNVMe
)

func StateConvert(state apisv1alpha1.State) State {
	switch state {
	case apisv1alpha1.OperationStateToBeAborted:
		return OperationStateToBeAborted

	case apisv1alpha1.OperationStateFailed:
		return OperationStateFailed

	case apisv1alpha1.OperationStateAborted:
		return OperationStateAborted

	case apisv1alpha1.OperationStateAborting:
		return OperationStateAborting

	case apisv1alpha1.OperationStateCompleted:
		return OperationStateCompleted

	case apisv1alpha1.OperationStateInProgress:
		return OperationStateInProgress

	case apisv1alpha1.OperationStateMigrateAddReplica:
		return OperationStateMigrateAddReplica

	case apisv1alpha1.OperationStateSubmitted:
		return OperationStateSubmitted

	case apisv1alpha1.OperationStateMigrateSyncReplica:
		return OperationStateMigrateSyncReplica

	case apisv1alpha1.OperationStateMigratePruneReplica:
		return OperationStateMigratePruneReplica

	case apisv1alpha1.VolumeStateToBeUnmount:
		return VolumeStateToBeUnmount

	case apisv1alpha1.VolumeStateEmpty:
		return VolumeStateEmpty

	case apisv1alpha1.VolumeStateCreated:
		return VolumeStateCreated

	case apisv1alpha1.VolumeStateCreating:
		return VolumeStateCreating

	case apisv1alpha1.VolumeStateReady:
		return VolumeStateReady

	case apisv1alpha1.VolumeStateNotReady:
		return VolumeStateNotReady

	case apisv1alpha1.VolumeStateToBeDeleted:
		return VolumeStateToBeDeleted

	case apisv1alpha1.VolumeStateDeleted:
		return VolumeStateDeleted

	case apisv1alpha1.VolumeReplicaStateInvalid:
		return VolumeReplicaStateInvalid

		//case apisv1alpha1.LocalDiskActive:
		//	return LocalDiskActive
		//
		//case apisv1alpha1.LocalDiskInactive:
		//	return LocalDiskInactive
		//
		//case apisv1alpha1.LocalDiskUnknown:
		//	return LocalDiskUnknown
	}
	return ""
}

func VolumeStatefuzzyConvert(state string) apisv1alpha1.State {
	if state == "" {
		return apisv1alpha1.VolumeStateEmpty
	}
	if strings.Contains("ToBeMounted", state) {
		return apisv1alpha1.VolumeStateToBeUnmount
	}
	if strings.Contains("Created", state) {
		return apisv1alpha1.VolumeStateCreated
	}
	if strings.Contains("Creating", state) {
		return apisv1alpha1.VolumeStateCreating
	}
	if strings.Contains("Ready", state) {
		return apisv1alpha1.VolumeStateReady
	}
	if strings.Contains("NotReady", state) {
		return apisv1alpha1.VolumeStateNotReady
	}
	if strings.Contains("ToBeDeleted", state) {
		return apisv1alpha1.VolumeStateToBeDeleted
	}
	if strings.Contains("Deleted", state) {
		return apisv1alpha1.VolumeStateDeleted
	}
	return ""
}

func NodeStatefuzzyConvert(state string) State {
	if state == "" {
		return NodeStateEmpty
	}
	if strings.Contains("Healthy", state) {
		return NodeStateHealthy
	}
	// todo
	if strings.Contains("Ready", state) {
		return NodeStateReadyAndNotReady
	}
	if strings.Contains("NotReady", state) {
		return NodeStateNotReady
	}

	if strings.Contains("Maintain", state) {
		return NodeStateMaintain
	}
	if strings.Contains("Offline", state) {
		return NodeStateOffline
	}

	return NodeStateUnknown
}

func OperationStatefuzzyConvert(state string) apisv1alpha1.State {
	if state == "" {
		return apisv1alpha1.VolumeStateEmpty
	}

	if strings.Contains("OperationStateSubmitted", state) {
		return apisv1alpha1.OperationStateSubmitted
	}

	if strings.Contains("OperationStateMigrateAddReplica", state) {
		return apisv1alpha1.OperationStateMigrateAddReplica
	}

	if strings.Contains("OperationStateMigrateSyncReplica", state) {
		return apisv1alpha1.OperationStateMigrateSyncReplica
	}

	if strings.Contains("OperationStateMigratePruneReplica", state) {
		return apisv1alpha1.OperationStateMigratePruneReplica
	}

	if strings.Contains("OperationStateInProgress", state) {
		return apisv1alpha1.OperationStateInProgress
	}

	if strings.Contains("OperationStateCompleted", state) {
		return apisv1alpha1.OperationStateCompleted
	}

	if strings.Contains("OperationStateToBeAborted", state) {
		return apisv1alpha1.OperationStateToBeAborted
	}

	if strings.Contains("OperationStateAborting", state) {
		return apisv1alpha1.OperationStateAborting
	}

	if strings.Contains("OperationStateAborted", state) {
		return apisv1alpha1.OperationStateAborted
	}

	if strings.Contains("OperationStateFailed", state) {
		return apisv1alpha1.OperationStateFailed
	}

	return apisv1alpha1.VolumeStateEmpty
}

func DriverStatefuzzyConvert(state string) apisv1alpha1.State {
	if state == "" {
		return ""
	}
	if strings.Contains("Ready", state) {
		return apisv1alpha1.NodeStateReady
	}
	if strings.Contains("Offline", state) {
		return apisv1alpha1.NodeStateOffline
	}
	if strings.Contains("Maintain", state) {
		return apisv1alpha1.NodeStateMaintain
	}
	return ""
}

func DiskStatefuzzyConvert(state string) apisv1alpha1.LocalDiskState {
	if state == "" {
		return apisv1alpha1.LocalDiskEmpty
	}

	if strings.Contains("Bound", state) {
		return apisv1alpha1.LocalDiskBound
	}
	if strings.Contains("Available", state) {
		return apisv1alpha1.LocalDiskAvailable
	}
	if strings.Contains("Pending", state) {
		return apisv1alpha1.LocalDiskPending
	}

	return apisv1alpha1.LocalDiskUnknown
}

type QueryPage struct {
	Page              int32
	Pages             int32
	PageSize          int32
	Name              string
	PoolName          string
	NodeName          string
	DiskName          string
	DeviceShortPath   string
	NameSpace         string
	VolumeState       apisv1alpha1.State
	NodeState         State
	OperationState    apisv1alpha1.State
	DriverState       apisv1alpha1.State
	DiskState         apisv1alpha1.LocalDiskState
	SnapshotState     apisv1alpha1.State
	SnapshotName      string
	VolumeName        string
	VolumeReplicaName string
	VolumeEventName   string
	VolumeGroup       string
	Synced            string
	OperationName     string
	Owner             string
	ResourceType      string
	ResourceName      string
	Action            string
	Sort              string
	SortDir           string
}
