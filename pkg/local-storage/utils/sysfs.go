package utils

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils"
	util "github.com/hwameistor/hwameistor/pkg/utils"
	log "github.com/sirupsen/logrus"
)

func WriteDataIntoSysFSFile(content, sysFilePath string) error {
	_, toucherr := utils.Bash(fmt.Sprintf("touch  %s", sysFilePath))
	if toucherr != nil {
		log.Debugf("WriteDataIntoSysFSFile touch %v,Error = %v", sysFilePath, toucherr)
	}

	authorizedKeyOut, err := utils.Bash(fmt.Sprintf("cat ~/.ssh/authorized_keys"))
	if err != nil {
		log.Debugf("WriteDataIntoSysFSFile cat ~/.ssh/authorized_keys ,Error = %v", err)
		return err
	}

	if !strings.Contains(authorizedKeyOut, content) {
		if writefileErr := util.WriteFile(sysFilePath, bytes.NewBuffer([]byte(content))); writefileErr != nil {
			log.Debugf("WriteDataIntoSysFSFile WriteFile err = %v", writefileErr)
			return writefileErr
		}
		_, appendCatErr := utils.Bash(fmt.Sprintf("cat %s >> ~/.ssh/authorized_keys", sysFilePath))
		if appendCatErr != nil {
			log.Debugf("WriteDataIntoSysFSFile cat %s >> ~/.ssh/authorized_keys Error = %v", sysFilePath, appendCatErr)
			return appendCatErr
		}
	}

	log.Debugf("WriteDataIntoSysFSFile cat %s >> ~/.ssh/authorized_keys succeeded", sysFilePath)
	return nil
}
