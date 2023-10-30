package udev

import (
	"reflect"
	"testing"
)

func TestDevice_ParseDeviceInfo(t *testing.T) {
	testCases := []struct {
		Description  string
		UdevInfo     string
		ExpectDevice *Device
	}{
		{
			Description: "It is a unPartition udevadm info result, should return empty string",
			UdevInfo:    "P: /devices/pci0000:00/0000:00:07.1/ata1/host1/target1:0:0/1:0:0:0/block/sdb\nN: sdb\nS: disk/by-id/ata-VMware_Virtual_IDE_Hard_Drive_00000000000000000001\nS: disk/by-id/wwn-0x5000c298825951d9\nS: disk/by-path/pci-0000:00:07.1-ata-1.0\nE: DEVLINKS=/dev/disk/by-id/ata-VMware_Virtual_IDE_Hard_Drive_00000000000000000001 /dev/disk/by-id/wwn-0x5000c298825951d9 /dev/disk/by-path/pci-0000:00:07.1-ata-1.0\nE: DEVNAME=/dev/sdb\nE: DEVPATH=/devices/pci0000:00/0000:00:07.1/ata1/host1/target1:0:0/1:0:0:0/block/sdb\nE: DEVTYPE=disk\nE: ID_ATA=1\nE: ID_ATA_FEATURE_SET_APM=1\nE: ID_ATA_FEATURE_SET_APM_ENABLED=0\nE: ID_ATA_FEATURE_SET_PM=1\nE: ID_ATA_FEATURE_SET_PM_ENABLED=1\nE: ID_BUS=ata\nE: ID_MODEL=VMware_Virtual_IDE_Hard_Drive\nE: ID_MODEL_ENC=VMware\\x20Virtual\\x20IDE\\x20Hard\\x20Drive\\x20\\x20\\x20\\x20\\x20\\x20\\x20\\x20\\x20\\x20\\x20\nE: ID_PART_TABLE_TYPE=dos\nE: ID_PATH=pci-0000:00:07.1-ata-1.0\nE: ID_PATH_TAG=pci-0000_00_07_1-ata-1_0\nE: ID_REVISION=00000001\nE: ID_SERIAL=VMware_Virtual_IDE_Hard_Drive_00000000000000000001\nE: ID_SERIAL_SHORT=00000000000000000001\nE: ID_TYPE=disk\nE: ID_WWN=0x5000c298825951d9\nE: ID_WWN_WITH_EXTENSION=0x5000c298825951d9\nE: MAJOR=8\nE: MINOR=16\nE: SUBSYSTEM=block\nE: TAGS=:systemd:\nE: USEC_INITIALIZED=645583\n",
			ExpectDevice: &Device{
				DevPath:       "/devices/pci0000:00/0000:00:07.1/ata1/host1/target1:0:0/1:0:0:0/block/sdb",
				DevName:       "/dev/sdb",
				DevType:       "disk",
				Major:         "8",
				Minor:         "16",
				SubSystem:     "block",
				Bus:           "ata",
				FSType:        "",
				Model:         "VMware_Virtual_IDE_Hard_Drive",
				WWN:           "0x5000c298825951d9",
				PartTableType: "dos",
				Serial:        "VMware_Virtual_IDE_Hard_Drive_00000000000000000001",
				Vendor:        "",
				IDType:        "disk",
				PartName:      "",
				Name:          "sdb",
				DevLinks: []string{
					"/dev/disk/by-id/ata-VMware_Virtual_IDE_Hard_Drive_00000000000000000001",
					"/dev/disk/by-id/wwn-0x5000c298825951d9",
					"/dev/disk/by-path/pci-0000:00:07.1-ata-1.0",
				},
			},
		},
		{
			Description: "It is a partition udevadm info result ,should return the partition name sdb1",
			UdevInfo:    "P: /devices/pci0000:00/0000:00:07.1/ata1/host1/target1:0:0/1:0:0:0/block/sdb/sdb1\nN: sdb1\nS: disk/by-id/ata-VMware_Virtual_IDE_Hard_Drive_00000000000000000001-part1\nS: disk/by-id/lvm-pv-uuid-QKLBih-32Tt-0JP7-uQKr-GzjD-cx0C-I8Vnno\nS: disk/by-id/wwn-0x5000c298825951d9-part1\nS: disk/by-path/pci-0000:00:07.1-ata-1.0-part1\nE: DEVLINKS=/dev/disk/by-id/ata-VMware_Virtual_IDE_Hard_Drive_00000000000000000001-part1 /dev/disk/by-id/lvm-pv-uuid-QKLBih-32Tt-0JP7-uQKr-GzjD-cx0C-I8Vnno /dev/disk/by-id/wwn-0x5000c298825951d9-part1 /dev/disk/by-path/pci-0000:00:07.1-ata-1.0-part1\nE: DEVNAME=/dev/sdb1\nE: DEVPATH=/devices/pci0000:00/0000:00:07.1/ata1/host1/target1:0:0/1:0:0:0/block/sdb/sdb1\nE: DEVTYPE=partition\nE: ID_ATA=1\nE: ID_ATA_FEATURE_SET_APM=1\nE: ID_ATA_FEATURE_SET_APM_ENABLED=0\nE: ID_ATA_FEATURE_SET_PM=1\nE: ID_ATA_FEATURE_SET_PM_ENABLED=1\nE: ID_BUS=ata\nE: ID_FS_TYPE=LVM2_member\nE: ID_FS_USAGE=raid\nE: ID_FS_UUID=QKLBih-32Tt-0JP7-uQKr-GzjD-cx0C-I8Vnno\nE: ID_FS_UUID_ENC=QKLBih-32Tt-0JP7-uQKr-GzjD-cx0C-I8Vnno\nE: ID_FS_VERSION=LVM2 001\nE: ID_MODEL=VMware_Virtual_IDE_Hard_Drive\nE: ID_MODEL_ENC=VMware\\x20Virtual\\x20IDE\\x20Hard\\x20Drive\\x20\\x20\\x20\\x20\\x20\\x20\\x20\\x20\\x20\\x20\\x20\nE: ID_PART_ENTRY_DISK=8:16\nE: ID_PART_ENTRY_NUMBER=1\nE: ID_PART_ENTRY_OFFSET=2048\nE: ID_PART_ENTRY_SCHEME=dos\nE: ID_PART_ENTRY_SIZE=31455232\nE: ID_PART_ENTRY_TYPE=0x83\nE: ID_PART_TABLE_TYPE=dos\nE: ID_PATH=pci-0000:00:07.1-ata-1.0\nE: ID_PATH_TAG=pci-0000_00_07_1-ata-1_0\nE: ID_REVISION=00000001\nE: ID_SERIAL=VMware_Virtual_IDE_Hard_Drive_00000000000000000001\nE: ID_SERIAL_SHORT=00000000000000000001\nE: ID_TYPE=disk\nE: ID_WWN=0x5000c298825951d9\nE: ID_WWN_WITH_EXTENSION=0x5000c298825951d9\nE: MAJOR=8\nE: MINOR=17\nE: SUBSYSTEM=block\nE: SYSTEMD_ALIAS=/dev/block/8:17\nE: SYSTEMD_READY=1\nE: SYSTEMD_WANTS=lvm2-pvscan@8:17.service\nE: TAGS=:systemd:",
			ExpectDevice: &Device{
				DevPath:       "/devices/pci0000:00/0000:00:07.1/ata1/host1/target1:0:0/1:0:0:0/block/sdb/sdb1",
				DevName:       "/dev/sdb1",
				DevType:       "partition",
				Major:         "8",
				Minor:         "17",
				SubSystem:     "block",
				Bus:           "ata",
				FSType:        "LVM2_member",
				Model:         "VMware_Virtual_IDE_Hard_Drive",
				WWN:           "0x5000c298825951d9",
				PartTableType: "dos",
				Serial:        "VMware_Virtual_IDE_Hard_Drive_00000000000000000001",
				Vendor:        "",
				IDType:        "disk",
				PartName:      "sdb1",
				Name:          "sdb1",
				DevLinks: []string{
					"/dev/disk/by-id/ata-VMware_Virtual_IDE_Hard_Drive_00000000000000000001-part1",
					"/dev/disk/by-id/lvm-pv-uuid-QKLBih-32Tt-0JP7-uQKr-GzjD-cx0C-I8Vnno",
					"/dev/disk/by-id/wwn-0x5000c298825951d9-part1",
					"/dev/disk/by-path/pci-0000:00:07.1-ata-1.0-part1",
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Description, func(t *testing.T) {
			info := parseUdevInfo(testCase.UdevInfo)
			d := &Device{}
			d.ParseDiskAttribute(info)
			// judge current device completely same as expect
			if !reflect.DeepEqual(d, testCase.ExpectDevice) {
				t.Fatal("Device should be the same with the ExpectDevice")
			}
		})
	}
}
