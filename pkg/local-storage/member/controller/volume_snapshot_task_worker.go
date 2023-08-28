package controller

import (
	"context"
	"fmt"
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sort"
)

func (m *manager) startVolumeSnapshotTaskWorker(stopCh <-chan struct{}) {
	m.logger.Debug("Volume Snapshot Worker is working now")
	go func() {
		for {
			task, shutdown := m.volumeSnapshotTaskQueue.Get()
			if shutdown {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Stop the Volume Snapshot worker")
				break
			}
			if err := m.processVolumeSnapshot(task); err != nil {
				m.logger.WithFields(log.Fields{"task": task, "attempts": m.volumeSnapshotTaskQueue.NumRequeues(task), "error": err.Error()}).Error("Failed to process Volume Snapshot task, retry later")
				m.volumeSnapshotTaskQueue.AddRateLimited(task)
			} else {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Completed a Volume Snapshot task.")
				m.volumeSnapshotTaskQueue.Forget(task)
			}
			m.volumeSnapshotTaskQueue.Done(task)
		}
	}()

	<-stopCh
	m.volumeSnapshotTaskQueue.Shutdown()
}

func (m *manager) processVolumeSnapshot(snapName string) error {
	logCtx := m.logger.WithFields(log.Fields{"Snapshot": snapName})
	logCtx.Debug("Working on a VolumeSnapshot task")
	snapshot := &apisv1alpha1.LocalVolumeSnapshot{}
	if err := m.apiClient.Get(context.TODO(), types.NamespacedName{Name: snapName}, snapshot); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("Failed to get Volume Snapshot from cache")
			return err
		}
		logCtx.Info("Not found the VolumeSnapshot from cache, should be deleted already")
		return nil
	}

	if snapshot.Spec.Delete && snapshot.Status.State != apisv1alpha1.VolumeReplicaStateToBeDeleted && snapshot.Status.State != apisv1alpha1.VolumeReplicaStateDeleted {
		snapshot.Status.State = apisv1alpha1.VolumeReplicaStateToBeDeleted
		return m.apiClient.Status().Update(context.TODO(), snapshot)
	}

	// log with namespace/name is enough
	logCtx = m.logger.WithFields(log.Fields{"Volume": snapshot.Spec.SourceVolume, "Snapshot": snapshot.Name, "Spec": snapshot.Spec, "Status": snapshot.Status})
	logCtx.Debug("Starting to process a Volume Snapshot")
	switch snapshot.Status.State {
	case "":
		return m.volumeSnapshotSubmit(snapshot)
	case apisv1alpha1.VolumeStateCreating:
		return m.volumeSnapshotCreate(snapshot)
	case apisv1alpha1.VolumeStateReady, apisv1alpha1.VolumeStateNotReady:
		return m.volumeSnapshotReadyOrNot(snapshot)
	case apisv1alpha1.VolumeStateToBeDeleted:
		return m.volumeSnapshotDelete(snapshot)
	case apisv1alpha1.VolumeStateDeleted:
		return m.volumeSnapshotCleanup(snapshot)
	default:
		logCtx.Error("Invalid state")
	}
	return fmt.Errorf("invalid state")
}

func (m *manager) volumeSnapshotSubmit(snapshot *apisv1alpha1.LocalVolumeSnapshot) error {
	logCtx := m.logger.WithFields(log.Fields{"Snapshot": snapshot.Name, "Spec": snapshot.Spec})
	logCtx.Debug("Submit a VolumeSnapshot")

	snapshot.Status.State = apisv1alpha1.VolumeStateCreating
	return m.apiClient.Status().Update(context.TODO(), snapshot)
}

func (m *manager) volumeSnapshotCreate(snapshot *apisv1alpha1.LocalVolumeSnapshot) error {
	logCtx := m.logger.WithFields(log.Fields{"Snapshot": snapshot.Name, "Spec": snapshot.Spec})
	logCtx.Debug("Create a VolumeSnapshot")

	snapshot.Status.State = apisv1alpha1.VolumeStateNotReady
	return m.apiClient.Status().Update(context.TODO(), snapshot)
}

func (m *manager) volumeSnapshotReadyOrNot(snapshot *apisv1alpha1.LocalVolumeSnapshot) error {
	logCtx := m.logger.WithFields(log.Fields{"Snapshot": snapshot.Name, "Spec": snapshot.Spec})
	logCtx.Debug("Check a VolumeSnapshot status in progress")

	// list all replicas snapshots, update status if all replicas snapshots are ready
	replicaSnapshotList := &apisv1alpha1.LocalVolumeReplicaSnapshotList{}
	if err := m.apiClient.List(context.TODO(), replicaSnapshotList); err != nil {
		logCtx.WithError(err).Error("Failed to list VolumeReplicas")
		return err
	}

	var volumeReplicaSnapshots, specifiedNodeReadySnapshots []string
	for _, replicaSnapshot := range replicaSnapshotList.Items {
		if replicaSnapshot.Spec.VolumeSnapshotName != snapshot.Name {
			continue
		}
		// collect all replicas snapshots belong to the volume snapshot
		volumeReplicaSnapshots = append(volumeReplicaSnapshots, replicaSnapshot.Name)

		// collect all specified node replicas ready snapshots
		if _, exist := utils.StrFind(snapshot.Spec.Accessibility.Nodes, replicaSnapshot.Spec.NodeName); !exist {
			continue
		}

		if replicaSnapshot.Status.State == apisv1alpha1.VolumeReplicaStateReady {
			specifiedNodeReadySnapshots = append(specifiedNodeReadySnapshots, replicaSnapshot.Name)
		}

		snapshot.Status.Attribute = replicaSnapshot.Status.Attribute
		snapshot.Status.CreationTime = replicaSnapshot.Status.CreationTime
		snapshot.Status.AllocatedCapacityBytes = replicaSnapshot.Status.AllocatedCapacityBytes
	}

	// keep sorted to avoid useless reconcile
	sort.Strings(volumeReplicaSnapshots)
	snapshot.Status.ReplicaSnapshots = volumeReplicaSnapshots

	if len(snapshot.Spec.Accessibility.Nodes) > len(specifiedNodeReadySnapshots) {
		logCtx.WithField("specifiedNodeReadySnapshots", specifiedNodeReadySnapshots).Debug("Not all VolumeReplicas Snapshots are ready")
		snapshot.Status.State = apisv1alpha1.VolumeStateNotReady
	} else {
		logCtx.WithField("specifiedNodeReadySnapshots", specifiedNodeReadySnapshots).Debug("All VolumeReplicas Snapshots are ready")
		snapshot.Status.State = apisv1alpha1.VolumeStateReady
	}

	return m.apiClient.Status().Update(context.TODO(), snapshot)
}

func (m *manager) volumeSnapshotCleanup(snapshot *apisv1alpha1.LocalVolumeSnapshot) error {
	logCtx := m.logger.WithFields(log.Fields{"Snapshot": snapshot.Name, "Spec": snapshot.Spec})
	logCtx.Debug("Cleanup a VolumeSnapshot")

	return m.apiClient.Delete(context.TODO(), snapshot)
}

func (m *manager) volumeSnapshotDelete(snapshot *apisv1alpha1.LocalVolumeSnapshot) error {
	logCtx := m.logger.WithFields(log.Fields{"Snapshot": snapshot.Name, "Spec": snapshot.Spec})
	logCtx.Debug("Delete a VolumeSnapshot")

	// check all volume replica snapshots in status.replicaSnapshots are deleted, try deleting if exist
	var existVolumeReplicaSnapshots []string
	for _, replicaSnapName := range snapshot.Status.ReplicaSnapshots {
		replicaSnapshot := &apisv1alpha1.LocalVolumeReplicaSnapshot{}
		if err := m.apiClient.Get(context.TODO(), types.NamespacedName{Name: replicaSnapName}, replicaSnapshot); err != nil {
			if !errors.IsNotFound(err) {
				logCtx.WithError(err).Error("Failed to get VolumeReplicaSnapshot from cache")
				return err
			}
			logCtx.WithField("volumeReplicaSnapshot", replicaSnapshot.Name).Info("Not found the VolumeReplicaSnapshot from cache, should be deleted already")
			continue
		}

		// try to set up delete true on the volume replica snapshot
		replicaSnapshot.Spec.Delete = true
		if err := m.apiClient.Update(context.TODO(), replicaSnapshot); err != nil {
			logCtx.WithField("volumeReplicaSnapshot", replicaSnapshot.Name).WithError(err).Error("Failed to delete VolumeReplicaSnapshot")
			return err
		}
		existVolumeReplicaSnapshots = append(existVolumeReplicaSnapshots, replicaSnapName)
	}

	if len(existVolumeReplicaSnapshots) > 0 {
		logCtx.WithField("existVolumeReplicaSnapshots", existVolumeReplicaSnapshots).Debug("Not all VolumeReplicas Snapshots are deleted")
		return fmt.Errorf("not all VolumeReplicas Snapshots are deleted")
	}

	snapshot.Status.State = apisv1alpha1.VolumeStateDeleted
	return m.apiClient.Status().Update(context.TODO(), snapshot)
}
