package types

import "path"

const (
	LocalDiskPoolPrefix = "LocalDisk_Pool"
	SysDevicePathPrefix = "/dev"
)

var (
	DefaultPoolClasses = []DevType{DevTypeHDD, DevTypeSSD, DevTypeNVMe}
)

// GetLocalDiskPoolName return LocalDisk_PoolHDD, LocalDisk_PoolSSD, LocalDisk_PoolNVMe
func GetLocalDiskPoolName(devType DevType) string {
	return LocalDiskPoolPrefix + devType
}

// GetLocalDiskPoolPath return /dev/LocalDisk_PoolHDD
func GetLocalDiskPoolPath(devType DevType) string {
	return path.Join(SysDevicePathPrefix, GetLocalDiskPoolName(devType))
}

func ComposePoolDevicePath(poolName, devName string) string {
	return path.Join(path.Join(SysDevicePathPrefix, poolName), devName)
}

func GetDefaultDiskPoolPath() (dps []string) {
	for _, poolClass := range DefaultPoolClasses {
		dps = append(dps, GetLocalDiskPoolPath(poolClass))
	}
	return
}
