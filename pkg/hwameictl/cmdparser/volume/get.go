package volume

import (
	"github.com/hwameistor/hwameistor/pkg/hwameictl/cmdparser/definitions"
	"github.com/spf13/cobra"
)

var volumeGet = &cobra.Command{
	Use:   "get",
	Short: definitions.CmdHelpMessages["volume"]["get"].Short,
	Long:  definitions.CmdHelpMessages["volume"]["get"].Long,
	RunE: func(cmd *cobra.Command, args []string) error {
		// root cmd will show help only
		return cmd.Help()
	},
}

func init() {
	// Volume get flags
	volumeGet.Flags().String("node", "", "Get Volumes by node")
	volumeGet.Flags().String("group", "", "Get Volumes by VolumeGroup")
}
