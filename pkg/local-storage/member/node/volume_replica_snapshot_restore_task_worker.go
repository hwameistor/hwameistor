package node

import (
	"context"
	"fmt"
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-storage/member/node/storage"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"path"
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
	replicaSnapshotRestore := &apisv1alpha1.LocalVolumeReplicaSnapshotRestore{}
	if err := m.apiClient.Get(context.TODO(), types.NamespacedName{Name: restoreName}, replicaSnapshotRestore); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("Failed to get VolumeReplicaSnapshotRestore from cache")
			return err
		}
		logCtx.Info("Not found the VolumeReplicaSnapshotRestore from cache, should be deleted already")
		return nil
	}

	if replicaSnapshotRestore.Spec.Abort &&
		replicaSnapshotRestore.Status.State != apisv1alpha1.OperationStateToBeAborted &&
		replicaSnapshotRestore.Status.State != apisv1alpha1.OperationStateAborting &&
		replicaSnapshotRestore.Status.State != apisv1alpha1.OperationStateAborted {

		replicaSnapshotRestore.Status.State = apisv1alpha1.OperationStateToBeAborted
		return m.apiClient.Status().Update(context.TODO(), replicaSnapshotRestore)
	}

	// log with namespace/name is enough
	logCtx = m.logger.WithFields(log.Fields{"TargetVolume": replicaSnapshotRestore.Spec.TargetVolume, "SnapshotRestore": replicaSnapshotRestore.Name, "Spec": replicaSnapshotRestore.Spec, "Status": replicaSnapshotRestore.Status})
	logCtx.Debug("Starting to process a VolumeReplicaSnapshotRestore")

	// state chain: (empty) -> Submitted -> Start -> InProgress -> Completed
	switch replicaSnapshotRestore.Status.State {
	case "":
		return m.volumeReplicaSnapshotRestoreSubmit(replicaSnapshotRestore)
	case apisv1alpha1.OperationStateSubmitted:
		return m.volumeReplicaSnapshotRestorePreCheck(replicaSnapshotRestore)
	case apisv1alpha1.OperationStateInProgress:
		return m.restoreVolumeFromSnapshot(replicaSnapshotRestore)
	case apisv1alpha1.OperationStateToBeAborted:
		return m.volumeReplicaSnapshotRestoreAbort(replicaSnapshotRestore)
	case apisv1alpha1.OperationStateAborted:
		return m.volumeReplicaSnapshotRestoreCleanup(replicaSnapshotRestore)
	case apisv1alpha1.OperationStateCompleted:
		// wait for VolumeSnapshotRestore confirm to delete
		m.logger.Info("VolumeReplicaSnapshotRestore is completed")
		return nil
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

	targetVolume := &apisv1alpha1.LocalVolume{}
	if err := m.apiClient.Get(context.Background(), client.ObjectKey{Name: snapshotRestore.Spec.TargetVolume}, targetVolume); err != nil {
		logCtx.WithError(err).Error("Failed to get target volume")
		return err
	}

	// consider data security, abort if target volume has been mounted
	// fixme: device path fetched by lvm will be better
	devicePath := path.Join("/dev", snapshotRestore.Spec.TargetPoolName, snapshotRestore.Spec.TargetVolume)
	if len(m.mounter.GetDeviceMountPoints(devicePath)) > 0 {
		err := fmt.Errorf(
			"target volume is already mounted, cannot %s from snapshot now", snapshotRestore.Spec.RestoreType)
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
	case apisv1alpha1.RestoreTypeRollback:
		err = m.rollbackSnapshot(snapshotRestore)
	case apisv1alpha1.RestoreTypeCreate:
		err = m.restoreSnapshot(snapshotRestore)
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

func (m *manager) rollbackSnapshot(snapshotRestore *apisv1alpha1.LocalVolumeReplicaSnapshotRestore) (err error) {
	logCtx := m.logger.WithFields(log.Fields{"ReplicaSnapshotRestore": snapshotRestore.Name, "TargetVolume": snapshotRestore.Spec.TargetVolume, "SourceSnapshot": snapshotRestore.Spec.SourceVolumeSnapshot})
	logCtx.Debug("Start rolling back snapshot")

	// 1. check whether volume snapshot is already merged

	volumeReplicaSnapshot := &apisv1alpha1.LocalVolumeReplicaSnapshot{}
	if err = m.apiClient.Get(context.Background(), client.ObjectKey{Name: snapshotRestore.Spec.SourceVolumeReplicaSnapshot}, volumeReplicaSnapshot); err != nil {
		if errors.IsNotFound(err) {
			// snapshot is already merged
			logCtx.WithField("VolumeReplicaSnapshot", volumeReplicaSnapshot).Info("volume replica snapshot is merged")
			return nil
		}
		logCtx.WithField("VolumeReplicaSnapshot", volumeReplicaSnapshot).Error("Failed to get volume replica snapshot")
		return err
	}

	isSnapshotInMerging := func() error {
		replicaSnapshotStatus := &apisv1alpha1.LocalVolumeReplicaSnapshotStatus{}
		if replicaSnapshotStatus, err = m.Storage().VolumeReplicaSnapshotManager().GetVolumeReplicaSnapshot(volumeReplicaSnapshot); err != nil {
			if err == storage.ErrorSnapshotNotFound {
				// snapshot is already merged, remove the snapshot from apiserver
				logCtx.WithField("VolumeReplicaSnapshot", volumeReplicaSnapshot).Info("volume replica snapshot is merged")
				volumeReplicaSnapshot.Spec.Delete = true
				return m.apiClient.Update(context.Background(), volumeReplicaSnapshot)
			}
			logCtx.WithField("VolumeReplicaSnapshot", volumeReplicaSnapshot).Error("Failed to get volume replica snapshot status")
			return err
		}

		if replicaSnapshotStatus.Attribute.Invalid {
			err = fmt.Errorf("snapshot is expiration, cannot be used to rollback")
			logCtx.WithField("VolumeReplicaSnapshot", volumeReplicaSnapshot).WithError(err).Error("Failed to rollback snapshot")
			return err
		}
		if replicaSnapshotStatus.Attribute.Merging {
			err = fmt.Errorf("snapshot is merging")
			logCtx.WithField("VolumeReplicaSnapshot", volumeReplicaSnapshot).Info(err.Error())
			return err
		}

		return nil
	}

	// 2. check whether volume snapshot is already merging
	if err = isSnapshotInMerging(); err != nil {
		return err
	}

	// 3. rollback the volumes snapshot to the source volume

	if err = m.Storage().VolumeReplicaSnapshotManager().RollbackVolumeReplicaSnapshot(snapshotRestore); err != nil {
		logCtx.WithError(err).Error("Failed to start rollback snapshot")
		return err
	}

	// 4. due to snapshot rollback is performed asynchronously, check again whether volume snapshot is already merging
	return isSnapshotInMerging()
}

func (m *manager) restoreSnapshot(snapshotRestore *apisv1alpha1.LocalVolumeReplicaSnapshotRestore) (err error) {
	logCtx := m.logger.WithFields(log.Fields{"ReplicaSnapshotRestore": snapshotRestore.Name, "TargetVolume": snapshotRestore.Spec.TargetVolume, "SourceSnapshot": snapshotRestore.Spec.SourceVolumeSnapshot})
	logCtx.Debug("Start restoring snapshot")

	// NOTE: different from snapshotRollback, restore snapshot won't cause snapshot deletion, so we just restore without judging snapshot is already merged
	if err = m.Storage().VolumeReplicaSnapshotManager().RestoreVolumeReplicaSnapshot(snapshotRestore); err != nil {
		logCtx.WithError(err).Error("Failed to restore snapshot")
	}
	return err
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
