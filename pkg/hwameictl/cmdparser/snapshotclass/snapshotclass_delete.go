package snapshotclass

import (
	"github.com/hwameistor/hwameistor/pkg/hwameictl/manager"
	"github.com/spf13/cobra"
)

var snapshotclassDelete = &cobra.Command{
	Use:     "delete",
	Args:    cobra.ExactArgs(1),
	Short:   "delete the Hwameistor's  snapshotclass.",
	Long:    "You can use 'hwameictl vsc delete' to delete the Hwameistor's snapshotclass.",
	Example: "hwameictl vsc delete example-vsc",
	RunE:    snapshotDeleteRunE,
}

func snapshotDeleteRunE(_ *cobra.Command, args []string) error {
	c, err := manager.NewSnapShotClassController()
	if err != nil {
		return err
	}
	return c.DeleteHwameistorSnapshotClass(args[0])
}
