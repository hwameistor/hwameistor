package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/labels"
)

type LocalVolumeMetricsCollector struct {
	dataCache *metricsCache

	statusMetricsDesc   *prometheus.Desc
	capacityMetricsDesc *prometheus.Desc
}

func newCollectorForLocalVolume(dataCache *metricsCache) prometheus.Collector {
	return &LocalVolumeMetricsCollector{
		dataCache: dataCache,

		statusMetricsDesc: prometheus.NewDesc(
			"hwameistor_localvolume_status",
			"The status summary of the localvolume.",
			[]string{"poolName", "volumeName", "type", "mountedOn", "status"},
			nil,
		),

		capacityMetricsDesc: prometheus.NewDesc(
			"hwameistor_localvolume_capacity",
			"The capacity of the localvolume.",
			[]string{"poolName", "volumeName", "type", "mountedOn", "kind"},
			nil,
		),
	}
}

func (mc *LocalVolumeMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(mc, ch)
}

func (mc *LocalVolumeMetricsCollector) Collect(ch chan<- prometheus.Metric) {

	log.Debug("Collecting metrics for LocalVolume ...")
	lvs, err := mc.dataCache.lvInformer.Lister().List(labels.NewSelector())
	if err != nil || len(lvs) == 0 {
		log.WithError(err).Debug("Not found LocalVolume")
		return
	}
	for _, vol := range lvs {
		poolName := unifiedPoolName(vol.Spec.PoolName)
		volType := "Unknown"
		if !vol.Spec.Convertible {
			volType = VolumeTypeNonHA
		} else if vol.Spec.ReplicaNumber > 1 {
			volType = VolumeTypeHA
		} else {
			volType = VolumeTypeConvertible
		}
		ch <- prometheus.MustNewConstMetric(mc.capacityMetricsDesc, prometheus.GaugeValue, float64(vol.Status.AllocatedCapacityBytes), poolName, vol.Name, volType, vol.Status.PublishedNodeName, "Allocated")
		ch <- prometheus.MustNewConstMetric(mc.capacityMetricsDesc, prometheus.GaugeValue, float64(vol.Status.UsedCapacityBytes), poolName, vol.Name, volType, vol.Status.PublishedNodeName, "Used")
		ch <- prometheus.MustNewConstMetric(mc.statusMetricsDesc, prometheus.GaugeValue, 1, poolName, vol.Name, volType, vol.Status.PublishedNodeName, string(vol.Status.State))
	}

	log.Debug("Collecting metrics for LocalDiskVolume ...")
	ldvs, err := mc.dataCache.ldvInformer.Lister().List(labels.NewSelector())
	if err != nil || len(ldvs) == 0 {
		log.WithError(err).Debug("Not found LocalDiskVolume")
		return
	}

	for _, vol := range ldvs {
		mountedOn := ""
		if len(vol.Status.MountPoints) > 0 && len(vol.Spec.Accessibility.Nodes) > 0 {
			mountedOn = vol.Spec.Accessibility.Nodes[0]
		}
		ch <- prometheus.MustNewConstMetric(mc.capacityMetricsDesc, prometheus.GaugeValue, float64(vol.Status.AllocatedCapacityBytes), vol.Spec.DiskType, vol.Name, "Disk", mountedOn, "Allocated")
		ch <- prometheus.MustNewConstMetric(mc.capacityMetricsDesc, prometheus.GaugeValue, float64(vol.Status.UsedCapacityBytes), vol.Spec.DiskType, vol.Name, "Disk", mountedOn, "Used")
		ch <- prometheus.MustNewConstMetric(mc.statusMetricsDesc, prometheus.GaugeValue, 1, vol.Spec.DiskType, vol.Name, "Disk", mountedOn, string(vol.Status.State))
	}
}
