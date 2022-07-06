package exechelper

import (
	"bytes"
)

// Executor is the interface for executing commands.
type Executor interface {
	RunCommand(params ExecParams) ExecResult
}

// ExecParams parameters to execute a command
type ExecParams struct {
	CmdName string
	CmdArgs []string
	Timeout int
}

// ExecResult result of executing a command
type ExecResult struct {
	OutBuf   *bytes.Buffer
	ErrBuf   *bytes.Buffer
	ExitCode int
	Error    error
}
