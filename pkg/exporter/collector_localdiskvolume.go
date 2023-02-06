package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
)

type LocalDiskVolumeMetricsCollector struct {
	dataCache *metricsCache

	statusMetricsDesc *prometheus.Desc
}

func newCollectorForLocalDiskVolume(dataCache *metricsCache) prometheus.Collector {
	return &LocalDiskVolumeMetricsCollector{
		dataCache: dataCache,
		statusMetricsDesc: prometheus.NewDesc(
			"hwameistor_localdiskvolume_status",
			"The status summary of the localdiskvolume.",
			[]string{"nodeName", "type", "status"},
			nil,
		),
	}
}

func (mc *LocalDiskVolumeMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(mc, ch)
}

func (mc *LocalDiskVolumeMetricsCollector) Collect(ch chan<- prometheus.Metric) {
}
