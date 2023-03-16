package csi

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	csi "github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/gofrs/uuid"

	"github.com/hwameistor/hwameistor/pkg/exechelper"
	"github.com/hwameistor/hwameistor/pkg/exechelper/nsexecutor"
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
	headers := []string{
		"Inodes",
		"IFree",
		"IUsed",
		"1B-blocks",
		"Avail",
		"Used",
	}
	dfFlags := []string{
		"--sync",
		"--block-size=1",
		"--output=itotal,iavail,iused,size,avail,used",
		mntPoint,
	}
	dfPath := "df"

	res := nsexecutor.New().RunCommand(exechelper.ExecParams{
		CmdName: dfPath,
		CmdArgs: dfFlags,
	})

	if res.ExitCode != 0 {
		return nil, res.Error
	}
	if res.OutBuf == nil {
		return nil, fmt.Errorf("no output")
	}

	for _, line := range strings.Split(res.OutBuf.String(), "\n") {
		line = strings.TrimSpace(line)
		// Skip for empty line or before header found.
		if len(line) == 0 {
			continue
		}

		if strings.Contains(line, headers[0]) {
			// skip the header line
			continue
		}
		// Split line into array
		fields := strings.Fields(line)
		if len(fields) != len(headers) {
			return nil, fmt.Errorf("invalid output")
		}
		inodeTotal, err := strconv.ParseInt(fields[0], 10, 64)
		if err != nil {
			return nil, err
		}
		inodeFree, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			return nil, err
		}
		inodeUsed, err := strconv.ParseInt(fields[2], 10, 64)
		if err != nil {
			return nil, err
		}
		capacityTotal, err := strconv.ParseInt(fields[3], 10, 64)
		if err != nil {
			return nil, err
		}
		capacityFree, err := strconv.ParseInt(fields[4], 10, 64)
		if err != nil {
			return nil, err
		}
		capacityUsed, err := strconv.ParseInt(fields[5], 10, 64)
		if err != nil {
			return nil, err
		}
		return &VolumeMetrics{
			TotalCapacityBytes: capacityTotal,
			UsedCapacityBytes:  capacityUsed,
			FreeCapacityBytes:  capacityFree,
			TotalINodeNumber:   inodeTotal,
			UsedINodeNumber:    inodeUsed,
			FreeINodeNumber:    inodeFree,
		}, nil
	}

	return nil, fmt.Errorf("not found")
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
