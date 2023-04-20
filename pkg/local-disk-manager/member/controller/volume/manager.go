package volume

import (
	"context"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/member/types"
)

const (
	TopologyNodeKey = "topology.disk.hwameistor.io/node"
)

type Manager interface {
	// CreateVolume when volume is not exist
	CreateVolume(name string, volumeRequest interface{}) (*types.Volume, error)

	// UpdateVolume
	UpdateVolume(name string, volumeRequest interface{}) (*types.Volume, error)

	// NodePublishVolume
	NodePublishVolume(ctx context.Context, volumeRequest interface{}) error

	// NodeUnpublishVolume
	NodeUnpublishVolume(ctx context.Context, name, targetPath string) error

	// DeleteVolume
	DeleteVolume(ctx context.Context, name string) error

	// GetVolumeInfo
	GetVolumeInfo(name string) (*types.Volume, error)

	// GetVolumeCapacities
	GetVolumeCapacities() interface{}

	// VolumeIsReady
	VolumeIsReady(name string) (bool, error)

	// VolumeIsExist
	VolumeIsExist(name string) (bool, error)
}
