package volume

import (
	"github.com/spf13/cobra"

	"github.com/hwameistor/hwameistor/pkg/hwameictl/manager"
)

var volumeExpandAbort = &cobra.Command{
	Use:     "abort {volumeName}",
	Args:    cobra.ExactArgs(1),
	Short:   "Abort the volume expand operation.",
	Long:    "Abort the volume expand operation.",
	Example: "hwameictl volume expand abort pvc-1187f716-db92-47ac-a5fc-44fd19047a81",
	RunE:    volumeExpandAbortRunE,
}

func volumeExpandAbortRunE(_ *cobra.Command, args []string) error {
	c, err := manager.NewLocalVolumeController()
	if err != nil {
		return err
	}

	_, err = c.CreateVolumeExpand(args[0], "", true)
	return err
}
