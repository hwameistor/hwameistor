package node

import (
	"context"
	log "github.com/sirupsen/logrus"
)

// startDiskTaskWorker starts a worker for handling  LocalDisk objects
func (m *nodeManager) startDiskTaskWorker(ctx context.Context) {
	m.logger.Info("Start LocalDisk worker now")
	go func() {
		for {
			task, shutdown := m.diskTaskQueue.Get()
			if shutdown {
				m.logger.Info("Stop the LocalDisk worker")
				break
			}
			if err := m.processLocalDisk(task); err != nil {
				m.logger.WithFields(log.Fields{"task": task, "error": err.Error()}).Error("Failed to process LocalDiskClaim task, retry later")
				m.diskTaskQueue.AddRateLimited(task)
			} else {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Completed a LocalDisk task")
				m.diskClaimTaskQueue.Forget(task)
			}
			m.diskClaimTaskQueue.Done(task)
		}
	}()

	// We are done, Stop Node Manager
	<-ctx.Done()
	m.diskTaskQueue.Shutdown()
}

func (m *nodeManager) processLocalDisk(disk string) error {
	return nil
}
