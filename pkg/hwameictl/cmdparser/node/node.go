package node

import (
	"github.com/spf13/cobra"
)

var Node = &cobra.Command{
	Use:   "node",
	Args:  cobra.ExactArgs(0),
	Short: "Manage the Hwameistor's storage nodes.",
	Long:  "Manage the Hwameistor's storage nodes.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Root cmd will show help only
		return cmd.Help()
	},
}

func init() {
	// Node sub commands
	Node.AddCommand(nodeGet, nodeList, nodeEnable)
}
