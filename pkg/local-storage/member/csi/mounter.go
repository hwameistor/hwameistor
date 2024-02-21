package csi

import (
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	utilexec "k8s.io/utils/exec"
	"k8s.io/utils/mount"

	"github.com/hwameistor/hwameistor/pkg/exechelper"
	"github.com/hwameistor/hwameistor/pkg/exechelper/nsexecutor"
)

// Mounter interface
//
//go:generate mockgen -source=mounter.go -destination=../../member/csi/mounter_mock.go  -package=csi
type Mounter interface {
	MountRawBlock(devPath string, mountpoint string) error
	BindMount(devPath string, mountpoint string) error
	FormatAndMount(devPath string, mountpoint string, fsType string, flags []string) error
	Unmount(mountpoint string) error
	GetDeviceMountPoints(devPath string) []string
}

type linuxMounter struct {
	mounter *mount.SafeFormatAndMount

	logger *log.Entry
}

// NewLinuxMounter creates a mounter
func NewLinuxMounter(logger *log.Entry) Mounter {
	return &linuxMounter{
		mounter: &mount.SafeFormatAndMount{
			Interface: mount.New("/bin/mount"),
			Exec:      utilexec.New(),
		},
		logger: logger,
	}
}

func (m *linuxMounter) FormatAndMount(devPath string, mountPoint string, fsType string, options []string) error {
	if err := makeDir(mountPoint); err != nil {
		m.logger.WithFields(log.Fields{"mountpoint": mountPoint, "error": err.Error()}).Error("Failed to create mountpoint directory")
		return err
	}

	return m.mounter.FormatAndMount(devPath, mountPoint, fsType, options)
}

func (m *linuxMounter) MountRawBlock(devPath string, mountPoint string) error {

	if err := makeFile(mountPoint); err != nil {
		m.logger.WithFields(log.Fields{"mountpoint": mountPoint, "error": err.Error()}).Error("Failed to create mountpoint file")
		return err
	}

	return m.mounter.Mount(devPath, mountPoint, "", []string{"bind"})
}

func (m *linuxMounter) BindMount(devPath string, mountPoint string) error {

	if err := makeDir(mountPoint); err != nil {
		m.logger.WithFields(log.Fields{"mountpoint": mountPoint, "error": err.Error()}).Error("Failed to create mountpoint directory")
		return err
	}

	return m.doBindMount(devPath, mountPoint)
}

func (m *linuxMounter) doBindMount(devPath string, mountPoint string) error {

	notMounted, err := m.mounter.IsLikelyNotMountPoint(mountPoint)
	if err != nil {
		m.logger.WithFields(log.Fields{"mountpoint": mountPoint, "error": err.Error()}).Error("Failed to check mountpoint")
		return err
	}
	if !notMounted {
		m.logger.WithFields(log.Fields{"mountpoint": mountPoint, "error": "wrong status of mountpoint"}).Error("Already mounted by others")
		return fmt.Errorf("wrong status of mountpoint")
	}

	if !m.isNotBindMountPoint(mountPoint) {
		// already mounted
		m.logger.WithFields(log.Fields{"devpath": devPath, "mountpoint": mountPoint}).Debug("Already bind mounted")
		return nil
	}

	err = m.bindMount(devPath, mountPoint)
	if err != nil {
		m.logger.WithFields(log.Fields{"devpath": devPath, "mountpoint": mountPoint}).WithError(err).Error("Failed to exec bind mount")
		return err
	}
	m.logger.WithFields(log.Fields{"devpath": devPath, "mountpoint": mountPoint}).Debug("Bind mounted successfully")
	return nil
}

func (m *linuxMounter) bindMount(devPath string, mountPoint string) error {
	params := exechelper.ExecParams{
		CmdName: "mount",
		CmdArgs: []string{"--bind", devPath, mountPoint},
	}
	result := nsexecutor.New().RunCommand(params)
	if result.ExitCode == 0 {
		return nil
	}
	return result.Error
}

func (m *linuxMounter) Unmount(mountPoint string) error {
	if exist, err := isPathExist(mountPoint); err != nil {
		m.logger.WithError(err).Errorf("failed to check if mountPoint %s exist", mountPoint)
		return err
	} else if !exist {
		m.logger.WithFields(log.Fields{"mountpoint": mountPoint}).Info("Already unmounted and deleted")
		return nil
	}

	notMounted, err := m.isNotMountPoint(mountPoint)
	if err != nil {
		m.logger.WithFields(log.Fields{"mountpoint": mountPoint, "error": err.Error()}).Error("Failed to check mountpoint")
		return err
	}
	if !notMounted {
		if err = m.mounter.Unmount(mountPoint); err != nil {
			m.logger.WithFields(log.Fields{"mountpoint": mountPoint}).WithError(err).Error("Failed to unmount")
			return err
		}
		m.logger.WithFields(log.Fields{"mountpoint": mountPoint}).Info("Succeed to unmount")
	}
	return removeFile(mountPoint)
}

func (m *linuxMounter) isNotMountPoint(mountPoint string) (bool, error) {
	// check for bind mountpoint firstly
	if !m.isNotBindMountPoint(mountPoint) {
		return false, nil
	}

	// check for regular mountpoint
	return m.mounter.IsLikelyNotMountPoint(mountPoint)
}

func (m *linuxMounter) isNotBindMountPoint(mountPoint string) bool {

	result := nsexecutor.New().RunCommand(exechelper.ExecParams{
		CmdName: "findmnt",
		CmdArgs: []string{"-n", "--output=source", "--target", mountPoint},
	})
	if result.ExitCode != 0 {
		return true
	}
	return !strings.HasPrefix(result.OutBuf.String(), "ramdisk")
}

func (m *linuxMounter) GetDeviceMountPoints(devPath string) []string {

	mps := []string{}
	result := nsexecutor.New().RunCommand(exechelper.ExecParams{
		CmdName: "findmnt",
		CmdArgs: []string{"-n", "--output=target", "--source", devPath},
	})
	if result.ExitCode == 0 {
		for _, mp := range strings.Split(result.OutBuf.String(), "\n") {
			if strings.Trim(mp, " ") != "" {
				mps = append(mps, mp)
			}
		}
	}
	return mps
}

func isPathExist(pathname string) (bool, error) {
	if _, err := os.Stat(pathname); err != nil {
		return false, err
	}
	// return true when this path is file or directory
	return true, nil
}

func makeDir(pathname string) error {
	err := os.MkdirAll(pathname, os.FileMode(0777))
	if err != nil && !os.IsExist(err) {
		return err
	}
	return nil
}

func removeDir(pathname string) error {
	return os.RemoveAll(pathname)
}

func makeFile(pathname string) error {
	f, err := os.OpenFile(pathname, os.O_CREATE, os.FileMode(0666))
	if err != nil && !os.IsExist(err) {
		return err
	}
	return f.Close()
}

func removeFile(pathname string) error {
	return os.Remove(pathname)
}
