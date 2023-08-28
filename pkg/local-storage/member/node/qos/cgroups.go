package qos

import (
	"fmt"
	"path/filepath"
	"syscall"

	"github.com/containerd/cgroups/v3"
	"github.com/hwameistor/hwameistor/pkg/exechelper"
	"github.com/hwameistor/hwameistor/pkg/exechelper/nsexecutor"
)

const (
	blkioPath = "/sys/fs/cgroup/blkio"
)

// VolumeCgroupsManager is the interface to configure QoS for a volume.
type VolumeCgroupsManager interface {
	// ConfigureQoSForDevice configures the QoS for a volume.
	ConfigureQoSForDevice(devPath string, iops, throughput int64) error
}

// NewVolumeCgroupsManager returns a VolumeCgroupsManager according to the cgroups mode.
func NewVolumeCgroupsManager() (VolumeCgroupsManager, error) {
	mode := cgroups.Mode()
	switch mode {
	case cgroups.Legacy:
		return &cgroupV1{nsexecutor.New()}, nil
	case cgroups.Unified, cgroups.Hybrid:
		// TODO: support cgroup v2
		return &noop{}, nil
	case cgroups.Unavailable:
		return &noop{}, fmt.Errorf("cgroups is not available")
	}
	return &noop{}, nil
}

var _ VolumeCgroupsManager = &cgroupV1{}

// cgroupV1 is the implementation of VolumeCgroupsManager for cgroup v1.
type cgroupV1 struct {
	exec exechelper.Executor
}

// ConfigureQoSForDevice configures the QoS for a volume.
func (c *cgroupV1) ConfigureQoSForDevice(devPath string, iops, throughput int64) error {
	major, minor, err := getDeviceNumber(devPath)
	if err != nil {
		return err
	}

	filename := filepath.Join(blkioPath, "blkio.throttle.read_bps_device")
	err = writeFile(c.exec, filename, fmt.Sprintf("%d:%d %d", major, minor, throughput))
	if err != nil {
		return err
	}

	filename = filepath.Join(blkioPath, "blkio.throttle.write_bps_device")
	err = writeFile(c.exec, filename, fmt.Sprintf("%d:%d %d", major, minor, throughput))
	if err != nil {
		return err
	}

	filename = filepath.Join(blkioPath, "blkio.throttle.read_iops_device")
	err = writeFile(c.exec, filename, fmt.Sprintf("%d:%d %d", major, minor, iops))
	if err != nil {
		return err
	}

	filename = filepath.Join(blkioPath, "blkio.throttle.write_iops_device")
	err = writeFile(c.exec, filename, fmt.Sprintf("%d:%d %d", major, minor, iops))
	if err != nil {
		return err
	}
	return nil
}

var _ VolumeCgroupsManager = &noop{}

type noop struct{}

func (n *noop) ConfigureQoSForDevice(devPath string, iops, throughput int64) error {
	return nil
}

// getDeviceNumber return the major and minor of a device according to the devicePath.
func getDeviceNumber(devicePath string) (uint64, uint64, error) {
	stat := syscall.Stat_t{}
	err := syscall.Stat(devicePath, &stat)
	if err != nil {
		return 0, 0, err
	}
	maj := uint64(stat.Rdev / 256)
	min := uint64(stat.Rdev % 256)
	return maj, min, nil
}

func writeFile(exec exechelper.Executor, filename, value string) error {
	result := exec.RunCommand(exechelper.ExecParams{
		CmdName: "sh",
		CmdArgs: []string{"-c", fmt.Sprintf("echo %s >> %s", value, filename)},
	})
	return result.Error
}
