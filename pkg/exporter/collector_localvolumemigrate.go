package exporter

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/labels"
)

type LocalVolumeMigrateMetricsCollector struct {
	dataCache *metricsCache

	durationMetricsDesc *prometheus.Desc
	statusMetricsDesc   *prometheus.Desc
}

func newCollectorForLocalVolumeMigrate(dataCache *metricsCache) prometheus.Collector {
	return &LocalVolumeMigrateMetricsCollector{
		dataCache: dataCache,

		durationMetricsDesc: prometheus.NewDesc(
			"hwameistor_localvolumemigrate_duration",
			"The duration of the localvolumemigrate operation.",
			[]string{"volumeName", "startTime", "from", "to"},
			nil,
		),
		statusMetricsDesc: prometheus.NewDesc(
			"hwameistor_localvolumemigrate_status",
			"The status of the localvolumemigrate operation.",
			[]string{"volumeName", "status"},
			nil,
		),
	}

}

func (mc *LocalVolumeMigrateMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(mc, ch)
}

func (mc *LocalVolumeMigrateMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	log.Debug("Collecting metrics for LocalVolumeMigrate ...")
	migrates, err := mc.dataCache.lvMigrateInformer.Lister().List(labels.NewSelector())
	if err != nil || len(migrates) == 0 {
		log.WithError(err).Debug("Not found LocalVolumeMigrate")
		return
	}

	for _, migrate := range migrates {
		ch <- prometheus.MustNewConstMetric(
			mc.durationMetricsDesc,
			prometheus.GaugeValue,
			time.Since(migrate.CreationTimestamp.Time).Seconds(),
			migrate.Spec.VolumeName,
			migrate.CreationTimestamp.String(),
			migrate.Spec.SourceNode,
			migrate.Status.TargetNode,
		)
		ch <- prometheus.MustNewConstMetric(mc.statusMetricsDesc, prometheus.GaugeValue, 1, migrate.Spec.VolumeName, string(migrate.Status.State))
	}
}
