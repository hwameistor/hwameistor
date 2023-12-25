package localvolumesnapshot

import (
	"github.com/hwameistor/hwameistor/pkg/hwameictl/manager"
	"github.com/spf13/cobra"
)

var localVolumeSnapshotRollback = &cobra.Command{
	Use:     "rollback",
	Args:    cobra.ExactArgs(1),
	Short:   "rollback the Hwameistor's  volume by lcoalVolumeSnapshot.",
	Long:    "You can use 'hwameictl lvs rollback [lvs-name} ' to rollback the Hwameistor's  volume by lcoalVolumeSnapshot..",
	Example: "hwameictl lvs rollback example-lvs",
	RunE:    localVolumeSnapshotRollbackRunE,
}

func localVolumeSnapshotRollbackRunE(_ *cobra.Command, args []string) error {
	c, err := manager.NewLocalVolumeSnapshotController()
	if err != nil {
		return err
	}
	return c.RollbackVolumeSnapshot(args[0])
}
