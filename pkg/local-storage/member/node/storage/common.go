package storage

import apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/local-storage/v1alpha1"

func mergeRegistryDiskMap(localDiskMap ...map[string]*apisv1alpha1.LocalDisk) map[string]*apisv1alpha1.LocalDisk {
	newLocalDiskMap := map[string]*apisv1alpha1.LocalDisk{}
	for _, m := range localDiskMap {
		for k, v := range m {
			newLocalDiskMap[k] = v
		}
	}
	return newLocalDiskMap
}
