package snapshot

import (
	"github.com/spf13/cobra"
)

var Snapshot = &cobra.Command{
	Use:   "vs",
	Args:  cobra.ExactArgs(0),
	Short: "Manage the Hwameistor's storage VolumeSnapshot.",
	Long:  "Manage the Hwameistor's storage VolumeSnapshot.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Root cmd will show help only
		return cmd.Help()
	},
}

func init() {
	// Node sub commands
	Snapshot.AddCommand(snapshotList)
	Snapshot.AddCommand(snapshotAdd)
	Snapshot.AddCommand(snapshotDelete)
	Snapshot.AddCommand(SnapshotRestore)
	Snapshot.AddCommand(SnapshotRollback)
}
