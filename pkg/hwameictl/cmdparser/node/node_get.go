package node

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/apiserver/api"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/formatter"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/manager"
)

var nodeGet = &cobra.Command{
	Use:     "get {nodeName}",
	Args:    cobra.ExactArgs(1),
	Short:   "Get the Hwameistor's storage node detail information.",
	Long:    "Get the Hwameistor's storage node detail information.",
	Example: "hwameictl node get worker-1",
	RunE:    nodeGetRunE,
}

func nodeGetRunE(_ *cobra.Command, args []string) error {
	nodeName := args[0]
	c, err := manager.NewLocalStorageNodeController()
	if err != nil {
		return err
	}

	node, err := c.GetStorageNode(nodeName)
	if err != nil {
		return err
	}

	// Print basic infos
	hddUtilization, ssdUtilization := getNodePoolUtilization(node)
	formatter.PrintParameters("Node parameters", []formatter.Parameter{
		{"NodeStatus", node.K8sNodeState},
		{"DriverStatus", node.LocalStorageNode.Status.State},
		{"HDDUtilization", hddUtilization},
		{"SSDUtilization", ssdUtilization},
	})

	// Print node's disks
	disks, err := c.LocalDiskListByNode(api.QueryPage{PageSize: -1, NodeName: nodeName})
	disksHeader := table.Row{"#", "Path", "Status", "Reserved", "Raid", "Type", "Owner", "TotalCapacity"}
	disksRows := make([]table.Row, len(disks.LocalDisks))
	for i, disk := range disks.LocalDisks {
		disksRows[i] = table.Row{i + 1, disk.Spec.DevicePath, disk.Status.State, disk.Spec.Reserved, disk.Spec.HasRAID,
			disk.Spec.DiskAttributes.Type, disk.Spec.Owner, formatter.FormatBytesToSize(disk.TotalCapacityBytes)}
	}
	formatter.PrintTable("Node disks", disksHeader, disksRows)

	// Print node's volumes
	volumeController, err := manager.NewLocalVolumeController()
	if err != nil {
		return err
	}
	lvs, err := volumeController.ListLocalVolume(api.QueryPage{PageSize: -1})
	if err != nil {
		return err
	}
	volumeHeader := table.Row{"#", "Name", "Status", "Replicas", "Group", "Capacity", "Convertible",
		"PVC", "CreateTime", "PublishedNode"}
	var volumeRows []table.Row

	for i, volume := range lvs.Volumes {
		// Filter by nodeName
		if checkVolumeLocateNode(volume, nodeName) {
			continue
		}

		volumeRows = append(volumeRows, table.Row{i, volume.Name, volume.Status.State, len(volume.Status.Replicas),
			volume.Spec.VolumeGroup, formatter.FormatBytesToSize(volume.Spec.RequiredCapacityBytes),
			volume.Spec.Config.Convertible, volume.Spec.PersistentVolumeClaimName,
			formatter.FormatTime(volume.CreationTimestamp.Time), volume.Status.PublishedNodeName,
		})
	}
	formatter.PrintTable("Node volumes", volumeHeader, volumeRows)
	return nil
}

func getNodePoolUtilization(node *api.StorageNode) (hddUtilization, ssdUtilization string) {
	// Calculate the hdd and ssd pool utilization
	hddUtilization = "N/A"
	ssdUtilization = "N/A"
	if hddPool, ok := node.LocalStorageNode.Status.Pools[api.PoolNamePrefix+apisv1alpha1.DiskClassNameHDD]; ok {
		hddUtilization = formatter.FormatPercentString(hddPool.UsedCapacityBytes, hddPool.TotalCapacityBytes)
	}
	if ssdPool, ok := node.LocalStorageNode.Status.Pools[api.PoolNamePrefix+apisv1alpha1.DiskClassNameSSD]; ok {
		ssdUtilization = formatter.FormatPercentString(ssdPool.UsedCapacityBytes, ssdPool.TotalCapacityBytes)
	}
	return
}

// Check the volume have a replica at the node
func checkVolumeLocateNode(volume *api.Volume, nodeName string) bool {
	for _, replica := range volume.Spec.Config.Replicas {
		if replica.Hostname == nodeName {
			return true
		}
	}
	return false
}
