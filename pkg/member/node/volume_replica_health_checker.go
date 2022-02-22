package node

import (
	log "github.com/sirupsen/logrus"
)

func (m *manager) startVolumeReplicaHealthChecker(stopCh <-chan struct{}) {
	m.logger.Debug("VolumeReplica health checker is working now")

	go func() {
		for {
			task, shutdown := m.healthCheckQueue.Get()
			if shutdown {
				m.logger.WithFields(log.Fields{"replica": task}).Debug("Stop the HealthChecker worker")
				break
			}
			if err := m.checkVolumeReplicaHealth(task); err != nil {
				m.logger.WithFields(log.Fields{"replica": task, "error": err.Error()}).Error("Failed to check health, retry later")
				m.healthCheckQueue.AddRateLimited(task)
			} else {
				m.logger.WithFields(log.Fields{"replica": task}).Debug("Completed a check health")
				m.healthCheckQueue.Forget(task)
			}
			m.healthCheckQueue.Done(task)
		}
	}()

	<-stopCh
	m.healthCheckQueue.Shutdown()
}

func (m *manager) checkVolumeReplicaHealth(replicaName string) error {
	m.volumeReplicaTaskQueue.Add(replicaName)
	return nil
}
