package api

import (
	"time"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

type StorageNodePool struct {
	apisv1alpha1.LocalPool
	// NodesName Pool所在节点
	NodeName string `json:"nodeName"`
}

type StoragePool struct {
	StorageNodePools []StorageNodePool `json:"items"`

	// Supported pool name: HDD_POOL, SSD_POOL, NVMe_POOL 存储池名称
	PoolName string `json:"poolName,omitempty"`

	// TotalCapacityBytes 存储池对应存储总容量
	TotalCapacityBytes int64 `json:"totalCapacityBytes"`

	// AllocatedCapacityBytes 存储池已经分配存储容量
	AllocatedCapacityBytes int64 `json:"allocatedCapacityBytes"`

	// NodesNames Pool所在节点列表
	NodeNames []string `json:"nodeNames"`

	// createTime 创建时间
	CreateTime time.Time `json:"createTime,omitempty"`
}

type StoragePoolList struct {
	// storagePools
	StoragePools []*StoragePool `json:"items"`
	// page 信息
	Page *Pagination `json:"pagination,omitempty"`
}

type NodeDiskListByPool struct {
	// PoolName 存储池名称
	PoolName string `json:"poolName,omitempty"`
	// nodeName 节点名称
	NodeName string `json:"nodeName,omitempty"`
	// localDisks 节点磁盘列表
	LocalDisks []*LocalDiskInfo `json:"items,omitempty"`
	// page 信息
	Page *Pagination `json:"pagination,omitempty"`
}

type StorageNodeListByPool struct {
	// StoragePoolName 存储池名称
	StoragePoolName string `json:"storagePoolName,omitempty"`
	// StorageNodes
	StorageNodes []*StorageNode `json:"items,omitempty"`
	// page 信息
	Page *Pagination `json:"pagination,omitempty"`
}

type StoragePoolNodesCollection struct {
	// 纳管节点列表
	ManagedNodeNames []string `json:"managedNodeNames"`
	// 存储池信息
	StoragePool StoragePool `json:"storagePool"`
}

type StoragePoolExpansionReqBody struct {
	NodeName string `json:"nodeName,omitempty"`
	// HDD/SSD/NVME
	DiskType string `json:"diskType,omitempty"`
	// local-storage/local-disk-manager
	Owner string `json:"owner,omitempty"`
}

type StoragePoolExpansionRspBody struct {
	Success bool `json:"success,omitempty"`
}
