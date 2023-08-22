package migrate

import (
	"github.com/spf13/cobra"

	"github.com/hwameistor/hwameistor/pkg/hwameictl/manager"
)

var VolumeMigrate = &cobra.Command{
	Use:   "migrate {volumeName} {sourceNode} {?targetNode}",
	Short: "Migrate a volume from a node to another node.",
	Long:  "",
	RunE:  volumeMigrateRunE,
}

func volumeMigrateRunE(cmd *cobra.Command, args []string) error {
	var targetNode string
	switch len(args) {
	case 2:
		// Auto select target node
	case 3:
		// With target node
		targetNode = args[2]
	default:
		// Missing or too many arguments
		return cmd.Help()
	}
	volumeName, sourceNode := args[0], args[1]

	c, err := manager.NewLocalVolumeController()
	if err != nil {
		return err
	}

	_, err = c.CreateVolumeMigrate(volumeName, sourceNode, targetNode, false)
	return err
}
