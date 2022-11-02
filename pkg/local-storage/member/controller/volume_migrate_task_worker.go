package controller

import (
	"context"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/wxnacy/wgo/arrays"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-storage/utils"
	"github.com/hwameistor/hwameistor/pkg/local-storage/utils/datacopy"
)

var (
	rcloneImageName = "daocloud.io/daocloud/hwameistor-migrate-rclone:v1.1.2"
)

func (m *manager) startVolumeMigrateTaskWorker(stopCh <-chan struct{}) {
	if value := os.Getenv("MIGRAGE_RCLONE_IMAGE"); len(value) > 0 {
		rcloneImageName = value
	}

	m.logger.Debug("VolumeMigrate Worker is working now")
	go func() {
		for {
			task, shutdown := m.volumeMigrateTaskQueue.Get()
			if shutdown {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Stop the VolumeMigrate worker")
				break
			}
			if err := m.processVolumeMigrate(task); err != nil {
				m.logger.WithFields(log.Fields{"task": task, "attempts": m.volumeMigrateTaskQueue.NumRequeues(task), "error": err.Error()}).Error("Failed to process VolumeMigrate task, retry later")
				m.volumeMigrateTaskQueue.AddRateLimited(task)
			} else {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Completed a VolumeMigrate task.")
				m.volumeMigrateTaskQueue.Forget(task)
			}
			m.volumeMigrateTaskQueue.Done(task)
		}
	}()

	<-stopCh
	m.volumeMigrateTaskQueue.Shutdown()
}

func (m *manager) processVolumeMigrate(vmName string) error {
	logCtx := m.logger.WithFields(log.Fields{"VolumeMigrate": vmName})
	logCtx.Debug("Working on a VolumeMigrate task")

	migrate := &apisv1alpha1.LocalVolumeMigrate{}
	if err := m.apiClient.Get(context.TODO(), types.NamespacedName{Name: vmName}, migrate); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("Failed to get VolumeMigrate from cache")
			return err
		}
		logCtx.Info("Not found the VolumeMigrate from cache, should be deleted already.")
		return nil
	}

	if migrate.Spec.Abort &&
		migrate.Status.State != apisv1alpha1.OperationStateToBeAborted &&
		migrate.Status.State != apisv1alpha1.OperationStateAborting &&
		migrate.Status.State != apisv1alpha1.OperationStateAborted &&
		migrate.Status.State != apisv1alpha1.OperationStateCompleted {

		migrate.Status.State = apisv1alpha1.OperationStateToBeAborted
		return m.apiClient.Status().Update(context.TODO(), migrate)
	}

	ctx := context.TODO()

	vol := &apisv1alpha1.LocalVolume{}
	if err := m.apiClient.Get(ctx, types.NamespacedName{Name: migrate.Spec.VolumeName}, vol); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("Failed to query volume")
		} else {
			logCtx.Info("Not found the volume")
		}
		migrate.Status.Message = err.Error()
		m.apiClient.Status().Update(ctx, migrate)
		return err
	}

	lvg := &apisv1alpha1.LocalVolumeGroup{}
	if err := m.apiClient.Get(ctx, types.NamespacedName{Name: vol.Spec.VolumeGroup}, lvg); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("VolumeMigrateSubmit: Failed to query lvg")
		} else {
			logCtx.Info("VolumeMigrateSubmit: Not found the lvg")
		}
		migrate.Status.Message = err.Error()
		m.apiClient.Status().Update(ctx, migrate)
		return err
	}

	// state chain: (empty) -> Submitted -> Start -> InProgress -> Completed

	logCtx = m.logger.WithFields(log.Fields{"migration": migrate.Name, "spec": migrate.Spec, "status": migrate.Status})
	logCtx.Debug("Starting to process a VolumeMigrate task")
	switch migrate.Status.State {
	case "":
		return m.volumeMigrateSubmit(migrate, vol, lvg)
	case apisv1alpha1.OperationStateSubmitted:
		return m.volumeMigrateStart(migrate, vol, lvg)
	case apisv1alpha1.OperationStateMigrateAddReplica:
		return m.volumeMigrateAddReplica(migrate, vol, lvg)
	case apisv1alpha1.OperationStateMigrateSyncReplica:
		return m.volumeMigrateSyncReplica(migrate, vol, lvg)
	case apisv1alpha1.OperationStateMigratePruneReplica:
		return m.volumeMigratePruneReplica(migrate, vol, lvg)
	case apisv1alpha1.OperationStateToBeAborted:
		return m.volumeMigrateAbort(migrate)
	case apisv1alpha1.OperationStateCompleted, apisv1alpha1.OperationStateAborted:
		return m.volumeMigrateCleanup(migrate)
	default:
		logCtx.Error("Invalid state/phase")
	}
	return fmt.Errorf("invalid state/phase")
}

func (m *manager) volumeMigrateSubmit(migrate *apisv1alpha1.LocalVolumeMigrate, vol *apisv1alpha1.LocalVolume, lvg *apisv1alpha1.LocalVolumeGroup) error {
	logCtx := m.logger.WithFields(log.Fields{"migration": migrate.Name, "spec": migrate.Spec})
	logCtx.Debug("Submit a VolumeMigrate")

	ctx := context.TODO()
	// if LV is still in use, waiting for it to be released
	if vol.Status.PublishedNodeName == migrate.Spec.SourceNode {
		logCtx.WithField("PublishedNode", vol.Status.PublishedNodeName).Warning("LocalVolume is still in use by source node, try it later")
		migrate.Status.Message = "Volume is still in use"
		m.apiClient.Status().Update(ctx, migrate)
		return fmt.Errorf("volume still in use")
	}

	if !vol.Spec.Convertible && vol.Status.PublishedRawBlock {
		logCtx.Warning("Can't migrate the unconvertable raw block volume")
		migrate.Status.Message = "Can't migrate the unconvertable raw block volume"
		m.apiClient.Status().Update(ctx, migrate)
		return fmt.Errorf("can't migrate the unconvertable raw block volume")
	}

	if len(migrate.Status.TargetNode) == 0 {
		logCtx.Debug("Selecting the target node for the migration")
		tgtNodeName, err := m.selectMigrateTargetNode(migrate, lvg)
		if err != nil {
			logCtx.WithError(err).Error("No valid target node for the migration")
			migrate.Status.Message = err.Error()
			m.apiClient.Status().Update(ctx, migrate)
			return err
		}
		migrate.Status.TargetNode = tgtNodeName
		migrate.Status.Message = "Selected target node"
		return m.apiClient.Status().Update(ctx, migrate)
	}

	volList := []*apisv1alpha1.LocalVolume{vol}
	if migrate.Spec.MigrateAllVols {
		vols, err := m.getAllVolumesInGroup(lvg)
		if err != nil {
			logCtx.WithField("LocalVolumeGroup", lvg.Name).WithError(err).Error("Failed to get all volumes in the group")
			migrate.Status.Message = err.Error()
			m.apiClient.Status().Update(ctx, migrate)
			return err
		}
		volList = vols
	}
	for i := range volList {
		if err := m.checkReplicasForVolume(volList[i]); err != nil {
			logCtx.WithField("LocalVolume", volList[i].Name).WithError(err).Error("Replicas are in problem")
			migrate.Status.Message = err.Error()
			m.apiClient.Status().Update(ctx, migrate)
			return err
		}
	}

	migrate.Status.OriginalReplicaNumber = vol.Spec.ReplicaNumber
	migrate.Status.State = apisv1alpha1.OperationStateSubmitted
	return m.apiClient.Status().Update(ctx, migrate)
}

func (m *manager) volumeMigrateStart(migrate *apisv1alpha1.LocalVolumeMigrate, vol *apisv1alpha1.LocalVolume, lvg *apisv1alpha1.LocalVolumeGroup) error {
	logCtx := m.logger.WithFields(log.Fields{"migration": migrate.Name, "spec": migrate.Spec, "status": migrate.Status})
	logCtx.Debug("Start setting target node")

	ctx := context.TODO()
	// set the target node in LocalVolumeGroup, so that the replica will be allocated base on it
	if arrays.ContainsString(lvg.Spec.Accessibility.Nodes, migrate.Status.TargetNode) == -1 {
		lvg.Spec.Accessibility.Nodes = append(lvg.Spec.Accessibility.Nodes, migrate.Status.TargetNode)
		if err := m.apiClient.Update(ctx, lvg); err != nil {
			migrate.Status.Message = fmt.Sprintf("Failed to set the target node: %s", err.Error())
			m.apiClient.Status().Update(ctx, migrate)
			return err
		}
	}

	migrate.Status.State = apisv1alpha1.OperationStateMigrateAddReplica
	return m.apiClient.Status().Update(ctx, migrate)
}

func (m *manager) volumeMigrateAddReplica(migrate *apisv1alpha1.LocalVolumeMigrate, vol *apisv1alpha1.LocalVolume, lvg *apisv1alpha1.LocalVolumeGroup) error {
	logCtx := m.logger.WithFields(log.Fields{"migration": migrate.Name, "spec": migrate.Spec, "status": migrate.Status})
	logCtx.Debug("Start adding replicas")

	ctx := context.TODO()

	lsNode := &apisv1alpha1.LocalStorageNode{}
	if err := m.apiClient.Get(ctx, types.NamespacedName{Name: migrate.Status.TargetNode}, lsNode); err != nil {
		logCtx.WithError(err).Error("Failed to fetch LocalStorageNode")
		migrate.Status.Message = "Failed to get LocalStorageNode"
		m.apiClient.Status().Update(ctx, migrate)
		return err
	}

	volList := []*apisv1alpha1.LocalVolume{vol}
	if migrate.Spec.MigrateAllVols {
		vols, err := m.getAllVolumesInGroup(lvg)
		if err != nil {
			logCtx.WithField("LocalVolumeGroup", lvg.Name).WithError(err).Error("Failed to get all volumes in the group")
			migrate.Status.Message = err.Error()
			m.apiClient.Status().Update(ctx, migrate)
			return err
		}
		volList = vols
	}

	for i := range volList {
		if volList[i].Spec.ReplicaNumber > migrate.Status.OriginalReplicaNumber {
			if err := m.checkReplicasForVolume(volList[i]); err != nil {
				logCtx.WithField("LocalVolume", volList[i].Name).WithError(err).Debug("Volume migration is still in process")
				migrate.Status.Message = fmt.Sprintf("In progress: %s", err.Error())
				m.apiClient.Status().Update(ctx, migrate)
				return err
			}
			logCtx.WithField("LocalVolume", volList[i].Name).Debug("Volume has added a new replica successfully")
			continue
		}
		volList[i].Spec.ReplicaNumber++
		conf, err := m.volumeScheduler.ConfigureVolumeOnAdditionalNodes(volList[i], []*apisv1alpha1.LocalStorageNode{lsNode})
		if err != nil {
			logCtx.WithField("volume", volList[i].Name).WithError(err).Error("Failed to configure LocalVolume")
			migrate.Status.Message = fmt.Sprintf("Failed to configure LocalVolume %s", volList[i].Name)
			m.apiClient.Status().Update(ctx, migrate)
			return err
		}
		volList[i].Spec.Config = conf
		if err := m.apiClient.Update(ctx, volList[i]); err != nil {
			logCtx.WithField("LocalVolume", volList[i].Name).WithError(err).Error("Failed to add a new replica to the volume")
			migrate.Status.Message = fmt.Sprintf("Failed to migrate volume %s", volList[i].Name)
			m.apiClient.Status().Update(ctx, migrate)
			return fmt.Errorf("failed to migrate volume %s", volList[i].Name)
		}
		migrate.Status.Message = fmt.Sprintf("Adding a new replica to volume %s", volList[i].Name)
		m.apiClient.Status().Update(ctx, migrate)
		return fmt.Errorf("still in progress")
	}

	logCtx.Debug("Successfully added the new replicas")
	migrate.Status.Message = "Successfully added the new replica"
	migrate.Status.State = apisv1alpha1.OperationStateMigrateSyncReplica
	return m.apiClient.Status().Update(ctx, migrate)
}

func (m *manager) volumeMigrateSyncReplica(migrate *apisv1alpha1.LocalVolumeMigrate, vol *apisv1alpha1.LocalVolume, lvg *apisv1alpha1.LocalVolumeGroup) error {
	logCtx := m.logger.WithFields(log.Fields{"migration": migrate.Name, "spec": migrate.Spec, "status": migrate.Status})
	logCtx.Debug("Start syncing replicas")

	ctx := context.TODO()

	volList := []*apisv1alpha1.LocalVolume{vol}
	if migrate.Spec.MigrateAllVols {
		vols, err := m.getAllVolumesInGroup(lvg)
		if err != nil {
			logCtx.WithField("LocalVolumeGroup", lvg.Name).WithError(err).Error("Failed to get all volumes in the group")
			migrate.Status.Message = err.Error()
			m.apiClient.Status().Update(ctx, migrate)
			return err
		}
		volList = vols
	}

	for i := range volList {
		if volList[i].Spec.Convertible {
			continue
		}
		// non-convertible volume
		if err := m.syncReplicaByRclone(migrate, volList[i]); err != nil {
			logCtx.WithField("LocalVolume", volList[i].Name).WithError(err).Error("Failed to synchronize replicas")
			migrate.Status.Message = err.Error()
			m.apiClient.Status().Update(ctx, migrate)
			return err
		}
	}

	migrate.Status.State = apisv1alpha1.OperationStateMigratePruneReplica
	return m.apiClient.Status().Update(ctx, migrate)

}

func (m *manager) syncReplicaByRclone(migrate *apisv1alpha1.LocalVolumeMigrate, vol *apisv1alpha1.LocalVolume) error {
	logCtx := m.logger.WithFields(log.Fields{"migration": migrate.Name, "volume": vol.Name})
	logCtx.Debug("Preparing the resources for rclone to execute")

	rcl, err := m.prepareForRClone(migrate, vol)
	if err != nil {
		return err
	}

	ctx := context.TODO()

	cmName := datacopy.GetConfigMapName(datacopy.RCloneConfigMapName, vol.Name)
	cm := &corev1.ConfigMap{}
	if err := m.apiClient.Get(context.TODO(), types.NamespacedName{Namespace: m.namespace, Name: cmName}, cm); err != nil {
		logCtx.WithField("configmap", cmName).Error("Not found the rclone configmap")
		return err
	}

	if ready := cm.Data[datacopy.RCloneConfigSourceNodeReadyKey]; ready != datacopy.RCloneTrue {
		logCtx.WithField(datacopy.RCloneConfigSourceNodeReadyKey, ready).Debug("Waiting for source mountpoint to be ready ...")
		return fmt.Errorf("source mountpoint is not ready")
	}
	if ready := cm.Data[datacopy.RCloneConfigRemoteNodeReadyKey]; ready != datacopy.RCloneTrue {
		logCtx.WithField(datacopy.RCloneConfigRemoteNodeReadyKey, ready).Debug("Waiting for remote mountpoint to be ready ...")
		return fmt.Errorf("remote mountpoint is not ready")
	}

	jobName := generateJobName(migrate.Name, vol.Spec.PersistentVolumeClaimName)
	syncJob := &batchv1.Job{}
	if err := m.apiClient.Get(ctx, types.NamespacedName{Namespace: m.namespace, Name: jobName}, syncJob); err != nil {
		if errors.IsNotFound(err) {
			logCtx.WithField("Job", jobName).Info("No job is created to sync replicas, create one ...")
			if err := rcl.StartRCloneJob(jobName, vol.Name, migrate.Spec.SourceNode, true, 0); err != nil {
				logCtx.WithField("LocalVolume", vol.Name).WithError(err).Error("Failed to start a job to sync replicas")
				return fmt.Errorf("failed to start a job to sync replicas for volume %s", vol.Name)
			}
			return fmt.Errorf("syncing replica still in progress")
		}
		logCtx.WithError(err).Error("Failed to get MigrateJob from cache")
		return err
	}

	// found the job, check the status
	isJobCompleted := false
	for _, cond := range syncJob.Status.Conditions {
		if cond.Type == batchv1.JobComplete && syncJob.Status.CompletionTime != nil && syncJob.Status.StartTime != nil {
			logCtx.WithFields(log.Fields{
				"Job":          syncJob.Name,
				"Namespace":    syncJob.Namespace,
				"StartTime":    syncJob.Status.StartTime.String(),
				"CompleteTime": syncJob.Status.CompletionTime.String(),
			}).Debug("The replicas have already been synchronized successfully")

			cm.Data[datacopy.RCloneConfigSyncDoneKey] = datacopy.RCloneTrue
			if err := m.apiClient.Update(ctx, cm, &client.UpdateOptions{Raw: &metav1.UpdateOptions{}}); err != nil {
				logCtx.WithField("configmap", cmName).WithError(err).Error("Failed to update rclone configmap")
				return err
			}
			// remove the finalizer will release the job
			syncJob.Finalizers = []string{}
			if err := m.apiClient.Update(ctx, syncJob); err != nil {
				logCtx.WithField("Job", syncJob).WithError(err).Error("Failed to remove finalizer")
				return err
			}
			if err := m.apiClient.Get(ctx, types.NamespacedName{Namespace: m.namespace, Name: jobName}, syncJob); err != nil {
				if !errors.IsNotFound(err) {
					logCtx.WithField("Job", syncJob).WithError(err).Error("Failed to fetch the job")
					return err
				}
			} else {
				if err := m.apiClient.Delete(ctx, syncJob); err != nil {
					logCtx.WithField("Job", syncJob).WithError(err).Error("Failed to cleanup the job")
					return err
				}
			}
			if err := m.apiClient.Delete(ctx, cm); err != nil {
				logCtx.WithField("configmap", cm.Name).WithError(err).Warning("Failed to cleanup the rclone configmap, just leak it")
			}
			isJobCompleted = true
			break
		}
	}
	if !isJobCompleted {
		return fmt.Errorf("waiting for the sync job to complete: %s", syncJob.Name)
	}

	logCtx.Debug("RClone has already been executed successfully")
	return nil
}

func (m *manager) prepareForRClone(migrate *apisv1alpha1.LocalVolumeMigrate, vol *apisv1alpha1.LocalVolume) (*datacopy.Rclone, error) {
	logCtx := m.logger.WithFields(log.Fields{"migration": migrate.Name, "volume": vol.Name})
	logCtx.Debug("Preparing the resources for rclone to execute")

	rcl := m.dataCopyManager.UseRclone(rcloneImageName, m.namespace)
	if err := m.makeSSHKeysForRClone(rcl); err != nil {
		logCtx.WithError(err).Error("Failed to create ssh keys for rclone")
		return nil, err
	}

	// Prepare the rclone's configuration, which should be created unique for each volume data copy
	if err := m.makeConfigForRClone(migrate.Status.TargetNode, migrate.Spec.SourceNode, vol.Name); err != nil {
		logCtx.WithError(err).Error("Failed to create rclone's config")
		return nil, err
	}

	logCtx.Debug("RClone is ready to execute")

	return rcl, nil
}

func (m *manager) volumeMigratePruneReplica(migrate *apisv1alpha1.LocalVolumeMigrate, vol *apisv1alpha1.LocalVolume, lvg *apisv1alpha1.LocalVolumeGroup) error {
	logCtx := m.logger.WithFields(log.Fields{"migration": migrate.Name, "spec": migrate.Spec, "status": migrate.Status})
	logCtx.Debug("Start pruning the replicas")

	ctx := context.TODO()

	if arrays.ContainsString(lvg.Spec.Accessibility.Nodes, migrate.Spec.SourceNode) != -1 {
		lvg.Spec.Accessibility.Nodes = utils.RemoveStringItem(lvg.Spec.Accessibility.Nodes, migrate.Spec.SourceNode)
		if err := m.apiClient.Update(ctx, lvg); err != nil {
			logCtx.WithField("LocalVolumeGroup", lvg.Name).WithError(err).Error("Failed to update LocalVolumeGroup")
			return err
		}
	}

	volList := []*apisv1alpha1.LocalVolume{vol}
	if migrate.Spec.MigrateAllVols {
		vols, err := m.getAllVolumesInGroup(lvg)
		if err != nil {
			logCtx.WithField("LocalVolumeGroup", lvg.Name).WithError(err).Error("Failed to get all volumes in the group")
			migrate.Status.Message = err.Error()
			m.apiClient.Status().Update(ctx, migrate)
			return err
		}
		volList = vols
	}

	for i := range volList {
		// New replica is added and synced successfully, will remove the to-be-migrated replica from Volume's config
		if volList[i].Spec.ReplicaNumber > migrate.Status.OriginalReplicaNumber {
			// prune the to-be-migrated replica
			replicas := []apisv1alpha1.VolumeReplica{}
			for j := range volList[i].Spec.Config.Replicas {
				if volList[i].Spec.Config.Replicas[j].Hostname != migrate.Spec.SourceNode {
					replicas = append(replicas, volList[i].Spec.Config.Replicas[j])
				}
			}
			volList[i].Spec.Config.Replicas = replicas
			volList[i].Spec.ReplicaNumber = migrate.Status.OriginalReplicaNumber
			if err := m.apiClient.Update(ctx, volList[i]); err != nil {
				logCtx.WithField("LocalVolume", volList[i].Name).WithError(err).Error("Failed to prune a replica")
				migrate.Status.Message = fmt.Sprintf("Failed to prune a replica of volume %s", volList[i].Name)
				m.apiClient.Status().Update(ctx, migrate)
				return fmt.Errorf("failed to prune a replica of volume %s", volList[i].Name)
			}
			migrate.Status.Message = fmt.Sprintf("Pruning a replica of volume %s", volList[i].Name)
			m.apiClient.Status().Update(ctx, migrate)
			return fmt.Errorf("pruning replicas in progress")
		}
		if err := m.checkReplicasForVolume(volList[i]); err != nil {
			logCtx.WithField("LocalVolume", volList[i].Name).WithError(err).Debug("Still pruning the replica")
			return err
		}
	}

	logCtx.Debug("Successfully prune the replicas to be migrated, and completed the volume migration task")
	migrate.Status.State = apisv1alpha1.OperationStateCompleted
	return m.apiClient.Status().Update(ctx, migrate)

}

func (m *manager) volumeMigrateAbort(migrate *apisv1alpha1.LocalVolumeMigrate) error {
	logCtx := m.logger.WithFields(log.Fields{"migration": migrate.Name, "spec": migrate.Spec, "status": migrate.Status})
	logCtx.Debug("Abort a VolumeMigrate")

	migrate.Status.State = apisv1alpha1.OperationStateAborted
	return m.apiClient.Status().Update(context.TODO(), migrate)
}

func (m *manager) volumeMigrateCleanup(migrate *apisv1alpha1.LocalVolumeMigrate) error {
	logCtx := m.logger.WithFields(log.Fields{"migration": migrate.Name, "spec": migrate.Spec, "status": migrate.Status})
	logCtx.Debug("Cleanup a VolumeMigrate")

	return m.apiClient.Delete(context.TODO(), migrate)
}

func (m *manager) checkReplicasForVolume(vol *apisv1alpha1.LocalVolume) error {
	if vol.Spec.Config == nil {
		return fmt.Errorf("invalid volume configuration")
	}
	// if len(vol.Status.PublishedNodeName) > 0 {
	// 	return fmt.Errorf("volume still in use")
	// }

	replicas, err := m.getReplicasForVolume(vol.Name)
	if err != nil {
		return err
	}
	if len(replicas) != int(vol.Spec.ReplicaNumber) {
		return fmt.Errorf("inconsistent replicas")
	}
	for _, replica := range replicas {
		if replica.Status.State != apisv1alpha1.VolumeReplicaStateReady {
			return fmt.Errorf("replica %s not ready", replica.Name)
		}
	}

	return nil
}

func (m *manager) selectMigrateTargetNode(migrate *apisv1alpha1.LocalVolumeMigrate, lvg *apisv1alpha1.LocalVolumeGroup) (string, error) {

	vols, err := m.getAllVolumesInGroup(lvg)
	if err != nil {
		return "", fmt.Errorf("failed to get all the volumes in the group")
	}

	validNodes := m.VolumeScheduler().GetNodeCandidates(vols)
	if len(validNodes) == 0 {
		return "", fmt.Errorf("no valid target node")
	}
	qualifiedNodes := []string{}
	for i := range validNodes {
		if migrate.Spec.SourceNode == validNodes[i].Name {
			// skip the source node
			continue
		}
		if arrays.ContainsString(migrate.Spec.TargetNodesSuggested, validNodes[i].Name) != -1 {
			// return the target node immediately if it's qualified
			return validNodes[i].Name, nil
		}
		qualifiedNodes = append(qualifiedNodes, validNodes[i].Name)
	}

	if len(qualifiedNodes) == 0 {
		return "", fmt.Errorf("no qualified target node")
	}

	if len(migrate.Spec.TargetNodesSuggested) == 0 {
		return qualifiedNodes[0], nil
	}
	return "", fmt.Errorf("invalid target node")
}

func (m *manager) getAllVolumesInGroup(lvg *apisv1alpha1.LocalVolumeGroup) ([]*apisv1alpha1.LocalVolume, error) {
	vols := []*apisv1alpha1.LocalVolume{}
	for _, v := range lvg.Spec.Volumes {
		vol := &apisv1alpha1.LocalVolume{}
		if err := m.apiClient.Get(context.TODO(), types.NamespacedName{Name: v.LocalVolumeName}, vol); err != nil {
			return vols, err
		}
		vols = append(vols, vol)
	}
	return vols, nil
}

func (m *manager) makeSSHKeysForRClone(rcl *datacopy.Rclone) error {
	ctx := context.TODO()
	// Prepare the public/private ssh keys for rclone to execute. The keys should be shared by all the rclone executions.
	// Don't update once it exists
	cm := &corev1.ConfigMap{}
	if err := m.apiClient.Get(ctx, types.NamespacedName{Namespace: m.namespace, Name: datacopy.RCloneKeyConfigMapName}, cm); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		cm = rcl.GenerateRcloneKeyConfigMap()
		if err := m.apiClient.Create(ctx, cm); err != nil {
			return err
		}
	}
	return nil
}

func (m *manager) makeConfigForRClone(targetNodeName, sourceNodeName, lvName string) error {
	ctx := context.TODO()

	cmName := datacopy.GetConfigMapName(datacopy.RCloneConfigMapName, lvName)

	cm := &corev1.ConfigMap{}
	if err := m.apiClient.Get(context.TODO(), types.NamespacedName{Namespace: m.namespace, Name: cmName}, cm); err == nil {
		m.logger.WithField("configmap", cmName).Debug("The config of rclone already exists")
		return nil
	}

	remoteNameData := "[remote]" + "\n"
	sourceNameData := "[source]" + "\n"
	typeData := "type = sftp" + "\n"
	remoteHostData := "host = " + targetNodeName + "\n"
	sourceHostData := "host = " + sourceNodeName + "\n"
	keyFileData := "key_file = /config/rclone/" + datacopy.RCloneCertKey + "\n"
	shellTypeData := "shell_type = unix" + "\n"
	md5sumCommandData := "md5sum_command = md5sum" + "\n"
	sha1sumCommandData := "sha1sum_command = sha1sum" + "\n"

	remoteConfig := remoteNameData + typeData + remoteHostData + keyFileData + shellTypeData + md5sumCommandData + sha1sumCommandData
	sourceConfig := sourceNameData + typeData + sourceHostData + keyFileData + shellTypeData + md5sumCommandData + sha1sumCommandData
	data := map[string]string{
		datacopy.RCloneConfigMapKey:         remoteConfig + sourceConfig,
		datacopy.RCloneConfigVolumeNameKey:  lvName,
		datacopy.RCloneConfigDstNodeNameKey: targetNodeName,
		datacopy.RCloneConfigSrcNodeNameKey: sourceNodeName,
	}
	cm = &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: m.namespace,
			Labels:    map[string]string{},
		},
		Data: data,
	}

	if err := m.apiClient.Create(ctx, cm); err != nil {
		m.logger.WithError(err).Error("Failed to create MigrateConfigmap")
		return err
	}

	return nil
}

func generateJobName(mName string, pvcName string) string {
	if len(mName) > 25 {
		mName = mName[:25]
	}
	if len(pvcName) > 25 {
		pvcName = pvcName[:25]
	}
	return fmt.Sprintf("%s-datacopy-%s", mName, pvcName)
}

func (m *manager) rclonePodGC(pod *corev1.Pod) error {
	if pod.Namespace == m.namespace && pod.Labels["app"] == datacopy.RcloneJobLabelApp && len(pod.OwnerReferences) == 0 && pod.Status.Phase == corev1.PodSucceeded {
		return m.apiClient.Delete(context.TODO(), pod)
	}
	return nil
}
