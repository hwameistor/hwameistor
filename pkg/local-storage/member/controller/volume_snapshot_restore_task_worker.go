package controller

import (
	"context"
	"fmt"
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-storage/utils"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

	snapshotRestore.Status.State = apisv1alpha1.OperationStateSubmitted
	return m.apiClient.Status().Update(context.Background(), snapshotRestore)
}

func (m *manager) volumeSnapshotRestoreStart(snapshotRestore *apisv1alpha1.LocalVolumeSnapshotRestore) error {
	logCtx := m.logger.WithFields(log.Fields{"SnapshotRestore": snapshotRestore.Name, "Spec": snapshotRestore.Spec})
	logCtx.Debug("Start a VolumeSnapshotRestore")

	sourceVolume, err := m.getSourceVolumeFromSnapshot(snapshotRestore.Spec.SourceVolumeSnapshot)
	if err != nil {
		logCtx.Error("Failed to get source volume from snapshot")
		return err
	}

	nodeVolumeReplicaSnapshot, err := m.getNodeVolumeReplicaSnapshot(snapshotRestore.Spec.SourceVolumeSnapshot)
	if err != nil {
		logCtx.Error("Failed to get volume replica snapshot from snapshot")
		return err
	}

	// create LocalVolumeReplicaSnapshotRestore on each node according to the topology of source volume
	for _, nodeName := range sourceVolume.Spec.Accessibility.Nodes {
		nodeReplicaSnapshot, ok := nodeVolumeReplicaSnapshot[nodeName]
		if !ok {
			err = fmt.Errorf("LocalVolumeReplicaSnapshot not found on node %s but it is accessible in the source LocalVolume topology", nodeName)
			logCtx.WithError(err).Error("Failed to create LocalVolumeReplicaSnapshot")
			return err
		}

		replicaSnapshotRestore := &apisv1alpha1.LocalVolumeReplicaSnapshotRestore{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("%s-%s", snapshotRestore.Name, utilrand.String(6)),
			},
			Spec: apisv1alpha1.LocalVolumeReplicaSnapshotRestoreSpec{
				LocalVolumeSnapshotRestoreSpec: snapshotRestore.Spec,
				NodeName:                       nodeName,
				SourceVolumeReplicaSnapshot:    nodeReplicaSnapshot,
				VolumeSnapshotRestore:          snapshotRestore.Name,
			},
		}

		if err = m.apiClient.Create(context.Background(), replicaSnapshotRestore); err != nil && !errors.IsAlreadyExists(err) {
			logCtx.WithField("replicaSnapshotRestore", replicaSnapshotRestore.Name).WithError(err).Error("Failed to create VolumeReplicaSnapshotRestore")
			return err
		}

		snapshotRestore.Status.VolumeReplicaSnapshotRestore = utils.AddUniqueStringItem(snapshotRestore.Status.VolumeReplicaSnapshotRestore, replicaSnapshotRestore.Name)
		logCtx.WithField("replicaSnapshotRestore", replicaSnapshotRestore.Name).WithError(err).Errorf("VolumeReplicaSnapshotRestore is created successfully on %s", nodeName)
	}

	snapshotRestore.Status.State = apisv1alpha1.OperationStateInProgress
	return m.apiClient.Status().Update(context.Background(), snapshotRestore)
}

func (m *manager) checkInProgressVolumeSnapshotRestore(snapshotRestore *apisv1alpha1.LocalVolumeSnapshotRestore) error {
	logCtx := m.logger.WithFields(log.Fields{"SnapshotRestore": snapshotRestore.Name, "Spec": snapshotRestore.Spec})
	logCtx.Debug("Check a InProgress VolumeSnapshotRestore")

	var (
		message        string
		completedCount int
	)

	for _, replicaSnapshotRestoreName := range snapshotRestore.Status.VolumeReplicaSnapshotRestore {
		replicaSnapshotRestore := &apisv1alpha1.LocalVolumeReplicaSnapshot{}
		if err := m.apiClient.Get(context.Background(), client.ObjectKey{Name: replicaSnapshotRestoreName}, replicaSnapshotRestore); err != nil {
			logCtx.WithField("ReplicaSnapshotRestore", replicaSnapshotRestoreName).WithError(err).Error("Failed to get VolumeReplicaSnapshotRestore")
			return err
		}

		if replicaSnapshotRestore.Status.Message != "" {
			message += fmt.Sprintf("%s: %s;", replicaSnapshotRestoreName, replicaSnapshotRestore.Status.Message)
		} else {
			message += fmt.Sprintf("%s is %s;", replicaSnapshotRestoreName, replicaSnapshotRestore.Status.State)
		}

		if replicaSnapshotRestore.Status.State == apisv1alpha1.OperationStateCompleted {
			completedCount++
		}
	}

	snapshotRestore.Status.Message = message
	if completedCount >= len(snapshotRestore.Status.VolumeReplicaSnapshotRestore) {
		snapshotRestore.Status.State = apisv1alpha1.OperationStateCompleted
	}

	return m.apiClient.Status().Update(context.Background(), snapshotRestore)
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

func (m *manager) getNodeVolumeReplicaSnapshot(volumeSnapshotName string) (map[string]string, error) {
	volumeSnapshot := &apisv1alpha1.LocalVolumeSnapshot{}
	if err := m.apiClient.Get(context.Background(), client.ObjectKey{Name: volumeSnapshotName}, volumeSnapshot); err != nil {
		return nil, err
	}

	nodeReplicaSnapshot := map[string]string{}
	for _, replicaSnapshot := range volumeSnapshot.Status.ReplicaSnapshots {
		volumeReplicaSnapshot := &apisv1alpha1.LocalVolumeReplicaSnapshot{}
		if err := m.apiClient.Get(context.Background(), client.ObjectKey{Name: replicaSnapshot}, volumeReplicaSnapshot); err != nil {
			return nil, err
		}
		nodeReplicaSnapshot[volumeReplicaSnapshot.Spec.NodeName] = volumeReplicaSnapshot.Name
	}

	return nodeReplicaSnapshot, nil
}
