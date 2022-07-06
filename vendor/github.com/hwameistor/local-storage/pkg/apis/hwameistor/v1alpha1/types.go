package v1alpha1

// State is state type of resources
type State string

// Phase is the phase of an operation
type Phase string

// states
const (
	NodeStateReady    State = "Ready"
	NodeStateMaintain State = "Maintain"
	NodeStateOffline  State = "Offline"

	VolumeStateCreating    State = "Creating"
	VolumeStateReady       State = "Ready"
	VolumeStateNotReady    State = "NotReady"
	VolumeStateToBeDeleted State = "ToBeDeleted"
	VolumeStateDeleted     State = "Deleted"

	VolumeReplicaStateInvalid     State = "Invalid"
	VolumeReplicaStateCreating    State = "Creating"
	VolumeReplicaStateReady       State = "Ready"
	VolumeReplicaStateNotReady    State = "NotReady"
	VolumeReplicaStateToBeDeleted State = "ToBeDeleted"
	VolumeReplicaStateDeleted     State = "Deleted"

	// ha replica state
	HAVolumeReplicaStateConsistent   State = "Consistent"
	HAVolumeReplicaStateInconsistent State = "Inconsistent"
	HAVolumeReplicaStateUp           State = "Up"
	HAVolumeReplicaStateDown         State = "Down"

	// purpose of the following CRDs is for operational job,
	// so, they will be in different state machine from volume/volumereplica
	OperationStateSubmitted   State = "Submitted"
	OperationStateInProgress  State = "InProgress"
	OperationStateCompleted   State = "Completed"
	OperationStateToBeAborted State = "ToBeAborted"
	OperationStateAborting    State = "Cancelled"
	OperationStateAborted     State = "Aborted"

	DiskStateAvailable State = "Available"
	DiskStateInUse     State = "InUse"
	DiskStateOffline   State = "Offline"

	LVMVolumeMaxCount int64 = 1000
	RAMVolumeMaxCount int64 = 1000

	VolumeKindLVM = "LVM"

	VolumeExpansionCapacityBytesMin int64 = 10 * 1024 * 1024 // 10MiB

	StoragePoolCapacityThresholdRatio = 0.85

	VolumeMigratePhaseMove  Phase = "Move"
	VolumeMigratePhasePrune Phase = "Prune"
)

// disk class
const (
	DiskClassNameHDD  = "HDD"
	DiskClassNameSSD  = "SSD"
	DiskClassNameNVMe = "NVMe"
)

// consts
const (
	PoolNamePrefix  = "LocalStorage_Pool"
	PoolNameForHDD  = PoolNamePrefix + DiskClassNameHDD
	PoolNameForSSD  = PoolNamePrefix + DiskClassNameSSD
	PoolNameForNVMe = PoolNamePrefix + DiskClassNameNVMe

	PoolTypeRegular = "REGULAR"
)

// consts
const (
	VolumeParameterPoolClassKey     = "poolClass"
	VolumeParameterPoolTypeKey      = "poolType"
	VolumeParameterReplicaNumberKey = "replicaNumber"
	VolumeParameterFSTypeKey        = "csi.storage.k8s.io/fstype"
	VolumeParameterConvertible      = "convertible"
)

// misc
const (
	CSIDriverName = "lvm.hwameistor.io"

	VendorName = "hwameistor.io"
)

// k8snode
const (
	StorageIPv4AddressAnnotationKeyEnv = "NODE_ANNOTATION_KEY_STORAGE_IPV4"
)

// localstorage local storage dev paths
const (
	DiskDevRootPath     = "/dev"
	AssigedDiskPool     = DiskDevRootPath + "/LocalStorage_DiskPool"
	AssigedDiskPoolHDD  = AssigedDiskPool + DiskClassNameHDD
	AssigedDiskPoolSSD  = AssigedDiskPool + DiskClassNameSSD
	AssigedDiskPoolNVMe = AssigedDiskPool + DiskClassNameNVMe
)

// // =========== For disk health check (smartctl) ===============

// // consts
// const (
// 	SmartCtlDeviceProtocolATA  = "ATA"
// 	SmartCtlDeviceProtocolSCSI = "SCSI"
// 	SmartCtlDeviceProtocolNVMe = "NVMe"

// 	SmartCtlDeviceProductVirtualDisk = "Virtual disk"
// )

// // NVMeSmartHealthDetailsInfo struct
// type NVMeSmartHealthDetailsInfo struct {
// 	CriticalWarning         int64 `json:"critical_warning"`
// 	Temperature             int64 `json:"temperature"`
// 	AvailableSpare          int64 `json:"available_spare"`
// 	AvailableSpareThreshold int64 `json:"available_spare_threshold"`
// 	PercentageUsed          int64 `json:"percentage_used"`
// 	DataUnitsRead           int64 `json:"data_units_read"`
// 	DataUnitsWritten        int64 `json:"data_units_written"`
// 	HostReads               int64 `json:"host_reads"`
// 	HostWrites              int64 `json:"host_writes"`
// 	ControllerBusyTime      int64 `json:"controller_busy_time"`
// 	PowerCycles             int64 `json:"power_cycles"`
// 	PowerOnHours            int64 `json:"power_on_hours"`
// 	UnsafeShutdowns         int64 `json:"unsafe_shutdowns"`
// 	MediaErrors             int64 `json:"media_errors"`
// 	NumErrLogEntries        int64 `json:"num_err_log_entries"`
// }

// // SCSISmartHealthDetailsInfo struct
// type SCSISmartHealthDetailsInfo struct {
// 	Read   *SCSIErrorCounter `json:"read,omitempty"`
// 	Write  *SCSIErrorCounter `json:"write,omitempty"`
// 	Verify *SCSIErrorCounter `json:"verify,omitempty"`
// }

// // SCSIErrorCounter struct
// type SCSIErrorCounter struct {
// 	ErrorsCorrectedByECCFast        int64  `json:"errors_corrected_by_eccfast"`
// 	ErrorsCorrectedByECCDelayed     int64  `json:"errors_corrected_by_eccdelayed"`
// 	ErrorsCorrectedByRereadRewrites int64  `json:"errors_corrected_by_rereads_rewrites"`
// 	TotalErrorsCorrected            int64  `json:"total_errors_corrected"`
// 	CorrectionAlgorithmInvocations  int64  `json:"correction_algorithm_invocations"`
// 	GigabytesProcessed              string `json:"gigabytes_processed"`
// 	TotalUncorrectedErrors          int64  `json:"total_uncorrected_errors"`
// }

// // ATASmartHealthDetailsInfo struct
// type ATASmartHealthDetailsInfo struct {
// 	AttributesTable []ATASmartHealthAttribute `json:"table,omitempty"`
// }

// // ATASmartHealthAttribute struct
// type ATASmartHealthAttribute struct {
// 	ID         int64                           `json:"id"`
// 	Name       string                          `json:"name"`
// 	Value      int64                           `json:"value"`
// 	Worst      int64                           `json:"worst"`
// 	Threshold  int64                           `json:"thresh"`
// 	WhenFailed string                          `json:"when_failed"`
// 	Flags      *ATASmartHealthAttributeFlag    `json:"flags"`
// 	Raw        *ATASmartHealthAttributeRawData `json:"raw"`
// }

// // ATASmartHealthAttributeFlag struct
// type ATASmartHealthAttributeFlag struct {
// 	Value         int64  `json:"value"`
// 	String        string `json:"string"`
// 	Prefailure    bool   `json:"prefailure"`
// 	UpdatedOnline bool   `json:"updated_online"`
// 	Performance   bool   `json:"performance"`
// 	ErrorRate     bool   `json:"error_rate"`
// 	EventCount    bool   `json:"event_count"`
// 	AutoKeep      bool   `json:"auto_keep"`
// }

// // ATASmartHealthAttributeRawData struct
// type ATASmartHealthAttributeRawData struct {
// 	Value  int64  `json:"value"`
// 	String string `json:"string"`
// }

// SystemMode of HA module
type SystemMode string

// misc
var (
	SystemModeDRBD SystemMode = "drbd"
)

// DRBDSystemConfig of HA module
type DRBDSystemConfig struct {
	StartPort int `json:"haStartPort"`
	EndPort   int `json:"haEndPort"`
}

// SystemConfig is volume HA related system configuration
type SystemConfig struct {
	Mode             SystemMode        `json:"mode"`
	DRBD             *DRBDSystemConfig `json:"drbd"`
	MaxHAVolumeCount int               `json:"maxVolumeCount"`
}
