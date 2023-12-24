package node

import (
	"github.com/hwameistor/hwameistor/pkg/exechelper"
	"github.com/hwameistor/hwameistor/pkg/exechelper/nsexecutor"
)

type NodeExecutor struct {
	executor exechelper.Executor
}

func NewNodeNSExecutor() *NodeExecutor {
	return &NodeExecutor{
		executor: nsexecutor.New(),
	}
}