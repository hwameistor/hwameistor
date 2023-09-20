package volume

import (
	"github.com/spf13/cobra"

	"github.com/hwameistor/hwameistor/pkg/hwameictl/manager"
)

var volumeExpand = &cobra.Command{
	Use:   "expand {volumeName} {targetCapacity}",
	Args:  cobra.ExactArgs(2),
	Short: "Expand a volume's capacity.",
	Long: "Expand a volume's capacity. The volume's storageclass should be with `allowVolumeExpansion: true`.\n" +
		"The targetCapacity should be like `256Mi` `20Gi`.",
	Example: "hwameictl volume expand pvc-1187f716-db92-47ac-a5fc-44fd19047a81 20Gi",
	RunE:    volumeExpandRunE,
}

func init() {
	// Sub commands
	volumeExpand.AddCommand(volumeExpandAbort, volumeExpandList)
}

func volumeExpandRunE(_ *cobra.Command, args []string) error {
	c, err := manager.NewLocalVolumeController()
	if err != nil {
		return err
	}

	_, err = c.CreateVolumeExpand(args[0], args[1], false)
	return err
}
