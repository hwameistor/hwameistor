package sys

import (
	"fmt"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// Device represents a blockdevice using its sysfs path.
type Device struct {
	// deviceName is the name of the device node sda, sdb, dm-0 etc
	deviceName string

	// Path of the blockdevice. eg: /dev/sda, /dev/dm-0
	path string

	// SysPath of the blockdevice. eg: /sys/devices/pci0000:00/0000:00:1f.2/ata1/host0/target0:0:0/0:0:0:0/block/sda/
	sysPath string
}

// NewDevice
func NewDevice(sysPath string, path string, devName string) *Device {
	if !strings.HasSuffix(sysPath, "/") {
		sysPath = sysPath + "/"
	}
	return &Device{sysPath: sysPath, path: path, deviceName: devName}
}

const (
	// DriveTypeHDD represents a rotating hard disk drive
	DriveTypeHDD = "HDD"

	// DriveTypeSSD represents a solid state drive
	DriveTypeSSD = "SSD"

	// DriveTypeUnknown is used when the drive type of the disk could not be determined.
	DriveTypeUnknown = "Unknown"
)

const (
	// BlockSubSystem is the key used to represent block subsystem in sysfs
	BlockSubSystem = "block"
	// NVMeSubSystem is the key used to represent nvme subsystem in sysfs
	NVMeSubSystem = "nvme"
	// sectorSize is the sector size as understood by the unix systems.
	// It is kept as 512 bytes. all entries in /sys/class/block/sda/size
	// are in 512 byte blocks
	sectorSize int64 = 512
)

// getParent gets the parent of this device if it has parent
func (s Device) getParent() (string, bool) {
	parts := strings.Split(s.sysPath, "/")

	var parentBlockDevice string
	ok := false

	// checking for block subsystem, return the next part after subsystem only
	// if the length is greater. This check is to avoid an index out of range panic.
	for i, part := range parts {
		if part == BlockSubSystem {
			// check if the length is greater to avoid panic. Also need to make sure that
			// the same device is not returned if the given device is a parent.
			if len(parts)-1 >= i+1 && s.deviceName != parts[i+1] {
				ok = true
				parentBlockDevice = parts[i+1]
			}
			return parentBlockDevice, ok
		}
	}

	// checking for nvme subsystem, return the 2nd item in hierarchy, which will be the
	// nvme namespace. Length checking is to avoid index out of range in case of malformed
	// links (extremely rare case)
	for i, part := range parts {
		if part == NVMeSubSystem {
			// check if the length is greater to avoid panic. Also need to make sure that
			// the same device is not returned if the given device is a parent.
			if len(parts)-1 >= i+2 && s.deviceName != parts[i+2] {
				ok = true
				parentBlockDevice = parts[i+2]
			}
			return parentBlockDevice, ok
		}
	}

	return parentBlockDevice, ok
}

// getPartitions gets the partitions of this device if it has any
func (s Device) GetPartitions() ([]string, bool) {

	// if partition file has value 0, or the file doesn't exist,
	// can return from there itself
	// partitionPath := s.SysPath + "partition"
	// if _, err := os.Stat(partitionPath); os.IsNotExist(err) {
	// }

	partitions := make([]string, 0)

	files, err := ioutil.ReadDir(s.sysPath)
	if err != nil {
		return nil, false
	}
	for _, file := range files {
		if strings.HasPrefix(file.Name(), s.deviceName) {
			partitions = append(partitions, file.Name())
		}
	}

	return partitions, true
}

// getHolders gets the devices that are held by this device
func (s Device) getHolders() ([]string, bool) {
	holderPath := s.sysPath + "holders/"
	holders := make([]string, 0)

	// check if holders are available for this device
	if _, err := os.Stat(holderPath); os.IsNotExist(err) {
		return nil, false
	}

	files, err := ioutil.ReadDir(holderPath)
	if err != nil {
		return nil, false
	}

	for _, file := range files {
		holders = append(holders, file.Name())
	}
	return holders, true
}

// getSlaves gets the devices to which this device is a slave. Or, the devices
// which holds this device
func (s Device) getSlaves() ([]string, bool) {
	slavePath := s.sysPath + "slaves/"
	slaves := make([]string, 0)

	// check if slaves are available for this device
	if _, err := os.Stat(slavePath); os.IsNotExist(err) {
		return nil, false
	}

	files, err := ioutil.ReadDir(slavePath)
	if err != nil {
		return nil, false
	}

	for _, file := range files {
		slaves = append(slaves, file.Name())
	}
	return slaves, true
}

// GetLogicalBlockSize gets the logical block size, the caller should handle if 0 LB size is returned
func (s Device) GetLogicalBlockSize() (int64, error) {
	logicalBlockSize, err := utils.ReadSysFSFileAsInt64(s.sysPath + "queue/logical_block_size")
	if err != nil {
		return 0, err
	}
	return logicalBlockSize, nil
}

// GetPhysicalBlockSize gets the physical block size of the device
func (s Device) GetPhysicalBlockSize() (int64, error) {
	physicalBlockSize, err := utils.ReadSysFSFileAsInt64(s.sysPath + "queue/physical_block_size")
	if err != nil {
		return 0, err
	}
	return physicalBlockSize, nil
}

// GetHardwareSectorSize gets the hardware sector size of the device
func (s Device) GetHardwareSectorSize() (int64, error) {
	hardwareSectorSize, err := utils.ReadSysFSFileAsInt64(s.sysPath + "queue/hw_sector_size")
	if err != nil {
		return 0, err
	}
	return hardwareSectorSize, nil
}

// GetDriveType gets the drive type of the device based on the rotational value. Can be HDD or SSD
func (s Device) GetDriveType() (string, error) {
	rotational, err := utils.ReadSysFSFileAsInt64(s.sysPath + "queue/rotational")
	if err != nil {
		return DriveTypeUnknown, err
	}

	if rotational == 1 {
		return DriveTypeHDD, nil
	} else if rotational == 0 {
		return DriveTypeSSD, nil
	}
	return DriveTypeUnknown, fmt.Errorf("undefined rotational value %d", rotational)
}

// GetCapacityInBytes gets the capacity of the device in bytes
func (s Device) GetCapacityInBytes() (int64, error) {
	// The size (/size) entry returns the `nr_sects` field of the block device structure.
	// Ref: https://elixir.bootlin.com/linux/v4.4/source/fs/block_dev.c#L1267
	//
	// Traditionally, in Unix disk size contexts, “sector” or “block” means 512 bytes,
	// regardless of what the manufacturer of the underlying hardware might call a “sector” or “block”
	// Ref: https://elixir.bootlin.com/linux/v4.4/source/fs/block_dev.c#L487
	//
	// Therefore, to get the capacity of the device it needs to always multiplied with 512
	numberOfBlocks, err := utils.ReadSysFSFileAsInt64(s.sysPath + "size")
	if err != nil {
		return 0, err
	} else if numberOfBlocks == 0 {
		return 0, fmt.Errorf("block count reported as zero")
	}
	return numberOfBlocks * sectorSize, nil

}

// GetDeviceType gets the device type, as shown in lsblk
// devtype should be prefilled by udev probe (DEVTYPE) as disk/part for this to work
//
// Ported from https://github.com/karelzak/util-linux/blob/master/misc-utils/lsblk.c
func (s Device) GetDeviceType(devType string) (string, error) {

	var result string

	if devType == BlockDeviceTypePartition {
		return BlockDeviceTypePartition, nil
	}

	// TODO may need to distinguish between normal partitions and partitions on DM devices. The original
	//  lsblk implementation does not have this distinction.
	if isDM(s.deviceName) {
		dmUuid, err := utils.ReadSysFSFileAsString(s.sysPath + "dm/uuid")
		if err != nil {
			return "", fmt.Errorf("unable to get DM_UUID, error: %v", err)
		}
		if len(dmUuid) > 0 {
			dmUuidPrefix := strings.Split(dmUuid, "-")[0]
			if len(dmUuidPrefix) != 0 {
				if len(dmUuidPrefix) > 4 && dmUuidPrefix[0:4] == "part" {
					result = BlockDeviceTypePartition
				} else {
					result = dmUuidPrefix
				}
			}
		}
		if len(result) == 0 {
			result = BlockDeviceTypeDMDevice
		}
	} else if len(s.deviceName) >= 4 && s.deviceName[0:4] == "loop" {
		result = BlockDeviceTypeLoop
	} else if len(s.deviceName) >= 2 && s.deviceName[0:2] == "md" {
		mdLevel, err := utils.ReadSysFSFileAsString(s.sysPath + "md/level")
		if err != nil {
			return "", fmt.Errorf("unable to get raid level, error: %v", err)
		}
		if len(mdLevel) != 0 {
			result = mdLevel
		} else {
			result = "md"
		}
	} else {
		// TODO Ideally should read device/type file and find the device type using blkdev_scsi_type_to_name()
		result = "disk"
	}
	return strings.ToLower(result), nil
}

func isDM(devName string) bool {
	return devName[0:3] == "dm-"
}

// NewSysFsDeviceFromDevPath is used to get sysfs device struct from the device devpath
// The sysfs device struct contains the device name along with the syspath
func NewSysFsDeviceFromDevPath(devPath string) (*Device, error) {
	devName := strings.Replace(devPath, "/dev/", "", 1)
	if len(devName) == 0 {
		return nil, fmt.Errorf("unable to create sysfs device from devPath for device: %s, error: device name empty", devPath)
	}

	sysPath, err := getDeviceSysPath(devPath)
	if err != nil {
		return nil, fmt.Errorf("unable to create sysfs device from devpath for device: %s, error: %v", devPath, err)
	}

	dev := &Device{
		deviceName: devName,
		path:       devPath,
		sysPath:    sysPath,
	}
	return dev, nil
}

var sysFSDirectoryPath = "/sys/"

// getDeviceSysPath gets the syspath struct for the given blockdevice.
// It is generated by evaluating the symlink in /sys/class/block.
func getDeviceSysPath(devicePath string) (string, error) {

	var blockDeviceSymLink string

	if strings.HasPrefix(devicePath, "/dev/") {
		blockDeviceName := strings.Replace(devicePath, "/dev/", "", 1)
		blockDeviceSymLink = sysFSDirectoryPath + "class/block/" + blockDeviceName
	} else {
		blockDeviceSymLink = devicePath
	}
	// after evaluation the syspath we get will be similar to
	// /sys/devices/pci0000:00/0000:00:1f.2/ata1/host0/target0:0:0/0:0:0:0/block/sda/
	sysPath, err := filepath.EvalSymlinks(blockDeviceSymLink)
	if err != nil {
		return "", err
	}

	return sysPath + "/", nil
}

const (
	// SparseBlockDeviceType is the sparse blockdevice type
	SparseBlockDeviceType = "sparse"
	// BlockDeviceType is the type for blockdevice.
	BlockDeviceType = "blockdevice"

	// BlockDevicePrefix is the prefix used in UUIDs
	BlockDevicePrefix = BlockDeviceType + "-"

	// The following blockdevice types correspond to the types as seen by the host system
	// BlockDeviceTypeDisk represents a disk type
	BlockDeviceTypeDisk = "disk"

	// BlockDeviceTypePartition represents a partition
	BlockDeviceTypePartition = "partition"

	// BlockDeviceTypeLoop represents a loop device
	BlockDeviceTypeLoop = "loop"

	// BlockDeviceTypeDMDevice is a dm device
	BlockDeviceTypeDMDevice = "dm"

	// BlockDeviceTypeLVM is an lvm device type
	BlockDeviceTypeLVM = "lvm"

	// BlockDeviceTypeCrypt is a LUKS volume
	BlockDeviceTypeCrypt = "crypt"

	// BlockDeviceTypeMultiPath is a multipath device
	BlockDeviceTypeMultiPath = "mpath"
)
