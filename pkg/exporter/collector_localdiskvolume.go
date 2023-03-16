package exporter

import (
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/labels"
)

type LocalDiskVolumeMetricsCollector struct {
	dataCache *metricsCache

	stateMetricsDesc    *prometheus.Desc
	capacityMetricsDesc *prometheus.Desc
}

func newCollectorForLocalDiskVolume(dataCache *metricsCache) prometheus.Collector {
	return &LocalDiskVolumeMetricsCollector{
		dataCache: dataCache,
		stateMetricsDesc: prometheus.NewDesc(
			"hwameistor_localdiskvolume_state_count",
			"The state summary of the localdiskvolume.",
			[]string{"type", "state"},
			nil,
		),
		capacityMetricsDesc: prometheus.NewDesc(
			"hwameistor_localdiskvolume_capacity",
			"The capacity of the localdiskvolume.",
			[]string{"volumeName", "kind"},
			nil,
		),
	}
}

func (mc *LocalDiskVolumeMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(mc, ch)
}

func (mc *LocalDiskVolumeMetricsCollector) Collect(ch chan<- prometheus.Metric) {

	log.Debug("Collecting metrics for LocalDiskVolume ...")
	volumes, err := mc.dataCache.ldvInformer.Lister().List(labels.NewSelector())
	if err != nil || len(volumes) == 0 {
		log.WithError(err).Debug("Not found LocalDiskVolume")
		return
	}

	volumeStatusCount := map[string]map[string]int64{
		apisv1alpha1.DiskClassNameHDD:  {},
		apisv1alpha1.DiskClassNameSSD:  {},
		apisv1alpha1.DiskClassNameNVMe: {},
	}
	for _, vol := range volumes {
		ch <- prometheus.MustNewConstMetric(mc.capacityMetricsDesc, prometheus.GaugeValue, float64(vol.Status.AllocatedCapacityBytes), vol.Name, "Allocated")
		ch <- prometheus.MustNewConstMetric(mc.capacityMetricsDesc, prometheus.GaugeValue, float64(vol.Status.UsedCapacityBytes), vol.Name, "Used")
		volumeStatusCount[vol.Spec.DiskType][string(vol.Status.State)]++
	}
	for typeName, typeCount := range volumeStatusCount {
		for stateName, count := range typeCount {
			ch <- prometheus.MustNewConstMetric(mc.stateMetricsDesc, prometheus.GaugeValue, float64(count), typeName, stateName)
		}
	}
}
