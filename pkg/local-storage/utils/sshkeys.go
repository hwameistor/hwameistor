package utils

import (
	"os"
	"strings"

	"github.com/hwameistor/hwameistor/pkg/local-storage/utils/datacopy"
)

const (
	sshAuthorizedKeysFilePath = "/root/.ssh/authorized_keys"
)

func AddPubKeyIntoAuthorizedKeys(pubkey string) error {
	txt, err := os.ReadFile(sshAuthorizedKeysFilePath)
	if err != nil {
		return err
	}

	newTxt := ""
	for _, line := range strings.Split(string(txt), "\n") {
		if str := strings.TrimSpace(line); len(str) != 0 {
			if strings.Contains(line, datacopy.RCloneKeyComment) {
				continue
			}
			newTxt += line + "\n"
		}
	}
	// add the rclone pub key at the end of keys file
	newTxt += pubkey + "\n"

	return os.WriteFile(sshAuthorizedKeysFilePath, []byte(newTxt), 0644)
}
