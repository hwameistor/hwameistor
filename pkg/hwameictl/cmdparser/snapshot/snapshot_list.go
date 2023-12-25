package snapshot

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	"github.com/hwameistor/hwameistor/pkg/hwameictl/formatter"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/manager"
)

var snapshotList = &cobra.Command{
	Use:     "list",
	Args:    cobra.ExactArgs(0),
	Short:   "List the Hwameistor's  snapshots.",
	Long:    "You can use 'hwameictl sc list' to obtain information about all storage snapshots.",
	Example: "hwameictl vs list",
	RunE:    snapshotListRunE,
}

func snapshotListRunE(_ *cobra.Command, _ []string) error {
	c, err := manager.NewVolumeSnapShotController()
	if err != nil {
		return err
	}

	vss, err := c.ListVolumeSnapShot()
	if err != nil {
		return err
	}

	scsHeader := table.Row{"#", "Name", "READYTOUSE", "SOURCEPVC", "RESTORESIZE", "SNAPSHOTCLASS", "LVSNAME"}
	scsRows := make([]table.Row, len(vss))
	for i, vs := range vss {
		scsRows[i] = table.Row{i + 1, vs.Name, *vs.Status.ReadyToUse, *vs.Spec.Source.PersistentVolumeClaimName,
			vs.Status.RestoreSize, *vs.Spec.VolumeSnapshotClassName, *vs.Status.BoundVolumeSnapshotContentName}
	}

	formatter.PrintTable("VolumeSnapshot", scsHeader, scsRows)
	return nil
}
