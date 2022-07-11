package manager

// IDiskAttribute
type IDiskAttribute interface {
	ParseDiskAttr() Attribute
}

// AttributeParser
type AttributeParser struct {
	IDiskAttribute
}

// Attribute for disk details
type Attribute struct {
	// DevicePath represents the disk hardware path.
	// The general format is like /sys/devices/pci0000:ae/0000:ae:02.0/0000:b1:00.0/host2/target2:1:0/2:1:0:0/block/sdc/sdc
	DevPath string `json:"devPath,omitempty"`

	// DevName the general format is /dev/sda
	DevName string `json:"devName,omitempty"`

	// DevType such as disk, partition
	DevType string `json:"devType,omitempty"`

	// Major represents drive used by the device
	Major string `json:"major,omitempty"`

	// Minor is used to distinguish different devices
	Minor string `json:"minor,omitempty"`

	// SubSystem identifies the device's system type, such as block
	SubSystem string `json:"subSystem,omitempty"`

	// Bus
	Bus string `json:"id_bus,omitempty"`

	// FS_TYPE
	FSType string `json:"id_fs_type,omitempty"`

	// Model
	Model string `json:"id_model,omitempty"`

	// WWN
	WWN string `json:"id_wwn,omitempty"`

	// PartTableType
	PartTableType string `json:"id_part_table_type,omitempty"`

	// Serial
	Serial string `json:"id_serial,omitempty"`

	// Vendor
	Vendor string `json:"id_vendor,omitempty"`

	// ID_TYPE
	IDType string `json:"id_type"`

	// Capacity of the disk in bytes
	Capacity int64 `json:"capacity,omitempty"`

	// DriverType such as HDD, SSD
	DriverType string `json:"driverType"`
}
