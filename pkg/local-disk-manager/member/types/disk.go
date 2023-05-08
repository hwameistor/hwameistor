package types

type DiskStatus = string

const (
	DiskStatusAvailable DiskStatus = "Available"
)

type DevType = string

const (
	DevTypeHDD  DevType = "HDD"
	DevTypeSSD  DevType = "SSD"
	DevTypeNVMe DevType = "NVMe"
)

// Disk all disk info about a disk
type Disk struct {
	// AttachNode represent where disk is attached
	AttachNode string `json:"attachNode,omitempty"`

	// Name unique identification for a disk
	Name string `json:"name,omitempty"`

	// DevPath
	DevPath string `json:"devPath,omitempty"`

	// Capacity
	Capacity int64 `json:"capacity,omitempty"`

	// DiskType SSD/HDD/NVME...
	DiskType DevType `json:"diskType,omitempty"`

	// Status
	Status DiskStatus `json:"status,omitempty"`
}
