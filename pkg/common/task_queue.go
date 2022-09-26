package common

import (
	"time"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/util/workqueue"
)

// TaskQueue is rate limiting queue for task
type TaskQueue struct {
	queue      workqueue.RateLimitingInterface
	logger     *log.Entry
	maxRetries int
}

// NewTaskQueue creates a queue
// maxRetries limit retry number, no limit if maxRetries=0
// 1s, 2s, 4s, 8s, 16s, 32s, 60s, 60s, 60s, ...
func NewTaskQueue(taskName string, maxRetries int) *TaskQueue {
	return &TaskQueue{
		queue: workqueue.NewNamedRateLimitingQueue(
			workqueue.NewItemExponentialFailureRateLimiter(time.Second, time.Minute),
			taskName,
		),
		maxRetries: maxRetries,
		logger:     log.WithField("TaskQueue", taskName),
	}
}

// Add a task into the queue
func (q *TaskQueue) Add(task string) {
	q.queue.Add(task)
}

// AddRateLimited a task
func (q *TaskQueue) AddRateLimited(task string) {
	if q.maxRetries > 0 && q.NumRequeues(task) > q.maxRetries {
		q.logger.WithField("Task", task).Infof("exceeds maxRetries(%d), drop it", q.maxRetries)
		return
	}
	q.queue.AddRateLimited(task)
}

// Get a task from queue. It's a blocking call
func (q *TaskQueue) Get() (string, bool) {
	item, shutdown := q.queue.Get()
	if item == nil {
		return "", true
	}
	return item.(string), shutdown
}

// Done completes the task and remove from the queue
func (q *TaskQueue) Done(task string) {
	q.queue.Done(task)
}

// NumRequeues get number of task retried
func (q *TaskQueue) NumRequeues(task string) int {
	return q.queue.NumRequeues(task)
}

// Forget cleanup the rate limit on the task
func (q *TaskQueue) Forget(task string) {
	q.queue.Forget(task)
}

// Shutdown the queue
func (q *TaskQueue) Shutdown() {
	q.queue.ShutDown()
}
