package lsblk

import (
	"encoding/json"
	"fmt"
	"github.com/hwameistor/hwameistor/pkg/local-storage/utils"
	log "github.com/sirupsen/logrus"

	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/disk/manager"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/udev"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils/sys"
)

type AttributeParser struct {
	*manager.DiskIdentify
}

// ParseDiskAttr
func (ap AttributeParser) ParseDiskAttr() manager.Attribute {
	// get basic info by udev
	uDevice := udev.NewDevice(ap.DevPath)
	if err := uDevice.ParseDeviceInfo(); err != nil {
		log.WithError(err).Errorf("Parse device by udev fail")
		return manager.Attribute{}
	}

	device := sys.NewDevice(ap.DevPath, uDevice.DevName, uDevice.Name)
	diskAttr := manager.Attribute{
		DevPath:   ap.DevPath,
		DevName:   uDevice.DevName,
		Major:     uDevice.Major,
		Minor:     uDevice.Minor,
		SubSystem: uDevice.SubSystem,
		Bus:       uDevice.Bus,
		FSType:    uDevice.FSType,
		Model:     uDevice.Model,
		WWN:       uDevice.WWN,
		Serial:    uDevice.Serial,
		Vendor:    uDevice.Vendor,
		IDType:    uDevice.IDType,
		DevLinks:  uDevice.DevLinks,
	}

	// Parse disk capacity
	if capacity, err := device.GetCapacityInBytes(); err != nil {
		log.WithError(err).Errorf("Parse disk %v capacity fail", ap.DevPath)
	} else {
		diskAttr.Capacity = capacity
	}

	// Parse disk type
	if devType, err := device.GetDeviceType(""); err != nil {
		log.WithError(err).Errorf("Parse disk %v type fail", ap.DevPath)
	} else {
		diskAttr.DevType = devType
	}

	// Parse driver type
	if driverType, err := device.GetDriveType(); err != nil {
		log.WithError(err).Errorf("Parse disk %v driver type fail", ap.DevPath)
	} else {
		diskAttr.DriverType = driverType
	}

	return diskAttr
}

// NewAttributeParser
func NewAttributeParser(disk *manager.DiskIdentify) *manager.AttributeParser {
	return &manager.AttributeParser{
		IDiskAttribute: AttributeParser{disk},
	}
}

// ListAllBlockDevices returns a list of all block devices converted from lsblk outputs
func ListAllBlockDevices() ([]manager.Attribute, error) {
	// internal use only for convert lsblk output
	type _lsblkJsonFormat struct {
		blockdevices []struct {
			Name   string `json:"name"`
			FSType string `json:"fstype"`
			Serial string `json:"serial"`
		}
	}

	var lo _lsblkJsonFormat
	output, err := utils.Bash(fmt.Sprintf("lsblk -Jbdo NAME,FSTYPE,SERIAL"))
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal([]byte(output), &lo); err != nil {
		return nil, err
	}

	var result []manager.Attribute
	for _, blockDevice := range lo.blockdevices {
		result = append(result, manager.Attribute{
			DevName: blockDevice.Name,
			FSType:  blockDevice.FSType,
			Serial:  blockDevice.Serial,
		})
	}

	return result, nil
}
