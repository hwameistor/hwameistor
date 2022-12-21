package api

import (
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

// StorageNode
type StorageNode struct {
	// name 节点名字
	Name string `json:"name,omitempty"`
	// ip 节点IP
	IP string `json:"ip,omitempty"`
	// node state 节点状态 运行中（Ready）,未就绪（NotReady）,未知（Unknown）
	NodeState State `json:"nodeState,omitempty"`
	// driver status 驱动状态  运行中（Ready）,维护中（Maintain）, 离线（Offline）
	DriverStatus State `json:"driverStatus,omitempty"`
	// totalDiskCount 总磁盘数
	TotalDiskCount int64 `json:"totalDiskCount,omitempty"`
	// usedDiskCount 已绑定磁盘数
	UsedDiskCount int64 `json:"usedDiskCount,omitempty"`
	//// freeCapacityBytes LSN可分配存储容量
	//FreeCapacityBytes int64 `json:"freeCapacityBytes,omitempty"`
	// totalHDDCapacityBytes HDD存储总容量
	TotalHDDCapacityBytes int64 `json:"totalHDDCapacityBytes,omitempty"`
	// totalSSDCapacityBytes SSD存储总容量
	TotalSSDCapacityBytes int64 `json:"totalSSDCapacityBytes,omitempty"`
	// allocatedHDDCapacityBytes HDD已经分配存储量
	AllocatedHDDCapacityBytes int64 `json:"allocatedHDDCapacityBytes,omitempty"`
	// allocatedSSDCapacityBytes SSD已经分配存储量
	AllocatedSSDCapacityBytes int64 `json:"allocatedSSDCapacityBytes,omitempty"`
	// IsRAID 是否Raid
	IsRAID bool `json:"isRaid,omitempty"`
}

type LocalDisksItemsList struct {
	// localDisks 节点磁盘列表
	LocalDisks []*LocalDisk `json:"localDisks,omitempty"`
}

// LocalDiskListByNode
type LocalDiskListByNode struct {
	// nodeName 节点名称
	NodeName string `json:"nodeName,omitempty"`
	//// localDisks 节点磁盘列表
	//LocalDisks []*LocalDisk `json:"items,omitempty"`
	// localDisks 节点磁盘列表
	LocalDisksItemsList LocalDisksItemsList `json:"items,omitempty"`
	// page 信息
	Page *Pagination `json:"pagination,omitempty"`
}

// StorageNodesItemsList
type StorageNodesItemsList struct {
	// localDisks 节点磁盘列表
	StorageNodes []*StorageNode `json:"storageNodes,omitempty"`
}

// StorageNodeList
type StorageNodeList struct {
	//// StorageNodes
	//StorageNodes []*StorageNode `json:"items,omitempty"`
	// StorageNodesItemsList
	StorageNodesItemsList StorageNodesItemsList `json:"items,omitempty"`
	// page 信息
	Page *Pagination `json:"pagination,omitempty"`
}

// YamlData
type YamlData struct {
	// yaml data
	Data string `json:"data,omitempty"`
}

// TargetNodeList
type TargetNodeList struct {
	// TargetNodeType
	TargetNodeType string `json:"targetNodeType,omitempty"`
	// TargetNodes
	TargetNodes []string `json:"targetNodes,omitempty"`
}

func ToStorageNodeResource(lsn apisv1alpha1.LocalStorageNode) *StorageNode {
	r := &StorageNode{}

	r.Name = lsn.Name
	r.DriverStatus = State(lsn.Status.State)

	return r
}
