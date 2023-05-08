package volume

import (
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/member/types"
	"k8s.io/kubernetes/pkg/volume/util/hostutil"
	"os"
	"path"
)

// Manager responsible for creating, updating, and deleting volumes on nodes
type Manager interface {
	// CreateVolume create volume from device exist in pool
	CreateVolume(name string, pool string, device string) error

	// DeleteVolume delete volume from pool and release bound disk
	DeleteVolume(name string, pool string) error

	// GetVolume return info about this volume
	GetVolume(name string) *types.Volume
}

type volume struct {
	hu hostutil.HostUtils
}

// CreateVolume create volume symlink for bound disk
func (v *volume) CreateVolume(volume string, pool string, device string) error {
	devicePath := path.Join("..", "disk", device)
	volumePath := types.ComposePoolVolumePath(pool, volume)
	exist, err := v.hu.PathExists(volumePath)
	if err != nil || exist {
		return err
	}
	return os.Symlink(devicePath, volumePath)
}

func (v *volume) DeleteVolume(volume string, pool string) error {
	volumePath := types.ComposePoolVolumePath(pool, volume)
	exist, err := v.hu.PathExists(volumePath)
	if err != nil || !exist {
		return err
	}
	return os.Remove(volumePath)
}

func (v *volume) GetVolume(name string) *types.Volume {
	return nil
}

func New() Manager {
	return &volume{
		hu: hostutil.NewHostUtil(),
	}
}
