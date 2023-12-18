package graph

import log "github.com/sirupsen/logrus"

func (m *manager) startGraphManagementTaskWorker(stopCh <-chan struct{}) {
	m.logger.Debug("GraphManagement Worker is working now")

	// worker for processing Pods
	go func() {
		for {
			task, shutdown := m.podTaskQueue.Get()
			if shutdown {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Stop the GraphManagement worker")
				break
			}
			if err := m.processPods(task); err != nil {
				m.logger.WithFields(log.Fields{"task": task, "error": err.Error()}).Error("Failed to process GraphManagement task, retry later")
				m.podTaskQueue.AddRateLimited(task)
			} else {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Completed a GraphManagement task.")
				m.podTaskQueue.Forget(task)
			}
			m.podTaskQueue.Done(task)
		}
	}()

	// worker for processing PersistentVolumeClaims
	go func() {
		for {
			task, shutdown := m.pvcTaskQueue.Get()
			if shutdown {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Stop the GraphManagement worker")
				break
			}
			if err := m.processPVCs(task); err != nil {
				m.logger.WithFields(log.Fields{"task": task, "error": err.Error()}).Error("Failed to process GraphManagement task, retry later")
				m.pvcTaskQueue.AddRateLimited(task)
			} else {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Completed a GraphManagement task.")
				m.pvcTaskQueue.Forget(task)
			}
			m.pvcTaskQueue.Done(task)
		}
	}()

	// worker for processing PersistentVolumes
	go func() {
		for {
			task, shutdown := m.pvTaskQueue.Get()
			if shutdown {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Stop the GraphManagement worker")
				break
			}
			if err := m.processPVs(task); err != nil {
				m.logger.WithFields(log.Fields{"task": task, "error": err.Error()}).Error("Failed to process GraphManagement task, retry later")
				m.pvTaskQueue.AddRateLimited(task)
			} else {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Completed a GraphManagement task.")
				m.pvTaskQueue.Forget(task)
			}
			m.pvTaskQueue.Done(task)
		}
	}()

	// worker for processing LocalVolumes
	go func() {
		for {
			task, shutdown := m.localVolumeTaskQueue.Get()
			if shutdown {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Stop the GraphManagement worker")
				break
			}
			if err := m.processLocalVolumes(task); err != nil {
				m.logger.WithFields(log.Fields{"task": task, "error": err.Error()}).Error("Failed to process GraphManagement task, retry later")
				m.localVolumeTaskQueue.AddRateLimited(task)
			} else {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Completed a GraphManagement task.")
				m.localVolumeTaskQueue.Forget(task)
			}
			m.localVolumeTaskQueue.Done(task)
		}
	}()

	// worker for processing LocalStorageNodes
	go func() {
		for {
			task, shutdown := m.storageNodeTaskQueue.Get()
			if shutdown {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Stop the GraphManagement worker")
				break
			}
			if err := m.processStorageNodes(task); err != nil {
				m.logger.WithFields(log.Fields{"task": task, "error": err.Error()}).Error("Failed to process GraphManagement task, retry later")
				m.storageNodeTaskQueue.AddRateLimited(task)
			} else {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Completed a GraphManagement task.")
				m.storageNodeTaskQueue.Forget(task)
			}
			m.storageNodeTaskQueue.Done(task)
		}
	}()

	<-stopCh
	m.podTaskQueue.Shutdown()
	m.pvcTaskQueue.Shutdown()
	m.pvTaskQueue.Shutdown()
	m.localVolumeTaskQueue.Shutdown()
	m.storageNodeTaskQueue.Shutdown()
}

func (m *manager) processPods(podNamespaceName string) error {
	return nil
}

func (m *manager) processPVCs(pvcNamespaceName string) error {
	return nil
}

func (m *manager) processPVs(pvName string) error {
	return nil
}

func (m *manager) processLocalVolumes(localVolumeName string) error {
	return nil
}

func (m *manager) processStorageNodes(storageNodeName string) error {
	return nil
}
