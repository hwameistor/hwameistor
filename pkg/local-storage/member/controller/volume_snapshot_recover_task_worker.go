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

func (m *manager) startVolumeSnapshotRecoverTaskWorker(stopCh <-chan struct{}) {
	m.logger.Debug("Volume Snapshot Recover Worker is working now")
	go func() {
		for {
			task, shutdown := m.volumeSnapshotRecoverTaskQueue.Get()
			if shutdown {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Stop the Volume Snapshot Recover worker")
				break
			}
			if err := m.processVolumeSnapshotRecover(task); err != nil {
				m.logger.WithFields(log.Fields{"task": task, "attempts": m.volumeSnapshotRecoverTaskQueue.NumRequeues(task), "error": err.Error()}).Error("Failed to process Volume Snapshot Recover task, retry later")
				m.volumeSnapshotRecoverTaskQueue.AddRateLimited(task)
			} else {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Completed a Volume Snapshot Recover task.")
				m.volumeSnapshotRecoverTaskQueue.Forget(task)
			}
			m.volumeSnapshotRecoverTaskQueue.Done(task)
		}
	}()

	<-stopCh
	m.volumeSnapshotRecoverTaskQueue.Shutdown()
}

func (m *manager) processVolumeSnapshotRecover(snapRecoverName string) error {
	logCtx := m.logger.WithFields(log.Fields{"VolumeSnapshotRecover": snapRecoverName})
	logCtx.Debug("Working on a VolumeSnapshotRecover task")
	snapshotRecover := &apisv1alpha1.LocalVolumeSnapshotRecover{}
	if err := m.apiClient.Get(context.TODO(), types.NamespacedName{Name: snapRecoverName}, snapshotRecover); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("Failed to get VolumeSnapshotRecover from cache")
			return err
		}

		m.lock.Lock()
		defer m.lock.Unlock()
		delete(m.replicaSnapRecoverRecords, snapRecoverName)

		logCtx.Info("Not found the VolumeSnapshotRecover from cache, should be deleted already")
		return nil
	}

	if snapshotRecover.Spec.Abort &&
		snapshotRecover.Status.State != apisv1alpha1.OperationStateToBeAborted &&
		snapshotRecover.Status.State != apisv1alpha1.OperationStateAborting &&
		snapshotRecover.Status.State != apisv1alpha1.OperationStateAborted &&
		snapshotRecover.Status.State != apisv1alpha1.OperationStateCompleted {

		snapshotRecover.Status.State = apisv1alpha1.OperationStateToBeAborted
		return m.apiClient.Status().Update(context.TODO(), snapshotRecover)
	}

	// log with namespace/name is enough
	logCtx = m.logger.WithFields(log.Fields{"TargetVolume": snapshotRecover.Spec.TargetVolume, "SnapshotRecover": snapshotRecover.Name, "Spec": snapshotRecover.Spec, "Status": snapshotRecover.Status})
	logCtx.Debug("Starting to process a VolumeSnapshotRecover")

	// state chain: (empty) -> Submitted -> Start -> InProgress -> Completed
	switch snapshotRecover.Status.State {
	case "":
		return m.volumeSnapshotRecoverSubmit(snapshotRecover)
	case apisv1alpha1.OperationStateSubmitted:
		return m.volumeSnapshotRecoverStart(snapshotRecover)
	case apisv1alpha1.OperationStateInProgress:
		return m.checkInProgressVolumeSnapshotRecover(snapshotRecover)
	case apisv1alpha1.OperationStateToBeAborted:
		return m.volumeSnapshotRecoverAbort(snapshotRecover)
	case apisv1alpha1.OperationStateCompleted, apisv1alpha1.OperationStateAborted:
		return m.volumeSnapshotRecoverCleanup(snapshotRecover)
	default:
		logCtx.Error("Invalid state/phase")
	}
	return fmt.Errorf("invalid state")
}

func (m *manager) volumeSnapshotRecoverSubmit(snapshotRecover *apisv1alpha1.LocalVolumeSnapshotRecover) error {
	logCtx := m.logger.WithFields(log.Fields{"SnapshotRecover": snapshotRecover.Name, "Spec": snapshotRecover.Spec})
	logCtx.Debug("Submit a VolumeSnapshotRecover")

	snapshotRecover.Status.State = apisv1alpha1.OperationStateSubmitted
	return m.apiClient.Status().Update(context.Background(), snapshotRecover)
}

func (m *manager) volumeSnapshotRecoverStart(snapshotRecover *apisv1alpha1.LocalVolumeSnapshotRecover) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	logCtx := m.logger.WithFields(log.Fields{"SnapshotRecover": snapshotRecover.Name, "Spec": snapshotRecover.Spec})
	logCtx.Debug("Start a VolumeSnapshotRecover")

	if err := m.volumeSnapshotRecoverPreCheck(snapshotRecover); err != nil {
		logCtx.Error("Failed to precheck volume snapshot recover")
		return err
	}

	sourceVolume, err := m.getSourceVolumeFromSnapshot(snapshotRecover.Spec.SourceVolumeSnapshot)
	if err != nil {
		logCtx.Error("Failed to get source volume from snapshot")
		return err
	}

	nodeVolumeReplicaSnapshot, err := m.getNodeVolumeReplicaSnapshot(snapshotRecover.Spec.SourceVolumeSnapshot)
	if err != nil {
		logCtx.Error("Failed to get volume replica snapshot from snapshot")
		return err
	}

	// create LocalVolumeReplicaSnapshotRecover on each node according to the topology of source volume
	var allVolumeReplicaSnapshotRecovers []string
	for _, nodeName := range sourceVolume.Spec.Accessibility.Nodes {
		// 1. check if replica snapshot recover has already created on this node
		if exist, nodeSnapRecover, err := m.isReplicaSnapshotRecoverExistOnNode(nodeName, snapshotRecover.Name); err != nil {
			logCtx.WithError(err).Errorf("Failed to judge if LocalVolumeReplicaSnapshot exist on node %s", nodeName)
			return err
		} else if exist {
			allVolumeReplicaSnapshotRecovers = append(allVolumeReplicaSnapshotRecovers, nodeSnapRecover.Name)
			logCtx.WithField("replicaSnapshotRecover", nodeSnapRecover.Name).Infof("VolumeReplicaSnapshotRecover is already exist on %s", nodeName)
			continue
		}

		// 2. start creating replica snapshot recover
		nodeReplicaSnapshot, ok := nodeVolumeReplicaSnapshot[nodeName]
		if !ok {
			err = fmt.Errorf("LocalVolumeReplicaSnapshot not found on node %s but it is accessible in the source LocalVolume topology", nodeName)
			logCtx.WithError(err).Error("Failed to create LocalVolumeReplicaSnapshot")
			return err
		}

		if m.replicaSnapRecoverRecords[snapshotRecover.Name] == nil {
			m.replicaSnapRecoverRecords[snapshotRecover.Name] = make(map[string]*apisv1alpha1.LocalVolumeReplicaSnapshotRecover)
		}
		replicaSnapshotRecover := &apisv1alpha1.LocalVolumeReplicaSnapshotRecover{
			ObjectMeta: metav1.ObjectMeta{
				Name:            fmt.Sprintf("%s-%s", snapshotRecover.Name, utilrand.String(6)),
				OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(snapshotRecover, snapshotRecover.GroupVersionKind())},
			},
			Spec: apisv1alpha1.LocalVolumeReplicaSnapshotRecoverSpec{
				NodeName:                    nodeName,
				SourceVolumeReplicaSnapshot: nodeReplicaSnapshot,
				VolumeSnapshotRecover:       snapshotRecover.Name,
				RecoverType:                 snapshotRecover.Spec.RecoverType,
				TargetVolume:                snapshotRecover.Spec.TargetVolume,
				TargetPoolName:              snapshotRecover.Spec.TargetPoolName,
				SourceVolumeSnapshot:        snapshotRecover.Spec.SourceVolumeSnapshot,
			},
		}

		if err = m.apiClient.Create(context.Background(), replicaSnapshotRecover); err != nil && !errors.IsAlreadyExists(err) {
			logCtx.WithField("replicaSnapshotRecover", replicaSnapshotRecover.Name).WithError(err).Error("Failed to create VolumeReplicaSnapshotRecover")
			return err
		}

		m.replicaSnapRecoverRecords[snapshotRecover.Name][nodeName] = replicaSnapshotRecover
		allVolumeReplicaSnapshotRecovers = append(allVolumeReplicaSnapshotRecovers, replicaSnapshotRecover.Name)

		logCtx.WithField("replicaSnapshotRecover", replicaSnapshotRecover.Name).WithError(err).Errorf("VolumeReplicaSnapshotRecover is created successfully on %s", nodeName)
	}

	snapshotRecover.Status.VolumeReplicaSnapshotRecover = allVolumeReplicaSnapshotRecovers
	snapshotRecover.Status.State = apisv1alpha1.OperationStateInProgress
	return m.apiClient.Status().Update(context.Background(), snapshotRecover)
}

func (m *manager) volumeSnapshotRecoverPreCheck(snapshotRecover *apisv1alpha1.LocalVolumeSnapshotRecover) error {
	logCtx := m.logger.WithFields(log.Fields{"SnapshotRecover": snapshotRecover.Name, "Spec": snapshotRecover.Spec})
	logCtx.Debug("precheck volumeSnapshotRecover")

	sourceVolume, err := m.getSourceVolumeFromSnapshot(snapshotRecover.Spec.SourceVolumeSnapshot)
	if err != nil {
		logCtx.Error("Failed to get source volume from snapshot")
		return err
	}

	switch snapshotRecover.Spec.RecoverType {
	case apisv1alpha1.RecoverTypeRollback:
		// compare if target volume is set correctly when recover type is rollback
		if (snapshotRecover.Spec.TargetVolume != "" && snapshotRecover.Spec.TargetVolume != sourceVolume.Name) ||
			(snapshotRecover.Spec.TargetPoolName != "" && snapshotRecover.Spec.TargetPoolName != sourceVolume.Spec.PoolName) {
			logCtx.WithFields(log.Fields{"originTargetPoolVolume": snapshotRecover.Spec.TargetPoolName + "/" + snapshotRecover.Spec.TargetVolume,
				"correctTargetPoolVolume": sourceVolume.Spec.PoolName + "/" + sourceVolume.Name}).Info("TargetPoolVolume is wrong, correct it with info in source volume")
		}
		snapshotRecover.Spec.TargetVolume = sourceVolume.Name
		snapshotRecover.Spec.TargetPoolName = sourceVolume.Spec.PoolName
		return m.apiClient.Update(context.Background(), snapshotRecover)

	case apisv1alpha1.RecoverTypeRestore:
		if snapshotRecover.Spec.TargetVolume == "" || snapshotRecover.Spec.TargetPoolName == "" {
			return fmt.Errorf("TargetVolume and TargetPoolName is required")
		}
	default:
		return fmt.Errorf("invalid revover type")
	}

	return nil
}

func (m *manager) checkInProgressVolumeSnapshotRecover(snapshotRecover *apisv1alpha1.LocalVolumeSnapshotRecover) error {
	logCtx := m.logger.WithFields(log.Fields{"SnapshotRecover": snapshotRecover.Name, "Spec": snapshotRecover.Spec})
	logCtx.Debug("Check a InProgress VolumeSnapshotRecover")

	var (
		message        string
		completedCount int
	)

	for _, replicaSnapshotRecoverName := range snapshotRecover.Status.VolumeReplicaSnapshotRecover {
		replicaSnapshotRecover := &apisv1alpha1.LocalVolumeReplicaSnapshotRecover{}
		if err := m.apiClient.Get(context.Background(), client.ObjectKey{Name: replicaSnapshotRecoverName}, replicaSnapshotRecover); err != nil {
			logCtx.WithField("ReplicaSnapshotRecover", replicaSnapshotRecoverName).WithError(err).Error("Failed to get VolumeReplicaSnapshotRecover")
			return err
		}

		if replicaSnapshotRecover.Status.Message != "" {
			message += fmt.Sprintf("%s: %s;", replicaSnapshotRecoverName, replicaSnapshotRecover.Status.Message)
		} else {
			message += fmt.Sprintf("%s is %s;", replicaSnapshotRecoverName, replicaSnapshotRecover.Status.State)
		}

		if replicaSnapshotRecover.Status.State == apisv1alpha1.OperationStateCompleted {
			completedCount++
		}
	}

	snapshotRecover.Status.Message = message
	if completedCount >= len(snapshotRecover.Status.VolumeReplicaSnapshotRecover) {
		if err := m.markTargetVolumeAsCompleted(snapshotRecover.Spec.TargetVolume); err != nil {
			logCtx.WithError(err).Error("Failed to mark target volume as completed")
			return err
		}
		snapshotRecover.Status.State = apisv1alpha1.OperationStateCompleted
	}

	return m.apiClient.Status().Update(context.Background(), snapshotRecover)
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

	anno[apisv1alpha1.VolumeSnapshotRecoverCompletedAnnoKey] = ""
	targetVolume.SetAnnotations(anno)

	return m.apiClient.Update(context.Background(), targetVolume)
}

func (m *manager) volumeSnapshotRecoverAbort(snapshotRecover *apisv1alpha1.LocalVolumeSnapshotRecover) error {
	logCtx := m.logger.WithFields(log.Fields{"SnapshotRecover": snapshotRecover.Name, "Spec": snapshotRecover.Spec})
	logCtx.Debug("Abort a VolumeSnapshotRecover")

	snapshotRecover.Status.State = apisv1alpha1.OperationStateAborted
	return m.apiClient.Status().Update(context.TODO(), snapshotRecover)
}

func (m *manager) volumeSnapshotRecoverCleanup(snapshotRecover *apisv1alpha1.LocalVolumeSnapshotRecover) error {
	logCtx := m.logger.WithFields(log.Fields{"SnapshotRecover": snapshotRecover.Name, "Spec": snapshotRecover.Spec})
	logCtx.Debug("Cleanup a VolumeSnapshotRecover")

	cleanedCount := 0
	for _, replicaSnapshotRecoverName := range snapshotRecover.Status.VolumeReplicaSnapshotRecover {
		replicaSnapshotRecover := &apisv1alpha1.LocalVolumeReplicaSnapshotRecover{}
		if err := m.apiClient.Get(context.Background(), client.ObjectKey{Name: replicaSnapshotRecoverName}, replicaSnapshotRecover); err != nil {
			if errors.IsNotFound(err) {
				cleanedCount++
				logCtx.WithField("ReplicaSnapshotRecover", replicaSnapshotRecoverName).Error("Cleanup VolumeReplicaSnapshotRecover successfully")
				continue
			}
			logCtx.WithField("ReplicaSnapshotRecover", replicaSnapshotRecoverName).WithError(err).Error("Failed to get VolumeReplicaSnapshotRecover")
			return err
		}

		if !replicaSnapshotRecover.Spec.Abort {
			replicaSnapshotRecover.Spec.Abort = true
			if err := m.apiClient.Update(context.Background(), replicaSnapshotRecover); err != nil {
				logCtx.WithField("ReplicaSnapshotRecover", replicaSnapshotRecoverName).Error("Failed to cleanup VolumeReplicaSnapshotRecover")
				return err
			}
			delete(m.replicaSnapRecoverRecords[snapshotRecover.Name], replicaSnapshotRecover.Spec.NodeName)
			logCtx.WithField("ReplicaSnapshotRecover", replicaSnapshotRecoverName).Error("Cleaning VolumeReplicaSnapshotRecover")
		}
	}

	if cleanedCount < len(snapshotRecover.Status.VolumeReplicaSnapshotRecover) {
		err := fmt.Errorf("remaining %d VolumeReplicaSnapshotRecover to clean", len(snapshotRecover.Status.VolumeReplicaSnapshotRecover)-cleanedCount)
		logCtx.WithError(err).Info("VolumeSnapshotRecover is deleting")
		return err
	}

	return m.apiClient.Delete(context.TODO(), snapshotRecover)
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

func (m *manager) isReplicaSnapshotRecoverExistOnNode(nodeName, volumeSnapshotRecoverName string) (bool, *apisv1alpha1.LocalVolumeReplicaSnapshotRecover, error) {
	replicaSnapshotRecover := &apisv1alpha1.LocalVolumeReplicaSnapshotRecoverList{}
	if err := m.apiClient.List(context.Background(), replicaSnapshotRecover); err != nil {
		m.logger.WithError(err).Errorf("failed to list replica snapshots on node %s", nodeName)
		return false, nil, err
	}

	// 1. check apiserver
	for _, replicaRecover := range replicaSnapshotRecover.Items {
		if replicaRecover.Spec.NodeName == nodeName {
			return true, replicaRecover.DeepCopy(), nil
		}
	}

	// 2. check local cache
	if records, ok := m.replicaSnapRecoverRecords[volumeSnapshotRecoverName]; ok {
		if volumeReplicaSnapshotRecover, ok := records[nodeName]; ok {
			return true, volumeReplicaSnapshotRecover, nil
		}
	}

	return false, nil, nil
}
