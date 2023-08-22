package disk

import (
	"github.com/spf13/cobra"

	"github.com/hwameistor/hwameistor/pkg/hwameictl/manager"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils"
)

var diskOwner = &cobra.Command{
	Use:     "owner {nodeName} {deviceDevPath} {owner}",
	Args:    cobra.ExactArgs(3),
	Short:   "Set the disk's owner.",
	Long:    "Set the disk's owner. The deviceDevPath just like 'sda', 'sdb'.",
	Example: "hwameictl disk owner worker-1 sdc local-disk-manager",
	RunE:    diskOwnerRunE,
}

func diskOwnerRunE(_ *cobra.Command, args []string) error {
	c, err := manager.NewLocalDiskController()
	if err != nil {
		return err
	}
	// Build the disk name like 'master-node-1-sdb'
	diskName := utils.ConvertNodeName(args[0]) + "-" + args[1]
	// Set the disk's owner
	return c.SetLocalDiskOwner(diskName, args[2])
}
