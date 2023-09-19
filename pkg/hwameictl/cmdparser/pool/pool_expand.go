package pool

import (
	"fmt"

	"github.com/spf13/cobra"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/manager"
)

var poolExpand = &cobra.Command{
	Use:   "expand {nodeName} {diskType} {owner}",
	Args:  cobra.ExactArgs(3),
	Short: "Expand the available storage capacity of this node by claiming a new disk.",
	Long: "Expand the available storage capacity of this node by claiming a new disk.\n" +
		"The diskType must be SSD or HDD or NVMe, and owner must be" +
		fmt.Sprintf("%s or %s.", apisv1alpha1.LocalStorage, apisv1alpha1.LocalDiskManager),
	Example: "hwameictl pool expand worker-1 HDD local-storage\n" +
		"hwameictl pool expand worker-2 SSD local-disk-manager",
	RunE: poolExpandRunE,
}

func poolExpandRunE(_ *cobra.Command, args []string) error {
	nodeName, diskType, owner := args[0], args[1], args[2]
	// Check the parameters

	if diskType != apisv1alpha1.DiskClassNameSSD && diskType != apisv1alpha1.DiskClassNameHDD && diskType != apisv1alpha1.DiskClassNameNVMe {
		return fmt.Errorf("diskType must be %s or %s or %s", apisv1alpha1.DiskClassNameSSD, apisv1alpha1.DiskClassNameHDD, apisv1alpha1.DiskClassNameNVMe)
	}
	if owner != apisv1alpha1.LocalStorage && owner != apisv1alpha1.LocalDiskManager {
		return fmt.Errorf("owner must be %s or %s", apisv1alpha1.LocalStorage, apisv1alpha1.LocalDiskManager)
	}

	c, err := manager.NewLocalStoragePoolController()
	if err != nil {
		return err
	}

	return c.ExpandStoragePool(nodeName, diskType, owner)
}
