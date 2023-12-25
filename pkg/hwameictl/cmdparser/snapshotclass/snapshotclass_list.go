package snapshotclass

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	"github.com/hwameistor/hwameistor/pkg/hwameictl/formatter"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/manager"
)

var snapshotclassList = &cobra.Command{
	Use:     "list",
	Args:    cobra.ExactArgs(0),
	Short:   "List the Hwameistor's  snapshotclasses.",
	Long:    "You can use 'hwameictl vsc list' to obtain information about all storage snapshotclasses.",
	Example: "hwameictl vsc list",
	RunE:    snapshotClassListRunE,
}

func snapshotClassListRunE(_ *cobra.Command, _ []string) error {
	c, err := manager.NewSnapShotClassController()
	if err != nil {
		return err
	}

	scs, err := c.ListSnapShotClass()
	if err != nil {
		return err
	}

	scsHeader := table.Row{"#", "NAME", "DRIVER", "DELETIONPOLICY", "SNAPSIZE"}
	scsRows := make([]table.Row, len(scs))
	for i, sc := range scs {
		scsRows[i] = table.Row{i + 1, sc.Name, sc.Driver, sc.DeletionPolicy, ""}
		if sc.Parameters != nil {
			s, ok := sc.Parameters["snapsize"]
			if ok {
				scsRows[i] = table.Row{i + 1, sc.Name, sc.Driver, sc.DeletionPolicy, s}
			}
		}
	}

	formatter.PrintTable("SnapshotClass", scsHeader, scsRows)
	return nil
}
