package graph

import log "github.com/sirupsen/logrus"

func (m *manager) startPVCTaskWorker() {
	m.logger.Debug("GraphManagement PVC Worker is working now")
	for {
		task, shutdown := m.pvcTaskQueue.Get()
		if shutdown {
			m.logger.WithFields(log.Fields{"task": task}).Debug("Stop the GraphManagement PVC worker")
			break
		}
		if err := m.processPVCs(task); err != nil {
			m.logger.WithFields(log.Fields{"task": task, "error": err.Error()}).Error("Failed to process GraphManagement PVC task, retry later")
			m.pvcTaskQueue.AddRateLimited(task)
		} else {
			m.logger.WithFields(log.Fields{"task": task}).Debug("Completed a GraphManagement PVC task.")
			m.pvcTaskQueue.Forget(task)
		}
		m.pvcTaskQueue.Done(task)
	}
}

func (m *manager) processPVCs(pvcNamespaceName string) error {
	return nil
}
