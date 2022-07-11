package manager

import (
	"os"

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
}

// GenerateUUID
func (disk DiskInfo) GenerateUUID() string {
	elementSet := disk.Attribute.Serial + disk.Attribute.Model + disk.Attribute.Vendor + disk.Attribute.WWN

	vitualDiskModels := []string{"EphemeralDisk", "QEMU_HARDDISK", "Virtual_disk"}
	for _, virtualModel := range vitualDiskModels {
		if virtualModel == disk.Attribute.Model {
			host, _ := os.Hostname()
			elementSet += host + disk.Attribute.DevName
		}
	}
	return utils.Hash(elementSet)
}
