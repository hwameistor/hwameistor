package volume

import (
	"github.com/spf13/cobra"

	"github.com/hwameistor/hwameistor/pkg/hwameictl/manager"
)

var volumeConvert = &cobra.Command{
	Use:   "convert {volumeName}",
	Args:  cobra.ExactArgs(1),
	Short: "Convert a volume to a HA(High Availability) volume.",
	Long: "Convert a volume to a HA(High Availability) volume.\n" +
		"The volume should be convertible and the replica number should be 1.",
	Example: "hwameictl volume convert pvc-1187f716-db92-47ac-a5fc-44fd19047a81",
	RunE:    volumeConvertRunE,
}

func init() {
	// Sub commands
	volumeConvert.AddCommand(volumeConvertAbort, volumeConvertList)
}

func volumeConvertRunE(_ *cobra.Command, args []string) error {
	c, err := manager.NewLocalVolumeController()
	if err != nil {
		return err
	}

	_, err = c.CreateVolumeConvert(args[0], false)
	return err
}
