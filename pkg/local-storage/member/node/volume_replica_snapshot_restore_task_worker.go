package node

import (
	"context"
	"fmt"
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (m *manager) startVolumeReplicaSnapshotRestoreTaskWorker(stopCh <-chan struct{}) {
	m.logger.Debug("Volume Replica Snapshot Restore Worker is working now")
	go func() {
		for {
			task, shutdown := m.volumeReplicaSnapshotRestoreTaskQueue.Get()
			if shutdown {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Stop the Volume Replica Snapshot Restore worker")
				break
			}
			if err := m.processVolumeReplicaSnapshotRestore(task); err != nil {
				m.logger.WithFields(log.Fields{"task": task, "attempts": m.volumeReplicaSnapshotRestoreTaskQueue.NumRequeues(task), "error": err.Error()}).Error("Failed to process Volume Replica Snapshot Restore task, retry later")
				m.volumeReplicaSnapshotRestoreTaskQueue.AddRateLimited(task)
			} else {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Completed a Volume Replica Snapshot Restore task.")
				m.volumeReplicaSnapshotRestoreTaskQueue.Forget(task)
			}
			m.volumeReplicaSnapshotRestoreTaskQueue.Done(task)
		}
	}()

	<-stopCh
	m.volumeReplicaSnapshotRestoreTaskQueue.Shutdown()
}

func (m *manager) processVolumeReplicaSnapshotRestore(restoreName string) error {
	logCtx := m.logger.WithFields(log.Fields{"VolumeReplicaSnapshotRestore": restoreName})
	logCtx.Debug("Working on a VolumeReplicaSnapshotRestore task")
	snapshotRestore := &apisv1alpha1.LocalVolumeReplicaSnapshotRestore{}
	if err := m.apiClient.Get(context.TODO(), types.NamespacedName{Name: restoreName}, snapshotRestore); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("Failed to get VolumeReplicaSnapshotRestore from cache")
			return err
		}
		logCtx.Info("Not found the VolumeReplicaSnapshotRestore from cache, should be deleted already")
		return nil
	}

	if snapshotRestore.Spec.Abort &&
		snapshotRestore.Status.State != apisv1alpha1.OperationStateToBeAborted &&
		snapshotRestore.Status.State != apisv1alpha1.OperationStateAborting &&
		snapshotRestore.Status.State != apisv1alpha1.OperationStateAborted {

		snapshotRestore.Status.State = apisv1alpha1.OperationStateToBeAborted
		return m.apiClient.Status().Update(context.TODO(), snapshotRestore)
	}

	// log with namespace/name is enough
	logCtx = m.logger.WithFields(log.Fields{"TargetVolume": snapshotRestore.Spec.TargetVolume, "SnapshotRestore": snapshotRestore.Name, "Spec": snapshotRestore.Spec, "Status": snapshotRestore.Status})
	logCtx.Debug("Starting to process a VolumeReplicaSnapshotRestore")

	// state chain: (empty) -> Submitted -> Start -> InProgress -> Completed
	switch snapshotRestore.Status.State {
	case "":
		return m.volumeReplicaSnapshotRestoreSubmit(snapshotRestore)
	case apisv1alpha1.OperationStateSubmitted:
		return m.volumeReplicaSnapshotRestorePreCheck(snapshotRestore)
	case apisv1alpha1.OperationStateInProgress:
		return m.restoreVolumeFromSnapshot(snapshotRestore)
	case apisv1alpha1.OperationStateToBeAborted:
		return m.volumeReplicaSnapshotRestoreAbort(snapshotRestore)
	case apisv1alpha1.OperationStateCompleted, apisv1alpha1.OperationStateAborted:
		return m.volumeReplicaSnapshotRestoreCleanup(snapshotRestore)
	default:
		logCtx.Error("Invalid state/phase")
	}
	return fmt.Errorf("invalid state")
}

func (m *manager) volumeReplicaSnapshotRestoreSubmit(snapshotRestore *apisv1alpha1.LocalVolumeReplicaSnapshotRestore) error {
	logCtx := m.logger.WithFields(log.Fields{"SnapshotRestore": snapshotRestore.Name, "Spec": snapshotRestore.Spec})
	logCtx.Debug("Submit a VolumeReplicaSnapshotRestore")

	snapshotRestore.Status.State = apisv1alpha1.OperationStateSubmitted
	return m.apiClient.Status().Update(context.Background(), snapshotRestore)
}

func (m *manager) volumeReplicaSnapshotRestorePreCheck(snapshotRestore *apisv1alpha1.LocalVolumeReplicaSnapshotRestore) error {
	logCtx := m.logger.WithFields(log.Fields{"SnapshotRestore": snapshotRestore.Name, "Spec": snapshotRestore.Spec})
	logCtx.Debug("PreCheck a VolumeReplicaSnapshotRestore")

	targetVolume := &apisv1alpha1.LocalVolumeSnapshot{}
	if err := m.apiClient.Get(context.Background(), client.ObjectKey{Name: snapshotRestore.Spec.TargetVolume}, targetVolume); err != nil {
		logCtx.WithError(err).Error("Failed to get target volume")
		return err
	}

	// consider data security, abort if target volume has been mounted

	if len(m.mounter.GetDeviceMountPoints(snapshotRestore.Spec.TargetVolume)) > 0 {
		err := fmt.Errorf("target volume is already mounted, cannot restore from snapshot now")
		logCtx.WithError(err).Error("Failed to check mount point for target volume")

		snapshotRestore.Status.Message = err.Error()
		_ = m.apiClient.Status().Update(context.Background(), snapshotRestore)
		return err
	}

	snapshotRestore.Status.State = apisv1alpha1.OperationStateInProgress
	return m.apiClient.Status().Update(context.Background(), snapshotRestore)
}

func (m *manager) restoreVolumeFromSnapshot(snapshotRestore *apisv1alpha1.LocalVolumeReplicaSnapshotRestore) error {
	logCtx := m.logger.WithFields(log.Fields{"SnapshotRestore": snapshotRestore.Name, "Spec": snapshotRestore.Spec})
	logCtx.Debug("Restoring a VolumeReplicaSnapshotRestore")

	var err error
	switch snapshotRestore.Spec.RestoreType {
	case apisv1alpha1.RestoreTypeMerge:
		err = m.restoreVolumeByMerge(snapshotRestore)
	case apisv1alpha1.RestoreTypeCreate:
		err = m.restoreVolumeByCreate(snapshotRestore)
	default:
		logCtx.Error("invalid restore type")
	}

	if err != nil {
		snapshotRestore.Status.Message = err.Error()
		_ = m.apiClient.Status().Update(context.Background(), snapshotRestore)
		return err
	}

	snapshotRestore.Status.Message = "restore volume successfully"
	snapshotRestore.Status.State = apisv1alpha1.OperationStateCompleted
	return m.apiClient.Status().Update(context.Background(), snapshotRestore)
}

func (m *manager) restoreVolumeByMerge(snapshotRestore *apisv1alpha1.LocalVolumeReplicaSnapshotRestore) error {

	return nil
}

func (m *manager) restoreVolumeByCreate(snapshotRestore *apisv1alpha1.LocalVolumeReplicaSnapshotRestore) error {
	return nil
}

func (m *manager) volumeReplicaSnapshotRestoreAbort(snapshotRestore *apisv1alpha1.LocalVolumeReplicaSnapshotRestore) error {
	logCtx := m.logger.WithFields(log.Fields{"SnapshotRestore": snapshotRestore.Name, "Spec": snapshotRestore.Spec})
	logCtx.Debug("Abort a VolumeReplicaSnapshotRestore")

	snapshotRestore.Status.State = apisv1alpha1.OperationStateAborted
	return m.apiClient.Status().Update(context.TODO(), snapshotRestore)
}

func (m *manager) volumeReplicaSnapshotRestoreCleanup(snapshotRestore *apisv1alpha1.LocalVolumeReplicaSnapshotRestore) error {
	logCtx := m.logger.WithFields(log.Fields{"SnapshotRestore": snapshotRestore.Name, "Spec": snapshotRestore.Spec})
	logCtx.Debug("Cleanup a VolumeReplicaSnapshotRestore")

	cleanedCount := 0
	for _, replicaSnapshotRestoreName := range snapshotRestore.Status.VolumeReplicaSnapshotRestore {
		replicaSnapshotRestore := &apisv1alpha1.LocalVolumeReplicaSnapshot{}
		if err := m.apiClient.Get(context.Background(), client.ObjectKey{Name: replicaSnapshotRestoreName}, replicaSnapshotRestore); err != nil {
			if errors.IsNotFound(err) {
				cleanedCount++
				logCtx.WithField("ReplicaSnapshotRestore", replicaSnapshotRestoreName).WithError(err).Error("Cleanup VolumeReplicaSnapshotRestore successfully")
				continue
			}
			logCtx.WithField("ReplicaSnapshotRestore", replicaSnapshotRestoreName).WithError(err).Error("Failed to get VolumeReplicaSnapshotRestore")
			return err
		}

		if !replicaSnapshotRestore.Spec.Delete {
			replicaSnapshotRestore.Spec.Delete = true
			if err := m.apiClient.Update(context.Background(), replicaSnapshotRestore); err != nil {
				logCtx.WithField("ReplicaSnapshotRestore", replicaSnapshotRestoreName).Error("Failed to cleanup VolumeReplicaSnapshotRestore")
				return err
			}
			logCtx.WithField("ReplicaSnapshotRestore", replicaSnapshotRestoreName).Error("Cleaning VolumeReplicaSnapshotRestore")
		}
	}

	if cleanedCount < len(snapshotRestore.Status.VolumeReplicaSnapshotRestore) {
		logCtx.Debugf("Remaining %d VolumeReplicaSnapshotRestore to clean", len(snapshotRestore.Status.VolumeReplicaSnapshotRestore)-cleanedCount)
		return nil
	}

	return m.apiClient.Delete(context.TODO(), snapshotRestore)
}

func (m *manager) getSourceVolumeFromSnapshot(volumeSnapshotName string) (*apisv1alpha1.LocalVolume, error) {
	volumeSnapshot := &apisv1alpha1.LocalVolumeSnapshot{}
	if err := m.apiClient.Get(context.Background(), client.ObjectKey{Name: volumeSnapshotName}, volumeSnapshot); err != nil {
		return nil, err
	}

	sourceVolume := &apisv1alpha1.LocalVolume{}
	if err := m.apiClient.Get(context.Background(), client.ObjectKey{Name: volumeSnapshot.Spec.SourceVolume}, volumeSnapshot); err != nil {
		return nil, err
	}

	return sourceVolume, nil
}
