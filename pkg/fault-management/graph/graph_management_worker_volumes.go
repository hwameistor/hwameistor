package graph

import (
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
)

func (m *manager) startLocalVolumeTaskWorker() {
	m.logger.Debug("GraphManagement LocalVolume Worker is working now")
	for {
		task, shutdown := m.localVolumeTaskQueue.Get()
		if shutdown {
			m.logger.WithFields(log.Fields{"task": task}).Debug("Stop the GraphManagement LocalVolume worker")
			break
		}
		if err := m.processLocalVolumes(task); err != nil {
			m.logger.WithFields(log.Fields{"task": task, "error": err.Error()}).Error("Failed to process GraphManagement LocalVolume task, retry later")
			m.localVolumeTaskQueue.AddRateLimited(task)
		} else {
			m.logger.WithFields(log.Fields{"task": task}).Debug("Completed a GraphManagement LocalVolume task.")
			m.localVolumeTaskQueue.Forget(task)
		}
		m.localVolumeTaskQueue.Done(task)
	}
}

func (m *manager) processLocalVolumes(localVolumeName string) error {
	logger := m.logger.WithField("localVolumeName", localVolumeName)
	logger.Debug("Processing localvolume")
	localVolume, err := m.localVolumeLister.Get(localVolumeName)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Debug("Not found local volume drop it")
			return nil
		}
		logger.WithError(err).Error("Failed to process localvolume")
		return err
	}

	// add/update localvolume as Vertex if necessary
	if err = m.Topology.AddLocalVolume(localVolume); err != nil {
		logger.WithError(err).Error("Failed to add localvolume to topology graph")
		return err
	}
	return nil
}
