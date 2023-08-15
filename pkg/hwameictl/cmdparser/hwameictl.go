package cmdparser

import (
	"github.com/hwameistor/hwameistor/pkg/hwameictl/cmdparser/definitions"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/cmdparser/volume"
	"github.com/spf13/cobra"
	"time"
)

var Hwameictl = &cobra.Command{
	Use:   "hwameictl",
	Short: definitions.CmdHelpMessages["hwameictl"]["hwameictl"].Short,
	Long:  definitions.CmdHelpMessages["hwameictl"]["hwameictl"].Long,
	RunE: func(cmd *cobra.Command, args []string) error {
		// root cmd will show help only
		return cmd.Help()
	},
}

func init() {
	// Hwameictl flags
	Hwameictl.Flags().StringVar(&definitions.Kubeconfig, "kubeconfig", "~/.kube/config", "Specify the kubeconfig file")
	Hwameictl.Flags().DurationVar(&definitions.Timeout, "timeout", 5*time.Second, "Set the request timeout")

	// Sub command
	Hwameictl.AddCommand(volume.Volume)
}
