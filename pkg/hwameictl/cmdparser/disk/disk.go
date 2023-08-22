package disk

import (
	"github.com/spf13/cobra"
)

var Disk = &cobra.Command{
	Use:   "disk",
	Args:  cobra.ExactArgs(0),
	Short: "Manage the disks.",
	Long:  "Manage the disks.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Root cmd will show help only
		return cmd.Help()
	},
}

func init() {
	// Disk sub commands
	Disk.AddCommand(diskList, diskOwner, diskReserve)
}
