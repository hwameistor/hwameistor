package storage

import udsv1alpha1 "github.com/HwameiStor/local-storage/pkg/apis/uds/v1alpha1"

func mergeRegistryDiskMap(localDiskMap ...map[string]*udsv1alpha1.LocalDisk) map[string]*udsv1alpha1.LocalDisk {
	newLocalDiskMap := map[string]*udsv1alpha1.LocalDisk{}
	for _, m := range localDiskMap {
		for k, v := range m {
			newLocalDiskMap[k] = v
		}
	}
	return newLocalDiskMap
}
