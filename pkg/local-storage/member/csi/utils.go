package csi

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	csi "github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/gofrs/uuid"
)

func newControllerServiceCapability(cap csi.ControllerServiceCapability_RPC_Type) *csi.ControllerServiceCapability {
	return &csi.ControllerServiceCapability{
		Type: &csi.ControllerServiceCapability_Rpc{
			Rpc: &csi.ControllerServiceCapability_RPC{
				Type: cap,
			},
		},
	}
}

func newNodeServiceCapability(cap csi.NodeServiceCapability_RPC_Type) *csi.NodeServiceCapability {
	return &csi.NodeServiceCapability{
		Type: &csi.NodeServiceCapability_Rpc{
			Rpc: &csi.NodeServiceCapability_RPC{
				Type: cap,
			},
		},
	}
}

func newPluginCapability(cap csi.PluginCapability_Service_Type) *csi.PluginCapability {
	return &csi.PluginCapability{
		Type: &csi.PluginCapability_Service_{
			Service: &csi.PluginCapability_Service{
				Type: cap,
			},
		},
	}
}

// parseEndpoint parse socket endpoint
func parseEndpoint(ep string) (string, string, error) {
	if strings.HasPrefix(strings.ToLower(ep), "unix://") || strings.HasPrefix(strings.ToLower(ep), "tcp://") {
		s := strings.SplitN(ep, "://", 2)
		if s[1] != "" {
			return s[0], s[1], nil
		}
	}
	return "", "", fmt.Errorf("invalid endpoint: %v", ep)
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func getVolumeMetrics(mntPoint string) (*VolumeMetrics, error) {
	var stats syscall.Statfs_t

	syscall.Sync()

	err := syscall.Statfs(mntPoint, &stats)

	if err != nil {
		return nil, err
	}

	return &VolumeMetrics{
		TotalCapacityBytes: int64(stats.Blocks * uint64(stats.Bsize)),
		UsedCapacityBytes:  int64((stats.Blocks - stats.Bfree) * uint64(stats.Bsize)),
		FreeCapacityBytes:  int64(stats.Bavail * uint64(stats.Bsize)),
		TotalINodeNumber:   int64(stats.Files),
		UsedINodeNumber:    int64(stats.Files - stats.Ffree),
		FreeINodeNumber:    int64(stats.Ffree),
	}, nil
}

func isStringInArray(str string, strs []string) bool {
	for _, s := range strs {
		if str == s {
			return true
		}
	}
	return false
}

func genUUID() string {
	return fmt.Sprint(uuid.Must(uuid.NewV4()))
}
