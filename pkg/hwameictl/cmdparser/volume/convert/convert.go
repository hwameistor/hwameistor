package convert

import (
	"github.com/spf13/cobra"

	"github.com/hwameistor/hwameistor/pkg/hwameictl/manager"
)

var VolumeConvert = &cobra.Command{
	Use:   "convert {volumeName}",
	Short: "",
	Long:  "",
	RunE:  volumeConvertRunE,
}

func volumeConvertRunE(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return cmd.Help()
	}

	c, err := manager.NewLocalVolumeController()
	if err != nil {
		return err
	}

	_, err = c.CreateVolumeConvert(args[0], false)
	return err
}
