package controller

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	coorv1 "k8s.io/api/coordination/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	apis "github.com/hwameistor/hwameistor/pkg/apis/hwameistor"
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-storage/utils"
)

func (m *manager) startNodeTaskWorker(stopCh <-chan struct{}) {
	m.logger.Debug("LocalNode Worker is working now")
	go func() {
		for {
			task, shutdown := m.nodeTaskQueue.Get()
			if shutdown {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Stop the node task worker")
				break
			}
			if err := m.processNode(task); err != nil {
				m.logger.WithFields(log.Fields{"task": task, "error": err.Error()}).Error("Failed to process Node task, retry later ...")
				m.nodeTaskQueue.AddRateLimited(task)
			} else {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Completed a Node task.")
				m.nodeTaskQueue.Forget(task)
			}
			m.nodeTaskQueue.Done(task)
		}
	}()

	<-stopCh
	m.nodeTaskQueue.Shutdown()
}

func (m *manager) processNode(nodeName string) error {
	logCtx := m.logger.WithFields(log.Fields{"Node": nodeName})
	logCtx.Debug("Working on a node task")
	node := &apisv1alpha1.LocalStorageNode{}
	if err := m.apiClient.Get(context.TODO(), types.NamespacedName{Name: nodeName}, node); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("Failed to get node from cache")
			return err
		}
		logCtx.Info("Not found the node from cache, should be deleted already.")
		delete(m.localNodes, nodeName)
		return nil
	}

	// checking for lease in order to clean up the noisy Node CRD which may be created for test or by mistake
	nodeLease := &coorv1.Lease{}
	nn := types.NamespacedName{
		Namespace: m.namespace,
		Name:      utils.SanitizeName(fmt.Sprintf("%s-%s", apis.NodeLeaseNamePrefix, nodeName)),
	}
	if err := m.apiClient.Get(context.TODO(), nn, nodeLease); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithField("lease", nn).WithError(err).Error("Failed to get node lease")
			return err
		}
		// not found the node lease, so the node should be a noisy. clean it up
		logCtx.WithField("lease", nn).Info("Not found the node lease, remove it")
		return m.apiClient.Delete(context.TODO(), node)
	}

	m.localNodes[nodeName] = node.Status.State
	return nil
}
