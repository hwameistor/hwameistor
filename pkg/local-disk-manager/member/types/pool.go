package types

import "path"

const (
	LocalDiskPoolPrefix = "LocalDisk_Pool"
	SysDeviceRoot       = "/dev"

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

// GetLocalDiskPoolPath return /dev/LocalDisk_PoolHDD
func GetLocalDiskPoolPath(devType DevType) string {
	return path.Join(SysDeviceRoot, GetLocalDiskPoolName(devType))
}

func GetPoolDiskPath(devType DevType) string {
	return path.Join(GetLocalDiskPoolPath(devType), diskSubPath)
}

func GetPoolVolumePath(devType DevType) string {
	return path.Join(GetLocalDiskPoolPath(devType), volumeSubPath)
}

func ComposePoolDevicePath(poolName, devName string) string {
	return path.Join(path.Join(SysDeviceRoot, poolName, diskSubPath), devName)
}

func ComposePoolVolumePath(poolName, volumeName string) string {
	return path.Join(path.Join(SysDeviceRoot, poolName, volumeSubPath), volumeName)
}

func GetDefaultDiskPoolPath() (dps []string) {
	for _, poolClass := range DefaultDevTypes {
		dps = append(dps, GetLocalDiskPoolPath(poolClass))
	}
	return
}
