package volume

import (
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	"github.com/hwameistor/hwameistor/pkg/apiserver/api"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/formatter"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/manager"
)

var volumeGet = &cobra.Command{
	Use:     "get {volumeName}",
	Args:    cobra.ExactArgs(1),
	Short:   "Get the Hwameistor volume's detail information.",
	Long:    "Get the Hwameistor volume's detail information.",
	Example: "hwameictl volume get pvc-1187f716-db92-47ac-a5fc-44fd19047a81",
	RunE:    volumeGetRunE,
}

func volumeGetRunE(_ *cobra.Command, args []string) error {
	volumeName := args[0]
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
	formatter.PrintParameters("Volume parameters", []formatter.Parameter{
		{"Status", lv.Status.State},
		{"Replicas", len(lv.Status.Replicas)},
		{"Used", formatter.FormatBytesToSize(lv.Status.UsedCapacityBytes)},
		{"Capacity", formatter.FormatBytesToSize(lv.Spec.RequiredCapacityBytes)},
		{"Convertible", lv.Spec.Convertible},
		{"PVC", lv.Spec.PersistentVolumeClaimName},
		{"PublishedNode", lv.Status.PublishedNodeName},
		{"Pool", lv.Spec.PoolName},
		{"Group", lv.Spec.VolumeGroup},
		{"FSType", lv.Status.PublishedFSType},
		{"CreateTime", formatter.FormatTime(lv.CreationTimestamp.Time)},
	})

	// Replicas info
	replicas, err := controller.GetVolumeReplicas(queryPage)
	if err != nil {
		return err
	}

	replicaHeader := table.Row{"#", "Name", "Status", "SyncState", "Node", "Capacity"}
	replicaRows := make([]table.Row, len(replicas.VolumeReplicas))
	for i, replica := range replicas.VolumeReplicas {
		replicaRows[i] = table.Row{i + 1, replica.Name, replica.Status.State, replica.Status.Synced,
			replica.Spec.NodeName, formatter.FormatBytesToSize(replica.Spec.RequiredCapacityBytes)}
	}

	formatter.PrintTable("Volume replicas", replicaHeader, replicaRows)

	// Operations info
	operations, err := controller.GetVolumeOperation(queryPage)
	if err != nil {
		return nil
	}

	operationHeader := table.Row{"#", "Name", "Type", "Status", "Description", "StartTime"}
	var operationRows []table.Row

	for _, expand := range operations.VolumeExpandOperations {
		operationRows = append(operationRows, table.Row{
			len(operationRows) + 1, expand.Name, "Expand", expand.Status.State,
			expand.Status.Message, formatter.FormatTime(expand.CreationTimestamp.Time),
		})
	}
	for _, migrate := range operations.VolumeMigrateOperations {
		operationRows = append(operationRows, table.Row{
			len(operationRows) + 1, migrate.Name, "Migrate", migrate.Status.State,
			migrate.Status.Message, formatter.FormatTime(migrate.CreationTimestamp.Time),
		})
	}
	for _, convert := range operations.VolumeConvertOperations {
		operationRows = append(operationRows, table.Row{
			len(operationRows) + 1, convert.Name, "Convert", convert.Status.State,
			convert.Status.Message, formatter.FormatTime(convert.CreationTimestamp.Time),
		})
	}

	formatter.PrintTable("Volume operations", operationHeader, operationRows)
	return nil
}
