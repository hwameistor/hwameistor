package pool

import (
	"github.com/spf13/cobra"
)

var Pool = &cobra.Command{
	Use:   "pool",
	Args:  cobra.ExactArgs(0),
	Short: "Manage the Hwameistor's storage pools.",
	Long:  "Manage the Hwameistor's storage pools.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Root cmd will show help only
		return cmd.Help()
	},
}

func init() {
	// Pool sub commands
	Pool.AddCommand(poolList, poolExpand)
}
