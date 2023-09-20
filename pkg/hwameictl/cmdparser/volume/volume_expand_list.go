package volume

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	"github.com/hwameistor/hwameistor/pkg/hwameictl/formatter"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/manager"
)

var volumeExpandList = &cobra.Command{
	Use:     "list",
	Args:    cobra.ExactArgs(0),
	Short:   "List all the expand operations.",
	Long:    "List all the expand operations.",
	Example: "hwameictl volume expand list",
	RunE:    volumeExpandListRunE,
}

func volumeExpandListRunE(_ *cobra.Command, _ []string) error {
	c, err := manager.NewLocalVolumeController()
	if err != nil {
		return err
	}

	// List all the expand operations
	expandList, err := c.ListVolumeExpand()
	if err != nil {
		return err
	}

	expandListHeader := table.Row{"#", "Name", "VolumeName", "RequireCapacity",
		"Abort", "State", "Message"}
	expandListRows := make([]table.Row, len(expandList.Items))
	for i, expand := range expandList.Items {
		expandListRows[i] = table.Row{i + 1, expand.Name, expand.Spec.VolumeName,
			formatter.FormatBytesToSize(expand.Spec.RequiredCapacityBytes), expand.Spec.Abort,
			expand.Status.State, expand.Status.Message}
	}

	formatter.PrintTable("Expand operation list", expandListHeader, expandListRows)
	return nil
}
