package snapshotclass

import (
	"github.com/hwameistor/hwameistor/pkg/hwameictl/manager"
	"github.com/spf13/cobra"
)

var snapsize string
var snapshotclassAdd = &cobra.Command{
	Use:   "add {sscName}",
	Args:  cobra.ExactArgs(1),
	Short: "add the Hwameistor's storage VolumeSnapshotClass",
	Long:  "You can use 'hwameictl vsc add' to add hwameistor-VolumeSnapshotClass.",
	Example: "hwameictl ssc add example-vsc \n" +
		"hwameictl sc add example-vsc --snapsize=1G",
	RunE: snapshotClassAddRunE,
}

func init() {
	// Volume list flags
	snapshotclassAdd.Flags().StringVar(&snapsize, "snapsize", "", "Specify the size at which to create a volume snapshot")
}

func snapshotClassAddRunE(_ *cobra.Command, args []string) error {
	c, err := manager.NewSnapShotClassController()
	if err != nil {
		return err
	}

	return c.AddHwameistorSnapshotClass(args[0], snapsize)
}
