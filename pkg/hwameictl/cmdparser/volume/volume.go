package volume

import (
	"github.com/hwameistor/hwameistor/pkg/hwameictl/cmdparser/definitions"
	"github.com/spf13/cobra"
)

var Volume = &cobra.Command{
	Use:   "volume",
	Short: definitions.CmdHelpMessages["volume"]["volume"].Short,
	Long:  definitions.CmdHelpMessages["volume"]["volume"].Long,
	RunE: func(cmd *cobra.Command, args []string) error {
		// root cmd will show help only
		return cmd.Help()
	},
}

func init() {
	// Volume sub command
	Volume.AddCommand(volumeGet)
}
