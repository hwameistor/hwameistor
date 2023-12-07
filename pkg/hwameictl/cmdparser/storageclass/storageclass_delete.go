package storageclass

import (
	"github.com/hwameistor/hwameistor/pkg/hwameictl/manager"
	"github.com/spf13/cobra"
)

var storageClassDelete = &cobra.Command{
	Use:     "delete",
	Args:    cobra.ExactArgs(1),
	Short:   "delete the Hwameistor's  storageClasses.",
	Long:    "You can use 'hwameictl sc delete' to delete the Hwameistor's storageClasses.",
	Example: "hwameictl sc delete example-sc",
	RunE:    storageClassDeleteRunE,
}

func storageClassDeleteRunE(_ *cobra.Command, args []string) error {
	c, err := manager.NewStorageClassController()
	if err != nil {
		return err
	}
	return c.DeleteHwameistorStroageClass(args[0])
}
