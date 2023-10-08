package types

import (
	"path"
	"strings"
)

const (
	LocalDiskPoolPrefix  = "LocalDisk_Pool"
	HwameiStorDeviceRoot = "/etc/hwameistor"
	SysDeviceRoot        = "/dev"

	// sub path store sub resources under LocalDiskPool
	diskSubPath   = "disk"
	volumeSubPath = "volume"
)

var (
	DefaultDevTypes = []DevType{DevTypeHDD, DevTypeSSD, DevTypeNVMe}
)

// GetLocalDiskPoolName return LocalDisk_PoolHDD, LocalDisk_PoolSSD, LocalDisk_PoolNVMe
func GetLocalDiskPoolName(devType DevType) string {
	return LocalDiskPoolPrefix + devType
}

// GetLocalDiskPoolPath return /etc/hwameistor/LocalDisk_PoolHDD
func GetLocalDiskPoolPath(devType DevType) string {
	return path.Join(HwameiStorDeviceRoot, GetLocalDiskPoolName(devType))
}

func GetPoolDiskPath(devType DevType) string {
	return path.Join(GetLocalDiskPoolPath(devType), diskSubPath)
}

func GetPoolVolumePath(devType DevType) string {
	return path.Join(GetLocalDiskPoolPath(devType), volumeSubPath)
}

func ComposePoolDevicePath(poolName, devName string) string {
	return path.Join(path.Join(HwameiStorDeviceRoot, poolName, diskSubPath), devName)
}

func ComposePoolVolumePath(poolName, volumeName string) string {
	return path.Join(path.Join(HwameiStorDeviceRoot, poolName, volumeSubPath), volumeName)
}

func GetLocalDiskPoolPathFromVolume(volumePath string) string {
	return strings.Split(volumePath, volumeSubPath)[0]
}

func GetDefaultDiskPoolPath() (dps []string) {
	for _, poolClass := range DefaultDevTypes {
		dps = append(dps, GetLocalDiskPoolPath(poolClass))
	}
	return
}
