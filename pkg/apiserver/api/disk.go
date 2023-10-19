package api

import (
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

// LocalDiskInfo is disk struct
type LocalDiskInfo struct {
	apisv1alpha1.LocalDisk

	// diskPathShort 磁盘路径简写
	DiskPathShort string `json:"diskPathShort,omitempty"`

	// TotalCapacityBytes 总容量
	TotalCapacityBytes int64 `json:"totalCapacityBytes,omitempty"`

	// AvailableCapacityBytes 可用容量
	AvailableCapacityBytes int64 `json:"availableCapacityBytes,omitempty"`

	//// Possible state: Claimed, UnClaimed, Inuse, Released, Reserved 状态
	//State State `json:"state,omitempty"`

	// LocalStoragePooLName 存储池名称
	LocalStoragePooLName string `json:"localStoragePooLName,omitempty"`
}

type DiskReqBody struct {
	Reserve bool `json:"reserve,omitempty"`
}

type DiskReservedRspBody struct {
	DiskReservedRsp DiskReservedRsp `json:"data,omitempty"`
}

type DiskReservedRsp struct {
	DiskName    string `json:"diskName,omitempty"`
	ReservedRsp State  `json:"reservedRsp,omitempty"`
}

type DiskRemoveReservedRsp struct {
	DiskName          string `json:"diskName,omitempty"`
	RemoveReservedRsp State  `json:"removeReservedRsp,omitempty"`
}

type DiskRemoveReservedRspBody struct {
	DiskRemoveReservedRsp DiskRemoveReservedRsp `json:"data,omitempty"`
}

type DiskOwnerReqBody struct {
	//[ local-storage]  [local-disk-manager] [system]
	Owner string `json:"owner,omitempty"`
}

type DiskOwnerRsp struct {
	DiskName string `json:"diskName,omitempty"`
	Owner    string `json:"owner,omitempty"`
}

type DiskOwnerRspBody struct {
	DiskOwnerRsp DiskOwnerRsp `json:"data,omitempty"`
}
