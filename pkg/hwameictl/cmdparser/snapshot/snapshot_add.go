package snapshot

import (
	"github.com/hwameistor/hwameistor/pkg/hwameictl/manager"
	"github.com/spf13/cobra"
)

var vsc, pvc, ns string
var snapshotAdd = &cobra.Command{
	Use:     "add {vsName}",
	Args:    cobra.ExactArgs(1),
	Short:   "add the Hwameistor's storage snapshot.",
	Long:    "You can use 'hwameictl ssc add' to add hwameistor-snapshotclass.",
	Example: "hwameictl vs add example-vs --vsc=example-volumesnapshotclass --pvc=local-storage-pvc-lvm --ns=default",
	RunE:    snapshotAddRunE,
}

func init() {
	snapshotAdd.Flags().StringVar(&vsc, "vsc", "", "volumeSnapshotClassName")
	snapshotAdd.Flags().StringVar(&pvc, "pvc", "", "persistentVolumeClaimName")
	snapshotAdd.Flags().StringVar(&ns, "ns", "default", "namespace")

}

func snapshotAddRunE(_ *cobra.Command, args []string) error {
	c, err := manager.NewVolumeSnapShotController()
	if err != nil {
		return err
	}

	return c.AddVolumeSnapshot(args[0], vsc, pvc, ns)
}
