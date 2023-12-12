package prediction

import (
	"encoding/json"
)

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
		Revision int         `json:"revision"`
		Table    []TableItem `json:"table"`
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

type TableItem struct {
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
}

type DiskTrainingData struct {
	//smart_5_raw,smart_9_raw,smart_187_raw,smart_188_raw,smart_193_raw,smart_194_raw,smart_197_raw,smart_198_raw,smart_241_raw,smart_242_raw,failure
	Node          string `csv:"node"`
	Disk          string `csv:"disk"`
	Smart_5_raw   int    `csv:"smart_5_raw"`
	Smart_9_raw   int    `csv:"smart_9_raw"`
	Smart_187_raw int    `csv:"smart_187_raw"`
	Smart_188_raw int    `csv:"smart_188_raw"`
	Smart_193_raw int    `csv:"smart_193_raw"`
	Smart_194_raw int    `csv:"smart_194_raw"`
	Smart_197_raw int    `csv:"smart_197_raw"`
	Smart_198_raw int    `csv:"smart_198_raw"`
	Smart_241_raw int    `csv:"smart_241_raw"`
	Smart_242_raw int    `csv:"smart_242_raw"`
}

type DiskPredictResult struct {
	Node    string `csv:"node"`
	Disk    string `csv:"disk"`
	Failure string `csv:"failure"`
}

const (
	Smart_5_raw   = "Reallocated_Sector_Ct"
	Smart_9_raw   = "Power_On_Hours"
	Smart_187_raw = "Reported_Uncorrect"
	Smart_188_raw = "Command_Timeout"
	Smart_193_raw = "Load_Cycle_Count"
	Smart_194_raw = "Temperature_Celsius"
	Smart_197_raw = "Current_Pending_Sector"
	Smart_198_raw = "Offline_Uncorrectable"
	Smart_241_raw = "Total_LBAs_Written"
	Smart_242_raw = "Total_LBAs_Read"
)
