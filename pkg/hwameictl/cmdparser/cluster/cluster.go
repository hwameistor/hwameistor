package cluster

import (
	"github.com/spf13/cobra"
)

var Cluster = &cobra.Command{
	Use:   "cluster",
	Args:  cobra.ExactArgs(0),
	Short: "Manage the cluster.",
	Long:  "Manage the cluster.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Root cmd will show help only
		return cmd.Help()
	},
}

func init() {
	// Cluster sub commands
	Cluster.AddCommand(install, event)
}
