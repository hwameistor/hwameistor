package storageclass

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	"github.com/hwameistor/hwameistor/pkg/apiserver/api"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/formatter"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/manager"
)

var storageClassList = &cobra.Command{
	Use:     "list",
	Args:    cobra.ExactArgs(0),
	Short:   "List the Hwameistor's  storageClasses.",
	Long:    "You can use 'hwameictl sc list' to obtain information about all storage storageClasses.",
	Example: "hwameictl sc list",
	RunE:    storageClassListRunE,
}

func storageClassListRunE(_ *cobra.Command, _ []string) error {
	c, err := manager.NewStorageClassController()
	if err != nil {
		return err
	}

	scs, err := c.ListHwameistorStroageClass(api.QueryPage{PageSize: -1})
	if err != nil {
		return err
	}

	scsHeader := table.Row{"#", "NAME", "PROVISIONER", "RECLAIMPOLICY", "VOLUMEBINDINGMODE", "ALLOWVOLUMEEXPANSION"}
	scsRows := make([]table.Row, len(scs))
	for i, sc := range scs {
		scsRows[i] = table.Row{i + 1, sc.Name, sc.Provisioner, *sc.ReclaimPolicy, *sc.VolumeBindingMode, *sc.AllowVolumeExpansion}
	}

	formatter.PrintTable("StorageClass", scsHeader, scsRows)
	return nil
}
