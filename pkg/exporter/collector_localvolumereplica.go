package exporter

import (
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
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
			"The status summary of the localvolumereplica.",
			[]string{"nodeName", "poolName", "status"},
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

	replicaStatusCount := map[string]map[string]map[string]int64{
		apisv1alpha1.DiskClassNameHDD:  {},
		apisv1alpha1.DiskClassNameSSD:  {},
		apisv1alpha1.DiskClassNameNVMe: {},
	}

	for _, replica := range replicas {
		poolName := unifiedPoolName(replica.Spec.PoolName)
		_, exist := replicaStatusCount[poolName][replica.Spec.NodeName]
		if !exist {
			replicaStatusCount[poolName][replica.Spec.NodeName] = map[string]int64{}
		}
		replicaStatusCount[poolName][replica.Spec.NodeName][string(replica.Status.State)]++

		ch <- prometheus.MustNewConstMetric(mc.capacityMetricsDesc, prometheus.GaugeValue, float64(replica.Status.AllocatedCapacityBytes), replica.Spec.NodeName, poolName, replica.Spec.VolumeName)
	}

	for poolName, nodeCount := range replicaStatusCount {
		for nodeName, stateCount := range nodeCount {
			for state, count := range stateCount {
				ch <- prometheus.MustNewConstMetric(mc.statusMetricsDesc, prometheus.GaugeValue, float64(count), nodeName, poolName, state)
			}
		}

	}
}
