package exporter

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/labels"
)

type LocalVolumeConvertMetricsCollector struct {
	dataCache *metricsCache

	durationMetricsDesc *prometheus.Desc
	statusMetricsDesc   *prometheus.Desc
}

func newCollectorForLocalVolumeConvert(dataCache *metricsCache) prometheus.Collector {
	return &LocalVolumeConvertMetricsCollector{
		dataCache: dataCache,
		durationMetricsDesc: prometheus.NewDesc(
			"hwameistor_localvolumeconvert_duration",
			"The duration of the localvolumeconvert operation.",
			[]string{"volumeName", "startTime"},
			nil,
		),
		statusMetricsDesc: prometheus.NewDesc(
			"hwameistor_localvolumeconvert_status",
			"The status of the localvolumeconvert operation.",
			[]string{"volumeName", "status"},
			nil,
		),
	}

}

func (mc *LocalVolumeConvertMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(mc, ch)
}

func (mc *LocalVolumeConvertMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	log.Debug("Collecting metrics for LocalVolumeConvert ...")
	converts, err := mc.dataCache.lvConvertInformer.Lister().List(labels.NewSelector())
	if err != nil || len(converts) == 0 {
		log.WithError(err).Debug("Not found LocalVolumeConvert")
		return
	}

	for _, convert := range converts {
		ch <- prometheus.MustNewConstMetric(
			mc.durationMetricsDesc,
			prometheus.GaugeValue,
			time.Since(convert.CreationTimestamp.Time).Seconds(),
			convert.Spec.VolumeName,
			convert.CreationTimestamp.String(),
		)
		ch <- prometheus.MustNewConstMetric(mc.statusMetricsDesc, prometheus.GaugeValue, 1, convert.Spec.VolumeName, string(convert.Status.State))
	}
}
