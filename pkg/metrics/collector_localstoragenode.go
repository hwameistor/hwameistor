package metrics

import (
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/labels"
)

type LocalStorageNodeMetricsCollector struct {
	dataCache *metricsCache

	capacityMetricsDesc    *prometheus.Desc
	volumeCountMetricsDesc *prometheus.Desc
	statusMetricsDesc      *prometheus.Desc
}

func newCollectorForLocalStorageNode(dataCache *metricsCache) prometheus.Collector {
	return &LocalStorageNodeMetricsCollector{
		dataCache: dataCache,
		capacityMetricsDesc: prometheus.NewDesc(
			"hwameistor_localstoragenode_capacity",
			"The storage capacity of the localstoragenode.",
			[]string{"nodeName", "poolName", "kind"},
			nil,
		),
		volumeCountMetricsDesc: prometheus.NewDesc(
			"hwameistor_localstoragenode_volumecount",
			"The volume count of the localstoragenode.",
			[]string{"nodeName", "poolName", "kind"},
			nil,
		),
		statusMetricsDesc: prometheus.NewDesc(
			"hwameistor_localstoragenode_status",
			"The status summary of the localstoragenode.",
			[]string{"status"},
			nil,
		),
	}
}

func (mc *LocalStorageNodeMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(mc, ch)
}

func (mc *LocalStorageNodeMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	log.Debug("Collecting metrics for LocalStorageNode ...")
	nodes, err := mc.dataCache.lsnInformer.Lister().List(labels.NewSelector())
	if err != nil || len(nodes) == 0 {
		log.WithError(err).Debug("Not found LocalStorageNode")
		return
	}
	nodeCount := map[string]int64{}
	nodeCount[string(apisv1alpha1.NodeStateReady)] = 0
	nodeCount[string(apisv1alpha1.NodeStateMaintain)] = 0
	nodeCount[string(apisv1alpha1.NodeStateOffline)] = 0

	for _, node := range nodes {
		nodeCount[string(node.Status.State)]++
		for poolName, pool := range node.Status.Pools {
			poolName = unifiedPoolName(poolName)
			ch <- prometheus.MustNewConstMetric(mc.capacityMetricsDesc, prometheus.GaugeValue, float64(pool.TotalCapacityBytes), node.Name, poolName, "Total")
			ch <- prometheus.MustNewConstMetric(mc.capacityMetricsDesc, prometheus.GaugeValue, float64(pool.FreeCapacityBytes), node.Name, poolName, "Free")
			ch <- prometheus.MustNewConstMetric(mc.capacityMetricsDesc, prometheus.GaugeValue, float64(pool.UsedCapacityBytes), node.Name, poolName, "Used")

			ch <- prometheus.MustNewConstMetric(mc.volumeCountMetricsDesc, prometheus.GaugeValue, float64(pool.TotalVolumeCount), node.Name, poolName, "Total")
			ch <- prometheus.MustNewConstMetric(mc.volumeCountMetricsDesc, prometheus.GaugeValue, float64(pool.FreeVolumeCount), node.Name, poolName, "Free")
			ch <- prometheus.MustNewConstMetric(mc.volumeCountMetricsDesc, prometheus.GaugeValue, float64(pool.UsedVolumeCount), node.Name, poolName, "Used")
		}
	}

	for state, count := range nodeCount {
		ch <- prometheus.MustNewConstMetric(mc.statusMetricsDesc, prometheus.GaugeValue, float64(count), state)
	}

}
