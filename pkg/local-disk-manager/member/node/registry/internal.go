package registry

import (
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/member/types"
	log "github.com/sirupsen/logrus"
	"k8s.io/kubernetes/pkg/volume/util/hostutil"
	"os"
	"path/filepath"
	"sync"
)

type localRegistry struct {
	// disks storage node disks managed by LocalDiskManager
	disks sync.Map

	// volumes storage node volumes managed by LocalDiskManager
	volumes sync.Map

	hu hostutil.HostUtils
}

func New() Manager {
	return &localRegistry{
		hu: hostutil.NewHostUtil(),
	}
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

func (r *localRegistry) discoveryDisks() {
	for _, poolClass := range types.DefaultPoolClasses {
		rootPath := types.GetLocalDiskPoolPath(poolClass)
		disks, err := discoveryDevices(rootPath)
		if err != nil {
			log.WithError(err).Errorf("Failed to discovery devices from %s", rootPath)
			os.Exit(1)
		}

		// store discovery disks
		r.disks.Store(poolClass, disks)
	}
}

func (r *localRegistry) discoveryVolumes() {}

var hu hostutil.HostUtils = hostutil.NewHostUtil()

func discoveryDevices(rootPath string) ([]string, error) {
	ok, err := hu.PathExists(rootPath)
	if err != nil || !ok {
		return nil, err
	}

	// walk the folder and discovery devices
	var discoveryDevices []string
	err = filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		actualPath, err := hu.EvalHostSymlinks(path)
		if err != nil {
			return err
		}
		ok, err := hu.PathIsDevice(actualPath)
		if err != nil {
			return err
		}
		if ok {
			log.Infof("Found disk %s exist in %s", info.Name(), rootPath)
			discoveryDevices = append(discoveryDevices, info.Name())
		} else {
			log.Debugf("Found %s but not a device, skip it", info.Name())
		}
		return nil
	})
	if err != nil {
		log.WithError(err).Error("Failed to discovery disks")
	}
	return discoveryDevices, err
}
