package storageclass

import (
	"github.com/spf13/cobra"
)

var StorageClass = &cobra.Command{
	Use:   "sc",
	Args:  cobra.ExactArgs(0),
	Short: "Manage the Hwameistor's storage StorageClass.",
	Long:  "Manage the Hwameistor's storage StorageClass.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Root cmd will show help only
		return cmd.Help()
	},
}

func init() {
	// Node sub commands
	StorageClass.AddCommand(storageClassList)
	StorageClass.AddCommand(storageClassAdd)
	StorageClass.AddCommand(storageClassDelete)
}
