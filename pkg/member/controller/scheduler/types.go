package scheduler

import (
	localstoragev1alpha1 "github.com/hwameistor/local-storage/pkg/apis/localstorage/v1alpha1"
)

type storageCollection struct {
	kinds map[string]storageKindCollection
}

type storageKindCollection struct {
	// for each kind (LVM, DISK, RAM),
	// collection of capacity of each node, nodeName -> capacity
	pools map[string]storagePoolCollection
}

type storagePoolCollection struct {
	// for each class(HDD, NVMe, SSD, RAM),
	// collection of capacity of each node, nodeName -> capacity
	capacities map[string]int64
	// collection of volume numbers of each node
	volumeCount map[string]int64
}

func newStorageCollection() *storageCollection {
	collection := &storageCollection{kinds: map[string]storageKindCollection{}}
	kinds := []string{localstoragev1alpha1.VolumeKindLVM, localstoragev1alpha1.VolumeKindDisk, localstoragev1alpha1.VolumeKindRAM}
	poolNames := []string{localstoragev1alpha1.PoolNameForHDD, localstoragev1alpha1.PoolNameForSSD, localstoragev1alpha1.PoolNameForNVMe, localstoragev1alpha1.PoolNameForRAM}
	for _, kind := range kinds {
		collection.kinds[kind] = storageKindCollection{pools: map[string]storagePoolCollection{}}
		for _, poolName := range poolNames {
			collection.kinds[kind].pools[poolName] = storagePoolCollection{
				capacities:  map[string]int64{},
				volumeCount: map[string]int64{},
			}
		}
	}

	return collection
}
