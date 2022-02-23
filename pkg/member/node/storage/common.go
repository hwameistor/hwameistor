package storage

import localstoragev1alpha1 "github.com/hwameistor/local-storage/pkg/apis/localstorage/v1alpha1"

func mergeRegistryDiskMap(localDiskMap ...map[string]*localstoragev1alpha1.LocalDisk) map[string]*localstoragev1alpha1.LocalDisk {
	newLocalDiskMap := map[string]*localstoragev1alpha1.LocalDisk{}
	for _, m := range localDiskMap {
		for k, v := range m {
			newLocalDiskMap[k] = v
		}
	}
	return newLocalDiskMap
}
