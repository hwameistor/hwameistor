package exporter

import "github.com/prometheus/client_golang/prometheus"

type LocalVolumeConvertMetricsCollector struct {
	dataCache *metricsCache
}

func newCollectorForLocalVolumeConvert(dataCache *metricsCache) prometheus.Collector {
	return &LocalVolumeConvertMetricsCollector{
		dataCache: dataCache,
	}

}

func (mc *LocalVolumeConvertMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(mc, ch)
}

func (mc *LocalVolumeConvertMetricsCollector) Collect(ch chan<- prometheus.Metric) {
}
