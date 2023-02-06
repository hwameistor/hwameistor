package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
)

type LocalDiskMetricsCollector struct {
	dataCache *metricsCache

	statusMetricsDesc *prometheus.Desc
}

func newCollectorForLocalDisk(dataCache *metricsCache) prometheus.Collector {
	return &LocalDiskMetricsCollector{
		dataCache: dataCache,
		statusMetricsDesc: prometheus.NewDesc(
			"hwameistor_localdisk_status",
			"The status summary of the localdisk.",
			[]string{"nodeName", "type", "status"},
			nil,
		),
	}
}

func (mc *LocalDiskMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(mc, ch)
}

func (mc *LocalDiskMetricsCollector) Collect(ch chan<- prometheus.Metric) {
}
