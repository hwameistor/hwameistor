package volumemanager

import "context"

const (
	TopologyNodeKey = "topology.disk.hwameistor.io/node"
)

// Volume
type Volume struct {
	// Name
	Name string `json:"name"`

	// Ready
	Ready bool `json:"ready"`

	// Exist
	Exist bool `json:"exist"`

	// Capacity
	Capacity int64 `json:"capacity"`

	// VolumeContext
	VolumeContext map[string]string

	// AttachNode
	AttachNode string `json:"attachNode"`
}

type VolumeManager interface {
	// CreateVolume when volume is not exist
	CreateVolume(name string, volumeRequest interface{}) (*Volume, error)

	// UpdateVolume
	UpdateVolume(name string, volumeRequest interface{}) (*Volume, error)

	// NodePublishVolume
	NodePublishVolume(ctx context.Context, volumeRequest interface{}) error

	// NodeUnpublishVolume
	NodeUnpublishVolume(ctx context.Context, name, targetPath string) error

	// DeleteVolume
	DeleteVolume(ctx context.Context, name string) error

	// GetVolumeInfo
	GetVolumeInfo(name string) (*Volume, error)

	// GetVolumeCapacities
	GetVolumeCapacities() interface{}

	// VolumeIsReady
	VolumeIsReady(name string) (bool, error)

	// VolumeIsExist
	VolumeIsExist(name string) (bool, error)
}
