package smart

import (
	"context"
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/disk/manager"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/smart/storage"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils/kubernetes"
)

const (
	DefaultCMName = "hwameistor-smart-result"
)

var (
	defaultConfigMapRW *storage.ConfigMap
)

func NewSMARTStorage() (*storage.ConfigMap, error) {
	if defaultConfigMapRW != nil {
		return defaultConfigMapRW, nil
	}

	kubeClient, err := kubernetes.NewClientSet()
	if err != nil {
		log.WithError(err).Error("Failed to create kubeClient")
		return nil, err
	}
	defaultConfigMapRW = storage.NewConfigMap(DefaultCMName, utils.GetNamespace()).SetKubeClient(kubeClient)
	return defaultConfigMapRW, nil
}

// collector collect stats by smartctl
type collector struct {
	syncPeriod time.Duration
}

func NewCollector() *collector {
	return &collector{}
}

func (c *collector) WithSyncPeriod(syncPeriod time.Duration) *collector {
	c.syncPeriod = syncPeriod
	return c
}

// StartTimerCollect  collect S.M.A.R.T result periodically and save to configmap
func (c *collector) StartTimerCollect(ctx context.Context) {
	log.WithField("syncPeriod", c.syncPeriod).Info("Start S.M.A.R.T timer collect")
	// timer trigger
	go wait.Until(c.collectAndSaveConfigMap, c.syncPeriod, ctx.Done())

	<-ctx.Done()
}

// collectAndSaveConfigMap triggered by timer and collect S.M.A.R.T result and save to configmap
func (c *collector) collectAndSaveConfigMap() {
	var (
		ns           = utils.GetNamespace()
		name         = DefaultCMName
		node         = utils.GetNodeName()
		err          error
		smartStorage *storage.ConfigMap
		logCtx       = log.Fields{
			"node":     node,
			"cmNsName": fmt.Sprintf("%s/%s", ns, name),
		}
	)
	log.WithFields(logCtx).Info("Start to collect S.M.A.R.T result")
	smartStorage, err = NewSMARTStorage()
	if err != nil {
		log.WithError(err).Error("Failed to create smart storage")
		return
	}

	// collect metrics on the node
	totalResult, err := c.Collect()
	if err != nil {
		log.WithError(err).Error("Failed to collect S.M.A.R.T result")
		return
	}

	// storage result to configmap: hwameistor/hwameistor-smart-result
	err = smartStorage.SetKV(node, (&totalResult).Marshal())
	if err != nil {
		log.WithError(err).WithFields(logCtx).Error("Failed to update S.M.A.R.T result")
		return
	}

	log.WithFields(logCtx).Info("Succeed to update S.M.A.R.T result")
}

// Collect all devices stats
func (c *collector) Collect() (TotalResult, error) {
	var (
		devices     []device
		devicesCtrs []*controller
		err         error
		totalResult TotalResult
	)

	if devices, err = ScanDevice(); err != nil {
		log.WithError(err).Error("Failed to scan devices")
		return nil, err
	}

	log.Infof("Find %d devices", len(devices))
	for _, dev := range devices {
		devicesCtrs = append(devicesCtrs, newSMARTControllerByDevice(dev))
	}

	for _, devSMARTCtr := range devicesCtrs {
		deviceResult := Result{Device: device{
			Name:     devSMARTCtr.Name,
			InfoName: fmt.Sprintf("%v", devSMARTCtr.Options),
		}}

		stats, err := devSMARTCtr.GetAllStats()
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"device":  devSMARTCtr.DevName,
				"options": devSMARTCtr.Options,
			}).Error("Failed to collect SMART stats")
			deviceResult.Error = err.Error()
			totalResult = append(totalResult, deviceResult)
			continue
		}

		err = json.Unmarshal([]byte(stats.Raw), &deviceResult)
		if err != nil {
			log.WithError(err).Error("Failed to convert SMART  result")
			deviceResult.Error = err.Error()
			totalResult = append(totalResult, deviceResult)
			continue
		}

		totalResult = append(totalResult, deviceResult)
	}

	return totalResult, nil
}

// newSMARTControllerByDevice according device type
func newSMARTControllerByDevice(dev device) *controller {
	ctr := NewSMARTController(&manager.DiskIdentify{
		DevPath: dev.Name,
		DevName: dev.Name,
		Name:    dev.Name,
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

func (t *TotalResult) Marshal() string {
	bytes, _ := json.Marshal(t)
	return string(bytes)
}

func (t *TotalResult) Unmarshal(data []byte) error {
	return json.Unmarshal(data, t)
}

type TotalResult []Result
type Result struct {
	JSONFormatVersion []int `json:"json_format_version"`
	Smartctl          struct {
		Version      []int    `json:"version"`
		SvnRevision  string   `json:"svn_revision"`
		PlatformInfo string   `json:"platform_info"`
		BuildInfo    string   `json:"build_info"`
		Argv         []string `json:"argv"`
		Messages     []struct {
			String   string `json:"string"`
			Severity string `json:"severity"`
		} `json:"messages"`
		ExitStatus int `json:"exit_status"`
	} `json:"smartctl"`
	Device struct {
		Name     string `json:"name"`
		InfoName string `json:"info_name"`
		Type     string `json:"type"`
		Protocol string `json:"protocol"`
	} `json:"device"`
	ModelFamily  string `json:"model_family"`
	ModelName    string `json:"model_name"`
	SerialNumber string `json:"serial_number"`
	Wwn          struct {
		Naa int   `json:"naa"`
		Oui int   `json:"oui"`
		ID  int64 `json:"id"`
	} `json:"wwn"`
	FirmwareVersion string `json:"firmware_version"`
	UserCapacity    struct {
		Blocks int   `json:"blocks"`
		Bytes  int64 `json:"bytes"`
	} `json:"user_capacity"`
	LogicalBlockSize  int `json:"logical_block_size"`
	PhysicalBlockSize int `json:"physical_block_size"`
	RotationRate      int `json:"rotation_rate"`
	FormFactor        struct {
		AtaValue int    `json:"ata_value"`
		Name     string `json:"name"`
	} `json:"form_factor"`
	InSmartctlDatabase bool `json:"in_smartctl_database"`
	AtaVersion         struct {
		String     string `json:"string"`
		MajorValue int    `json:"major_value"`
		MinorValue int    `json:"minor_value"`
	} `json:"ata_version"`
	SataVersion struct {
		String string `json:"string"`
		Value  int    `json:"value"`
	} `json:"sata_version"`
	InterfaceSpeed struct {
		Max struct {
			SataValue      int    `json:"sata_value"`
			String         string `json:"string"`
			UnitsPerSecond int    `json:"units_per_second"`
			BitsPerUnit    int    `json:"bits_per_unit"`
		} `json:"max"`
		Current struct {
			SataValue      int    `json:"sata_value"`
			String         string `json:"string"`
			UnitsPerSecond int    `json:"units_per_second"`
			BitsPerUnit    int    `json:"bits_per_unit"`
		} `json:"current"`
	} `json:"interface_speed"`
	LocalTime struct {
		TimeT   int    `json:"time_t"`
		Asctime string `json:"asctime"`
	} `json:"local_time"`
	SmartStatus struct {
		Passed bool `json:"passed"`
	} `json:"smart_status"`
	AtaSmartData struct {
		OfflineDataCollection struct {
			Status struct {
				Value  int    `json:"value"`
				String string `json:"string"`
			} `json:"status"`
			CompletionSeconds int `json:"completion_seconds"`
		} `json:"offline_data_collection"`
		SelfTest struct {
			Status struct {
				Value  int    `json:"value"`
				String string `json:"string"`
				Passed bool   `json:"passed"`
			} `json:"status"`
			PollingMinutes struct {
				Short      int `json:"short"`
				Extended   int `json:"extended"`
				Conveyance int `json:"conveyance"`
			} `json:"polling_minutes"`
		} `json:"self_test"`
		Capabilities struct {
			Values                        []int `json:"values"`
			ExecOfflineImmediateSupported bool  `json:"exec_offline_immediate_supported"`
			OfflineIsAbortedUponNewCmd    bool  `json:"offline_is_aborted_upon_new_cmd"`
			OfflineSurfaceScanSupported   bool  `json:"offline_surface_scan_supported"`
			SelfTestsSupported            bool  `json:"self_tests_supported"`
			ConveyanceSelfTestSupported   bool  `json:"conveyance_self_test_supported"`
			SelectiveSelfTestSupported    bool  `json:"selective_self_test_supported"`
			AttributeAutosaveEnabled      bool  `json:"attribute_autosave_enabled"`
			ErrorLoggingSupported         bool  `json:"error_logging_supported"`
			GpLoggingSupported            bool  `json:"gp_logging_supported"`
		} `json:"capabilities"`
	} `json:"ata_smart_data"`
	AtaSctCapabilities struct {
		Value                         int  `json:"value"`
		ErrorRecoveryControlSupported bool `json:"error_recovery_control_supported"`
		FeatureControlSupported       bool `json:"feature_control_supported"`
		DataTableSupported            bool `json:"data_table_supported"`
	} `json:"ata_sct_capabilities"`
	AtaSmartAttributes struct {
		Revision int `json:"revision"`
		Table    []struct {
			ID         int    `json:"id"`
			Name       string `json:"name"`
			Value      int    `json:"value"`
			Worst      int    `json:"worst"`
			Thresh     int    `json:"thresh"`
			WhenFailed string `json:"when_failed"`
			Flags      struct {
				Value         int    `json:"value"`
				String        string `json:"string"`
				Prefailure    bool   `json:"prefailure"`
				UpdatedOnline bool   `json:"updated_online"`
				Performance   bool   `json:"performance"`
				ErrorRate     bool   `json:"error_rate"`
				EventCount    bool   `json:"event_count"`
				AutoKeep      bool   `json:"auto_keep"`
			} `json:"flags"`
			Raw struct {
				Value  int    `json:"value"`
				String string `json:"string"`
			} `json:"raw"`
		} `json:"table"`
	} `json:"ata_smart_attributes"`
	PowerOnTime struct {
		Hours int `json:"hours"`
	} `json:"power_on_time"`
	PowerCycleCount int `json:"power_cycle_count"`
	Temperature     struct {
		Current int `json:"current"`
	} `json:"temperature"`
	AtaSmartErrorLog struct {
		Summary struct {
			Revision int `json:"revision"`
			Count    int `json:"count"`
		} `json:"summary"`
	} `json:"ata_smart_error_log"`
	AtaSmartSelfTestLog struct {
		Standard struct {
			Revision int `json:"revision"`
			Count    int `json:"count"`
		} `json:"standard"`
	} `json:"ata_smart_self_test_log"`
	AtaSmartSelectiveSelfTestLog struct {
		Revision int `json:"revision"`
		Table    []struct {
			LbaMin int `json:"lba_min"`
			LbaMax int `json:"lba_max"`
			Status struct {
				Value  int    `json:"value"`
				String string `json:"string"`
			} `json:"status"`
		} `json:"table"`
		Flags struct {
			Value                int  `json:"value"`
			RemainderScanEnabled bool `json:"remainder_scan_enabled"`
		} `json:"flags"`
		PowerUpScanResumeMinutes int `json:"power_up_scan_resume_minutes"`
	} `json:"ata_smart_selective_self_test_log"`
	Error string `json:"error"`
}
