package volume

import "github.com/spf13/cobra"

var volumeMigrate = &cobra.Command{
	Use:   "migrate {volumeName} {sourceNode} {?targetNode}",
	Short: "",
	Long:  "",
	RunE:  volumeMigrateRunE,
}

func volumeMigrateRunE(cmd *cobra.Command, args []string) error {
	return nil
}
