package exporter

import "github.com/prometheus/client_golang/prometheus"

type LocalVolumeMigrateMetricsCollector struct {
	dataCache *metricsCache
}

func newCollectorForLocalVolumeMigrate(dataCache *metricsCache) prometheus.Collector {
	return &LocalVolumeMigrateMetricsCollector{
		dataCache: dataCache,
	}

}

func (mc *LocalVolumeMigrateMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(mc, ch)
}

func (mc *LocalVolumeMigrateMetricsCollector) Collect(ch chan<- prometheus.Metric) {
}
