package localvolumesnapshot

import (
	hwameistorapi "github.com/hwameistor/hwameistor/pkg/apiserver/api"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/formatter"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/manager"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

var volumeName string
var localVolumeSnapshotList = &cobra.Command{
	Use:   "list ",
	Args:  cobra.ExactArgs(0),
	Short: "List LocalVolumeSnapshots .",
	Long:  "List LocalVolumeSnapshots .",
	Example: "hwameictl lvs list \n" +
		"hwameictl lvs list --volume example-volume",
	RunE: localVolumeSnapshotListRunE,
}

func init() {
	localVolumeSnapshotList.Flags().StringVar(&volumeName, "volume", "", "volumeName")

}

func localVolumeSnapshotListRunE(_ *cobra.Command, _ []string) error {
	c, err := manager.NewLocalVolumeSnapshotController()
	if err != nil {
		return err
	}

	var queryPage hwameistorapi.QueryPage
	queryPage.PageSize = -1
	if volumeName != "" {
		queryPage.VolumeName = volumeName
	}
	snapshotList, err := c.ListLocalSnapshot(queryPage)
	if err != nil {
		return err
	}

	snapshotListHeader := table.Row{"#", "Name", "CAPACITY", "SOURCEVOLUME", "STATE", "MERGING", "INVALID"}
	snapshotListRows := make([]table.Row, len(snapshotList.Snapshots))
	for i, s := range snapshotList.Snapshots {
		snapshotListRows[i] = table.Row{i + 1, s.Name, s.Spec.RequiredCapacityBytes,
			s.Spec.SourceVolume, s.Status.State, s.Status.Attribute.Merging, s.Status.Attribute.Invalid}
	}

	formatter.PrintTable("Volume LocalVolumeSnapshot list", snapshotListHeader, snapshotListRows)
	return nil
}
