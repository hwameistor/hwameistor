package disk

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/hwameistor/hwameistor/pkg/apiserver/api"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/manager"
)

var diskReserve = &cobra.Command{
	Use:     "reserve {nodeName} {deviceDevPath} {reserve}",
	Args:    cobra.ExactArgs(3),
	Short:   "Set a disk to be reserved.",
	Long:    "Set a disk to be reserved.",
	Example: "hwameictl disk reserve worker-1 sdb true",
	RunE:    diskReserveRunE,
}

func diskReserveRunE(_ *cobra.Command, args []string) error {
	c, err := manager.NewLocalStorageNodeController()
	if err != nil {
		return err
	}

	// Build the query parameters
	parameters := api.QueryPage{
		NodeName:        args[0],
		DeviceShortPath: args[1],
	}

	switch args[2] {
	case "true":
		// Set the disk reserved
		_, err = c.ReserveStorageNodeDisk(parameters)
		return err
	case "false":
		// Set the disk unreserved
		_, err = c.RemoveReserveStorageNodeDisk(parameters)
		return err
	}
	return fmt.Errorf("the `reserve` parameter should be true/false")
}
