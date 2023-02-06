package exporter

import (
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

const (
	VolumeTypeNonHA       = "NonHA"
	VolumeTypeConvertible = "Convertible"
	VolumeTypeHA          = "HA"
)

func unifiedPoolName(poolName string) string {
	if poolName == apisv1alpha1.DiskClassNameHDD || poolName == apisv1alpha1.PoolNameForHDD {
		return apisv1alpha1.DiskClassNameHDD
	}
	if poolName == apisv1alpha1.DiskClassNameSSD || poolName == apisv1alpha1.PoolNameForSSD {
		return apisv1alpha1.DiskClassNameSSD
	}
	if poolName == apisv1alpha1.DiskClassNameNVMe || poolName == apisv1alpha1.PoolNameForNVMe {
		return apisv1alpha1.DiskClassNameNVMe
	}
	return "Unknown"
}
