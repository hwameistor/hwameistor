package snapshotclass

import (
	"github.com/spf13/cobra"
)

var SnapshotClass = &cobra.Command{
	Use:   "vsc",
	Args:  cobra.ExactArgs(0),
	Short: "Manage the Hwameistor's storage snapshotclass.",
	Long:  "Manage the Hwameistor's storage snapshotclass.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Root cmd will show help only
		return cmd.Help()
	},
}

func init() {
	// Node sub commands
	SnapshotClass.AddCommand(snapshotclassList)
	SnapshotClass.AddCommand(snapshotclassAdd)
	SnapshotClass.AddCommand(snapshotclassDelete)
}
