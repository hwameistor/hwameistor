package volume

import (
	"fmt"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/cmdparser/definitions"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/manager"
	"github.com/spf13/cobra"
)

var volumeGet = &cobra.Command{
	Use:   "get",
	Short: "",
	Long:  "",
	RunE:  volumeGetRunE,
}

func init() {
	// Volume get flags
	volumeGet.Flags().String("node", "", "Filter Volumes by node name")
	volumeGet.Flags().String("group", "", "Filter Volumes by group name")
}

func volumeGetRunE(cmd *cobra.Command, args []string) error {
	m, e := manager.NewServerManager(definitions.Kubeconfig)
	fmt.Println("e:", e)
	fmt.Println("list:")
	fmt.Println(m.LocalDiskController().ListLocalDisk())
	fmt.Println("list end")
	return nil
}
