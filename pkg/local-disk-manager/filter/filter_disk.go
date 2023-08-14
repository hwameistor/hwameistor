package filter

import (
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"strings"

	v1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils/sys"
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

func (ld *LocalDiskFilter) OwnerMatch(owner string) *LocalDiskFilter {
	if (ld.localDisk.Spec.Owner == "" && owner == "local-storage") /* Only for local-storage */ ||
		ld.localDisk.Spec.Owner == owner {
		ld.setResult(TRUE)
	} else {
		log.Infof("disk %s owner(%s) mismatch, owner in localdiskclaim is %s", ld.localDisk.Name, ld.localDisk.Spec.Owner, owner)
		ld.setResult(FALSE)
	}

	return ld
}

// HasBoundWith indicates disk has already bound with the claim
// https://github.com/hwameistor/hwameistor/issues/315
func (ld *LocalDiskFilter) HasBoundWith(claimUID types.UID) bool {
	// since the ldc will be deleted after consumed, so the ldc(s) instance with the same name can be applied multiple times.
	// so, we need to filter more when ClaimRef has already exist
	if ld.localDisk.Spec.ClaimRef != nil {
		if ld.localDisk.Spec.ClaimRef.UID == claimUID {
			return true
		}
	}

	return false
}

func (ld *LocalDiskFilter) IsNameFormatMatch() *LocalDiskFilter {
	// since v0.12.0, we use localdisk- as the localdisk prefix, so those old localdisk won't be matched
	if strings.HasPrefix(ld.localDisk.Name, v1alpha1.LocalDiskObjectPrefix) {
		ld.setResult(TRUE)
	} else {
		ld.setResult(FALSE)
	}

	return ld
}

func (ld *LocalDiskFilter) GetTotalResult() bool {
	return ld.Result == TRUE
}

func (ld *LocalDiskFilter) setResult(result Bool) {
	ld.Result &= result
}
