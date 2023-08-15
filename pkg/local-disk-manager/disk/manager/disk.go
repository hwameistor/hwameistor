package manager

import (
	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"os"
	"strings"

	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils"
)

// DiskInfo
type DiskInfo struct {
	// DiskIdentify
	DiskIdentify `json:"diskIdentify,omitempty"`

	// Attribute
	Attribute Attribute `json:"attribute,omitempty"`

	// Partition
	Partitions []PartitionInfo `json:"partition,omitempty"`

	// Raid
	Raid RaidInfo `json:"raid,omitempty"`

	// Smart
	Smart SmartInfo `json:"smart"`
}

// GenerateUUID generates a UUID for the disk
// If the serial number exists, it is used first. If it does not exist, it is generated using by-path path
func (disk DiskInfo) GenerateUUID() string {
	// NOTES: in virtual environments, model can be changed after creation filesystem on it e.g. lvm
	var elementSet = disk.Attribute.Vendor
	if disk.Attribute.Serial != "" {
		elementSet += disk.Attribute.Serial + disk.Attribute.WWN
	} else {
		hostName, _ := os.Hostname()

		foundIDPath := false
		for _, devLink := range disk.Attribute.DevLinks {
			if strings.Contains(devLink, v1alpha1.LinkByPath) {
				elementSet += hostName + devLink
				foundIDPath = true
			}
		}

		// don't generate a UUID if the device has no element to identify
		if !foundIDPath {
			return ""
		}
	}

	return utils.Hash(elementSet)
}
