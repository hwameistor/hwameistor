package localvolumesnapshot

import (
	"github.com/spf13/cobra"
)

var LocalSnapshot = &cobra.Command{
	Use:   "lvs",
	Args:  cobra.ExactArgs(0),
	Short: "Manage the Hwameistor's storage LocalVolumeSnapshot.",
	Long:  "Manage the Hwameistor's storage LocalVolumeSnapshot.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Root cmd will show help only
		return cmd.Help()
	},
}

func init() {
	// Node sub commands
	LocalSnapshot.AddCommand(localVolumeSnapshotList)
	LocalSnapshot.AddCommand(localVolumeSnapshotRollback)
}
