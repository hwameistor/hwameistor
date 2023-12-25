package api

import (
	k8sv1 "k8s.io/api/core/v1"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

type StorageNode struct {
	LocalStorageNode apisv1alpha1.LocalStorageNode `json:"localStorageNode,omitempty"`
	LocalDiskNode    apisv1alpha1.LocalDiskNode    `json:"localDiskNode,omitempty"`
	TotalDisk        int                           `json:"totalDisk,omitempty"`
	K8sNode          *k8sv1.Node
	K8sNodeState     State `json:"k8SNodeState,omitempty"`
}

type LocalDisksItemsList struct {
	// localDisks 节点磁盘列表
	LocalDisks []*LocalDiskInfo `json:"items"`
}

type LocalDiskListByNode struct {
	// nodeName 节点名称
	NodeName string `json:"nodeName,omitempty"`
	// diskPathShort 磁盘路径简写
	DiskPathShort string `json:"diskPathShort,omitempty"`
	// localDisks 节点磁盘列表
	LocalDisks []*LocalDiskInfo `json:"items"`
	// page 信息
	Page *Pagination `json:"pagination,omitempty"`
}

type LocalDiskList struct {
	// LocalDisks 集群磁盘列表
	LocalDisks []*apisv1alpha1.LocalDisk `json:"items"`
}

type LocalDiskNodeList struct {
	// LocalDiskNodes 集群磁盘组列表
	LocalDiskNodes []*apisv1alpha1.LocalDiskNode `json:"items"`
}

type StorageNodesItemsList struct {
	// localDisks 节点磁盘列表
	StorageNodes []*StorageNode `json:"items"`
}

type StorageNodeList struct {
	// StorageNodes
	StorageNodes []*StorageNode `json:"items"`
	// page 信息
	Page *Pagination `json:"pagination,omitempty"`
}

type YamlData struct {
	// yaml data
	Data string `json:"data,omitempty"`
}

type TargetNodeList struct {
	// TargetNodes
	TargetNodes []string `json:"targetNodes,omitempty"`
}

type NodeUpdateReqBody struct {
	Enable *bool `json:"enable,omitempty"`
}

type NodeUpdateRspBody struct {
	Success bool `json:"success"`
}
