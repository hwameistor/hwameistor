package volume

import (
	"github.com/spf13/cobra"
)

var Volume = &cobra.Command{
	Use:   "volume",
	Args:  cobra.ExactArgs(0),
	Short: "Manage the Hwameistor's volumes.",
	Long: "Manage the Hwameistor's volumes.Hwameistor provides LVM-based data volumes,\n" +
		"which offer read and write performance comparable to that of native disks.\n" +
		"These data volumes also provide advanced features such as data volume expansion,\n" +
		"migration, high availability, and more.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// root cmd will show help only
		return cmd.Help()
	},
}

func init() {
	// Volume sub commands
	Volume.AddCommand(volumeGet, volumeList, volumeMigrate, volumeConvert, volumeExpand)
}
