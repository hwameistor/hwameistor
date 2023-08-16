package volume

import (
	"fmt"
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
	c, err := manager.NewLocalVolumeController()
	if err != nil {
		return err
	}
	fmt.Println(c.GetLocalVolume("pvc-29f19e12-145d-453e-89ca-6bf9d1cf852a"))
	return nil
}
