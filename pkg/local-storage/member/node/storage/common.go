package storage

import apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"

func mergeRegistryDiskMap(localDiskMap ...map[string]*apisv1alpha1.LocalDevice) map[string]*apisv1alpha1.LocalDevice {
	newLocalDiskMap := map[string]*apisv1alpha1.LocalDevice{}
	for _, m := range localDiskMap {
		for k, v := range m {
			newLocalDiskMap[k] = v
		}
	}
	return newLocalDiskMap
}
