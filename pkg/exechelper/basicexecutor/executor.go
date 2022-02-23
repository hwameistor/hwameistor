package basicexecutor

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/hwameistor/local-storage/pkg/exechelper"
	log "github.com/sirupsen/logrus"
)

type basicExecutor struct {
	formatRegex *regexp.Regexp
}

const (
	defaultExecTimeout = 30

	exitCodeTimeout    = 124
	exitCodeErrDefault = 1
	exitCodeSuccess    = 0
)

// New creates a new basicExecutor instance, which implements
// exechelper.Executor interface
func New() exechelper.Executor {
	return &basicExecutor{}
}

func (e *basicExecutor) squashString(str string) string {
	if e.formatRegex == nil {
		e.formatRegex = regexp.MustCompile("[\t\n\r]+")
	}
	return e.formatRegex.ReplaceAllString(str, " ")
}

// RunCommand run a command, and get result
func (e *basicExecutor) RunCommand(params exechelper.ExecParams) exechelper.ExecResult {
	log.WithFields(log.Fields{"params": params}).Debug("Running command")

	// Create a new timeout context
	if params.Timeout == 0 {
		params.Timeout = defaultExecTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(params.Timeout))
	defer cancel()

	outbuf, errbuf := bytes.NewBufferString(""), bytes.NewBufferString("")
	cmd := exec.CommandContext(ctx, params.CmdName, params.CmdArgs...)
	cmd.Stdout = outbuf
	cmd.Stderr = errbuf
	err := cmd.Run()

	result := exechelper.ExecResult{
		OutBuf:   bytes.NewBufferString(strings.TrimSuffix(outbuf.String(), "\n")),
		ErrBuf:   bytes.NewBufferString(strings.TrimSuffix(errbuf.String(), "\n")),
		ExitCode: exitCodeSuccess,
		Error:    err,
	}

	if ctx.Err() == context.DeadlineExceeded {
		result.ExitCode = exitCodeTimeout
		result.Error = fmt.Errorf("Command %s %s timed out after %d seconds", params.CmdName, params.CmdArgs, params.Timeout)
		err = result.Error
	}

	if err != nil {
		// try to get the exit code
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			result.ExitCode = ws.ExitStatus()
		} else {
			// failed to get exit code, use default code
			result.ExitCode = exitCodeErrDefault
		}
		result.Error = errors.New(e.squashString(err.Error()))
	}

	log.WithFields(log.Fields{
		"command": params.CmdName,
		"args":    params.CmdArgs,
		"timeout": params.Timeout,
		"stdout":  result.OutBuf.String(),
		"stderr":  result.ErrBuf.String(),
		"error":   result.Error,
	}).Debug("Finished running command")

	return result
}
