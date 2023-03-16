package exporter

import (
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/labels"
)

type LocalVolumeMetricsCollector struct {
	dataCache *metricsCache

	typeMetricsDesc     *prometheus.Desc
	statusMetricsDesc   *prometheus.Desc
	capacityMetricsDesc *prometheus.Desc
}

func newCollectorForLocalVolume(dataCache *metricsCache) prometheus.Collector {
	return &LocalVolumeMetricsCollector{
		dataCache: dataCache,
		typeMetricsDesc: prometheus.NewDesc(
			"hwameistor_localvolume_type_count",
			"The type of the localvolume.",
			[]string{"poolName", "type"},
			nil,
		),

		statusMetricsDesc: prometheus.NewDesc(
			"hwameistor_localvolume_status_count",
			"The status summary of the localvolume.",
			[]string{"poolName", "type", "status"},
			nil,
		),

		capacityMetricsDesc: prometheus.NewDesc(
			"hwameistor_localvolume_capacity",
			"The capacity of the localvolume.",
			[]string{"poolName", "volumeName", "type"},
			nil,
		),
	}
}

func (mc *LocalVolumeMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(mc, ch)
}

func (mc *LocalVolumeMetricsCollector) Collect(ch chan<- prometheus.Metric) {

	log.Debug("Collecting metrics for LocalVolume ...")
	volumes, err := mc.dataCache.lvInformer.Lister().List(labels.NewSelector())
	if err != nil || len(volumes) == 0 {
		log.WithError(err).Debug("Not found LocalVolume")
		return
	}
	volumeTypeCount := map[string]map[string]int64{
		apisv1alpha1.DiskClassNameHDD: {
			VolumeTypeNonHA:       0,
			VolumeTypeConvertible: 0,
			VolumeTypeHA:          0,
		},
		apisv1alpha1.DiskClassNameSSD: {
			VolumeTypeNonHA:       0,
			VolumeTypeConvertible: 0,
			VolumeTypeHA:          0,
		},
		apisv1alpha1.DiskClassNameNVMe: {
			VolumeTypeNonHA:       0,
			VolumeTypeConvertible: 0,
			VolumeTypeHA:          0,
		},
	}
	volumeStatusCount := map[string]map[string]map[string]int64{
		apisv1alpha1.DiskClassNameHDD: {
			VolumeTypeNonHA:       {},
			VolumeTypeHA:          {},
			VolumeTypeConvertible: {},
		},
		apisv1alpha1.DiskClassNameSSD: {
			VolumeTypeNonHA:       {},
			VolumeTypeHA:          {},
			VolumeTypeConvertible: {},
		},
		apisv1alpha1.DiskClassNameNVMe: {
			VolumeTypeNonHA:       {},
			VolumeTypeHA:          {},
			VolumeTypeConvertible: {},
		},
	}
	for _, vol := range volumes {
		poolName := unifiedPoolName(vol.Spec.PoolName)
		if !vol.Spec.Convertible {
			volumeTypeCount[poolName][VolumeTypeNonHA]++
			volumeStatusCount[poolName][VolumeTypeNonHA][string(vol.Status.State)]++
		} else if vol.Spec.ReplicaNumber > 1 {
			volumeTypeCount[poolName][VolumeTypeHA]++
			volumeStatusCount[poolName][VolumeTypeHA][string(vol.Status.State)]++
		} else {
			volumeTypeCount[poolName][VolumeTypeConvertible]++
			volumeStatusCount[poolName][VolumeTypeConvertible][string(vol.Status.State)]++
		}
		ch <- prometheus.MustNewConstMetric(mc.capacityMetricsDesc, prometheus.GaugeValue, float64(vol.Status.AllocatedCapacityBytes), poolName, vol.Name, "Allocated")
		ch <- prometheus.MustNewConstMetric(mc.capacityMetricsDesc, prometheus.GaugeValue, float64(vol.Status.UsedCapacityBytes), poolName, vol.Name, "Used")
	}
	for poolName, volumeCount := range volumeTypeCount {
		for volType, count := range volumeCount {
			ch <- prometheus.MustNewConstMetric(mc.typeMetricsDesc, prometheus.GaugeValue, float64(count), poolName, volType)
		}
	}
	for poolName, typeCount := range volumeStatusCount {
		for volType, statusCount := range typeCount {
			for state, count := range statusCount {
				ch <- prometheus.MustNewConstMetric(mc.statusMetricsDesc, prometheus.GaugeValue, float64(count), poolName, volType, state)
			}
		}
	}

}
