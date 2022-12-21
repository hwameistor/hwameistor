package csi

import (
	"errors"
	"strings"

	"github.com/hwameistor/hwameistor/pkg/exechelper"
)

func (p *plugin) getFSTypeByMountPoint(mntPath string) (string, error) {
	result := p.cmdExecutor.RunCommand(exechelper.ExecParams{
		CmdName: "findmnt",
		CmdArgs: []string{"-n", "--output=fstype", "--target", mntPath},
	})
	if result.Error != nil {
		return "", result.Error
	}
	if result.ExitCode != 0 {
		return "", errors.New(result.ErrBuf.String())
	}
	return strings.TrimSuffix(result.OutBuf.String(), "\n"), nil
}

func (p *plugin) getDeviceByMountPoint(mntPath string) (string, error) {
	result := p.cmdExecutor.RunCommand(exechelper.ExecParams{
		CmdName: "findmnt",
		CmdArgs: []string{"-n", "--output=source", "--target", mntPath},
	})
	if result.Error != nil {
		return "", result.Error
	}
	if result.ExitCode != 0 {
		return "", errors.New(result.ErrBuf.String())
	}
	return strings.TrimSuffix(result.OutBuf.String(), "\n"), nil
}

func (p *plugin) expandXFSByMountPoint(mntPath string) error {

	result := p.cmdExecutor.RunCommand(exechelper.ExecParams{
		CmdName: "xfs_growfs",
		CmdArgs: []string{"-d", mntPath},
	})
	if result.Error != nil {
		return result.Error
	}
	if result.ExitCode != 0 {
		return errors.New(result.ErrBuf.String())
	}
	return nil
}

// expandEXTByDevice expands the ext2/3/4 filesystem size
func (p *plugin) expandEXTByDevice(devPath string) error {
	result := p.cmdExecutor.RunCommand(exechelper.ExecParams{
		CmdName: "resize2fs",
		CmdArgs: []string{devPath},
	})
	if result.Error != nil {
		return result.Error
	}
	if result.ExitCode != 0 {
		return errors.New(result.ErrBuf.String())
	}
	return nil
}

func (p *plugin) expandFileSystemByMountPoint(mntPath string) error {
	// Check for filesystem type
	fsType, err := p.getFSTypeByMountPoint(mntPath)
	if err != nil {
		return err
	}

	// 3. resize file system
	if fsType == "xfs" {
		return p.expandXFSByMountPoint(mntPath)
	}

	if strings.HasPrefix(fsType, "ext") {
		devPath, err := p.getDeviceByMountPoint(mntPath)
		if err != nil {
			return err
		}
		return p.expandEXTByDevice(devPath)
	}

	return errors.New("unsupport Filesystem type")
}
