package qos

import (
	"fmt"
	"path/filepath"
	"syscall"

	"github.com/containerd/cgroups/v3"

	"github.com/hwameistor/hwameistor/pkg/exechelper"
	"github.com/hwameistor/hwameistor/pkg/exechelper/nsexecutor"
)

// VolumeCgroupsManager is the interface to configure QoS for a volume.
type VolumeCgroupsManager interface {
	// ConfigureQoSForDevice configures the QoS for a volume.
	ConfigureQoSForDevice(devPath string, iops, throughput int64) error
}

// NewVolumeCgroupsManager returns a VolumeCgroupsManager according to the cgroups mode.
func NewVolumeCgroupsManager() (VolumeCgroupsManager, error) {
	exec := nsexecutor.New()
	mode := cgroups.Mode()
	switch mode {
	case cgroups.Legacy:
		return &cgroupV1{exec: exec, blkioPath: "/sys/fs/cgroup/blkio"}, nil
	case cgroups.Hybrid:
		// Is systemd driver?
		ioPath := "/sys/fs/cgroup/unified/kubepods.slice/io.max"
		err := exec.RunCommand(exechelper.ExecParams{
			CmdName: "stat",
			CmdArgs: []string{ioPath},
		}).Error
		if err == nil {
			return &cgroupV2{exec, ioPath}, nil
		}
		// Is cgroup driver?
		ioPath = "/sys/fs/cgroup/unified/kubepods/io.max"
		err = exec.RunCommand(exechelper.ExecParams{
			CmdName: "stat",
			CmdArgs: []string{ioPath},
		}).Error
		if err == nil {
			return &cgroupV2{exec, ioPath}, nil
		}
		return &noop{}, nil
	case cgroups.Unified:
		// Is systemd driver?
		ioPath := "/sys/fs/cgroup/kubepods.slice/io.max"
		err := exec.RunCommand(exechelper.ExecParams{
			CmdName: "stat",
			CmdArgs: []string{ioPath},
		}).Error
		if err == nil {
			return &cgroupV2{exec, ioPath}, nil
		}
		// Is cgroup driver?
		ioPath = "/sys/fs/cgroup/kubepods/io.max"
		err = exec.RunCommand(exechelper.ExecParams{
			CmdName: "stat",
			CmdArgs: []string{ioPath},
		}).Error
		if err == nil {
			return &cgroupV2{exec, ioPath}, nil
		}
		return &noop{}, nil
	case cgroups.Unavailable:
		return &noop{}, fmt.Errorf("cgroups is not available")
	}
	return &noop{}, nil
}

var _ VolumeCgroupsManager = &cgroupV1{}

// cgroupV1 is the implementation of VolumeCgroupsManager for cgroup v1.
type cgroupV1 struct {
	blkioPath string
	exec      exechelper.Executor
}

// ConfigureQoSForDevice configures the QoS for a volume.
func (c *cgroupV1) ConfigureQoSForDevice(devPath string, iops, throughput int64) error {
	major, minor, err := getDeviceNumber(devPath)
	if err != nil {
		return err
	}

	filename := filepath.Join(c.blkioPath, "blkio.throttle.read_bps_device")
	err = writeFile(c.exec, filename, fmt.Sprintf("%d:%d %d", major, minor, throughput))
	if err != nil {
		return err
	}

	filename = filepath.Join(c.blkioPath, "blkio.throttle.write_bps_device")
	err = writeFile(c.exec, filename, fmt.Sprintf("%d:%d %d", major, minor, throughput))
	if err != nil {
		return err
	}

	filename = filepath.Join(c.blkioPath, "blkio.throttle.read_iops_device")
	err = writeFile(c.exec, filename, fmt.Sprintf("%d:%d %d", major, minor, iops))
	if err != nil {
		return err
	}

	filename = filepath.Join(c.blkioPath, "blkio.throttle.write_iops_device")
	err = writeFile(c.exec, filename, fmt.Sprintf("%d:%d %d", major, minor, iops))
	if err != nil {
		return err
	}
	return nil
}

var _ VolumeCgroupsManager = &noop{}

type noop struct{}

func (n *noop) ConfigureQoSForDevice(string, int64, int64) error {
	return nil
}

// cgroupV2 is the implementation of VolumeCgroupsManager for cgroup v2.
type cgroupV2 struct {
	exec         exechelper.Executor
	iolimitsPath string
}

// ConfigureQoSForDevice configures the QoS for a volume.
func (c *cgroupV2) ConfigureQoSForDevice(devPath string, iops, throughput int64) error {
	major, minor, err := getDeviceNumber(devPath)
	if err != nil {
		return err
	}

	if throughput == 0 {
		err = writeFile(c.exec, c.iolimitsPath, fmt.Sprintf("%d:%d rbps=max wbps=max", major, minor))
		if err != nil {
			return err
		}
	} else {
		err = writeFile(c.exec, c.iolimitsPath, fmt.Sprintf("%d:%d rbps=%d wbps=%d", major, minor, throughput, throughput))
		if err != nil {
			return err
		}
	}

	if iops == 0 {
		err = writeFile(c.exec, c.iolimitsPath, fmt.Sprintf("%d:%d riops=max wiops=max", major, minor))
		if err != nil {
			return err
		}
	} else {
		err = writeFile(c.exec, c.iolimitsPath, fmt.Sprintf("%d:%d riops=%d wiops=%d", major, minor, iops, iops))
		if err != nil {
			return err
		}
	}
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
