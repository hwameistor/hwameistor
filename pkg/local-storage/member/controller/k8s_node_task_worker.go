package controller

import (
	"context"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

func (m *manager) startK8sNodeTaskWorker(stopCh <-chan struct{}) {
	m.logger.Debug("K8s Node Worker is working now")
	go func() {
		for {
			task, shutdown := m.k8sNodeTaskQueue.Get()
			if shutdown {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Stop the k8s node task worker")
				break
			}
			if err := m.processK8sNode(task); err != nil {
				m.logger.WithFields(log.Fields{"task": task, "error": err.Error()}).Error("Failed to process K8s Node task, retry later ...")
				m.k8sNodeTaskQueue.AddRateLimited(task)
			} else {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Completed a K8s Node task.")
				m.k8sNodeTaskQueue.Forget(task)
			}
			m.k8sNodeTaskQueue.Done(task)
		}
	}()

	<-stopCh
	m.k8sNodeTaskQueue.Shutdown()
}

func (m *manager) processK8sNode(nodeName string) error {
	logCtx := m.logger.WithFields(log.Fields{"K8sNode": nodeName})
	logCtx.Debug("Working on a k8s node task")
	node := &corev1.Node{}
	if err := m.apiClient.Get(context.TODO(), types.NamespacedName{Name: nodeName}, node); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("Failed to get k8s node from cache")
			return err
		}
		logCtx.Info("Not found the k8s node from cache, should be deleted already.")
		return nil
	}

	for _, cond := range node.Status.Conditions {
		if cond.Type == corev1.NodeReady && cond.Status == corev1.ConditionUnknown {
			return m.degradeVolumeReplicasForNode(nodeName)
		}
	}

	return nil
}

func (m *manager) degradeVolumeReplicasForNode(nodeName string) error {
	m.logger.WithField("node", nodeName).Debug("Start to degrade volume replicas")
	replicaList := &apisv1alpha1.LocalVolumeReplicaList{}
	if err := m.apiClient.List(context.TODO(), replicaList); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	for _, replica := range replicaList.Items {
		if replica.Spec.NodeName != nodeName || replica.Status.State != apisv1alpha1.VolumeReplicaStateReady {
			continue
		}
		m.logger.WithFields(log.Fields{"replica": replica.Name, "node": nodeName}).Debug("Degrade VolumeReplica because of unknown K8s node status")
		replica.Status.State = apisv1alpha1.VolumeReplicaStateNotReady
		if err := m.apiClient.Status().Update(context.TODO(), &replica); err != nil {
			m.logger.WithFields(log.Fields{"replica": replica.Name, "node": nodeName}).WithError(err).Error("Failed to degrade VolumeReplica")
			return err
		}
	}
	return nil
}
