package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type CollectorManager struct {
}

func NewCollectorManager() *CollectorManager {
	return &CollectorManager{}
}

func (mc *CollectorManager) Register(stopCh <-chan struct{}) {

	cache := newCache()
	cache.run(stopCh)

	registry := prometheus.NewRegistry()

	registry.MustRegister(newCollectorForLocalStorageNode(cache))
	registry.MustRegister(newCollectorForLocalVolume(cache))
	registry.MustRegister(newCollectorForLocalVolumeReplica(cache))
	registry.MustRegister(newCollectorForLocalDisk(cache))
	registry.MustRegister(newCollectorForLocalDiskVolume(cache))
	registry.MustRegister(NewSMARTCollector())

}
