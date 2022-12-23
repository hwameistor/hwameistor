package api

import (
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"time"
)

type StoragePool struct {
	apisv1alpha1.LocalPool

	// AllocatedCapacityBytes 存储池已经分配存储容量
	AllocatedCapacityBytes int64 `json:"allocatedCapacityBytes"`

	// NodesNames Pool所在节点列表
	NodeNames []string `json:"nodeNames"`

	// createTime 创建时间
	CreateTime time.Time `json:"createTime,omitempty"`
}

// StoragePoolList
type StoragePoolList struct {
	// storagePools
	StoragePools []*StoragePool `json:"storagePools"`
	// page 信息
	Page *Pagination `json:"page,omitempty"`
}

// NodeDiskListByPool
type NodeDiskListByPool struct {
	// PoolName 存储池名称
	PoolName string `json:"poolName,omitempty"`
	// nodeName 节点名称
	NodeName string `json:"nodeName,omitempty"`
	// localDisks 节点磁盘列表
	LocalDisks []*LocalDiskInfo `json:"localDisks,omitempty"`
	// page 信息
	Page *Pagination `json:"page,omitempty"`
}

// StorageNodeListByPool
type StorageNodeListByPool struct {
	// StoragePoolName 存储池名称
	StoragePoolName string `json:"storagePoolName,omitempty"`
	// StorageNodes
	StorageNodes []*StorageNode `json:"storageNodes,omitempty"`
	// page 信息
	Page *Pagination `json:"page,omitempty"`
}

type StoragePoolNodesCollection struct {
	// 纳管节点列表
	ManagedNodeNames []string `json:"managedNodeNames"`
	// 存储池信息
	StoragePool StoragePool `json:"storagePool"`
}
