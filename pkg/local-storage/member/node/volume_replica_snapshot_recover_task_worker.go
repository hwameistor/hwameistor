package node

import (
	"context"
	"fmt"
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-storage/member/node/storage"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (m *manager) startVolumeReplicaSnapshotRecoverTaskWorker(stopCh <-chan struct{}) {
	m.logger.Debug("Volume Replica Snapshot Recover Worker is working now")
	go func() {
		for {
			task, shutdown := m.volumeReplicaSnapshotRecoverTaskQueue.Get()
			if shutdown {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Stop the Volume Replica Snapshot Recover worker")
				break
			}
			if err := m.processVolumeReplicaSnapshotRecover(task); err != nil {
				m.logger.WithFields(log.Fields{"task": task, "attempts": m.volumeReplicaSnapshotRecoverTaskQueue.NumRequeues(task), "error": err.Error()}).Error("Failed to process Volume Replica Snapshot Recover task, retry later")
				m.volumeReplicaSnapshotRecoverTaskQueue.AddRateLimited(task)
			} else {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Completed a Volume Replica Snapshot Recover task.")
				m.volumeReplicaSnapshotRecoverTaskQueue.Forget(task)
			}
			m.volumeReplicaSnapshotRecoverTaskQueue.Done(task)
		}
	}()

	<-stopCh
	m.volumeReplicaSnapshotRecoverTaskQueue.Shutdown()
}

func (m *manager) processVolumeReplicaSnapshotRecover(recoverName string) error {
	logCtx := m.logger.WithFields(log.Fields{"VolumeReplicaSnapshotRecover": recoverName})
	logCtx.Debug("Working on a VolumeReplicaSnapshotRecover task")
	replicaSnapshotRecover := &apisv1alpha1.LocalVolumeReplicaSnapshotRecover{}
	if err := m.apiClient.Get(context.TODO(), types.NamespacedName{Name: recoverName}, replicaSnapshotRecover); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("Failed to get VolumeReplicaSnapshotRecover from cache")
			return err
		}
		logCtx.Info("Not found the VolumeReplicaSnapshotRecover from cache, should be deleted already")
		return nil
	}

	if replicaSnapshotRecover.Spec.Abort &&
		replicaSnapshotRecover.Status.State != apisv1alpha1.OperationStateToBeAborted &&
		replicaSnapshotRecover.Status.State != apisv1alpha1.OperationStateAborting &&
		replicaSnapshotRecover.Status.State != apisv1alpha1.OperationStateAborted {

		replicaSnapshotRecover.Status.State = apisv1alpha1.OperationStateToBeAborted
		return m.apiClient.Status().Update(context.TODO(), replicaSnapshotRecover)
	}

	// log with namespace/name is enough
	logCtx = m.logger.WithFields(log.Fields{"TargetVolume": replicaSnapshotRecover.Spec.TargetVolume, "SnapshotRecover": replicaSnapshotRecover.Name, "Spec": replicaSnapshotRecover.Spec, "Status": replicaSnapshotRecover.Status})
	logCtx.Debug("Starting to process a VolumeReplicaSnapshotRecover")

	// state chain: (empty) -> Submitted -> Start -> InProgress -> Completed
	switch replicaSnapshotRecover.Status.State {
	case "":
		return m.volumeReplicaSnapshotRecoverSubmit(replicaSnapshotRecover)
	case apisv1alpha1.OperationStateSubmitted:
		return m.volumeReplicaSnapshotRecoverPreCheck(replicaSnapshotRecover)
	case apisv1alpha1.OperationStateInProgress:
		return m.recoverVolumeFromSnapshot(replicaSnapshotRecover)
	case apisv1alpha1.OperationStateToBeAborted:
		return m.volumeReplicaSnapshotRecoverAbort(replicaSnapshotRecover)
	case apisv1alpha1.OperationStateAborted:
		return m.volumeReplicaSnapshotRecoverCleanup(replicaSnapshotRecover)
	case apisv1alpha1.OperationStateCompleted:
		// wait for VolumeSnapshotRecover confirm to delete
		m.logger.Info("VolumeReplicaSnapshotRecover is completed")
		return nil
	default:
		logCtx.Error("Invalid state/phase")
	}
	return fmt.Errorf("invalid state")
}

func (m *manager) volumeReplicaSnapshotRecoverSubmit(snapshotRecover *apisv1alpha1.LocalVolumeReplicaSnapshotRecover) error {
	logCtx := m.logger.WithFields(log.Fields{"SnapshotRecover": snapshotRecover.Name, "Spec": snapshotRecover.Spec})
	logCtx.Debug("Submit a VolumeReplicaSnapshotRecover")

	snapshotRecover.Status.State = apisv1alpha1.OperationStateSubmitted
	return m.apiClient.Status().Update(context.Background(), snapshotRecover)
}

func (m *manager) volumeReplicaSnapshotRecoverPreCheck(snapshotRecover *apisv1alpha1.LocalVolumeReplicaSnapshotRecover) error {
	logCtx := m.logger.WithFields(log.Fields{"SnapshotRecover": snapshotRecover.Name, "Spec": snapshotRecover.Spec})
	logCtx.Debug("PreCheck a VolumeReplicaSnapshotRecover")

	targetVolume := &apisv1alpha1.LocalVolume{}
	if err := m.apiClient.Get(context.Background(), client.ObjectKey{Name: snapshotRecover.Spec.TargetVolume}, targetVolume); err != nil {
		logCtx.WithError(err).Error("Failed to get target volume")
		return err
	}

	// consider data security, abort if target volume has been mounted

	if len(m.mounter.GetDeviceMountPoints(snapshotRecover.Spec.TargetVolume)) > 0 {
		err := fmt.Errorf("target volume is already mounted, cannot recover from snapshot now")
		logCtx.WithError(err).Error("Failed to check mount point for target volume")

		snapshotRecover.Status.Message = err.Error()
		_ = m.apiClient.Status().Update(context.Background(), snapshotRecover)
		return err
	}

	snapshotRecover.Status.State = apisv1alpha1.OperationStateInProgress
	return m.apiClient.Status().Update(context.Background(), snapshotRecover)
}

func (m *manager) recoverVolumeFromSnapshot(snapshotRecover *apisv1alpha1.LocalVolumeReplicaSnapshotRecover) error {
	logCtx := m.logger.WithFields(log.Fields{"SnapshotRecover": snapshotRecover.Name, "Spec": snapshotRecover.Spec})
	logCtx.Debug("Recovering a VolumeReplicaSnapshotRecover")

	var err error
	switch snapshotRecover.Spec.RecoverType {
	case apisv1alpha1.RecoverTypeRollback:
		err = m.rollbackSnapshot(snapshotRecover)
	case apisv1alpha1.RecoverTypeRestore:
		err = m.restoreSnapshot(snapshotRecover)
	default:
		logCtx.Error("invalid recover type")
	}

	if err != nil {
		snapshotRecover.Status.Message = err.Error()
		_ = m.apiClient.Status().Update(context.Background(), snapshotRecover)
		return err
	}

	snapshotRecover.Status.Message = "recover volume successfully"
	snapshotRecover.Status.State = apisv1alpha1.OperationStateCompleted
	return m.apiClient.Status().Update(context.Background(), snapshotRecover)
}

func (m *manager) rollbackSnapshot(snapshotRecover *apisv1alpha1.LocalVolumeReplicaSnapshotRecover) (err error) {
	logCtx := m.logger.WithFields(log.Fields{"ReplicaSnapshotRecover": snapshotRecover.Name, "TargetVolume": snapshotRecover.Spec.TargetVolume, "SourceSnapshot": snapshotRecover.Spec.SourceVolumeSnapshot})
	logCtx.Debug("Start rolling back snapshot")

	// 1. check whether volume snapshot is already merged

	volumeReplicaSnapshot := &apisv1alpha1.LocalVolumeReplicaSnapshot{}
	if err = m.apiClient.Get(context.Background(), client.ObjectKey{Name: snapshotRecover.Spec.SourceVolumeReplicaSnapshot}, volumeReplicaSnapshot); err != nil {
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

	if err = m.Storage().VolumeReplicaSnapshotManager().RollbackVolumeReplicaSnapshot(snapshotRecover); err != nil {
		logCtx.WithError(err).Error("Failed to start rollback snapshot")
		return err
	}

	// 4. due to snapshot rollback is performed asynchronously, check again whether volume snapshot is already merging
	return isSnapshotInMerging()
}

func (m *manager) restoreSnapshot(snapshotRecover *apisv1alpha1.LocalVolumeReplicaSnapshotRecover) (err error) {
	logCtx := m.logger.WithFields(log.Fields{"ReplicaSnapshotRecover": snapshotRecover.Name, "TargetVolume": snapshotRecover.Spec.TargetVolume, "SourceSnapshot": snapshotRecover.Spec.SourceVolumeSnapshot})
	logCtx.Debug("Start restoring snapshot")

	// 1. check whether volume snapshot is already merged

	volumeReplicaSnapshot := &apisv1alpha1.LocalVolumeReplicaSnapshot{}
	if err = m.apiClient.Get(context.Background(), client.ObjectKey{Name: snapshotRecover.Spec.SourceVolumeReplicaSnapshot}, volumeReplicaSnapshot); err != nil {
		if errors.IsNotFound(err) {
			// snapshot is already merged
			logCtx.WithField("VolumeReplicaSnapshot", volumeReplicaSnapshot).Info("volume replica snapshot is restored")
			return nil
		}
		logCtx.WithField("VolumeReplicaSnapshot", volumeReplicaSnapshot).Error("Failed to get volume replica snapshot")
		return err
	}

	if _, err = m.Storage().VolumeReplicaSnapshotManager().GetVolumeReplicaSnapshot(volumeReplicaSnapshot); err != nil {
		if err == storage.ErrorSnapshotNotFound {
			// snapshot is already merged, remove the snapshot from apiserver
			logCtx.WithField("VolumeReplicaSnapshot", volumeReplicaSnapshot).Info("volume replica snapshot is restored")
			volumeReplicaSnapshot.Spec.Delete = true
			return m.apiClient.Update(context.Background(), volumeReplicaSnapshot)
		}
		logCtx.WithField("VolumeReplicaSnapshot", volumeReplicaSnapshot).Error("Failed to get volume replica snapshot status")
		return err
	}

	// 2. todo: check whether volume snapshot is already restoring

	// 3. rollback the volumes snapshot to the source volume

	if err = m.Storage().VolumeReplicaSnapshotManager().RestoreVolumeReplicaSnapshot(snapshotRecover); err != nil {
		logCtx.WithError(err).Error("Failed to start restore snapshot")
	}
	return err
}

func (m *manager) volumeReplicaSnapshotRecoverAbort(snapshotRecover *apisv1alpha1.LocalVolumeReplicaSnapshotRecover) error {
	logCtx := m.logger.WithFields(log.Fields{"SnapshotRecover": snapshotRecover.Name, "Spec": snapshotRecover.Spec})
	logCtx.Debug("Abort a VolumeReplicaSnapshotRecover")

	snapshotRecover.Status.State = apisv1alpha1.OperationStateAborted
	return m.apiClient.Status().Update(context.TODO(), snapshotRecover)
}

func (m *manager) volumeReplicaSnapshotRecoverCleanup(snapshotRecover *apisv1alpha1.LocalVolumeReplicaSnapshotRecover) error {
	logCtx := m.logger.WithFields(log.Fields{"SnapshotRecover": snapshotRecover.Name, "Spec": snapshotRecover.Spec})
	logCtx.Debug("Cleanup a VolumeReplicaSnapshotRecover")

	return m.apiClient.Delete(context.TODO(), snapshotRecover)
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
