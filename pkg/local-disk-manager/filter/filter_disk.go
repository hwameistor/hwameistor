package filter

import (
	v1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils/sys"
	v1 "k8s.io/api/core/v1"
)

type Bool int

const (
	FALSE Bool = 0
	TRUE  Bool = 1
)

type LocalDiskFilter struct {
	LocalDisk v1alpha1.LocalDisk
	Result    Bool
}

// NewLocalDiskFilter
func NewLocalDiskFilter(ld v1alpha1.LocalDisk) LocalDiskFilter {
	return LocalDiskFilter{
		LocalDisk: ld,
		Result:    TRUE,
	}
}

// Init
func (ld *LocalDiskFilter) Init() *LocalDiskFilter {
	ld.Result = TRUE
	return ld
}

// Unclaimed
func (ld *LocalDiskFilter) Unclaimed() *LocalDiskFilter {
	if ld.LocalDisk.Status.State == v1alpha1.LocalDiskUnclaimed {
		ld.setResult(TRUE)
	} else {
		ld.setResult(FALSE)
	}

	return ld
}

// NodeMatch
func (ld *LocalDiskFilter) NodeMatch(wantNode string) *LocalDiskFilter {
	if wantNode == ld.LocalDisk.Spec.NodeName {
		ld.setResult(TRUE)
	} else {
		ld.setResult(FALSE)
	}
	return ld
}

// Unique
func (ld *LocalDiskFilter) Unique(diskRefs []*v1.ObjectReference) *LocalDiskFilter {
	for _, disk := range diskRefs {
		if disk.Name == ld.LocalDisk.Name {
			ld.setResult(FALSE)
			return ld
		}
	}

	ld.setResult(TRUE)
	return ld
}

// Capacity
func (ld *LocalDiskFilter) Capacity(cap int64) *LocalDiskFilter {
	if ld.LocalDisk.Spec.Capacity >= cap {
		ld.setResult(TRUE)
	} else {
		ld.setResult(FALSE)
	}

	return ld
}

// DiskType
func (ld *LocalDiskFilter) DiskType(diskType string) *LocalDiskFilter {
	if ld.LocalDisk.Spec.DiskAttributes.Type == diskType {
		ld.setResult(TRUE)
	} else {
		ld.setResult(FALSE)
	}

	return ld
}

// DevType
func (ld *LocalDiskFilter) DevType() *LocalDiskFilter {
	if ld.LocalDisk.Spec.DiskAttributes.DevType == sys.BlockDeviceTypeDisk {
		ld.setResult(TRUE)
	} else {
		ld.setResult(FALSE)
	}

	return ld
}

// NoPartition
func (ld *LocalDiskFilter) NoPartition() *LocalDiskFilter {
	if len(ld.LocalDisk.Spec.PartitionInfo) > 0 {
		ld.setResult(FALSE)
	} else {
		ld.setResult(TRUE)
	}

	return ld
}

// HasBoundWith indicates disk has already bound with the claim
// https://github.com/hwameistor/hwameistor/issues/315
func (ld *LocalDiskFilter) HasBoundWith(claimName string) bool {
	if ld.LocalDisk.Spec.ClaimRef != nil {
		if ld.LocalDisk.Spec.ClaimRef.Name == claimName {
			return true
		}
	}

	return false
}

// Capacity
func (ld *LocalDiskFilter) GetTotalResult() bool {
	return ld.Result == TRUE
}

// setResult
func (ld *LocalDiskFilter) setResult(result Bool) {
	ld.Result &= result
}
