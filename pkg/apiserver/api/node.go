package api

import (
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

// StorageNode
type StorageNode struct {
	LocalStorageNode apisv1alpha1.LocalStorageNode `json:"localStorageNode,omitempty"`
	LocalDiskNode    apisv1alpha1.LocalDiskNode    `json:"localDiskNode,omitempty"`

	K8sNodeState State `json:"k8SNodeState,omitempty"`
	//// driver status 驱动状态  运行中（Ready）,维护中（Maintain）, 离线（Offline）
	//DriverStatus State `json:"driverStatus,omitempty"`
	//// totalDiskCount 总磁盘数
	//TotalDiskCount int64 `json:"totalDiskCount,omitempty"`
	//// usedDiskCount 已绑定磁盘数
	//UsedDiskCount int64 `json:"usedDiskCount,omitempty"`
}

type LocalDisksItemsList struct {
	// localDisks 节点磁盘列表
	LocalDisks []*LocalDiskInfo `json:"items"`
}

// LocalDiskListByNode
type LocalDiskListByNode struct {
	// nodeName 节点名称
	NodeName string `json:"nodeName,omitempty"`
	// localDisks 节点磁盘列表
	LocalDisks []*LocalDiskInfo `json:"items"`
	// page 信息
	Page *Pagination `json:"pagination,omitempty"`
}

// StorageNodesItemsList
type StorageNodesItemsList struct {
	// localDisks 节点磁盘列表
	StorageNodes []*StorageNode `json:"items"`
}

// StorageNodeList
type StorageNodeList struct {
	// StorageNodes
	StorageNodes []*StorageNode `json:"items"`
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
	// TargetNodes
	TargetNodes []string `json:"targetNodes,omitempty"`
}

type NodeReserveReqBody struct {
	NodeName string `json:"nodeName,omitempty"`
	DiskName string `json:"diskName,omitempty"`
}

type NodeRemoveReserveReqBody struct {
	NodeName string `json:"nodeName,omitempty"`
	DiskName string `json:"diskName,omitempty"`
}
