package node

import (
	"context"
	"fmt"
	"strings"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	diskmonitor "github.com/hwameistor/hwameistor/pkg/local-storage/member/node/diskmonitor"
	log "github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func (m *manager) startLocalDiskTaskWorker(stopCh <-chan struct{}) {

	m.logger.Debug("localDisk Worker is working now")
	go func() {
		for {
			task, shutdown := m.localDiskTaskQueue.Get()
			if shutdown {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Stop the localDisk worker")
				break
			}
			if err := m.processLocalDisk(task); err != nil {
				m.logger.WithFields(log.Fields{"task": task, "error": err.Error()}).Error("Failed to process localDisk task, retry later")
				m.localDiskTaskQueue.AddRateLimited(task)
			} else {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Completed a localDisk task.")
				m.localDiskTaskQueue.Forget(task)
			}
			m.localDiskTaskQueue.Done(task)
		}
	}()

	<-stopCh
	m.volumeReplicaTaskQueue.Shutdown()
}

func (m *manager) processLocalDisk(localDiskNameSpacedName string) error {
	m.logger.Debug("processLocalDisk start ...")
	logCtx := m.logger.WithFields(log.Fields{"localDisk": localDiskNameSpacedName})
	logCtx.Debug("Working on a localDisk task")
	splitRes := strings.Split(localDiskNameSpacedName, "/")
	var diskName string
	if len(splitRes) >= 2 {
		// nameSpace = splitRes[0]
		diskName = splitRes[1]
	}

	localDisk := &apisv1alpha1.LocalDisk{}
	if err := m.apiClient.Get(context.TODO(), types.NamespacedName{Name: diskName}, localDisk); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("Failed to get localDisk from cache, retry it later ...")
			return err
		}
		logCtx.Info("Not found the localDisk from cache, should be deleted already.")
		return nil
	}

	m.logger.Debugf("Required node name %s, current node name %s.", localDisk.Spec.NodeName, m.name)
	if localDisk.Spec.NodeName != m.name {
		return nil
	}

	switch localDisk.Spec.State {
	case apisv1alpha1.LocalDiskInactive:
		logCtx.Debug("LocalDiskInactive, todo ...")
		// 构建离线的event
		event := &diskmonitor.DiskEvent{}
		m.diskEventQueue.Add(event)
		return nil

	case apisv1alpha1.LocalDiskActive:
		logCtx.Debug("LocalDiskActive ...")
		return nil

	case apisv1alpha1.LocalDiskUnknown:
		logCtx.Debug("LocalDiskUnknown ...")
		return nil

	default:
		logCtx.Error("Invalid localDisk state")
	}

	switch localDisk.Status.State {
	case apisv1alpha1.LocalDiskAvailable:
		logCtx.Debug("LocalDiskAvailable ...")
		return nil

	case apisv1alpha1.LocalDiskBound:
		logCtx.Debug("LocalDiskBound ...")
		return nil

	default:
		logCtx.Error("Invalid localDisk state")
	}

	return fmt.Errorf("invalid localDisk state")
}
