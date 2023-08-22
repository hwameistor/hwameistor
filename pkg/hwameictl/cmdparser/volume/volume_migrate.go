package volume

import (
	"github.com/spf13/cobra"

	"github.com/hwameistor/hwameistor/pkg/hwameictl/manager"
)

var volumeMigrate = &cobra.Command{
	Use:   "migrate {volumeName} {sourceNode} {?targetNode}",
	Args:  cobra.RangeArgs(2, 3),
	Short: "Migrate a volume from sourceNode to targetNode.",
	Long: "Migrate a volume from sourceNode to targetNode. If targetNode is empty,\n" +
		"Hwameistor will automatically select a valid targetNode if one exists.\n" +
		"Volume Migration is an important operation and maintenance management function of Hwameistor.\n" +
		"Application-mounted data volumes can be unmounted and migrated from a node with errors or an alert\n" +
		"indicating an impending errors to a healthy node. ",
	Example: "hwameictl volume migrate pvc-1187f716-db92-47ac-a5fc-44fd19047a82 worker-1 worker-2",
	RunE:    volumeMigrateRunE,
}

func init() {
	// Sub commands
	volumeMigrate.AddCommand(volumeMigrateAbort, volumeMigrateList)
}

func volumeMigrateRunE(_ *cobra.Command, args []string) error {
	var targetNode string
	switch len(args) {
	case 2:
		// Auto select target node
	case 3:
		// With target node
		targetNode = args[2]
	}
	volumeName, sourceNode := args[0], args[1]

	c, err := manager.NewLocalVolumeController()
	if err != nil {
		return err
	}

	_, err = c.CreateVolumeMigrate(volumeName, sourceNode, targetNode, false)
	return err
}
