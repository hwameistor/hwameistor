package volume

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	"github.com/hwameistor/hwameistor/pkg/apiserver/api"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/formatter"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/manager"
)

// Read from volume list flags
var nodeName, groupName string

var volumeList = &cobra.Command{
	Use:   "list",
	Args:  cobra.ExactArgs(0),
	Short: "List the Hwameistor volumes.",
	Long: "You can use 'hwameictl volume list' to obtain information about all volumes.\n" +
		"Additionally, you can use the '--node' or '--group' options to filter the results.\n" +
		"Furthermore, you can use 'hwameictl volume get {volumeName}' to acquire detailed \n" +
		"information about a specific volume.",
	Example: "hwameictl volume list\n" +
		"hwameictl volume list --node worker-1\n" +
		"hwameictl volume list --group 2369664d-c8b6-4781-a0e7-96468dae6634",
	RunE: volumeListRunE,
}

func init() {
	// Volume list flags
	volumeList.Flags().StringVar(&nodeName, "node", "", "Filter Volumes by node name")
	volumeList.Flags().StringVar(&groupName, "group", "", "Filter Volumes by group name")
}

func volumeListRunE(_ *cobra.Command, _ []string) error {
	c, err := manager.NewLocalVolumeController()
	if err != nil {
		return err
	}

	// Get all volume
	lvs, err := c.ListLocalVolume(api.QueryPage{PageSize: -1})
	if err != nil {
		return err
	}
	volumeHeader := table.Row{"#", "Name", "Status", "Replicas", "Group", "Capacity", "Convertible",
		"PVC", "CreateTime", "PublishedNode"}
	var volumeRows []table.Row

	// Collect rows data
	for i, volume := range lvs.Volumes {
		// Filter by groupName and nodeName
		if (groupName != "" && volume.Spec.VolumeGroup != groupName) ||
			(nodeName != "" && checkVolumeLocateNode(volume, nodeName)) {
			continue
		}

		volumeRows = append(volumeRows, table.Row{i, volume.Name, volume.Status.State, len(volume.Status.Replicas),
			volume.Spec.VolumeGroup, formatter.FormatBytesToSize(volume.Spec.RequiredCapacityBytes),
			volume.Spec.Config.Convertible, volume.Spec.PersistentVolumeClaimName,
			formatter.FormatTime(volume.CreationTimestamp.Time), volume.Status.PublishedNodeName,
		})
	}
	formatter.PrintTable("Volumes", volumeHeader, volumeRows)
	return nil
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
