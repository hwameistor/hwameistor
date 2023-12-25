package snapshot

import (
	"github.com/hwameistor/hwameistor/pkg/hwameictl/manager"
	"github.com/spf13/cobra"
)

var snapshotDelete = &cobra.Command{
	Use:     "delete",
	Args:    cobra.ExactArgs(1),
	Short:   "delete the Hwameistor's  snapshot.",
	Long:    "You can use 'hwameictl vsc delete' to delete the Hwameistor's snapshot.",
	Example: "hwameictl vs delete example-vs",
	RunE:    snapshotDeleteRunE,
}

func init() {
	snapshotDelete.Flags().StringVar(&ns, "ns", "default", "namespace")
}

func snapshotDeleteRunE(_ *cobra.Command, args []string) error {
	c, err := manager.NewVolumeSnapShotController()
	if err != nil {
		return err
	}
	return c.DeleteVolumeSnapshot(args[0], ns)
}
