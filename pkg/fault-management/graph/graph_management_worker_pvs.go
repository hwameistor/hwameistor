package graph

import (
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
)

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
	logger := m.logger.WithField("pvName", pvName)
	logger.Debug("Processing pv")

	pv, err := m.pvLister.Get(pvName)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Debug("Not found pv, may be deleted from the cache already")
			return nil
		}
		return err
	}

	sc, err := m.fetchSC(pv.Spec.StorageClassName)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Debug("Can not determine provisioned by hwameistor whether or not because of not found sc drop it")
			return nil
		}
		return err
	}

	if !isHwameiStorVolume(sc.Provisioner) {
		logger.WithFields(log.Fields{"provisioner": sc.Provisioner, "volume": pvName}).Debug("Not hwameistor volume, drop it")
		return nil
	}

	// add/update pv as Vertex if necessary
	if err = m.Topology.AddPV(pv); err != nil {
		logger.WithError(err).Error("Failed to add pv to topology graph")
		return err
	}
	return nil
}