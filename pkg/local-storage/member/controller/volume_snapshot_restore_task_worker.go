package controller

import (
	"context"
	"fmt"
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
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
			if err := m.processVolumeSnapshotRestore(task); err != nil {
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

func (m *manager) processVolumeSnapshotRestore(snapRestoreName string) error {
	logCtx := m.logger.WithFields(log.Fields{"VolumeSnapshotRestore": snapRestoreName})
	logCtx.Debug("Working on a VolumeSnapshotRestore task")
	snapshotRestore := &apisv1alpha1.LocalVolumeSnapshotRestore{}
	if err := m.apiClient.Get(context.TODO(), types.NamespacedName{Name: snapRestoreName}, snapshotRestore); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("Failed to get VolumeSnapshotRestore from cache")
			return err
		}

		m.lock.Lock()
		defer m.lock.Unlock()
		delete(m.replicaSnapRestoreRecords, snapRestoreName)

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
	m.lock.Lock()
	defer m.lock.Unlock()

	logCtx := m.logger.WithFields(log.Fields{"SnapshotRestore": snapshotRestore.Name, "Spec": snapshotRestore.Spec})
	logCtx.Debug("Start a VolumeSnapshotRestore")

	if err := m.volumeSnapshotRestorePreCheck(snapshotRestore); err != nil {
		logCtx.Error("Failed to precheck volume snapshot restore")
		return err
	}

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
	var allVolumeReplicaSnapshotRestores []string
	for _, nodeName := range sourceVolume.Spec.Accessibility.Nodes {
		// 1. check if replica snapshot restore has already created on this node
		if exist, nodeSnapRestore, err := m.isReplicaSnapshotRestoreExistOnNode(nodeName, snapshotRestore.Name); err != nil {
			logCtx.WithError(err).Errorf("Failed to judge if LocalVolumeReplicaSnapshot exist on node %s", nodeName)
			return err
		} else if exist {
			allVolumeReplicaSnapshotRestores = append(allVolumeReplicaSnapshotRestores, nodeSnapRestore.Name)
			logCtx.WithField("replicaSnapshotRestore", nodeSnapRestore.Name).Infof("VolumeReplicaSnapshotRestore is already exist on %s", nodeName)
			continue
		}

		// 2. start creating replica snapshot restore
		nodeReplicaSnapshot, ok := nodeVolumeReplicaSnapshot[nodeName]
		if !ok {
			err = fmt.Errorf("LocalVolumeReplicaSnapshot not found on node %s but it is accessible in the source LocalVolume topology", nodeName)
			logCtx.WithError(err).Error("Failed to create LocalVolumeReplicaSnapshot")
			return err
		}

		if m.replicaSnapRestoreRecords[snapshotRestore.Name] == nil {
			m.replicaSnapRestoreRecords[snapshotRestore.Name] = make(map[string]*apisv1alpha1.LocalVolumeReplicaSnapshotRestore)
		}
		replicaSnapshotRestore := &apisv1alpha1.LocalVolumeReplicaSnapshotRestore{
			ObjectMeta: metav1.ObjectMeta{
				Name:            fmt.Sprintf("%s-%s", snapshotRestore.Name, utilrand.String(6)),
				OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(snapshotRestore, snapshotRestore.GroupVersionKind())},
			},
			Spec: apisv1alpha1.LocalVolumeReplicaSnapshotRestoreSpec{
				NodeName:                    nodeName,
				SourceVolumeReplicaSnapshot: nodeReplicaSnapshot,
				VolumeSnapshotRestore:       snapshotRestore.Name,
				RestoreType:                 snapshotRestore.Spec.RestoreType,
				TargetVolume:                snapshotRestore.Spec.TargetVolume,
				TargetPoolName:              snapshotRestore.Spec.TargetPoolName,
				SourceVolumeSnapshot:        snapshotRestore.Spec.SourceVolumeSnapshot,
			},
		}

		if err = m.apiClient.Create(context.Background(), replicaSnapshotRestore); err != nil && !errors.IsAlreadyExists(err) {
			logCtx.WithField("replicaSnapshotRestore", replicaSnapshotRestore.Name).WithError(err).Error("Failed to create VolumeReplicaSnapshotRestore")
			return err
		}

		m.replicaSnapRestoreRecords[snapshotRestore.Name][nodeName] = replicaSnapshotRestore
		allVolumeReplicaSnapshotRestores = append(allVolumeReplicaSnapshotRestores, replicaSnapshotRestore.Name)

		logCtx.WithField("replicaSnapshotRestore", replicaSnapshotRestore.Name).WithError(err).Errorf("VolumeReplicaSnapshotRestore is created successfully on %s", nodeName)
	}

	snapshotRestore.Status.VolumeReplicaSnapshotRestore = allVolumeReplicaSnapshotRestores
	snapshotRestore.Status.State = apisv1alpha1.OperationStateInProgress
	return m.apiClient.Status().Update(context.Background(), snapshotRestore)
}

func (m *manager) volumeSnapshotRestorePreCheck(snapshotRestore *apisv1alpha1.LocalVolumeSnapshotRestore) error {
	logCtx := m.logger.WithFields(log.Fields{"SnapshotRestore": snapshotRestore.Name, "Spec": snapshotRestore.Spec})
	logCtx.Debug("precheck volumeSnapshotRestore")

	sourceVolume, err := m.getSourceVolumeFromSnapshot(snapshotRestore.Spec.SourceVolumeSnapshot)
	if err != nil {
		logCtx.Error("Failed to get source volume from snapshot")
		return err
	}

	switch snapshotRestore.Spec.RestoreType {
	case apisv1alpha1.RestoreTypeRollback:
		// compare if target volume is set correctly when restore type is rollback
		if (snapshotRestore.Spec.TargetVolume != "" && snapshotRestore.Spec.TargetVolume != sourceVolume.Name) ||
			(snapshotRestore.Spec.TargetPoolName != "" && snapshotRestore.Spec.TargetPoolName != sourceVolume.Spec.PoolName) {
			logCtx.WithFields(log.Fields{"originTargetPoolVolume": snapshotRestore.Spec.TargetPoolName + "/" + snapshotRestore.Spec.TargetVolume,
				"correctTargetPoolVolume": sourceVolume.Spec.PoolName + "/" + sourceVolume.Name}).Info("TargetPoolVolume is wrong, correct it with info in source volume")
		}
		snapshotRestore.Spec.TargetVolume = sourceVolume.Name
		snapshotRestore.Spec.TargetPoolName = sourceVolume.Spec.PoolName
		return m.apiClient.Update(context.Background(), snapshotRestore)

	case apisv1alpha1.RestoreTypeCreate:
		if snapshotRestore.Spec.TargetVolume == "" || snapshotRestore.Spec.TargetPoolName == "" {
			return fmt.Errorf("TargetVolume and TargetPoolName is required")
		}
	default:
		return fmt.Errorf("invalid restore type")
	}

	return nil
}

func (m *manager) checkInProgressVolumeSnapshotRestore(snapshotRestore *apisv1alpha1.LocalVolumeSnapshotRestore) error {
	logCtx := m.logger.WithFields(log.Fields{"SnapshotRestore": snapshotRestore.Name, "Spec": snapshotRestore.Spec})
	logCtx.Debug("Check a InProgress VolumeSnapshotRestore")

	var (
		message        string
		completedCount int
	)

	for _, replicaSnapshotRestoreName := range snapshotRestore.Status.VolumeReplicaSnapshotRestore {
		replicaSnapshotRestore := &apisv1alpha1.LocalVolumeReplicaSnapshotRestore{}
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
		if err := m.markTargetVolumeAsCompleted(snapshotRestore.Spec.TargetVolume); err != nil {
			logCtx.WithError(err).Error("Failed to mark target volume as completed")
			return err
		}
		snapshotRestore.Status.State = apisv1alpha1.OperationStateCompleted
	}

	return m.apiClient.Status().Update(context.Background(), snapshotRestore)
}

func (m *manager) markTargetVolumeAsCompleted(targetVolumeName string) error {
	targetVolume := &apisv1alpha1.LocalVolume{}
	if err := m.apiClient.Get(context.Background(), client.ObjectKey{Name: targetVolumeName}, targetVolume); err != nil {
		return err
	}

	var anno map[string]string
	if anno = targetVolume.GetAnnotations(); anno == nil {
		anno = make(map[string]string)
	}

	anno[apisv1alpha1.VolumeSnapshotRestoreCompletedAnnoKey] = ""
	targetVolume.SetAnnotations(anno)

	return m.apiClient.Update(context.Background(), targetVolume)
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
		replicaSnapshotRestore := &apisv1alpha1.LocalVolumeReplicaSnapshotRestore{}
		if err := m.apiClient.Get(context.Background(), client.ObjectKey{Name: replicaSnapshotRestoreName}, replicaSnapshotRestore); err != nil {
			if errors.IsNotFound(err) {
				cleanedCount++
				logCtx.WithField("ReplicaSnapshotRestore", replicaSnapshotRestoreName).Error("Cleanup VolumeReplicaSnapshotRestore successfully")
				continue
			}
			logCtx.WithField("ReplicaSnapshotRestore", replicaSnapshotRestoreName).WithError(err).Error("Failed to get VolumeReplicaSnapshotRestore")
			return err
		}

		if !replicaSnapshotRestore.Spec.Abort {
			replicaSnapshotRestore.Spec.Abort = true
			if err := m.apiClient.Update(context.Background(), replicaSnapshotRestore); err != nil {
				logCtx.WithField("ReplicaSnapshotRestore", replicaSnapshotRestoreName).Error("Failed to cleanup VolumeReplicaSnapshotRestore")
				return err
			}
			delete(m.replicaSnapRestoreRecords[snapshotRestore.Name], replicaSnapshotRestore.Spec.NodeName)
			logCtx.WithField("ReplicaSnapshotRestore", replicaSnapshotRestoreName).Error("Cleaning VolumeReplicaSnapshotRestore")
		}
	}

	if cleanedCount < len(snapshotRestore.Status.VolumeReplicaSnapshotRestore) {
		err := fmt.Errorf("remaining %d VolumeReplicaSnapshotRestore to clean", len(snapshotRestore.Status.VolumeReplicaSnapshotRestore)-cleanedCount)
		logCtx.WithError(err).Info("VolumeSnapshotRestore is deleting")
		return err
	}

	return m.apiClient.Delete(context.TODO(), snapshotRestore)
}

func (m *manager) getSourceVolumeFromSnapshot(volumeSnapshotName string) (*apisv1alpha1.LocalVolume, error) {
	volumeSnapshot := &apisv1alpha1.LocalVolumeSnapshot{}
	if err := m.apiClient.Get(context.Background(), client.ObjectKey{Name: volumeSnapshotName}, volumeSnapshot); err != nil {
		return nil, err
	}

	sourceVolume := &apisv1alpha1.LocalVolume{}
	if err := m.apiClient.Get(context.Background(), client.ObjectKey{Name: volumeSnapshot.Spec.SourceVolume}, sourceVolume); err != nil {
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

func (m *manager) isReplicaSnapshotRestoreExistOnNode(nodeName, volumeSnapshotRestoreName string) (bool, *apisv1alpha1.LocalVolumeReplicaSnapshotRestore, error) {
	replicaSnapshotRestore := &apisv1alpha1.LocalVolumeReplicaSnapshotRestoreList{}
	if err := m.apiClient.List(context.Background(), replicaSnapshotRestore); err != nil {
		m.logger.WithError(err).Errorf("failed to list replica snapshots on node %s", nodeName)
		return false, nil, err
	}

	// 1. check apiserver
	for _, replicaRestore := range replicaSnapshotRestore.Items {
		if replicaRestore.Spec.NodeName == nodeName && replicaRestore.Spec.VolumeSnapshotRestore == volumeSnapshotRestoreName {
			return true, replicaRestore.DeepCopy(), nil
		}
	}

	// 2. check local cache
	if records, ok := m.replicaSnapRestoreRecords[volumeSnapshotRestoreName]; ok {
		if volumeReplicaSnapshotRestore, ok := records[nodeName]; ok {
			return true, volumeReplicaSnapshotRestore, nil
		}
	}

	return false, nil, nil
}
