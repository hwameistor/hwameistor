package volume

import (
	"fmt"
	"github.com/hwameistor/hwameistor/pkg/apiserver/api"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/formatter"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/manager"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/utils"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

// Read from volume get flags
var node, group string

var volumeGet = &cobra.Command{
	Use:   "get",
	Short: "Get information about volumes.",
	Long: "You can use \"hwameictl volume get\" to obtain information about all volumes.\n" +
		"Additionally, you can use the \"--node\" or \"--group\" options to filter the results.\n" +
		"Furthermore, you can use \"hwameictl volume get {volumeName}\" to acquire detailed \n" +
		"information about a specific volume.",
	RunE: volumeGetRunE,
}

func init() {
	// Volume get flags
	volumeGet.Flags().StringVar(&node, "node", "", "Filter Volumes by node name")
	volumeGet.Flags().StringVar(&group, "group", "", "Filter Volumes by group name")
}

func volumeGetRunE(cmd *cobra.Command, args []string) error {
	switch len(args) {
	case 0:
		// Get All volume info
		return volumeGetAll(node, group)
	case 1:
		// Get one volume detail info
		return volumeGetOne(args[0])
	default:
		// Too many arguments
		return cmd.Help()
	}
}

func volumeGetOne(volumeName string) error {
	queryPage := api.QueryPage{
		PageSize:   -1,
		VolumeName: volumeName,
	}

	controller, err := manager.NewLocalVolumeController()
	if err != nil {
		return err
	}

	lv, err := controller.GetLocalVolume(volumeName)
	if err != nil {
		return err
	}
	if lv == nil {
		return fmt.Errorf("volume %s is not exists", volumeName)
	}
	// Volume infos

	// Replicas info
	replicas, err := controller.GetVolumeReplicas(queryPage)
	if err != nil {
		return err
	}

	replicaHeader := table.Row{"#", "Name", "Status", "SyncState", "Node", "Capacity"}
	replicaRows := make([]table.Row, len(replicas.VolumeReplicas))
	for i, replica := range replicas.VolumeReplicas {
		replicaRows[i] = table.Row{i, replica.Name, replica.Status.State, replica.Status.Synced,
			replica.Spec.NodeName, replica.Spec.RequiredCapacityBytes}
	}

	formatter.PrintTable("Volume Replicas", replicaHeader, replicaRows)

	// Operations info
	operations, err := controller.GetVolumeOperation(queryPage)
	if err != nil {
		return nil
	}

	operationHeader := table.Row{"#", "EventName", "EventType", "Status", "Description", "StartTime"}
	var operationRows []table.Row

	for _, expand := range operations.VolumeExpandOperations {
		operationRows = append(operationRows, table.Row{
			len(operationRows) + 1, expand.Name, "Expand", expand.Status.State,
			expand.Status.Message, utils.FormatTime(expand.CreationTimestamp.Time),
		})
	}

	for _, migrate := range operations.VolumeMigrateOperations {
		operationRows = append(operationRows, table.Row{
			len(operationRows) + 1, migrate.Name, "Migrate", migrate.Status.State,
			migrate.Status.Message, utils.FormatTime(migrate.CreationTimestamp.Time),
		})
	}

	for _, convert := range operations.VolumeConvertOperations {
		operationRows = append(operationRows, table.Row{
			len(operationRows) + 1, convert.Name, "Convert", convert.Status.State,
			convert.Status.Message, utils.FormatTime(convert.CreationTimestamp.Time),
		})
	}

	formatter.PrintTable("Volume Operations", operationHeader, operationRows)

	return nil
}

func volumeGetAll(nodeName, groupName string) error {
	// todo: filter by nodeName, groupName
	controller, err := manager.NewLocalVolumeController()
	if err != nil {
		return err
	}

	// Get all volume
	lvs, err := controller.ListLocalVolume(api.QueryPage{PageSize: -1})
	if err != nil {
		return err
	}
	header := table.Row{"#", "Name", "Status", "Replicas", "VolumeGroup", "Capacity", "Convertible",
		"PVC", "CreateTime", "PublishedNode"}
	rows := make([]table.Row, len(lvs.Volumes))

	// Collect rows data
	for i, volume := range lvs.Volumes {
		rows[i] = table.Row{i, volume.Name, volume.Status.State, len(volume.Status.Replicas), volume.Spec.VolumeGroup,
			volume.Spec.RequiredCapacityBytes, volume.Spec.Config.Convertible,
			volume.Spec.PersistentVolumeClaimName, utils.FormatTime(volume.CreationTimestamp.Time),
			volume.Status.PublishedNodeName,
		}
	}

	formatter.PrintTable("Local Volumes", header, rows)

	return nil
}
