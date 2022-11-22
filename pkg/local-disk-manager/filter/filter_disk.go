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
	localDisk *v1alpha1.LocalDisk
	Result    Bool
}

func NewLocalDiskFilter(ld *v1alpha1.LocalDisk) LocalDiskFilter {
	return LocalDiskFilter{
		localDisk: ld,
		Result:    TRUE,
	}
}

func (ld *LocalDiskFilter) Init() *LocalDiskFilter {
	ld.Result = TRUE
	return ld
}

func (ld *LocalDiskFilter) Available() *LocalDiskFilter {
	if ld.localDisk.Status.State == v1alpha1.LocalDiskAvailable {
		ld.setResult(TRUE)
	} else {
		ld.setResult(FALSE)
	}

	return ld
}

func (ld *LocalDiskFilter) HasNotReserved() *LocalDiskFilter {
	if !ld.localDisk.Spec.Reserved {
		ld.setResult(TRUE)
	} else {
		ld.setResult(FALSE)
	}

	return ld
}

func (ld *LocalDiskFilter) NodeMatch(wantNode string) *LocalDiskFilter {
	if wantNode == ld.localDisk.Spec.NodeName {
		ld.setResult(TRUE)
	} else {
		ld.setResult(FALSE)
	}
	return ld
}

func (ld *LocalDiskFilter) Unique(diskRefs []*v1.ObjectReference) *LocalDiskFilter {
	for _, disk := range diskRefs {
		if disk.Name == ld.localDisk.Name {
			ld.setResult(FALSE)
			return ld
		}
	}

	ld.setResult(TRUE)
	return ld
}

func (ld *LocalDiskFilter) Capacity(cap int64) *LocalDiskFilter {
	if ld.localDisk.Spec.Capacity >= cap {
		ld.setResult(TRUE)
	} else {
		ld.setResult(FALSE)
	}

	return ld
}

func (ld *LocalDiskFilter) DiskType(diskType string) *LocalDiskFilter {
	if ld.localDisk.Spec.DiskAttributes.Type == diskType {
		ld.setResult(TRUE)
	} else {
		ld.setResult(FALSE)
	}

	return ld
}

func (ld *LocalDiskFilter) DevType() *LocalDiskFilter {
	if ld.localDisk.Spec.DiskAttributes.DevType == sys.BlockDeviceTypeDisk {
		ld.setResult(TRUE)
	} else {
		ld.setResult(FALSE)
	}

	return ld
}

func (ld *LocalDiskFilter) NoPartition() *LocalDiskFilter {
	if len(ld.localDisk.Spec.PartitionInfo) > 0 {
		ld.setResult(FALSE)
	} else {
		ld.setResult(TRUE)
	}

	return ld
}

// HasBoundWith indicates disk has already bound with the claim
// https://github.com/hwameistor/hwameistor/issues/315
func (ld *LocalDiskFilter) HasBoundWith(claimName string) bool {
	if ld.localDisk.Spec.ClaimRef != nil {
		if ld.localDisk.Spec.ClaimRef.Name == claimName {
			return true
		}
	}

	return false
}

func (ld *LocalDiskFilter) GetTotalResult() bool {
	return ld.Result == TRUE
}

func (ld *LocalDiskFilter) setResult(result Bool) {
	ld.Result &= result
}
