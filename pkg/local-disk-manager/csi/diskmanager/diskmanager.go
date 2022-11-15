package diskmanager

type DiskStatus = string

const (
	DiskStatusInUse     DiskStatus = "InUse"
	DiskStatusFree      DiskStatus = "Free"
	DiskStatusReserved  DiskStatus = "Reserved"
	DiskStatusUnclaimed DiskStatus = "Available"
	DiskStatusReleased  DiskStatus = "Released"
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
	DiskType string `json:"diskType,omitempty"`

	// Status
	Status DiskStatus `json:"status,omitempty"`
}

// DiskManager manage all disks in cluster
// The operation here needs to ensure thread safety
type DiskManager interface {
	// GetClusterDisks list all disks by node
	GetClusterDisks() (map[string][]*Disk, error)

	// GetNodeDisks list all disk located on the node
	GetNodeDisks(node string) ([]*Disk, error)

	// ClaimDisk UpdateDiskStatus mark disk to TobeMount/Free/InUse... status
	ClaimDisk(name string) error

	// FilterFreeDisks filter matchable free disks
	FilterFreeDisks([]Disk) (bool, error)

	// ReserveDiskForVolume reserve a disk for the volume
	ReserveDiskForVolume(disk Disk, pvc string) error

	// UnReserveDiskForPVC update related disk to release status
	UnReserveDiskForPVC(pvc string) error

	// ReleaseDisk setup disk to release status
	ReleaseDisk(disk string) error

	// GetReservedDiskByPVC get disk reserved by the volume
	GetReservedDiskByPVC(pvc string) (*Disk, error)
}
