package disk

import (
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/member/types"
)

// Manager manage all disks in cluster
// The operation here needs to ensure thread safety
type Manager interface {
	// GetNodeDisks list all disk located on the node
	GetNodeDisks(node string) ([]types.Disk, error)
}
