package graph

func (m *manager) startGraphManagementTaskWorker(stopCh <-chan struct{}) {
	m.logger.Debug("GraphManagement Worker is working now")

	go m.startPodTaskWorker()
	go m.startPVCTaskWorker()
	go m.startPVTaskWorker()
	go m.startLocalVolumeTaskWorker()
	go m.startStorageNodeTaskWorker()

	<-stopCh
	// notify all workers done
	m.podTaskQueue.Shutdown()
	m.pvcTaskQueue.Shutdown()
	m.pvTaskQueue.Shutdown()
	m.localVolumeTaskQueue.Shutdown()
	m.storageNodeTaskQueue.Shutdown()
}
