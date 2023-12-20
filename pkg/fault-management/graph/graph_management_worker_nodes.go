package graph

import (
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
)

func (m *manager) startStorageNodeTaskWorker() {
	m.logger.Debug("GraphManagement LocalStorageNode Worker is working now")
	for {
		task, shutdown := m.storageNodeTaskQueue.Get()
		if shutdown {
			m.logger.WithFields(log.Fields{"task": task}).Debug("Stop the GraphManagement LocalStorageNode worker")
			break
		}
		if err := m.processStorageNodes(task); err != nil {
			m.logger.WithFields(log.Fields{"task": task, "error": err.Error()}).Error("Failed to process GraphManagement LocalStorageNode task, retry later")
			m.storageNodeTaskQueue.AddRateLimited(task)
		} else {
			m.logger.WithFields(log.Fields{"task": task}).Debug("Completed a GraphManagement LocalStorageNode task.")
			m.storageNodeTaskQueue.Forget(task)
		}
		m.storageNodeTaskQueue.Done(task)
	}
}

func (m *manager) processStorageNodes(storageNodeName string) error {
	logger := m.logger.WithField("storageNodeName", storageNodeName)
	logger.Debug("Processing storage nodes")
	storageNode, err := m.storageNodeLister.Get(storageNodeName)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Debug("Not found storage node drop it")
			return nil
		}
		logger.WithError(err).Error("Failed to process storage node")
		return err
	}

	// add/update storage node as Vertex if necessary
	if err = m.Topology.AddStorageNode(storageNode); err != nil {
		logger.WithError(err).Error("Failed to add storage node to topology graph")
		return err
	}

	// add/update storage pool under this node as Vertex if necessary
	if err = m.Topology.AddStoragePool(storageNode); err != nil {
		logger.WithError(err).Error("Failed to add storage pool to topology graph")
		return err
	}

	return nil
}
