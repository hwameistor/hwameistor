package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/labels"
)

type LocalVolumeReplicaMetricsCollector struct {
	dataCache *metricsCache

	statusMetricsDesc   *prometheus.Desc
	capacityMetricsDesc *prometheus.Desc
}

func newCollectorForLocalVolumeReplica(dataCache *metricsCache) prometheus.Collector {
	return &LocalVolumeReplicaMetricsCollector{
		dataCache: dataCache,
		statusMetricsDesc: prometheus.NewDesc(
			"hwameistor_localvolumereplica_status",
			"The status of the localvolumereplica.",
			[]string{"nodeName", "poolName", "volumeName", "status"},
			nil,
		),

		capacityMetricsDesc: prometheus.NewDesc(
			"hwameistor_localvolumereplica_capacity",
			"The capacity of the localvolumereplica.",
			[]string{"nodeName", "poolName", "volumeName"},
			nil,
		),
	}
}

func (mc *LocalVolumeReplicaMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(mc, ch)
}

func (mc *LocalVolumeReplicaMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	log.Debug("Collecting metrics for LocalVolumeReplica ...")
	replicas, err := mc.dataCache.lvrInformer.Lister().List(labels.NewSelector())
	if err != nil || len(replicas) == 0 {
		log.WithError(err).Debug("Not found LocalVolumeReplica")
		return
	}

	for _, replica := range replicas {
		poolName := unifiedPoolName(replica.Spec.PoolName)
		ch <- prometheus.MustNewConstMetric(mc.capacityMetricsDesc, prometheus.GaugeValue, float64(replica.Status.AllocatedCapacityBytes), replica.Spec.NodeName, poolName, replica.Spec.VolumeName)
		ch <- prometheus.MustNewConstMetric(mc.statusMetricsDesc, prometheus.GaugeValue, 1, replica.Spec.NodeName, poolName, replica.Spec.VolumeName, string(replica.Status.State))
	}
}
