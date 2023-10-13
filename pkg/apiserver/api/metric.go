package api

import (
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"time"

	apiv1alpha1 "github.com/hwameistor/hwameistor-operator/api/v1alpha1"
)

type BaseMetric struct {
	// 高可用卷数目
	HighAvailableVolumeNum int64 `json:"highAvailableVolumeNum"`
	// 非高可用卷数目
	NonHighAvailableVolumeNum int64 `json:"nonHighAvailableVolumeNum"`
	// 本地卷总数
	LocalVolumeNum int64 `json:"localVolumeNum"`
	// 总容量
	TotalCapacityBytes int64 `json:"totalCapacityBytes"`
	// 已分配容量
	AllocatedCapacityBytes int64 `json:"allocatedCapacityBytes"`
	// 已预留容量
	ReservedCapacityBytes int64 `json:"reservedCapacityBytes"`
	// 可分配容量
	FreeCapacityBytes int64 `json:"freeCapacityBytes"`
	// 总磁盘数
	TotalDiskNum int64 `json:"totalDiskNum"`
	// 绑定磁盘
	BoundedDiskNum int64 `json:"boundedDiskNum"`
	// 健康磁盘
	HealthyDiskNum int64 `json:"healthyDiskNum"`
	// 错误磁盘
	UnHealthyDiskNum int64 `json:"unHealthyDiskNum"`
	// 总节点数
	TotalNodeNum int64 `json:"totalNodeNum"`
	// 纳管节点数
	ClaimedNodeNum int64 `json:"claimedNodeNum"`
}

// 存储池资源使用
type StoragePoolUse struct {
	// 存储池名字
	Name string `json:"name"`
	// 总容量
	TotalCapacityBytes int64 `json:"totalCapacityBytes"`
	// 已分配容量
	AllocatedCapacityBytes int64 `json:"allocatedCapacityBytes"`
}

// 存储池资源监控
type StoragePoolUseMetric struct {
	StoragePoolsUse []StoragePoolUse `json:"storagePoolsUse"`
}

// 节点存储使用率
type NodeStorageUse struct {
	// 存储节点名字
	Name string `json:"name"`
	// 总容量
	TotalCapacityBytes int64 `json:"totalCapacityBytes"`
	// 已分配容量
	AllocatedCapacityBytes int64 `json:"allocatedCapacityBytes"`
}

// 节点存储TOP5 使用率监控
type NodeStorageUseMetric struct {
	// 存储池类型 SSD HDD
	StoragePoolClass string `json:"storagePoolClass"`
	// 节点存储TOP5 使用率列表 5条列表上限
	NodeStoragesUse []NodeStorageUse `json:"nodeStoragesUse"`
}

// 组件状态
type ModuleState struct {
	// 组件名称
	Name string `json:"name"`
	// 组件状态 运行中 未就绪
	State State `json:"state"`
}

// 组件状态监控
type ModuleStatus struct {
	apiv1alpha1.ClusterStatus

	ModulesStatus []ModuleState `json:"modulesStatus"`
}

// 操作记录
type Operation struct {
	// 事件名称
	EventName string `json:"eventName"`
	// 事件类型
	EventType string `json:"eventType"`
	// 操作对象
	LocalVolumeName string `json:"localVolumeName"`
	// 状态
	Status State `json:"status"`
	// 详细描述
	Description string `json:"description"`
	// 开始时间
	StartTime time.Time `json:"startTime"`
	// 结束时间
	EndTime time.Time `json:"endTime"`
}

// 操作记录列表
type OperationMetric struct {
	OperationList []Operation `json:"items"`
	// page 信息
	Page *Pagination `json:"pagination,omitempty"`
}

type StorageCapacityCollection struct {
	// 总容量
	TotalCapacityBytes int64 `json:"totalCapacityBytes"`
	// 已分配容量
	AllocatedCapacityBytes int64 `json:"allocatedCapacityBytes"`
	// 已预留容量
	ReservedCapacityBytes int64 `json:"reservedCapacityBytes"`
	// 可使用容量
	FreeCapacityBytes int64 `json:"freeCapacityBytes"`
}

type StorageNodesCollection struct {
	// 总节点数
	TotalNodesNum int64 `json:"totalNodesNum"`
	// 纳管节点数
	ManagedNodesNum int64 `json:"managedNodesNum"`
}

type VolumeCollection struct {
	// 本地卷总数
	TotalVolumesNum int64 `json:"totalVolumesNum"`
	// 高可用卷
	HAVolumeNum int64 `json:"HAVolumeNum"`
	// 非高可用卷
	NonHAVolumeNum int64 `json:"nonHAVolumeNum"`
}

type DiskCollection struct {
	// 磁盘总数
	TotalDisksNum int64 `json:"totalDisksNum"`
	// 健康磁盘数目
	HealthyDiskNum int64 `json:"healthyDiskNum"`
	// 错误磁盘数目
	ErrorDiskNum int64 `json:"errorDiskNum"`
	// 绑定磁盘数目
	BoundedDiskNum int64 `json:"boundedDiskNum"`
}

type ModuleStatusCollection struct {
	// 组件状态
	ModuleStatus map[string]State `json:"moduleStatus"`
}

type StoragePoolUseCollection struct {
	// 存储池资源使用
	StoragePoolUseMap map[string]StoragePoolUse `json:"storagePoolUseMap"`
}

// 节点存储使用率
type NodeStorageUseRatio struct {
	// 节点名字
	Name string
	// 总容量
	TotalCapacityBytes int64
	// 已分配容量
	AllocatedCapacityBytes int64
	// 存储比率
	CapacityBytesRatio int64
}

type NodeStorageUseRatios []*NodeStorageUseRatio

func (p NodeStorageUseRatios) Len() int {
	return len(p)
}
func (p NodeStorageUseRatios) Less(i, j int) bool {
	return p[i].CapacityBytesRatio < p[j].CapacityBytesRatio
}
func (p NodeStorageUseRatios) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

type NodeStorageUseCollection struct {
	// 存储节点资源使用率
	NodeStorageUseRatios []*NodeStorageUseRatio `json:"nodeStorageUseRatios"`
}

type EventList struct {
	Event []*Event `json:"items"`
	// page 信息
	Page *Pagination `json:"pagination,omitempty"`
}

type Event struct {
	apisv1alpha1.Event
}

type EventActionList struct {
	EventActions []*EventAction `json:"items"`
	// page 信息
	Page *Pagination `json:"pagination,omitempty"`
}

type EventAction struct {
	EventRecord  apisv1alpha1.EventRecord `json:"eventRecord"`
	ResourceName string                   `json:"resourceName"`
	ResourceType string                   `json:"resourceType"`
}
