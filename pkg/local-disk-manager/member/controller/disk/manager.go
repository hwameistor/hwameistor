package disk

import (
	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/member/types"
)

// Manager manage all disks in cluster
// The operation here needs to ensure thread safety
type Manager interface {
	// GetNodeDisks list all disk located on the node
	GetNodeDisks(node string) ([]types.Disk, error)

	GetNodeAvailableDisks(node string) ([]types.Disk, error)

	MarkNodeDiskInuse(node string, disk *types.Disk) error

	MarkNodeDiskAvailable(node string, disk *types.Disk) error

	NodeIsReady(node string) (bool, error)

	ListLocalDiskByNodeDevicePath(nodeName, devicePath string) ([]v1alpha1.LocalDisk, error)
}
