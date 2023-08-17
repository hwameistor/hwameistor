package controller

import (
	"context"
	"fmt"
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func (m *manager) startVolumeSnapshotRestoreTaskWorker(stopCh <-chan struct{}) {
	m.logger.Debug("Volume Snapshot Restore Worker is working now")
	go func() {
		for {
			task, shutdown := m.volumeSnapshotRestoreTaskQueue.Get()
			if shutdown {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Stop the Volume Snapshot Restore worker")
				break
			}
			if err := m.processVolumeSnapshot(task); err != nil {
				m.logger.WithFields(log.Fields{"task": task, "attempts": m.volumeSnapshotRestoreTaskQueue.NumRequeues(task), "error": err.Error()}).Error("Failed to process Volume Snapshot Restore task, retry later")
				m.volumeSnapshotRestoreTaskQueue.AddRateLimited(task)
			} else {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Completed a Volume Snapshot Restore task.")
				m.volumeSnapshotRestoreTaskQueue.Forget(task)
			}
			m.volumeSnapshotRestoreTaskQueue.Done(task)
		}
	}()

	<-stopCh
	m.volumeSnapshotRestoreTaskQueue.Shutdown()
}

func (m *manager) processVolumeSnapshotRestore(restoreName string) error {
	logCtx := m.logger.WithFields(log.Fields{"VolumeSnapshotRestore": restoreName})
	logCtx.Debug("Working on a VolumeSnapshotRestore task")
	snapshotRestore := &apisv1alpha1.LocalVolumeSnapshotRestore{}
	if err := m.apiClient.Get(context.TODO(), types.NamespacedName{Name: restoreName}, snapshotRestore); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("Failed to get VolumeSnapshotRestore from cache")
			return err
		}
		logCtx.Info("Not found the VolumeSnapshotRestore from cache, should be deleted already")
		return nil
	}

	if snapshotRestore.Spec.Abort &&
		snapshotRestore.Status.State != apisv1alpha1.OperationStateToBeAborted &&
		snapshotRestore.Status.State != apisv1alpha1.OperationStateAborting &&
		snapshotRestore.Status.State != apisv1alpha1.OperationStateAborted &&
		snapshotRestore.Status.State != apisv1alpha1.OperationStateCompleted {

		snapshotRestore.Status.State = apisv1alpha1.OperationStateToBeAborted
		return m.apiClient.Status().Update(context.TODO(), snapshotRestore)
	}

	// log with namespace/name is enough
	logCtx = m.logger.WithFields(log.Fields{"TargetVolume": snapshotRestore.Spec.TargetVolume, "SnapshotRestore": snapshotRestore.Name, "Spec": snapshotRestore.Spec, "Status": snapshotRestore.Status})
	logCtx.Debug("Starting to process a VolumeSnapshotRestore")

	// state chain: (empty) -> Submitted -> Start -> InProgress -> Completed
	switch snapshotRestore.Status.State {
	case "":
		return m.volumeSnapshotRestoreSubmit(snapshotRestore)
	case apisv1alpha1.OperationStateSubmitted:
		return m.volumeSnapshotRestoreStart(snapshotRestore)
	case apisv1alpha1.OperationStateInProgress:
		return m.checkInProgressVolumeSnapshotRestore(snapshotRestore)
	case apisv1alpha1.OperationStateToBeAborted:
		return m.volumeSnapshotRestoreAbort(snapshotRestore)
	case apisv1alpha1.OperationStateCompleted, apisv1alpha1.OperationStateAborted:
		return m.volumeSnapshotRestoreCleanup(snapshotRestore)
	default:
		logCtx.Error("Invalid state/phase")
	}
	return fmt.Errorf("invalid state")
}

func (m *manager) volumeSnapshotRestoreSubmit(snapshotRestore *apisv1alpha1.LocalVolumeSnapshotRestore) error {
	logCtx := m.logger.WithFields(log.Fields{"SnapshotRestore": snapshotRestore.Name, "Spec": snapshotRestore.Spec})
	logCtx.Debug("Submit a VolumeSnapshotRestore")
	return nil
}

func (m *manager) volumeSnapshotRestoreStart(snapshotRestore *apisv1alpha1.LocalVolumeSnapshotRestore) error {
	logCtx := m.logger.WithFields(log.Fields{"SnapshotRestore": snapshotRestore.Name, "Spec": snapshotRestore.Spec})
	logCtx.Debug("Start a VolumeSnapshotRestore")
	return nil
}

func (m *manager) checkInProgressVolumeSnapshotRestore(snapshotRestore *apisv1alpha1.LocalVolumeSnapshotRestore) error {
	logCtx := m.logger.WithFields(log.Fields{"SnapshotRestore": snapshotRestore.Name, "Spec": snapshotRestore.Spec})
	logCtx.Debug("Check a InProgress VolumeSnapshotRestore")
	return nil
}

func (m *manager) volumeSnapshotRestoreAbort(snapshotRestore *apisv1alpha1.LocalVolumeSnapshotRestore) error {
	logCtx := m.logger.WithFields(log.Fields{"SnapshotRestore": snapshotRestore.Name, "Spec": snapshotRestore.Spec})
	logCtx.Debug("Abort a VolumeSnapshotRestore")

	snapshotRestore.Status.State = apisv1alpha1.OperationStateAborted
	return m.apiClient.Status().Update(context.TODO(), snapshotRestore)
}

func (m *manager) volumeSnapshotRestoreCleanup(snapshotRestore *apisv1alpha1.LocalVolumeSnapshotRestore) error {
	logCtx := m.logger.WithFields(log.Fields{"SnapshotRestore": snapshotRestore.Name, "Spec": snapshotRestore.Spec})
	logCtx.Debug("Cleanup a VolumeSnapshotRestore")

	return m.apiClient.Delete(context.TODO(), snapshotRestore)
}
