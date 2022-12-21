package api

import (
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

// LocalDisk is disk struct
type LocalDisk struct {
	// e.g. /dev/sdb 磁盘路径
	DevPath string `json:"devPath,omitempty"`

	// Supported: HDD, SSD, NVMe, RAM 磁盘类型
	Class string `json:"type,omitempty"`

	// HasRAID 是否Raid
	HasRAID bool `json:"hasRaid,omitempty"`

	// TotalCapacityBytes 总容量
	TotalCapacityBytes int64 `json:"totalCapacityBytes,omitempty"`

	// AvailableCapacityBytes 可用容量
	AvailableCapacityBytes int64 `json:"availableCapacityBytes,omitempty"`

	// Possible state: Claimed, UnClaimed, Inuse, Released, Reserved 状态
	State State `json:"state,omitempty"`

	// LocalStoragePooLName 存储池名称
	LocalStoragePooLName string `json:"localStoragePooLName,omitempty"`
}

type ReservedStatus struct {
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

func ToLocalDiskResource(ld apisv1alpha1.LocalDisk) *LocalDisk {
	r := &LocalDisk{}

	r.DevPath = ld.Spec.DevicePath
	r.State = State(ld.Status.State)
	r.HasRAID = ld.Spec.HasRAID
	r.Class = ld.Spec.DiskAttributes.Type
	r.TotalCapacityBytes = ld.Spec.Capacity
	// todo r.LocalStoragePooLName = ld.Spec.DiskAttributes.Type

	return r
}
