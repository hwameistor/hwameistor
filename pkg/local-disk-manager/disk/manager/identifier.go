package manager

import "strings"

// DiskIdentify
type DiskIdentify struct {
	// DevPath such as /sys/devices/pci0000:00/0000:00:15.0/0000:03:00.0/host0/target0:0:1/0:0:1:0/block/sda
	DevPath string `json:"devPath"`

	// DevName such as /dev/sda
	DevName string `json:"devName"`

	// Name such as sda
	Name string `json:"name"`
}

// NewDiskIdentify
func NewDiskIdentify(devPath string) *DiskIdentify {
	return &DiskIdentify{DevPath: devPath}
}

func NewDiskIdentifyWithName(devPath, devName string) *DiskIdentify {
	if strings.HasPrefix(devName, "/dev") {
		devName = strings.Replace(devName, "/dev/", "", 1)
	}
	return &DiskIdentify{DevPath: devPath, Name: devName}
}

// SetPath
func (disk *DiskIdentify) SetPath(devPath string) {
	disk.DevPath = devPath
}

// SetName
func (disk *DiskIdentify) SetName(devName string) {
	disk.DevName = devName
}
