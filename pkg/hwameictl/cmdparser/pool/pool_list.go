package pool

import (
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	"github.com/hwameistor/hwameistor/pkg/apiserver/api"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/formatter"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/manager"
)

var poolList = &cobra.Command{
	Use:     "list",
	Args:    cobra.ExactArgs(0),
	Short:   "List the Hwameistor storage pools.",
	Long:    "You can use 'hwameictl pool list' to obtain information about all pools.",
	Example: "hwameictl pool list",
	RunE:    poolListRunE,
}

func poolListRunE(_ *cobra.Command, _ []string) error {
	c, err := manager.NewLocalStoragePoolController()
	if err != nil {
		return err
	}

	pools, err := c.StoragePoolList(api.QueryPage{PageSize: -1})
	if err != nil {
		return err
	}

	for _, pool := range pools.StoragePools {
		// Print pool's basic info
		poolType := pool.StorageNodePools[0].Class
		formatter.PrintParameters(fmt.Sprintf("%s storage pool", poolType), []formatter.Parameter{
			{"Type", poolType},
			{"StorageUsage", formatter.FormatPercentString(pool.AllocatedCapacityBytes, pool.TotalCapacityBytes)},
			{"AllocatedCapacity", formatter.FormatBytesToSize(pool.AllocatedCapacityBytes)},
			{"TotalCapacity", formatter.FormatBytesToSize(pool.TotalCapacityBytes)},
			{"NodeCount", len(pool.NodeNames)},
		})

		// Print pool's nodes info
		poolNodesHeader := table.Row{"#", "Name", "UsedCapacity", "TotalCapacity", "StorageUsage", "Disks", "Volumes", "VolumeLimit"}
		poolNodesRows := make([]table.Row, len(pool.StorageNodePools))
		for i, node := range pool.StorageNodePools {
			poolNodesRows[i] = table.Row{i + 1, node.NodeName,
				formatter.FormatBytesToSize(node.UsedCapacityBytes), formatter.FormatBytesToSize(node.TotalCapacityBytes),
				formatter.FormatPercentString(node.UsedCapacityBytes, node.TotalCapacityBytes),
				len(node.Disks), len(node.Volumes), node.TotalVolumeCount}
		}
		formatter.PrintTable(fmt.Sprintf("%s storage pool nodes", poolType), poolNodesHeader, poolNodesRows)
	}
	return nil
}
