package nsexecutor

import (
	"github.com/hwameistor/local-storage/pkg/exechelper"
	"github.com/hwameistor/local-storage/pkg/exechelper/basicexecutor"
)

type nsenterExecutor struct {
	pExecutor exechelper.Executor
}

const nsenterCommand = "nsenter"

// use pid 1 as a target, and enter their mount and pid namespace along with borrowing their root directory
// TODO: Make these configable so that we aren't *always* running everything at max permissions
var nsenterArgs = []string{"--mount=/proc/1/ns/mnt", "--ipc=/proc/1/ns/ipc", "--net=/proc/1/ns/net", "--uts=/proc/1/ns/uts", "--"}

// New creates a new nsenterExecutor instance, which implements
// exechelper.Executor interface by wrapping over top of a basic
// executor
func New() exechelper.Executor {
	return &nsenterExecutor{
		pExecutor: basicexecutor.New(),
	}
}

// RunCommand runs a command to completion, and get returns
func (e *nsenterExecutor) RunCommand(params exechelper.ExecParams) exechelper.ExecResult {
	command := append([]string{params.CmdName}, params.CmdArgs...)
	combinedArgs := append(nsenterArgs, command...)
	params.CmdName = nsenterCommand
	params.CmdArgs = combinedArgs
	return e.pExecutor.RunCommand(params)
}
