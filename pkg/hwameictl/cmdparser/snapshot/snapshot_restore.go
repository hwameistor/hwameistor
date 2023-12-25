package snapshot

import (
	"github.com/hwameistor/hwameistor/pkg/hwameictl/manager"
	"github.com/spf13/cobra"
)

var pvcName, storage string

var SnapshotRestore = &cobra.Command{
	Use:     "restore",
	Args:    cobra.ExactArgs(1),
	Short:   "restore the Hwameistor's  volume by VolumeSnapshot.",
	Long:    "You can use 'hwameictl vs restore [vs-name} ' to rollback the Hwameistor's  volume by VolumeSnapshot..",
	Example: "hwameictl vs restore example-vs  --pvc=example-pvc --storage=1Gi",
	RunE:    VolumeSnapshotRestoreRunE,
}

func init() {
	SnapshotRestore.Flags().StringVar(&pvcName, "pvc", "", "persistentVolumeClaimName")
	SnapshotRestore.Flags().StringVar(&storage, "storage", "", "storage capacity")
	SnapshotRestore.Flags().StringVar(&ns, "ns", "default", "namespace")

}

func VolumeSnapshotRestoreRunE(_ *cobra.Command, args []string) error {
	c, err := manager.NewVolumeSnapShotController()
	if err != nil {
		return err
	}
	return c.RestoreVolumeSnapshot(args[0], pvcName, storage, ns)
}
