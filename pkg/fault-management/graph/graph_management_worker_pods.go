package graph

import log "github.com/sirupsen/logrus"

func (m *manager) startPodTaskWorker() {
	m.logger.Debug("GraphManagement Pod Worker is working now")
	for {
		task, shutdown := m.podTaskQueue.Get()
		if shutdown {
			m.logger.WithFields(log.Fields{"task": task}).Debug("Stop the GraphManagement Pod worker")
			break
		}
		if err := m.processPods(task); err != nil {
			m.logger.WithFields(log.Fields{"task": task, "error": err.Error()}).Error("Failed to process GraphManagement Pod task, retry later")
			m.podTaskQueue.AddRateLimited(task)
		} else {
			m.logger.WithFields(log.Fields{"task": task}).Debug("Completed a GraphManagement Pod task.")
			m.podTaskQueue.Forget(task)
		}
		m.podTaskQueue.Done(task)
	}
}

func (m *manager) processPods(podNamespaceName string) error {
	return nil
}
