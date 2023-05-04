package exporter

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type CollectorManager struct {
	cache *metricsCache
}

func NewCollectorManager() *CollectorManager {
	return &CollectorManager{}
}

func (mc *CollectorManager) Run(stopCh <-chan struct{}) {
	mc.cache = newCache()
	mc.cache.run(stopCh)

	newRegister := prometheus.NewRegistry()
	newRegister.MustRegister(newCollectorForLocalStorageNode(mc.cache))
	newRegister.MustRegister(newCollectorForLocalVolume(mc.cache))
	newRegister.MustRegister(newCollectorForLocalVolumeReplica(mc.cache))
	newRegister.MustRegister(newCollectorForLocalVolumeConvert(mc.cache))
	newRegister.MustRegister(newCollectorForLocalVolumeExpand(mc.cache))
	newRegister.MustRegister(newCollectorForLocalVolumeMigrate(mc.cache))
	newRegister.MustRegister(newCollectorForLocalDisk(mc.cache))
	newRegister.MustRegister(NewSMARTCollector())

	http.Handle("/metrics", promhttp.HandlerFor(newRegister, promhttp.HandlerOpts{}))
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		panic(err)
	}
}
