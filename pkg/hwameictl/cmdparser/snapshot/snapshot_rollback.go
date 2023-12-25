package snapshot

import (
	"github.com/hwameistor/hwameistor/pkg/hwameictl/manager"
	"github.com/spf13/cobra"
)

var SnapshotRollback = &cobra.Command{
	Use:     "rollback",
	Args:    cobra.ExactArgs(1),
	Short:   "rollback the Hwameistor's  volume by VolumeSnapshot.",
	Long:    "You can use 'hwameictl vs rollback [vs-name} ' to rollback the Hwameistor's  volume by VolumeSnapshot..",
	Example: "hwameictl vs rollback example-vs",
	RunE:    VolumeSnapshotRollbackRunE,
}

func init() {
	SnapshotRollback.Flags().StringVar(&ns, "ns", "default", "namespace")
}

func VolumeSnapshotRollbackRunE(_ *cobra.Command, args []string) error {
	c, err := manager.NewVolumeSnapShotController()
	if err != nil {
		return err
	}
	return c.RollbackVolumeSnapshot(args[0], ns)
}
