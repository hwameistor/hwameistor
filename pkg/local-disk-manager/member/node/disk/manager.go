package disk

import (
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/member/types"
)

// Manager manage all disks in cluster
// The operation here needs to ensure thread safety
type Manager interface {
	// GetClusterDisks list all disks by node
	GetClusterDisks() (map[string][]*types.Disk, error)

	// GetNodeDisks list all disk located on the node
	GetNodeDisks(node string) ([]*types.Disk, error)

	// ClaimDisk UpdateDiskStatus mark disk to TobeMount/Free/InUse... status
	ClaimDisk(name string) error

	// FilterFreeDisks filter matchable free disks
	FilterFreeDisks([]types.Disk) (bool, error)

	// ReserveDiskForVolume reserve a disk for the volume
	ReserveDiskForVolume(disk types.Disk, pvc string) error

	// UnReserveDiskForPVC update related disk to release status
	UnReserveDiskForPVC(pvc string) error

	// ReleaseDisk setup disk to release status
	ReleaseDisk(disk string) error

	// GetReservedDiskByPVC get disk reserved by the volume
	GetReservedDiskByPVC(pvc string) (*types.Disk, error)

	/* Bellow is New Fund */

}
