package encrypt

import (
	"github.com/hwameistor/hwameistor/pkg/exechelper"
	"github.com/hwameistor/hwameistor/pkg/exechelper/basicexecutor"
	"github.com/hwameistor/hwameistor/pkg/exechelper/nsexecutor"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
	"strings"
)

var _ Encryptor = &LUKS{}

type LUKS struct {
	nsCmdExec    exechelper.Executor
	basicCmdExec exechelper.Executor
	logger       *log.Entry
}

func NewLUKS() *LUKS {
	return &LUKS{
		nsCmdExec:    nsexecutor.New(),
		basicCmdExec: basicexecutor.New(),
		logger:       log.New().WithField("Module", "encrypt/LUKS"),
	}
}
func (lk *LUKS) EncryptVolume(volumePath string, secret string) error {
	lk.logger.WithField("volumePath", volumePath).Debug("Encrypting volume with LUKS")

	// check if volume exists
	checkVolume := exechelper.ExecParams{
		CmdName: "lvs",
		CmdArgs: []string{"--noheadings", "--readonly", "-o", "lv_name", volumePath},
		Timeout: 0,
	}
	res := lk.nsCmdExec.RunCommand(checkVolume)
	if res.Error != nil {
		lk.logger.WithError(res.Error).Error("Failed to check if volume exists, volume might not exist")
		return res.Error
	}

	// setup encrypt volume
	fh := FileHandler{}
	if err := fh.WriteToFile(secret); err != nil {
		_ = fh.DeleteFile()
		lk.logger.WithError(err).Error("Failed to write secret to file")
		return err
	}
	defer fh.DeleteFile()

	encryptVolume := exechelper.ExecParams{
		CmdName: "cryptsetup",
		CmdArgs: []string{"-q", "-s", "512", "luksFormat", volumePath, fh.FilePath},
	}
	res = lk.basicCmdExec.RunCommand(encryptVolume)
	if res.Error != nil {
		lk.logger.WithError(res.Error).Error("Failed to encrypt volume")
		return res.Error
	}

	lk.logger.WithField("volumePath", volumePath).Debug("Encrypted volume successfully")
	return nil
}

func (lk *LUKS) IsVolumeEncrypted(volumePath string) (bool, error) {
	checkVolumeEncrypted := exechelper.ExecParams{
		CmdName: "cryptsetup",
		CmdArgs: []string{"isLuks", volumePath},
		Timeout: 0,
	}
	res := lk.basicCmdExec.RunCommand(checkVolumeEncrypted)
	if res.Error != nil && res.ExitCode != 1 {
		lk.logger.WithError(res.Error).Error("Failed to check if volume encrypted")
		return false, res.Error
	}

	return res.ExitCode == 0, nil
}

func (lk *LUKS) DecryptVolume(volumePath string, secret string) error {
	//TODO implement me
	panic("implement me")
}

func (lk *LUKS) OpenVolume(volumePath string, secret string) (string, error) {
	ss := strings.Split(volumePath, "/")
	volumeName := ss[len(ss)-1]
	volumeEncryptPath := volumeName + "-encrypt"

	fh := FileHandler{}
	if err := fh.WriteToFile(secret); err != nil {
		_ = fh.DeleteFile()
		lk.logger.WithError(err).Error("Failed to write secret to file")
		return "", err
	}
	defer fh.DeleteFile()

	openVolume := exechelper.ExecParams{
		CmdName: "cryptsetup",
		CmdArgs: []string{"--allow-discards", "luksOpen", "-d", fh.FilePath, volumePath, volumeEncryptPath},
	}
	res := lk.basicCmdExec.RunCommand(openVolume)
	if res.Error != nil && !strings.Contains(res.ErrBuf.String(), "already exists") {
		lk.logger.WithError(res.Error).Error("Failed to open encrypted volume")
		return "", res.Error
	}

	return path.Join("/dev/mapper", volumeEncryptPath), nil
}

func (lk *LUKS) CloseVolume(volumePath string) error {
	closeVolume := exechelper.ExecParams{
		CmdName: "cryptsetup",
		CmdArgs: []string{"luksClose", volumePath},
	}
	res := lk.basicCmdExec.RunCommand(closeVolume)
	if res.Error != nil && !strings.Contains(res.ErrBuf.String(), "is not active") {
		lk.logger.WithError(res.Error).Error("Failed to close encrypted volume")
		return res.Error
	}

	return nil
}

type FileHandler struct {
	FilePath string
}

func (f *FileHandler) WriteToFile(content string) error {
	file, err := os.CreateTemp(os.TempDir(), "encrypt-*.txt")
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(content)
	if err != nil {
		return err
	}

	f.FilePath = file.Name()
	return nil
}

func (f *FileHandler) DeleteFile() error {
	return os.Remove(f.FilePath)
}
