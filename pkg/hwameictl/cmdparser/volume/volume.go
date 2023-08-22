package volume

import (
	"github.com/hwameistor/hwameistor/pkg/hwameictl/cmdparser/volume/convert"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/cmdparser/volume/get"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/cmdparser/volume/migrate"
	"github.com/spf13/cobra"
)

var Volume = &cobra.Command{
	Use:   "volume",
	Short: "Manage the hwameistor's Volumes.",
	Long:  "Manage the hwameistor's Volumes.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// root cmd will show help only
		return cmd.Help()
	},
}

func init() {
	// Volume sub command
	Volume.AddCommand(get.VolumeGet)
	Volume.AddCommand(migrate.VolumeMigrate)
	Volume.AddCommand(convert.VolumeConvert)
}
