package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/labels"
)

type LocalDiskMetricsCollector struct {
	dataCache *metricsCache

	capacityMetricsDesc *prometheus.Desc
}

func newCollectorForLocalDisk(dataCache *metricsCache) prometheus.Collector {
	return &LocalDiskMetricsCollector{
		dataCache: dataCache,
		capacityMetricsDesc: prometheus.NewDesc(
			"hwameistor_localdisk_capacity",
			"The capacity of the localdisk.",
			[]string{"nodeName", "type", "devPath", "reserved", "owner", "status"},
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

	for _, disk := range disks {
		reserved := "false"
		if disk.Spec.Reserved {
			reserved = "true"
		}
		owner := disk.Spec.Owner
		if len(owner) == 0 {
			owner = "Unknown"
		}
		ch <- prometheus.MustNewConstMetric(
			mc.capacityMetricsDesc, prometheus.GaugeValue,
			float64(disk.Spec.Capacity),
			disk.Spec.NodeName, disk.Spec.DiskAttributes.Type, disk.Spec.DevicePath, reserved, owner, string(disk.Status.State),
		)
	}
}
