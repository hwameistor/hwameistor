package udev

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils"
)

type Device struct {
	// DevPath represents the disk hardware path.
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

	// Bus represents the bus type of the device, such as USB, SATA
	Bus string `json:"id_bus,omitempty"`

	// FSType represents the filesystem type such as ext4, ntfs
	FSType string `json:"id_fs_type,omitempty"`

	// Model represents the specific model of the storage device, usually specified by the manufacturer
	Model string `json:"id_model,omitempty"`

	// WWN represents the World Wide Name(WWN) of the device.
	// The general format is like 5001b444a89e5acd.
	WWN string `json:"id_wwn,omitempty"`

	// PartTableType represents the partition table type, such as gpt or mbr
	PartTableType string `json:"id_part_table_type,omitempty"`

	// Serial represents the Serial Number(SN) of the device.
	// The general format is like 162061400553
	Serial string `json:"id_serial,omitempty"`

	// Vendor represents the manufacturer of the device
	Vendor string `json:"id_vendor,omitempty"`

	// IDType specifies the detailed type of the device according to udev rules, such as 'cd', 'disk', or 'partition',
	// providing a finer classification than DevType, usually the values of IDType and DevType are the same
	IDType string `json:"id_type"`

	// PartName such as EFI System Partition
	PartName string `json:"partName"`

	// Name is the name of the device node sda, sdb, dm-0 etc
	Name string `json:"name"`

	// DevLinks is a symbolic link array for the device, containing all symbolic links for the device
	DevLinks []string `json:"devLinks"`
}

func NewDevice(devPath string) *Device {
	return &Device{DevPath: devPath}
}

func NewDeviceWithName(devPath, devName string) *Device {
	return &Device{DevName: devName, DevPath: devPath}
}

// FilterDisk filter out disks that are virtual or can't identify themselves
func (d *Device) FilterDisk() bool {
	// disk with no identity will be filter out
	if d.Serial == "" {
		foundIDLink := false
		for _, devLink := range d.DevLinks {
			// by-path symlink will be used to identify disk when serial is empty, this mustn't be empty!
			if strings.Contains(devLink, v1alpha1.LinkByPath) {
				foundIDLink = true
				break
			}
		}

		if !foundIDLink {
			return false
		}
	}

	// virtual block device like loop device will be filter out
	if strings.Contains(d.DevPath, "/virtual/") {
		return false
	}

	// For some disk(ex AliCloud HDD Disk), IDType may be empty
	return (d.IDType == "disk" || d.IDType == "") &&
		d.DevType == "disk"
}

func (d *Device) ParseDeviceInfo() error {
	info, err := d.Info()
	if err != nil {
		return err
	}
	return d.ParseDiskAttribute(info)
}

// Info gets detailed information about the device using udevadm
func (d *Device) Info() (map[string]interface{}, error) {
	var out string
	var err error
	if d.DevPath != "" {
		out, err = utils.Bash(fmt.Sprintf("udevadm info -p %s --query=all", d.DevPath))
	} else {
		out, err = utils.Bash(fmt.Sprintf("udevadm info -n %s --query=all", d.DevName))
	}

	if err != nil {
		return nil, err
	}
	return parseUdevInfo(out), nil
}

func (d *Device) ParseDiskAttribute(info map[string]interface{}) error {
	// Why do we need to convert the map information into JSON data
	// instead of directly converting the map into a structure
	//
	// The main reason is that if the udev field is converted into a structure,
	// each key in the structure must be consistent with the udev information,
	// which will make the disk DiskAttribute structure information difficult to understand
	if idType, ok := info["DEVTYPE"]; ok && idType == "partition" {
		if _, ok := info["NAME"]; ok {
			d.PartName = info["NAME"].(string)
		}
	}
	jsonStr, err := json.Marshal(info)
	if err != nil {
		return err
	}

	return json.Unmarshal(jsonStr, d)
}

func parseUdevInfo(udevInfo string) map[string]interface{} {
	udevItems := make(map[string]interface{})
	for _, info := range utils.ConvertShellOutputs(udevInfo) {
		if info == "" {
			continue
		}

		switch info[0] {
		// ENV
		case 'E':
			items := strings.Split(strings.Replace(info, "E: ", "", 1), "=")
			if len(items) != 2 {
				continue
			}
			if items[0] == "DEVLINKS" {
				udevItems[items[0]] = strings.Split(items[1], " ")
				continue
			}
			udevItems[items[0]] = items[1]

		case 'N':
			info = strings.Replace(info, "N: ", "", 1)
			udevItems["NAME"] = info

		default:
			continue
		}
	}

	return udevItems
}
