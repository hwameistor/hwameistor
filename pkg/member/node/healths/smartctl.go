package healths

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/hwameistor/local-storage/pkg/exechelper"
	"github.com/hwameistor/local-storage/pkg/exechelper/basicexecutor"
	"github.com/hwameistor/local-storage/pkg/utils"
)

// DiskChecker interface
type DiskChecker interface {
	IsDiskHealthy(devPath string) (bool, error)

	GetLocalDisksAll() ([]DeviceInfo, error)

	CheckHealthForLocalDisk(device *DeviceInfo) (*DiskCheckResult, error)
}

type smartCtlr struct {
	cmdExec exechelper.Executor

	cmdName string
}

// NewSmartCtl creates an instance of smartctl
func NewSmartCtl() DiskChecker {
	return &smartCtlr{
		cmdName: "smartctl",
		cmdExec: basicexecutor.New(),
	}
}

func (sc *smartCtlr) IsDiskHealthy(devPath string) (bool, error) {
	checkRes, err := sc.checkForDisk(&DeviceInfo{Name: devPath})
	if err != nil {
		return false, err
	}
	if checkRes.Device == nil {
		return false, fmt.Errorf("not found device info")
	}
	if checkRes.IsVirtualDisk() {
		// always in healthy for virtual disk
		return true, nil
	}
	if checkRes.SmartStatus != nil {
		return checkRes.SmartStatus.Passed, nil
	}
	return false, fmt.Errorf("failed to check")
}

func (sc *smartCtlr) GetLocalDisksAll() ([]DeviceInfo, error) {
	return sc.scanForLocalDisks()
}

func (sc *smartCtlr) CheckHealthForLocalDisk(device *DeviceInfo) (*DiskCheckResult, error) {
	return sc.checkForDisk(device)
}

func (sc *smartCtlr) scanForLocalDisks() ([]DeviceInfo, error) {

	result := sc.cmdExec.RunCommand(exechelper.ExecParams{
		CmdName: sc.cmdName,
		CmdArgs: []string{"--scan", "--json"},
	})
	if result.ExitCode != 0 {
		return []DeviceInfo{}, result.Error
	}

	scanResult := &SmartCtlScanResult{}
	if err := json.Unmarshal(result.OutBuf.Bytes(), scanResult); err != nil {
		return []DeviceInfo{}, err
	}

	// filter out non-local disks (i.e. iscsi disks)
	pciDisks, err := utils.GetPCIDisks(sc.cmdExec)
	if err != nil {
		return []DeviceInfo{}, err
	}
	devices := []DeviceInfo{}
	for i, device := range scanResult.Devices {
		if _, exists := pciDisks[filepath.Base(device.InfoName)]; exists || strings.Contains(device.Type, ",") {
			// pci disk or RAID slave disk
			devices = append(devices, scanResult.Devices[i])
		}
	}
	return devices, nil
}

func (sc *smartCtlr) checkForDisk(device *DeviceInfo) (*DiskCheckResult, error) {
	cmdArgs := []string{"-x", device.Name, "--json"}
	if device.Type != "" {
		cmdArgs = append(cmdArgs, "-d", device.Type)
	}
	result := sc.cmdExec.RunCommand(exechelper.ExecParams{
		CmdName: sc.cmdName,
		CmdArgs: cmdArgs,
	})

	checkResult := &DiskCheckResult{}
	err := json.Unmarshal(result.OutBuf.Bytes(), checkResult)
	return checkResult, err

}
