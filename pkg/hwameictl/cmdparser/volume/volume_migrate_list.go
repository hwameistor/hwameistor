package volume

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	"github.com/hwameistor/hwameistor/pkg/hwameictl/formatter"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/manager"
)

var volumeMigrateList = &cobra.Command{
	Use:     "list",
	Args:    cobra.ExactArgs(0),
	Short:   "List all the migrate operations.",
	Long:    "List all the migrate operations.",
	Example: "hwameictl volume migrate list",
	RunE:    volumeMigrateListRunE,
}

func volumeMigrateListRunE(_ *cobra.Command, _ []string) error {
	c, err := manager.NewLocalVolumeController()
	if err != nil {
		return err
	}
	// List all the migrate operations
	migrateList, err := c.ListVolumeMigrate()
	if err != nil {
		return err
	}

	migrateListHeader := table.Row{"#", "Name", "VolumeName", "SourceNode", "TargetNode",
		"Abort", "State", "Message"}
	migrateListRows := make([]table.Row, len(migrateList.Items))
	for i, migrate := range migrateList.Items {
		migrateListRows[i] = table.Row{i + 1, migrate.Name, migrate.Spec.VolumeName,
			migrate.Spec.SourceNode, migrate.Status.TargetNode,
			migrate.Spec.Abort, migrate.Status.State, migrate.Status.Message}
	}

	formatter.PrintTable("Migrate operation list", migrateListHeader, migrateListRows)
	return nil
}
