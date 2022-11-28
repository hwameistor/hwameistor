package lsblk

import (
	"fmt"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/disk/manager"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/udev"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

const (
	PartType = "part"
)

type LSBlk struct {
	*manager.DiskIdentify
}

// HasPartition
func (lsb LSBlk) HasPartition() bool {
	log.Infof("Parse disk %v", lsb.DevPath)
	return true
}

// ParsePartitionInfo
func (lsb LSBlk) ParsePartitionInfo() []manager.PartitionInfo {
	log.Debugf("Parse disk %v", lsb.Name)
	if partitions, err := lsb.partitionInfo(); err != nil {
		log.WithError(err).Errorf("Parse partition info fail")
		return nil
	} else {
		return partitions
	}
}

func (lsb LSBlk) partitionInfo() ([]manager.PartitionInfo, error) {
	var devicePath string
	splitDevicePath := strings.Split(lsb.Name, "/")
	if len(splitDevicePath) == 1 || len(splitDevicePath) == 0 {
		devicePath = fmt.Sprintf("/dev/%s", lsb.Name)
	} else {
		devicePath = lsb.Name
	}

	output, err := utils.Bash(fmt.Sprintf("lsblk %s --bytes --pairs --output NAME,SIZE,TYPE,PKNAME,FSTYPE", devicePath))
	if err != nil {
		return nil, err
	}

	var partitions []manager.PartitionInfo
	for _, item := range utils.ConvertShellOutputs(output) {
		props := utils.ParseKeyValuePairString(item)
		switch props["NAME"] {
		case lsb.Name:
			if props["FSTYPE"] != "" {
				partitions = append(partitions, manager.PartitionInfo{
					Name:       lsb.Name,
					Filesystem: props["FSTYPE"],
				})
			}

		default:
			if !(props["PKNAME"] == lsb.Name && props["TYPE"] == PartType) {
				continue
			}

			p := manager.PartitionInfo{Name: props["NAME"]}
			p.Size, err = strconv.ParseUint(props["SIZE"], 10, 64)
			if err != nil {
				return nil, err
			}

			device := udev.NewDeviceWithName("", fmt.Sprintf("/dev/%s", props["NAME"]))
			if err = device.ParseDeviceInfo(); err != nil {
				return nil, err
			}

			p.Label = device.PartName
			p.Filesystem = device.FSType

			partitions = append(partitions, p)
		}
	}

	return partitions, nil
}

// NewPartitionParser
func NewPartitionParser(disk *manager.DiskIdentify) *manager.PartitionParser {
	return &manager.PartitionParser{
		IPartition: LSBlk{disk},
	}
}
