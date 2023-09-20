package volume

import (
	"github.com/spf13/cobra"

	"github.com/hwameistor/hwameistor/pkg/hwameictl/manager"
)

var volumeConvertAbort = &cobra.Command{
	Use:     "abort {volumeName}",
	Args:    cobra.ExactArgs(1),
	Short:   "Abort a volume's convert task.",
	Long:    "Abort a volume's convert task.",
	Example: "hwameictl volume convert abort pvc-1187f716-db92-47ac-a5fc-44fd19047a81",
	RunE:    volumeConvertAbortRunE,
}

func volumeConvertAbortRunE(_ *cobra.Command, args []string) error {
	c, err := manager.NewLocalVolumeController()
	if err != nil {
		return err
	}

	_, err = c.CreateVolumeConvert(args[0], true)
	return err
}
