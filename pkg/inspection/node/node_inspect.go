package node

import (
	// "sort"
	"strings"

	"github.com/hwameistor/hwameistor/pkg/exechelper"
	log "github.com/sirupsen/logrus"
	// v1 "k8s.io/api/core/v1"
)

type NodeInspector struct {
	nodeExecutor *NodeExecutor
}

func NewNodeInspector() *NodeInspector {
	return &NodeInspector{
		nodeExecutor: NewNodeNSExecutor(),
	}
}

func (i *NodeInspector) Inspect() {
	i.kernelReleaseInfo()
	i.drbdVersion()
	i.kubeletArgs()
}

func (i *NodeInspector) kernelReleaseInfo() {
	params := exechelper.ExecParams{
		CmdName: "uname",
		CmdArgs: []string{"-r"},
	}

	res := i.nodeExecutor.executor.RunCommand(params)
	if res.ExitCode != 0 {
		// return nil, res.Error
		log.WithError(res.Error).Error("err occured when run uname -r")
		return
	}

	log.Infof("result: %v", res.OutBuf.String())
	return
}

// cat /proc/drbd ?
func (i *NodeInspector) drbdVersion() {
	params := exechelper.ExecParams{
		CmdName: "drbdadm",
		CmdArgs: []string{"--version"},
	}

	res := i.nodeExecutor.executor.RunCommand(params)
	if res.ExitCode != 0 {
		// return nil, res.Error
		log.WithError(res.Error).Error("err occured when run drbdadm --version")
		return
	}

	log.Infof("result: %v", res.OutBuf.String())
	return
}

// pidof kubelet | xargs ps -o cmd --no-headers | cat
func (i *NodeInspector) kubeletArgs() {
	// params := exechelper.ExecParams{
	// 	CmdName: "pidof kubelet | xargs ps -o cmd --no-headers | cat",
	// }

	kubeletPID := i.pidof("kubelet")
	params := exechelper.ExecParams{
		// CmdName: "pidof",
		// CmdArgs: []string{"kubelet", "|", "xargs", "ps", "-o", "cmd", "--no-headers", "|", "cat"},
		CmdName: "ps",
		CmdArgs: []string{"-o", "cmd", "--no-headers", kubeletPID},
	}

	res := i.nodeExecutor.executor.RunCommand(params)
	if res.ExitCode != 0 {
		// return nil, res.Error
		log.WithError(res.Error).Error("err occured when run ps")
		return
	}

	log.Infof("result: %v", res.OutBuf.String())
	return
}

func (i *NodeInspector) pidof(programName string) string {
	params := exechelper.ExecParams{
		CmdName: "pidof",
		CmdArgs: []string{programName},
	}

	res := i.nodeExecutor.executor.RunCommand(params)
	if res.ExitCode != 0 {
		// return nil, res.Error
		log.WithError(res.Error).Errorf("err occured when run pidof %v", programName)
		return ""
	}

	outString := res.OutBuf.String()
	log.Infof("result: %v", outString)
	return outString
}

func (i NodeInspector) exec(cmd string) {
	tokens := strings.Split(cmd, " ")
	args := []string{}
	if len(tokens) > 1 {
		args = tokens[1:]
	}

	params := exechelper.ExecParams{
		CmdName: tokens[0],
		CmdArgs: args,
	}

	log.Infof("to run %v %v", params.CmdName, params.CmdArgs)

	res := i.nodeExecutor.executor.RunCommand(params)
	if res.ExitCode != 0 {
		// return nil, res.Error
		log.WithError(res.Error).Errorf("err occured when run %v %v", params.CmdName, params.CmdArgs)
		// return
	}

	log.Infof("result: %v", res.OutBuf.String())
	// return
}

// func (i *NodeInspector) Inspect(cm *v1.ConfigMap) {
// 	// i.kernelReleaseInfo()

// 	keys := make([]string, 0, len(cm.Data))
// 	for k := range cm.Data {
// 		keys = append(keys, k)
// 	}
	
// 	sort.Strings(keys)
// 	log.Info("keys sorted:")
// 	for _, str := range keys {
// 		log.Info(str)
// 	}
// 	// for k, v := range cm.Data {
// 	// 	log.Infof("to inspect %v", k)
// 	// 	i.exec(v)
// 	// }
// 	for _, mapKey := range keys {
// 		log.Infof("to inspect %v", mapKey)
// 		i.exec(cm.Data[mapKey])
// 	}
// }