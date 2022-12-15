package smart

import (
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/disk/manager"
	log "github.com/sirupsen/logrus"
	"strings"
)

// collector collect stats by smartctl
type collector struct {
}

func NewCollector() *collector {
	return &collector{}
}

// Collect all devices stats
func (*collector) Collect() {
	var (
		devices     []device
		devicesCtrs []*controller
		err         error
	)

	if devices, err = DiskScan(); err != nil {
		log.WithError(err).Error("Failed to scan devices")
		return
	}

	log.Infof("Find %d devices", len(devices))
	for _, dev := range devices {
		devicesCtrs = append(devicesCtrs, newSMARTControllerByDevice(dev))
	}

	for _, devSMARTCtr := range devicesCtrs {
		stats, err := devSMARTCtr.GetAllStats()
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"device":  devSMARTCtr.DevName,
				"options": devSMARTCtr.Options,
			}).Error("Failed to collect SMART stats")
			continue
		}

		log.WithFields(log.Fields{
			"device":  devSMARTCtr.DevName,
			"options": devSMARTCtr.Options,
		}).Infof("ata_smart_data.capabilities.exec_offline_immediate_supported: %v",
			stats.Get("ata_smart_data.capabilities.exec_offline_immediate_supported").Bool())
	}
}

// newSMARTControllerByDevice according device type
func newSMARTControllerByDevice(dev device) *controller {
	ctr := NewSMARTController(&manager.DiskIdentify{
		DevPath: dev.Name,
		DevName: dev.Name,
	})

	// if device type is mega raid, tell smartctl this type
	if strings.Contains(dev.Type, ",") {
		ctr.Options = []string{
			"--device",
			dev.Type,
		}
	}

	return ctr
}
