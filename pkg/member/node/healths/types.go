package healths

import (
	"fmt"

	localstoragev1alpha1 "github.com/hwameistor/local-storage/pkg/apis/localstorage/v1alpha1"
)

// SmartCtlScanResult is result of "smartctl --scan -j"
type SmartCtlScanResult struct {
	Devices []DeviceInfo `json:"devices,omitempty"`
}

// DeviceInfo struct
type DeviceInfo struct {
	Name     string `json:"name,omitempty"`
	InfoName string `json:"info_name,omitempty"`
	Type     string `json:"type,omitempty"`
	Protocol string `json:"protocol,omitempty"`
}

// DeviceTypeInfo struct
type DeviceTypeInfo struct {
	SCSIValue int64  `json:"scsi_value,omitempty"`
	Name      string `json:"name,omitempty"`
}

// SmartStatus struct
type SmartStatus struct {
	Passed bool `json:"passed,omitempty"`
}

// TemperatureStatus struct
type TemperatureStatus struct {
	Current int64 `json:"current,omitempty"`
}

// PowerOnTimeStatus struct
type PowerOnTimeStatus struct {
	Hours   int64 `json:"hours,omitempty"`
	Minutes int64 `json:"minutes,omitempty"`
}

// UserCapacityInfo struct
type UserCapacityInfo struct {
	Bytes  int64 `json:"bytes,omitempty"`
	Blocks int64 `json:"blocks,omitempty"`
}

// NVMePCIVendorInfo struct
type NVMePCIVendorInfo struct {
	ID          int64 `json:"id,omitempty"`
	SubSystemID int64 `json:"subsystem_id,omitempty"`
}

func (v NVMePCIVendorInfo) String() string {
	return fmt.Sprintf("%d/%d", v.ID, v.SubSystemID)
}

// FormFactorInfo struct - disk size
type FormFactorInfo struct {
	ATAValue int64  `json:"ata_value,omitempty"`
	Name     string `json:"name,omitempty"`
}

// ATAVersionInfo struct
type ATAVersionInfo struct {
	String     string `json:"string,omitempty"`
	MajorValue int64  `json:"major_value,omitempty"`
	MinorValue int64  `json:"minor_value,omitempty"`
}

// SATAVersionInfo struct
type SATAVersionInfo struct {
	String string `json:"string,omitempty"`
	Value  int64  `json:"value,omitempty"`
}

// InterfaceSpeedInfo struct
type InterfaceSpeedInfo struct {
	Max     *SATASpeedInfo `json:"max,omitempty"`
	Current *SATASpeedInfo `json:"current,omitempty"`
}

// SATASpeedInfo struct
type SATASpeedInfo struct {
	SATAValue      int64  `json:"sata_value,omitempty"`
	String         string `json:"string,omitempty"`
	UnitsPerSecond int64  `json:"units_per_second,omitempty"`
	BitsPerUnit    int64  `json:"bits_per_unit,omitempty"`
}

// DiskCheckResult struct
type DiskCheckResult struct {
	Device       *DeviceInfo       `json:"device,omitempty"`
	UserCapacity *UserCapacityInfo `json:"user_capacity,omitempty"`
	DeviceType   *DeviceTypeInfo   `json:"device_type,omitempty"`

	Vendor          string          `json:"vendor,omitempty"`
	Product         string          `json:"product,omitempty"`
	ModelName       string          `json:"model_name,omitempty"`
	Revision        string          `json:"revision,omitempty"`
	SerailNumber    string          `json:"serial_number,omitempty"`
	FirmwareVersion string          `json:"firmware_version,omitempty"`
	RotationRate    int64           `json:"rotation_rate,omitempty"`
	FormFactor      *FormFactorInfo `json:"form_factor,omitempty"`

	// for PCI disk like NVMe
	PCIVendor *NVMePCIVendorInfo `json:"nvme_pci_vendor,omitempty"`

	// for ATA/SATA disk, like SSD
	ATAVersion     *ATAVersionInfo     `json:"ata_version,omitempty"`
	SATAVersion    *SATAVersionInfo    `json:"sata_version,omitempty"`
	InterfaceSpeed *InterfaceSpeedInfo `json:"interface_speed,omitempty"`

	localstoragev1alpha1.SmartCheck
}

// IsVirtualDisk check if it's a virtual disk
func (d DiskCheckResult) IsVirtualDisk() bool {
	return d.Product == localstoragev1alpha1.SmartCtlDeviceProductVirtualDisk
}
