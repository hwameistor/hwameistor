package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/labels"
)

type LocalDiskMetricsCollector struct {
	dataCache *metricsCache

	claimStateMetricsDesc *prometheus.Desc
}

func newCollectorForLocalDisk(dataCache *metricsCache) prometheus.Collector {
	return &LocalDiskMetricsCollector{
		dataCache: dataCache,
		claimStateMetricsDesc: prometheus.NewDesc(
			"hwameistor_localdisk_claimstate",
			"The state summary of the localdisk.",
			[]string{"nodeName", "type", "claimstate"},
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
	stateCount := map[string]map[string]map[string]int64{}

	for _, disk := range disks {
		if _, exists := stateCount[disk.Spec.NodeName]; !exists {
			stateCount[disk.Spec.NodeName] = map[string]map[string]int64{}
		}
		if _, exists := stateCount[disk.Spec.NodeName][disk.Spec.DiskAttributes.Type]; !exists {
			stateCount[disk.Spec.NodeName][disk.Spec.DiskAttributes.Type] = map[string]int64{}
		}
		if _, exists := stateCount[disk.Spec.NodeName][disk.Spec.DiskAttributes.Type][string(disk.Status.State)]; !exists {
			stateCount[disk.Spec.NodeName][disk.Spec.DiskAttributes.Type][string(disk.Status.State)] = 0
		}

		stateCount[disk.Spec.NodeName][disk.Spec.DiskAttributes.Type][string(disk.Status.State)]++
	}

	for nodeName, nodeCount := range stateCount {
		for diskType, typeCount := range nodeCount {
			for claimState, count := range typeCount {
				ch <- prometheus.MustNewConstMetric(mc.claimStateMetricsDesc, prometheus.GaugeValue, float64(count), nodeName, diskType, claimState)
			}
		}
	}
}
