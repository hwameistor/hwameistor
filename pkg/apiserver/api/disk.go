package api

import apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"

// LocalDiskInfo is disk struct
type LocalDiskInfo struct {
	apisv1alpha1.LocalDisk

	// TotalCapacityBytes 总容量
	TotalCapacityBytes int64 `json:"totalCapacityBytes,omitempty"`

	// AvailableCapacityBytes 可用容量
	AvailableCapacityBytes int64 `json:"availableCapacityBytes,omitempty"`

	//// Possible state: Claimed, UnClaimed, Inuse, Released, Reserved 状态
	//State State `json:"state,omitempty"`

	// LocalStoragePooLName 存储池名称
	LocalStoragePooLName string `json:"localStoragePooLName,omitempty"`
}

type ReservedStatus struct {
}

// DiskReqBody
type DiskReqBody struct {
	Reserve bool `json:"reserve,omitempty"`
}

// DiskReservedRspBody
type DiskReservedRspBody struct {
	DiskReservedRsp DiskReservedRsp `json:"data,omitempty"`
}

// DiskReservedRsp
type DiskReservedRsp struct {
	// DiskName
	DiskName string `json:"diskName,omitempty"`
	// ReservedRsp
	ReservedRsp State `json:"reservedRsp,omitempty"`
}

// DiskRemoveReservedRsp
type DiskRemoveReservedRsp struct {
	// DiskName
	DiskName string `json:"diskName,omitempty"`
	// RemoveReservedRsp
	RemoveReservedRsp State `json:"removeReservedRsp,omitempty"`
}

// DiskRemoveReservedRspBody
type DiskRemoveReservedRspBody struct {
	DiskRemoveReservedRsp DiskRemoveReservedRsp `json:"data,omitempty"`
}
