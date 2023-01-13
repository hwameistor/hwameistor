package utils

import (
	"bytes"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

func Bash(cmd string) (string, error) {
	var (
		stdout  bytes.Buffer
		stderr  bytes.Buffer
		execCmd *exec.Cmd
	)

	execCmd = exec.Command("bash", "-c", cmd)
	execCmd.Stderr = &stderr
	execCmd.Stdout = &stdout

	log.Info(execCmd.String())
	err := execCmd.Run()
	return stdout.String(), err
}
