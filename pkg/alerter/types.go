package alerter

import (
	localstorageinformers "github.com/hwameistor/local-storage/pkg/apis/client/informers/externalversions"
)

// misc
const (
	SeverityCritical = iota
	SeverityWarning
	SeverityNotice

	ModuleDisk                   = "PhyDisk"
	ModuleVolume                 = "LocalVolume"
	ModuleVolumeReplica          = "LocalVolumeReplica"
	ModuleNode                   = "LocalStorageNode"
	ModuleVolumeExpansion        = "LocalVolumeExpansion"
	ModuleVolumeReplicaExpansion = "LocalVolumeReplicaExpansion"
)

// Alerter interface
type Alerter interface {
	Run(informerFactory localstorageinformers.SharedInformerFactory, stopCh <-chan struct{})
}

// disk consts
const (
	DiskEventOffline                 = "Offline"
	DiskEventSmartCheckFailed        = "SmartCheckFailed"
	DiskEventHighTemperature         = "HighTemperature"
	DiskEventSmartAttributesAbnormal = "SmartAttributesAbnormal"

	DiskTemperatureThreshold = 60

	scsiDiskSmartAttributeTotalCorrectedErrorsThreshold   = 100
	scsiDiskSmartAttributeTotalUncorrectedErrorsThreshold = 50

	nvmeDiskSmartAttributeCriticalWarningThreshold  = 10
	nvmeDiskSmartAttributeMediaErrorsThreshold      = 10
	nvmeDiskSmartAttributeNumErrLogEntriesThreshold = 50
	nvmeDiskSmartAttributeUnsafeShutdownsThreshold  = 20
)

type ataDiskSmartAttribute struct {
	name            string
	id              int64
	threshold       int64
	isSmallerBetter bool // true: check value > threshold; false: check value < threshold
}

var ataDiskSmartAttributesToCheck = []ataDiskSmartAttribute{
	{
		id:              5,
		threshold:       0,
		name:            "Reallocated_Sector_Ct",
		isSmallerBetter: true,
	},
	{
		id:              10,
		threshold:       5,
		name:            "Spin_Retry_Count",
		isSmallerBetter: true,
	},
	{
		id:              12,
		threshold:       28,
		name:            "Power_Cycle_Count",
		isSmallerBetter: true,
	},
	{
		id:              187,
		threshold:       0,
		name:            "Reported_Uncorrectable_Errors",
		isSmallerBetter: true,
	},
	{
		id:              188,
		threshold:       0,
		name:            "Command_Timeout",
		isSmallerBetter: true,
	},
	{
		id:              197,
		threshold:       0,
		name:            "Current_Pending_Sector_Count",
		isSmallerBetter: true,
	},
	{
		id:              198,
		threshold:       0,
		name:            "Offline_Uncorrectable",
		isSmallerBetter: true,
	},
}

const ()

// storage node consts
const (
	NodeEventOffline = "Offline"
)
