package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/labels"
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
	log.Debug("Collecting metrics for LocalDisk ...")
	disks, err := mc.dataCache.ldInformer.Lister().List(labels.NewSelector())
	if err != nil || len(disks) == 0 {
		log.WithError(err).Debug("Not found LocalDisk")
		return
	}

	// nodename.type.state = count
	statusCount := map[string]map[string]map[string]int64{}

	for _, disk := range disks {
		if _, exists := statusCount[disk.Spec.NodeName]; !exists {
			statusCount[disk.Spec.NodeName] = map[string]map[string]int64{}
		}
		if _, exists := statusCount[disk.Spec.NodeName][disk.Spec.DiskAttributes.Type]; !exists {
			statusCount[disk.Spec.NodeName][disk.Spec.DiskAttributes.Type] = map[string]int64{}
		}
		if _, exists := statusCount[disk.Spec.NodeName][disk.Spec.DiskAttributes.Type][string(disk.Status.State)]; !exists {
			statusCount[disk.Spec.NodeName][disk.Spec.DiskAttributes.Type][string(disk.Status.State)] = 0
		}

		statusCount[disk.Spec.NodeName][disk.Spec.DiskAttributes.Type][string(disk.Status.State)]++
	}

	for nodeName, nodeCount := range statusCount {
		for diskType, typeCount := range nodeCount {
			for status, count := range typeCount {
				ch <- prometheus.MustNewConstMetric(mc.statusMetricsDesc, prometheus.GaugeValue, float64(count), nodeName, diskType, status)
			}
		}
	}
}
