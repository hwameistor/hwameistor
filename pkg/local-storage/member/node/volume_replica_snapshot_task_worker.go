package node

import (
	"context"
	"fmt"
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-storage/member/node/storage"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func (m *manager) startVolumeReplicaSnapshotTaskWorker(stopCh <-chan struct{}) {
	m.logger.Debug("VolumeReplica Snapshot Worker is working now")
	go func() {
		for {
			task, shutdown := m.volumeReplicaSnapshotTaskQueue.Get()
			if shutdown {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Stop the VolumeReplica Snapshot worker")
				break
			}
			if err := m.processReplicaSnapshot(task); err != nil {
				m.logger.WithFields(log.Fields{"task": task, "attempts": m.volumeReplicaSnapshotTaskQueue.NumRequeues(task), "error": err.Error()}).Error("Failed to process VolumeReplica Snapshot task, retry later")
				m.volumeReplicaSnapshotTaskQueue.AddRateLimited(task)
			} else {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Completed a VolumeReplica Snapshot task.")
				m.volumeReplicaSnapshotTaskQueue.Forget(task)
			}
			m.volumeReplicaSnapshotTaskQueue.Done(task)
		}
	}()

	<-stopCh
	m.volumeReplicaSnapshotTaskQueue.Shutdown()
}

func (m *manager) processReplicaSnapshot(replicaSnapName string) error {
	logCtx := m.logger.WithFields(log.Fields{"ReplicaSnapshot": replicaSnapName})
	logCtx.Debug("Working on a VolumeReplica Snapshot task")
	replicaSnapshot := &apisv1alpha1.LocalVolumeReplicaSnapshot{}
	if err := m.apiClient.Get(context.TODO(), types.NamespacedName{Name: replicaSnapName}, replicaSnapshot); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("Failed to get VolumeReplica Snapshot from cache")
			return err
		}
		logCtx.Info("Not found the VolumeReplica Snapshot from cache, should be deleted already.")
		return nil
	}

	if replicaSnapshot.Spec.Delete && replicaSnapshot.Status.State != apisv1alpha1.VolumeReplicaStateToBeDeleted && replicaSnapshot.Status.State != apisv1alpha1.VolumeReplicaStateDeleted {
		replicaSnapshot.Status.State = apisv1alpha1.VolumeReplicaStateToBeDeleted
		return m.apiClient.Status().Update(context.TODO(), replicaSnapshot)
	}

	// log with namespace/name is enough
	logCtx = m.logger.WithFields(log.Fields{"Volume": replicaSnapshot.Spec.SourceVolume, "Snapshot": replicaSnapshot.Name, "Spec": replicaSnapshot.Spec, "Status": replicaSnapshot.Status})
	logCtx.Debug("Starting to process a Volume Snapshot")
	switch replicaSnapshot.Status.State {
	case "":
		return m.volumeReplicaSnapshotSubmit(replicaSnapshot)
	case apisv1alpha1.VolumeStateCreating:
		return m.volumeReplicaSnapshotCreate(replicaSnapshot)
	case apisv1alpha1.VolumeStateReady, apisv1alpha1.VolumeStateNotReady:
		return m.volumeReplicaSnapshotReadyOrNot(replicaSnapshot)
	case apisv1alpha1.VolumeStateToBeDeleted:
		return m.volumeReplicaSnapshotDelete(replicaSnapshot)
	case apisv1alpha1.VolumeStateDeleted:
		return m.volumeReplicaSnapshotCleanup(replicaSnapshot)
	default:
		logCtx.Error("Invalid state")
	}
	return fmt.Errorf("invalid state")
}

func (m *manager) volumeReplicaSnapshotSubmit(snapshot *apisv1alpha1.LocalVolumeReplicaSnapshot) error {
	logCtx := m.logger.WithFields(log.Fields{"Snapshot": snapshot.Name, "Spec": snapshot.Spec})
	logCtx.Debug("Submit a VolumeReplica Snapshot")

	snapshot.Status.State = apisv1alpha1.VolumeStateCreating
	return m.apiClient.Status().Update(context.TODO(), snapshot)
}

func (m *manager) volumeReplicaSnapshotCreate(snapshot *apisv1alpha1.LocalVolumeReplicaSnapshot) error {
	logCtx := m.logger.WithFields(log.Fields{"Snapshot": snapshot.Name, "Spec": snapshot.Spec})
	logCtx.Debug("Create a VolumeReplica Snapshot")

	snapshotExistOnHost := func() (bool, error) {
		_, err := m.storageMgr.VolumeReplicaSnapshotManager().GetVolumeReplicaSnapshot(snapshot)
		if err != nil && err != storage.ErrorSnapshotNotFound {
			logCtx.WithError(err).Error("Failed to get VolumeReplica Snapshot from host")
			return false, err
		}
		return !(err == storage.ErrorSnapshotNotFound), nil
	}

	exist, err := snapshotExistOnHost()
	if err != nil {
		return err
	}

	if !exist {
		if err = m.storageMgr.VolumeReplicaSnapshotManager().CreateVolumeReplicaSnapshot(snapshot); err != nil {
			logCtx.WithError(err).Error("Failed to create VolumeReplica Snapshot")
			return err
		}
	}

	snapshot.Status.State = apisv1alpha1.VolumeStateNotReady
	return m.apiClient.Status().Update(context.TODO(), snapshot)
}

func (m *manager) volumeReplicaSnapshotReadyOrNot(snapshot *apisv1alpha1.LocalVolumeReplicaSnapshot) error {
	logCtx := m.logger.WithFields(log.Fields{"Snapshot": snapshot.Name, "Spec": snapshot.Spec})
	logCtx.Debug("Check a VolumeReplica Snapshot status")

	snapRealStatus, err := m.storageMgr.VolumeReplicaSnapshotManager().GetVolumeReplicaSnapshot(snapshot)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get VolumeReplica Snapshot from host")

		// keep monitoring the snapshot status until no error happens
		snapshot.Status.State = apisv1alpha1.VolumeStateNotReady
		snapshot.Status.Message = err.Error()
		if err = m.apiClient.Status().Update(context.TODO(), snapshot); err != nil {
			return err
		}
		return err
	}

	snapshot.Status = *snapRealStatus
	if snapshot.Status.State != apisv1alpha1.NodeStateReady {
		if err = m.apiClient.Status().Update(context.TODO(), snapshot); err != nil {
			return err
		}
		err = fmt.Errorf(snapshot.Status.Message)
		return err
	}

	return m.apiClient.Status().Update(context.TODO(), snapshot)
}

func (m *manager) volumeReplicaSnapshotCleanup(snapshot *apisv1alpha1.LocalVolumeReplicaSnapshot) error {
	logCtx := m.logger.WithFields(log.Fields{"Snapshot": snapshot.Name, "Spec": snapshot.Spec})
	logCtx.Debug("Cleanup a VolumeReplica Snapshot")

	if err := m.apiClient.Delete(context.TODO(), snapshot); err != nil {
		return err
	}

	// clean up the records in cache when the replica snapshot is deleted from the cluster
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.replicaSnapshotsRecords, snapshot.Spec.VolumeSnapshotName)
	return nil
}

func (m *manager) volumeReplicaSnapshotDelete(snapshot *apisv1alpha1.LocalVolumeReplicaSnapshot) error {
	logCtx := m.logger.WithFields(log.Fields{"Snapshot": snapshot.Name, "Spec": snapshot.Spec})
	logCtx.Debug("Delete a VolumeReplica Snapshot")

	if _, err := m.storageMgr.VolumeReplicaSnapshotManager().GetVolumeReplicaSnapshot(snapshot); err == nil {
		// delete the volume replica snapshot from the node
		err = m.storageMgr.VolumeReplicaSnapshotManager().DeleteVolumeReplicaSnapshot(snapshot)
		if err != nil {
			logCtx.WithError(err).Error("Failed to delete VolumeReplica Snapshot")
			return err
		}
	} else if err != storage.ErrorSnapshotNotFound {
		logCtx.WithError(err).Error("Failed to get VolumeReplica Snapshot from host")
		return err
	}

	snapshot.Status.State = apisv1alpha1.VolumeStateDeleted
	return m.apiClient.Status().Update(context.TODO(), snapshot)
}
