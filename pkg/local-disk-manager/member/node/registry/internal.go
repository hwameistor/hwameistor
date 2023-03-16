package registry

import (
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/member/types"
	"sync"
)

type localRegistry struct {
	// disks storage node disks managed by LocalDiskManager
	disks sync.Map

	// volumes storage node volumes managed by LocalDiskManager
	volumes sync.Map
}

func New() Manager {
	return &localRegistry{}
}

// DiscoveryResources discovery disks and volumes
func (r *localRegistry) DiscoveryResources() {
	r.discoveryDisks()
	r.discoveryVolumes()
}

// ListDisks list all registered disks
func (r *localRegistry) ListDisks() []types.Disk {
	return nil
}

func (r *localRegistry) ListDisksByType(devType types.DevType) []types.Disk {
	return nil
}

func (r *localRegistry) GetDiskByPath(devPath string) types.Disk {
	return types.Disk{}
}

// ListVolumes list all registered volumes
func (r *localRegistry) ListVolumes() []types.Volume {
	return nil
}

// ListVolumesByType list all registered volumes
func (r *localRegistry) ListVolumesByType(devType types.DevType) []types.Volume {
	return nil
}

func (r *localRegistry) GetVolumeByName() types.Volume {
	return types.Volume{}
}

func (r *localRegistry) discoveryDisks() {}

func (r *localRegistry) discoveryVolumes() {}
