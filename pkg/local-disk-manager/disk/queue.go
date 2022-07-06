package disk

import "github.com/hwameistor/hwameistor/pkg/local-disk-manager/disk/manager"

// Push
func (ctr Controller) Push(disk manager.Event) {
	ctr.diskQueue <- disk
}

// Pop
func (ctr Controller) Pop() manager.Event {
	return <-ctr.diskQueue
}
