package graph

import log "github.com/sirupsen/logrus"

func (m *manager) startPVTaskWorker() {
	m.logger.Debug("GraphManagement PV Worker is working now")
	for {
		task, shutdown := m.pvTaskQueue.Get()
		if shutdown {
			m.logger.WithFields(log.Fields{"task": task}).Debug("Stop the GraphManagement PV worker")
			break
		}
		if err := m.processPVs(task); err != nil {
			m.logger.WithFields(log.Fields{"task": task, "error": err.Error()}).Error("Failed to process GraphManagement PV task, retry later")
			m.pvTaskQueue.AddRateLimited(task)
		} else {
			m.logger.WithFields(log.Fields{"task": task}).Debug("Completed a GraphManagement PV task.")
			m.pvTaskQueue.Forget(task)
		}
		m.pvTaskQueue.Done(task)
	}
}

func (m *manager) processPVs(pvName string) error {
	return nil
}
