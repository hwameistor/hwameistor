package volume

import (
	"github.com/spf13/cobra"

	"github.com/hwameistor/hwameistor/pkg/hwameictl/manager"
)

var volumeMigrateAbort = &cobra.Command{
	Use:     "abort {volumeName}",
	Args:    cobra.ExactArgs(1),
	Short:   "Abort a volume's migrate operation.",
	Long:    "Abort a volume's migrate operation.",
	Example: "hwameictl volume migrate abort pvc-1187f716-db92-47ac-a5fc-44fd19047a82",
	RunE:    volumeMigrateAbortRunE,
}

func volumeMigrateAbortRunE(_ *cobra.Command, args []string) error {
	c, err := manager.NewLocalVolumeController()
	if err != nil {
		return err
	}

	_, err = c.CreateVolumeMigrate(args[0], "", "", true)
	return err
}
