package exporter

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/labels"
)

type LocalVolumeExpandMetricsCollector struct {
	dataCache *metricsCache

	durationMetricsDesc *prometheus.Desc
	statusMetricsDesc   *prometheus.Desc
}

func newCollectorForLocalVolumeExpand(dataCache *metricsCache) prometheus.Collector {
	return &LocalVolumeExpandMetricsCollector{
		dataCache: dataCache,
		durationMetricsDesc: prometheus.NewDesc(
			"hwameistor_localvolumeexpand_duration",
			"The duration of the localvolumeexpand operation.",
			[]string{"volumeName", "startTime"},
			nil,
		),
		statusMetricsDesc: prometheus.NewDesc(
			"hwameistor_localvolumeexpand_status",
			"The status of the localvolumeexpand operation.",
			[]string{"volumeName", "status"},
			nil,
		),
	}

}

func (mc *LocalVolumeExpandMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(mc, ch)
}

func (mc *LocalVolumeExpandMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	log.Debug("Collecting metrics for LocalVolumeExpand ...")
	expands, err := mc.dataCache.lvExpandInformer.Lister().List(labels.NewSelector())
	if err != nil || len(expands) == 0 {
		log.WithError(err).Debug("Not found LocalVolumeExpand")
		return
	}

	for _, expand := range expands {
		ch <- prometheus.MustNewConstMetric(
			mc.durationMetricsDesc,
			prometheus.GaugeValue,
			time.Since(expand.CreationTimestamp.Time).Seconds(),
			expand.Spec.VolumeName,
			expand.CreationTimestamp.String(),
		)
		ch <- prometheus.MustNewConstMetric(mc.statusMetricsDesc, prometheus.GaugeValue, 1, expand.Spec.VolumeName, string(expand.Status.State))
	}
}
