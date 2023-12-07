package storageclass

import (
	"fmt"
	"github.com/hwameistor/hwameistor/pkg/apiserver/manager/hwameistor"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/manager"
	"github.com/spf13/cobra"
)

//convertible: "false"
//csi.storage.k8s.io/fstype: xfs
//poolClass: HDD
//poolType: REGULAR
//provision-iops-on-creation: "100"
//provision-throughput-on-creation: 1Mi
//replicaNumber: "1"
//striped: "true"
//volumeKind: LVM

var provisioner string
var convertible, striped string
var iops, replicaNumber string
var fstype, throughput, poolClass, poolType, volumeKind, diskType string
var storageClassAdd = &cobra.Command{
	Use:   "add {scName}",
	Args:  cobra.ExactArgs(1),
	Short: "add the Hwameistor's storage storageClasses.",
	Long:  "You can use 'hwameictl sc add' to add hwameistor-storageclass.",
	Example: "hwameictl sc add example-sc \n" +
		"hwameictl sc add example-sc --iops=100 --throughput=1Mi \n" +
		"hwameictl sc add example-sc --replicaNumber=2 \n" +
		"hwameictl sc add example-sc --poolClass=SSD \n" +
		"hwameictl sc add example-sc --provisioner=disk.hwameistor.io --diskType=HDD",
	RunE: storageClassAddRunE,
}

func init() {
	// Volume list flags
	storageClassAdd.Flags().StringVar(&provisioner, "provisioner", "lvm.hwameistor.io",
		"provisioner:{lvm.hwameistor.io,disk.hwameistor.io}")
	storageClassAdd.Flags().StringVar(&convertible, "convertible", "false", "convertible")
	storageClassAdd.Flags().StringVar(&striped, "striped", "true", "striped")
	storageClassAdd.Flags().StringVar(&replicaNumber, "replicaNumber", "1", "replicaNumber:1,2")
	storageClassAdd.Flags().StringVar(&fstype, "fstype", "xfs", "fstype")
	storageClassAdd.Flags().StringVar(&poolClass, "poolClass", "HDD", "poolClass:HDD,SSD,NVMe")
	storageClassAdd.Flags().StringVar(&poolType, "poolType", "REGULAR", "poolType")
	storageClassAdd.Flags().StringVar(&volumeKind, "volumeKind", "LVM", "volumeKind")
	storageClassAdd.Flags().StringVar(&volumeKind, "diskType", "HDD", "diskType")
	storageClassAdd.Flags().StringVar(&iops, "iops", "", "provision-iops-on-creation")
	storageClassAdd.Flags().StringVar(&throughput, "throughput", "", "provision-throughput-on-creation")
}

func storageClassAddRunE(_ *cobra.Command, args []string) error {
	c, err := manager.NewStorageClassController()
	if err != nil {
		return err
	}
	parameters := make(map[string]string)
	if provisioner == hwameistor.DiskHwameistor {
		parameters["diskType"] = diskType
	} else if provisioner == hwameistor.LvmHwameistor {
		m := map[string]string{
			"convertible":   convertible,
			"striped":       striped,
			"replicaNumber": replicaNumber,
			"fstype":        fstype,
			"poolClass":     poolClass,
			"poolType":      poolType,
			"volumeKind":    volumeKind,
		}
		if iops != "" {
			m["provision-iops-on-creation"] = iops
		}
		if throughput != "" {
			m["provision-throughput-on-creation"] = throughput
		}
		parameters = m
	} else {
		return fmt.Errorf("provisioner Only supported: lvm.hwameistor.io, disk.hwameistor.io")
	}

	return c.AddHwameistorStroageClass(args[0], provisioner, parameters)
}
