package utils

import (
	"bytes"
	log "github.com/sirupsen/logrus"
	"os/exec"
	"strings"
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

// ConvertShellOutputs
func ConvertShellOutputs(outputs string) []string {
	var result []string
	if len(outputs) == 0 {
		return result
	}

	start := 0
	for _, index := range GetAllIndex(outputs, "\n") {
		result = append(result, outputs[start:index])
		start = index + 1
	}

	if !strings.HasSuffix(outputs, "\n") {
		result = append(result, outputs[strings.LastIndex(outputs, "\n")+1:])
	}

	return result
}

// GetAllIndex
func GetAllIndex(s string, substr string) []int {
	var indexes []int

	start := 0
	end := len(s)

	for start < end {
		if index := strings.Index(s[start:end], substr); index > -1 {
			indexes = append(indexes, start+index)
			start = start + index + len(substr)
		} else {
			break
		}
	}

	return indexes
}
