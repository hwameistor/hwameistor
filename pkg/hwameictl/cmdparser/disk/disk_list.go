package disk

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	"github.com/hwameistor/hwameistor/pkg/hwameictl/formatter"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/manager"
)

var nodeName string

var diskList = &cobra.Command{
	Use:   "list",
	Args:  cobra.ExactArgs(0),
	Short: "List the disks' infos.",
	Long: "You can use 'hwameictl disk list' to obtain information about all disks.\n" +
		"Furthermore, you can use 'hwameictl disk list --node {nodeName}' to acquire detailed \n" +
		"information about a specific node.",
	Example: "hwameictl disk list\n" +
		"hwameictl disk list --node worker-1",
	RunE: diskListRunE,
}

func init() {
	// Disk list flags
	diskList.Flags().StringVar(&nodeName, "node", "", "Filter Volumes by node name")
}

func diskListRunE(_ *cobra.Command, _ []string) error {
	c, err := manager.NewLocalDiskController()
	if err != nil {
		return err
	}

	disks, err := c.ListLocalDisk()
	if err != nil {
		return err
	}

	disksHeader := table.Row{"#", "DevPath", "Status", "Node", "Reserved", "Raid", "DiskType",
		"Capacity", "Owner", "Partitioned", "Protocol"}
	var disksRows []table.Row
	index := 0
	for _, disk := range disks.Items {
		if nodeName == "" || nodeName == disk.Spec.NodeName {
			index++
			disksRows = append(disksRows, table.Row{index, disk.Spec.DevicePath, disk.Status.State, disk.Spec.NodeName,
				disk.Spec.Reserved, disk.Spec.HasRAID, disk.Spec.DiskAttributes.Type, formatter.FormatBytesToSize(disk.Spec.Capacity),
				disk.Spec.Owner, disk.Spec.HasPartition, disk.Spec.DiskAttributes.Protocol})
		}
	}

	formatter.PrintTable("Disks", disksHeader, disksRows)
	return nil
}
