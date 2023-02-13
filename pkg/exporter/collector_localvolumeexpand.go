package exporter

import "github.com/prometheus/client_golang/prometheus"

type LocalVolumeExpandMetricsCollector struct {
	dataCache *metricsCache
}

func newCollectorForLocalVolumeExpand(dataCache *metricsCache) prometheus.Collector {
	return &LocalVolumeExpandMetricsCollector{
		dataCache: dataCache,
	}

}

func (mc *LocalVolumeExpandMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(mc, ch)
}

func (mc *LocalVolumeExpandMetricsCollector) Collect(ch chan<- prometheus.Metric) {
}
