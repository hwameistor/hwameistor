package cmdparser

import (
	"github.com/hwameistor/hwameistor/pkg/hwameictl/cmdparser/localvolumesnapshot"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/cmdparser/snapshotclass"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/cmdparser/storageclass"
	"io"
	"time"

	"github.com/hwameistor/hwameistor/pkg/hwameictl/cmdparser/cluster"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/cmdparser/definitions"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/cmdparser/disk"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/cmdparser/node"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/cmdparser/pool"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/cmdparser/snapshot"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/cmdparser/volume"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var Hwameictl = &cobra.Command{
	Use:   "hwameictl",
	Args:  cobra.ExactArgs(0),
	Short: "Hwameictl is the command-line tool for Hwameistor.",
	Long: "Hwameictl is a tool that can manage all Hwameistor resources and their entire lifecycle.\n" +
		"Complete documentation is available at https://hwameistor.io/",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Root cmd will show help only
		return cmd.Help()
	},
}

func init() {
	// Hwameictl flags
	Hwameictl.PersistentFlags().BoolVar(&definitions.Debug, "debug", false, "Enable debug mode")
	Hwameictl.PersistentFlags().StringVar(&definitions.KubeConfigPath, "kubeconfig", definitions.DefaultKubeConfigPath, "Specify the kubeconfig file")
	Hwameictl.PersistentFlags().DurationVar(&definitions.Timeout, "timeout", 3*time.Second, "Set the request timeout")

	// Sub commands
	Hwameictl.AddCommand(volume.Volume, node.Node, pool.Pool, disk.Disk,
		storageclass.StorageClass, snapshotclass.SnapshotClass, snapshot.Snapshot,
		localvolumesnapshot.LocalSnapshot, cluster.Cluster)

	// Disable debug mode
	if definitions.Debug == false {
		log.SetOutput(io.Discard)
	}
}
