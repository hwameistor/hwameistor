package disk

import (
	"encoding/json"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/disk/manager"
)

// Push pushes item to the work queue
func (ctr *Controller) Push(disk manager.Event) {
	bytes, _ := json.Marshal(disk)
	ctr.diskQueue.Add(string(bytes))
}

// PushRateLimited pushes an item to the work queue after the rate limiter says it's ok
func (ctr *Controller) PushRateLimited(disk manager.Event) {
	bytes, _ := json.Marshal(disk)
	ctr.diskQueue.AddRateLimited(string(bytes))
}

// Pop pops an item from the work queue, it will block until it returns an item
func (ctr *Controller) Pop() manager.Event {
	eventString, _ := ctr.diskQueue.Get()
	event := manager.Event{}
	_ = json.Unmarshal([]byte(eventString), &event)
	return event
}

// Done completes the task and remove from the queue
func (ctr *Controller) Done(disk manager.Event) {
	bytes, _ := json.Marshal(disk)
	ctr.diskQueue.Done(string(bytes))
}

// Forget completes the task and remove from the queue
func (ctr *Controller) Forget(disk manager.Event) {
	bytes, _ := json.Marshal(disk)
	ctr.diskQueue.Forget(string(bytes))
}
