package node

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	"github.com/hwameistor/hwameistor/pkg/apiserver/api"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/formatter"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/manager"
)

var nodeList = &cobra.Command{
	Use:     "list",
	Args:    cobra.ExactArgs(0),
	Short:   "List the Hwameistor's storage nodes.",
	Long:    "You can use 'hwameictl node list' to obtain information about all storage nodes.",
	Example: "hwameictl node list",
	RunE:    nodeListRunE,
}

func nodeListRunE(_ *cobra.Command, _ []string) error {
	c, err := manager.NewLocalStorageNodeController()
	if err != nil {
		return err
	}

	nodes, err := c.ListLocalStorageNode(api.QueryPage{PageSize: -1})
	if err != nil {
		return err
	}

	nodeHeader := table.Row{"#", "Name", "DriverStatus", "NodeStatus", "HDDUtilization", "SSDUtilization"}
	nodeRows := make([]table.Row, len(nodes))
	for i, node := range nodes {
		hddUtilization, ssdUtilization := getNodePoolUtilization(node)
		nodeRows[i] = table.Row{i + 1, node.LocalStorageNode.Name, node.LocalStorageNode.Status.State,
			node.K8sNodeState, hddUtilization, ssdUtilization}
	}

	formatter.PrintTable("Nodes", nodeHeader, nodeRows)
	return nil
}
