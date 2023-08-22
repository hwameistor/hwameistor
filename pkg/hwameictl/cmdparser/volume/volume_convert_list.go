package volume

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	"github.com/hwameistor/hwameistor/pkg/hwameictl/formatter"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/manager"
)

var volumeConvertList = &cobra.Command{
	Use:     "list",
	Args:    cobra.ExactArgs(0),
	Short:   "List all the convert operations.",
	Long:    "List all the convert operations.",
	Example: "hwameictl volume convert list",
	RunE:    volumeConvertListRunE,
}

func volumeConvertListRunE(_ *cobra.Command, _ []string) error {
	c, err := manager.NewLocalVolumeController()
	if err != nil {
		return err
	}
	// List all the convert operations
	convertList, err := c.ListVolumeConvert()
	if err != nil {
		return err
	}

	convertListHeader := table.Row{"#", "Name", "VolumeName", "ReplicaNumber", "Abort",
		"Statue", "Message"}
	convertListRows := make([]table.Row, len(convertList.Items))
	for i, convert := range convertList.Items {
		convertListRows[i] = table.Row{i + 1, convert.Name, convert.Spec.VolumeName,
			convert.Spec.ReplicaNumber, convert.Spec.Abort, convert.Status.State,
			convert.Status.Message}
	}

	formatter.PrintTable("Convert operation list", convertListHeader, convertListRows)
	return nil
}
