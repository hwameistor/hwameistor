package diskmonitor

import (
	"time"

	"k8s.io/client-go/util/workqueue"
)

// EventQueue is rate limiting queue for event
type EventQueue struct {
	queue workqueue.RateLimitingInterface
}

// NewEventQueue creates a queue
func NewEventQueue(queueName string) *EventQueue {
	return &EventQueue{
		queue: workqueue.NewNamedRateLimitingQueue(
			workqueue.NewItemExponentialFailureRateLimiter(time.Second, 16*time.Second),
			queueName,
		),
	}
}

// Add a event into the queue
func (q *EventQueue) Add(event *DiskEvent) {
	q.queue.Add(event)
}

// AddRateLimited a event
func (q *EventQueue) AddRateLimited(event *DiskEvent) {
	q.queue.AddRateLimited(event)
}

// Get a event from queue. It's a blocking call
func (q *EventQueue) Get() (*DiskEvent, bool) {
	item, shutdown := q.queue.Get()
	if item == nil {
		return nil, true
	}
	return item.(*DiskEvent), shutdown
}

// Done completes the event and remove from the queue
func (q *EventQueue) Done(event *DiskEvent) {
	q.queue.Done(event)
}

// NumRequeues get number of event retried
func (q *EventQueue) NumRequeues(event *DiskEvent) int {
	return q.queue.NumRequeues(event)
}

// Forget cleanup the rate limit on the event
func (q *EventQueue) Forget(event *DiskEvent) {
	q.queue.Forget(event)
}

// Shutdown the queue
func (q *EventQueue) Shutdown() {
	q.queue.ShutDown()
}
