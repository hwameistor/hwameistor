package node

import (
	"fmt"
	"path/filepath"
	"syscall"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/exechelper"
	"github.com/hwameistor/hwameistor/pkg/exechelper/nsexecutor"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/resource"
)

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

// writeBlkioFile writes value into the given filename.
func (m *manager) writeBlkioFile(filename, value string) error {
	filePath := filepath.Join("/sys/fs/cgroup/blkio", filename)
	cmdExecutor := nsexecutor.New()
	result := cmdExecutor.RunCommand(exechelper.ExecParams{
		CmdName: "sh",
		CmdArgs: []string{"-c", fmt.Sprintf("echo %s >> %s", value, filePath)},
	})
	return result.Error
}

func (m *manager) configureQoS(replica *apisv1alpha1.LocalVolumeReplica) error {
	logCtx := m.logger.WithFields(log.Fields{"replica": replica.Name, "spec": replica.Spec, "status": replica.Status})
	logCtx.Debug("Configure Volume QoS")

	logCtx.Debug("Ensure Volume Qos")
	storagePath := replica.Status.StoragePath
	if len(storagePath) == 0 {
		storagePath = replica.Status.DevicePath
	}
	major, minor, err := getDeviceNumber(storagePath)
	if err != nil {
		m.logger.WithError(err).Error("Failed to get device number")
		return err
	}
	m.logger.Debugf("Device number: %d:%d", major, minor)

	throughput := resource.MustParse("0")
	if replica.Spec.VolumeQoS.Throughput != "" {
		throughput, err = resource.ParseQuantity(replica.Spec.VolumeQoS.Throughput)
		if err != nil {
			m.logger.WithError(err).Error("Failed to parse throughput")
			return err
		}
	}
	err = m.writeBlkioFile("blkio.throttle.read_bps_device", fmt.Sprintf("%d:%d %d", major, minor, throughput.Value()))
	if err != nil {
		return err
	}
	err = m.writeBlkioFile("blkio.throttle.write_bps_device", fmt.Sprintf("%d:%d %d", major, minor, throughput.Value()))
	if err != nil {
		return err
	}

	iops := resource.MustParse("0")
	if replica.Spec.VolumeQoS.IOPS != "" {
		iops, err = resource.ParseQuantity(replica.Spec.VolumeQoS.IOPS)
		if err != nil {
			m.logger.WithError(err).Error("Failed to parse iops")
			return err
		}
	}
	err = m.writeBlkioFile("blkio.throttle.read_iops_device", fmt.Sprintf("%d:%d %d", major, minor, iops.Value()))
	if err != nil {
		return err
	}
	err = m.writeBlkioFile("blkio.throttle.write_iops_device", fmt.Sprintf("%d:%d %d", major, minor, iops.Value()))
	if err != nil {
		return err
	}

	return nil
}
