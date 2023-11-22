package filter

import (
	"strings"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils/sys"
)

type Bool int

const (
	FALSE Bool = 0
	TRUE  Bool = 1
)

type LocalDiskFilter struct {
	localDisk     *v1alpha1.LocalDisk
	Result        Bool
	FailedMessage map[string]bool
}

func NewLocalDiskFilter(ld *v1alpha1.LocalDisk) LocalDiskFilter {
	return LocalDiskFilter{
		localDisk:     ld,
		Result:        TRUE,
		FailedMessage: make(map[string]bool),
	}
}

func (ld *LocalDiskFilter) Init() *LocalDiskFilter {
	ld.Result = TRUE
	ld.FailedMessage = make(map[string]bool)
	return ld
}

func (ld *LocalDiskFilter) Available() *LocalDiskFilter {
	if ld.localDisk.Status.State == v1alpha1.LocalDiskAvailable {
		ld.setResult(TRUE)
	} else {
		log.Debugf("disk %s state %s mismatch", ld.localDisk.Name, ld.localDisk.Status.State)
		ld.setResult(FALSE)
	}

	return ld
}

func (ld *LocalDiskFilter) HasNotReserved() *LocalDiskFilter {
	if !ld.localDisk.Spec.Reserved {
		ld.setResult(TRUE)
	} else {
		// only record reserved events when disk is Available
		if ld.localDisk.Status.State == v1alpha1.LocalDiskAvailable {
			ld.FailedMessage[v1alpha1.LocalDiskAssignFailReasonHasReserved] = true
		}
		log.Debugf("disk %s state is reserved", ld.localDisk.Name)
		ld.setResult(FALSE)
	}

	return ld
}

func (ld *LocalDiskFilter) NodeMatch(wantNode string) *LocalDiskFilter {
	if wantNode == ld.localDisk.Spec.NodeName {
		ld.setResult(TRUE)
	} else {
		log.Debugf("disk %s nodeName(%s) missmatch, wantNode %s", ld.localDisk.Name, ld.localDisk.Spec.NodeName, wantNode)
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
		// only record reserved events when disk is Available
		if ld.localDisk.Status.State == v1alpha1.LocalDiskAvailable {
			ld.FailedMessage[v1alpha1.LocalDiskAssignFailReasonInsufficientCapacity] = true
		}
		log.Debugf("disk %s capacity(%d) missmatch, want capacity %d", ld.localDisk.Name, ld.localDisk.Spec.Capacity, cap)
		ld.setResult(FALSE)
	}

	return ld
}

func (ld *LocalDiskFilter) DiskType(diskType string) *LocalDiskFilter {
	// result is true when diskType is empty string
	if diskType == "" {
		ld.setResult(TRUE)
		return ld
	}
	if ld.localDisk.Spec.DiskAttributes.Type == diskType {
		ld.setResult(TRUE)
	} else {
		// only record reserved events when disk is Available
		if ld.localDisk.Status.State == v1alpha1.LocalDiskAvailable {
			ld.FailedMessage[v1alpha1.LocalDiskAssignFailReasonDiskTypeUnMatch] = true
		}
		log.Debugf("disk %s type(%s) missmatch, wantType %s", ld.localDisk.Name, ld.localDisk.Spec.DiskAttributes.Type, diskType)
		ld.setResult(FALSE)
	}

	return ld
}

func (ld *LocalDiskFilter) DevType() *LocalDiskFilter {
	if ld.localDisk.Spec.DiskAttributes.DevType == sys.BlockDeviceTypeDisk {
		ld.setResult(TRUE)
	} else {
		// only record reserved events when disk is Available
		if ld.localDisk.Status.State == v1alpha1.LocalDiskAvailable {
			ld.FailedMessage[v1alpha1.LocalDiskAssignFailReasonDiskIsNotBlockDevice] = true
		}
		log.Debugf("disk %s devType(%s) missmatch, wantType %s", ld.localDisk.Name, ld.localDisk.Spec.DiskAttributes.DevType, sys.BlockDeviceTypeDisk)
		ld.setResult(FALSE)
	}

	return ld
}

func (ld *LocalDiskFilter) NoPartition() *LocalDiskFilter {
	if len(ld.localDisk.Spec.PartitionInfo) > 0 {
		log.Debugf("disk %s is already partitioned(%d)", ld.localDisk.Name, len(ld.localDisk.Spec.PartitionInfo))
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
		// only record reserved events when disk is Available
		if ld.localDisk.Status.State == v1alpha1.LocalDiskAvailable {
			ld.FailedMessage[v1alpha1.LocalDiskAssignFailReasonOwnerUnMatch] = true
		}
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
		// only record reserved events when disk is Available
		if ld.localDisk.Status.State == v1alpha1.LocalDiskAvailable {
			ld.FailedMessage[v1alpha1.LocalDiskAssignFailReasonLocalDiskNameFormatUnMatch] = true
		}
		log.Debugf("disk %s has deprecated format of name, this disk can be safely removed", ld.localDisk.Name)
		ld.setResult(FALSE)
	}

	return ld
}

func (ld *LocalDiskFilter) LdNameMatch(ldNames []string) *LocalDiskFilter {
	if len(ldNames) == 0 {
		ld.setResult(TRUE)
		return ld
	}

	found := false

	for _, ldName := range ldNames {
		if ldName == ld.localDisk.Name {
			found = true
			break
		}
	}

	if found {
		ld.setResult(TRUE)
	} else {
		log.Debugf("disk %s is not specified in diskClaim", ld.localDisk.Name)
		ld.setResult(FALSE)
	}
	return ld
}

func (ld *LocalDiskFilter) DevPathMatch(devPaths []string) *LocalDiskFilter {
	if len(devPaths) == 0 {
		ld.setResult(TRUE)
		return ld
	}

	found := false

	for _, devName := range devPaths {
		if devName == ld.localDisk.Spec.DevicePath {
			found = true
			break
		}
	}

	if found {
		ld.setResult(TRUE)
	} else {
		// only record reserved events when disk is Available
		if ld.localDisk.Status.State == v1alpha1.LocalDiskAvailable {
			ld.FailedMessage[v1alpha1.LocalDiskAssignFailReasonDevPathUnMatch] = true
		}
		log.Debugf("disk %s devPath %s is not specified in diskClaim", ld.localDisk.Name, ld.localDisk.Spec.DevicePath)
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
