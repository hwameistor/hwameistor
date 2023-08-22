package node

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/hwameistor/hwameistor/pkg/hwameictl/manager"
)

var nodeEnable = &cobra.Command{
	Use:   "enable {nodeName} {enable}",
	Args:  cobra.ExactArgs(2),
	Short: "Enable/Disable a Hwameistor's storage node.",
	Long: "Enable/Disable a Hwameistor's storage node.\n" +
		"This operation may takes a few minutes.",
	Example: "hwameictl node enable worker-1 false",
	RunE:    nodeEnableRunE,
}

func nodeEnableRunE(_ *cobra.Command, args []string) error {
	c, err := manager.NewLocalStorageNodeController()
	if err != nil {
		return err
	}

	switch args[1] {
	case "true":
		return c.UpdateLocalStorageNode(args[0], true)
	case "false":
		return c.UpdateLocalStorageNode(args[0], false)
	}
	return fmt.Errorf("the 'enable' parameter should be true/false")
}
