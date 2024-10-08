package encrypt

import (
	"github.com/hwameistor/hwameistor/pkg/exechelper"
	"github.com/hwameistor/hwameistor/pkg/exechelper/nsexecutor"
	log "github.com/sirupsen/logrus"
	"strings"
)

var _ Encryptor = &LUKS{}

type LUKS struct {
	cmdExec exechelper.Executor
	logger  *log.Entry
}

func NewLUKS() *LUKS {
	return &LUKS{
		cmdExec: nsexecutor.New(),
		logger:  log.New().WithField("Module", "encrypt/LUKS"),
	}
}
func (lk *LUKS) EncryptVolume(volumeGroupName string /* volumeGroup/volumeName */, secret string) error {
	lk.logger.WithField("volumeName", volumeGroupName).Debug("Encrypting volume with LUKS")

	// check if volume exists
	checkVolume := exechelper.ExecParams{
		CmdName: "lvs",
		CmdArgs: []string{"--noheadings", "--readonly", "-o", "lv_name", volumeGroupName},
		Timeout: 0,
	}
	res := lk.cmdExec.RunCommand(checkVolume)
	if res.Error != nil {
		lk.logger.WithError(res.Error).Error("Failed to check if volume exists, volume might not exist")
		return res.Error
	}

	// setup encrypt volume
	volumePathQuery := exechelper.ExecParams{
		CmdName: "lvs",
		CmdArgs: []string{"--noheadings", "--readonly", "-o", "lv_path", volumeGroupName},
		Timeout: 0,
	}
	res = lk.cmdExec.RunCommand(volumePathQuery)
	if res.Error != nil {
		lk.logger.WithError(res.Error).Error("Failed to get volume path")
		return res.Error
	}
	volumePath := res.OutBuf.String()

	encryptVolume := exechelper.ExecParams{
		CmdName: "cryptsetup",
		CmdArgs: []string{"-q", "-s", "512", "luksFormat", volumePath, secret},
	}
	res = lk.cmdExec.RunCommand(encryptVolume)
	if res.Error != nil {
		lk.logger.WithError(res.Error).Error("Failed to encrypt volume")
		return res.Error
	}

	lk.logger.WithField("volumeName", volumeGroupName).Debug("Encrypted volume successfully")
	return nil
}

func (lk *LUKS) DecryptVolume(volumeName string, secret string) error {
	//TODO implement me
	panic("implement me")
}

func (lk *LUKS) OpenVolume(volumeGroupName string, secret string) (string, error) {
	volumePathQuery := exechelper.ExecParams{
		CmdName: "lvs",
		CmdArgs: []string{"--noheadings", "--readonly", "-o", "lv_path", volumeGroupName},
		Timeout: 0,
	}

	res := lk.cmdExec.RunCommand(volumePathQuery)
	if res.Error != nil {
		lk.logger.WithError(res.Error).Error("Failed to get volume path")
		return "", res.Error
	}
	volumePath := res.OutBuf.String()
	volumeName := strings.Split(volumeGroupName, "/")[1]
	volumeEncryptPath := volumeName + "-encrypt"

	openVolume := exechelper.ExecParams{
		CmdName: "cryptsetup",
		CmdArgs: []string{"--allow-discards", "luksOpen", "-d", secret, volumePath, volumeEncryptPath},
	}
	res = lk.cmdExec.RunCommand(openVolume)
	if res.Error != nil {
		lk.logger.WithError(res.Error).Error("Failed to open encrypted volume")
		return "", res.Error
	}

	return volumeEncryptPath, nil
}

func (lk *LUKS) CloseVolume(volumeName string) error {
	//TODO implement me
	panic("implement me")
}
