package scheduler

import (
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

type storageCollection struct {
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
	collection := &storageCollection{pools: map[string]storagePoolCollection{}}
	poolNames := []string{apisv1alpha1.PoolNameForHDD, apisv1alpha1.PoolNameForSSD, apisv1alpha1.PoolNameForNVMe}
	for _, poolName := range poolNames {
		collection.pools[poolName] = storagePoolCollection{
			capacities:  map[string]int64{},
			volumeCount: map[string]int64{},
		}
	}

	return collection
}
