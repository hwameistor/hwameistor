package types

import "path"

const (
	LocalDiskPoolPrefix = "LocalDisk_Pool"
	SysDevicePathPrefix = "/dev"
)

// GetLocalDiskPoolName return LocalDisk_PoolHDD, LocalDisk_PoolSSD, LocalDisk_PoolNVMe
func GetLocalDiskPoolName(devType DevType) string {
	return LocalDiskPoolPrefix + devType
}

// GetLocalDiskPoolPath return /dev/LocalDisk_PoolHDD
func GetLocalDiskPoolPath(devType DevType) string {
	return path.Join(SysDevicePathPrefix, GetLocalDiskPoolName(devType))
}
