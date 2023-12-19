package graph

import (
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/types"
	"strings"
)

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
	logger := m.logger.WithField("pvcNamespaceName", pvcNamespaceName)
	logger.Debug("Processing pvc")

	namespace := strings.Split(pvcNamespaceName, "/")[0]
	name := strings.Split(pvcNamespaceName, "/")[1]
	pvc, err := m.pvcLister.PersistentVolumeClaims(namespace).Get(name)
	if err != nil {
		logger.WithError(err).Error("Failed to process pvc")
		return err
	}

	sc, err := m.fetchSC(*pvc.Spec.StorageClassName)
	if err != nil {
		return err
	}

	if !isHwameiStorVolume(sc.Provisioner) {
		logger.WithFields(log.Fields{"provisioner": sc.Provisioner, "pvcNamespacedName": types.NamespacedName{
			Namespace: pvc.Namespace,
			Name:      pvc.Name,
		}.String()}).Debug("not hwameistor volume, drop it")
		return nil
	}

	// add/update pvc as Vertex if necessary
	if err = m.Topology.AddPVC(pvc); err != nil {
		logger.WithError(err).Error("Failed to add pvc to topology graph")
		return err
	}
	return nil
}
