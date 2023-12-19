package graph

import (
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/types"
	"strings"
)

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
	logger := m.logger.WithField("podNamespaceName", podNamespaceName)
	logger.Debug("Processing pod")

	namespace := strings.Split(podNamespaceName, "/")[0]
	name := strings.Split(podNamespaceName, "/")[1]
	pod, err := m.podLister.Pods(namespace).Get(name)
	if err != nil {
		logger.WithError(err).Error("Failed to process pod")
		return err
	}

	// find out which volume(s) that provisioned by hwameistor
	var hwameistorVolumes []string
	for _, volume := range pod.Spec.Volumes {
		if volume.PersistentVolumeClaim == nil {
			continue
		}

		var (
			sc  *storagev1.StorageClass
			pvc *v1.PersistentVolumeClaim
		)

		if pvc, err = m.fetchPVC(pod.Namespace, volume.PersistentVolumeClaim.ClaimName); err != nil {
			return err
		}
		if sc, err = m.fetchSC(*pvc.Spec.StorageClassName); err != nil {
			return err
		}

		if isHwameiStorVolume(sc.Provisioner) {
			hwameistorVolumes = append(hwameistorVolumes, types.NamespacedName{
				Namespace: pod.Namespace,
				Name:      volume.PersistentVolumeClaim.ClaimName,
			}.String())
		}
	}

	// add/update pod as Vertex if necessary
	if err = m.Topology.AddPod(pod, hwameistorVolumes...); err != nil {
		logger.WithError(err).Error("Failed to add pod to topology graph")
		return err
	}
	return nil
}

func (m *manager) fetchPVC(ns, name string) (*v1.PersistentVolumeClaim, error) {
	if pvc, err := m.pvcLister.PersistentVolumeClaims(ns).Get(name); err != nil {
		return nil, err
	} else {
		return pvc, nil
	}
}

func (m *manager) fetchSC(name string) (*storagev1.StorageClass, error) {
	if sc, err := m.scLister.Get(name); err != nil {
		return nil, err
	} else {
		return sc, nil
	}
}
