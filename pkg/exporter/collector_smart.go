package exporter

import (
	"fmt"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/smart"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/smart/storage"
)

var _ prometheus.Collector = &SMARTCollector{}
var smartAttrsMaps sync.Map

type SMARTCollector struct {
	// ch used by prometheus
	ch chan<- prometheus.Metric
	// result struct SMART data
	result map[string]smart.TotalResult
}

// NewSMARTCollector collector SMART metrics by smartctl
func NewSMARTCollector() prometheus.Collector {
	return &SMARTCollector{result: make(map[string]smart.TotalResult)}
}

func (sc *SMARTCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(sc, ch)
}

func (sc *SMARTCollector) Collect(ch chan<- prometheus.Metric) {
	log.Info("Collecting metrics for S.M.A.R.T")
	smartStorage, err := smart.NewSMARTStorage()
	if err != nil {
		log.WithError(err).Error("Failed to collect metrics for S.M.A.R.T")
		return
	}

	//  setup data
	sc.setupMetricsChan(ch)
	sc.setupResult(smartStorage)

	// start collect metrics
	sc.collectAtaAttributes()

	log.Info("Succeed collect S.M.A.R.T metrics")
}

// collectAtaAttributes collect all ATA attributes
func (sc *SMARTCollector) collectAtaAttributes() {
	for nodeName, nodeResult := range sc.result {
		for _, diskResult := range nodeResult {
			for _, attr := range diskResult.AtaSmartAttributes.Table {
				if val, ok := smartAttrsMaps.Load("ata_smart_attributes"); ok {
					sc.ch <- prometheus.MustNewConstMetric(
						val.(*prometheus.Desc),
						prometheus.GaugeValue,
						float64(attr.Value),
						diskResult.Device.InfoName,
						nodeName,
						attr.Name,
						attr.Flags.String,
						flagsLong(diskResult, attr.Name),
						"value",
						fmt.Sprintf("%d", attr.ID),
					)
					sc.ch <- prometheus.MustNewConstMetric(
						val.(*prometheus.Desc),
						prometheus.GaugeValue,
						float64(attr.Worst),
						diskResult.Device.InfoName,
						nodeName,
						attr.Name,
						attr.Flags.String,
						flagsLong(diskResult, attr.Name),
						"worst",
						fmt.Sprintf("%d", attr.ID),
					)
					sc.ch <- prometheus.MustNewConstMetric(
						val.(*prometheus.Desc),
						prometheus.GaugeValue,
						float64(attr.Thresh),
						diskResult.Device.InfoName,
						nodeName,
						attr.Name,
						attr.Flags.String,
						flagsLong(diskResult, attr.Name),
						"thresh",
						fmt.Sprintf("%d", attr.ID),
					)
					sc.ch <- prometheus.MustNewConstMetric(
						val.(*prometheus.Desc),
						prometheus.GaugeValue,
						float64(attr.Raw.Value),
						diskResult.Device.InfoName,
						nodeName,
						attr.Name,
						attr.Flags.String,
						flagsLong(diskResult, attr.Name),
						"raw",
						fmt.Sprintf("%d", attr.ID),
					)
				}
			}
		}
	}
}

func (sc *SMARTCollector) setupMetricsChan(ch chan<- prometheus.Metric) {
	sc.ch = ch
}

func (sc *SMARTCollector) setupResult(smartStorage *storage.ConfigMap) {
	// get SMART result from configmap
	data, err := smartStorage.Read()
	if err != nil {
		log.WithError(err).Error("Failed to read S.M.A.R.T data from configmap")
		sc.result = nil
		return
	}

	// convert all nodes SMART result
	for node := range data {
		nodeResult := &smart.TotalResult{}
		err = nodeResult.Unmarshal([]byte(data[node]))
		if err != nil {
			log.WithField("nodeName", node).WithError(err).Error("Failed to Unmarshal SMART result from configmap")
			continue
		}

		sc.result[node] = *nodeResult
	}
}

func flagsLong(attr smart.Result, tableName string) string {
	for _, table := range attr.AtaSmartAttributes.Table {
		if table.Name == tableName {
			var result []string
			if table.Flags.Prefailure {
				result = append(result, "prefailure")
			}
			if table.Flags.UpdatedOnline {
				result = append(result, "updated_online")
			}
			if table.Flags.Performance {
				result = append(result, "performance")
			}
			if table.Flags.AutoKeep {
				result = append(result, "auto_keep")
			}
			if table.Flags.ErrorRate {
				result = append(result, "error_rate")
			}
			if table.Flags.EventCount {
				result = append(result, "event_count")
			}
			return strings.Join(result, ",")
		}
	}
	return ""
}

func init() {
	setupAttrsMaps()
}

func setupAttrsMaps() {
	smartAttrsMaps.Store("ata_smart_attributes", prometheus.NewDesc(
		"ata_smart_attribute",
		"device attributes",
		[]string{
			"device",
			"node",
			"attribute_name",
			"attribute_flags_short",
			"attribute_flags_long",
			"attribute_value_type",
			"attribute_id",
		},
		nil,
	))
}
