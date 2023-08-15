package volume

import (
	"github.com/spf13/cobra"
)

var Volume = &cobra.Command{
	Use:   "volume",
	Short: "Manage the hwameistor's LocalVolumes.",
	Long:  "Manage the hwameistor's LocalVolumes.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// root cmd will show help only
		return cmd.Help()
	},
}

func init() {
	// Volume sub command
	Volume.AddCommand(volumeGet)
}
