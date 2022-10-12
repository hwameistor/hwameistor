package controller

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/wxnacy/wgo/arrays"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-storage/utils"
)

const (
	imageRegistry          = "daocloud.io/daocloud"
	rcloneImageVersion     = "v1.1.2"
	rcloneImageName        = imageRegistry + "/" + "hwameistor-migrate-rclone" + ":" + rcloneImageVersion
	rcloneConfigMapName    = "migrate-rclone-config"
	rcloneConfigMapKey     = "rclone.conf"
	rcloneCertKey          = "rclonemerged"
	rcloneKeyConfigMapName = "rclone-key-config"
)

func (m *manager) startVolumeMigrateTaskWorker(stopCh <-chan struct{}) {
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
		return m.apiClient.Status().Update(ctx, migrate)
	}

	lvg := &apisv1alpha1.LocalVolumeGroup{}
	if err := m.apiClient.Get(ctx, types.NamespacedName{Name: vol.Spec.VolumeGroup}, lvg); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("VolumeMigrateSubmit: Failed to query lvg")
		} else {
			logCtx.Info("VolumeMigrateSubmit: Not found the lvg")
		}
		migrate.Status.Message = err.Error()
		return m.apiClient.Status().Update(ctx, migrate)
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
		if vol.Spec.Convertible {
			return m.volumeMigrateSyncReplica(migrate, vol, lvg)
		} else {
			return m.nonConvertibleVolumeMigrateSyncReplica(migrate, vol, lvg)
		}
	case apisv1alpha1.OperationStateMigratePruneReplica:
		return m.volumeMigratePruneReplica(migrate, vol, lvg)
	case apisv1alpha1.OperationStateCompleted:
		return m.volumeMigrateCleanup(migrate)
	case apisv1alpha1.OperationStateToBeAborted:
		return m.volumeMigrateAbort(migrate)
	case apisv1alpha1.OperationStateAborted:
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

	if len(migrate.Status.TargetNodeName) == 0 {
		logCtx.Debug("Selecting the target node for the migration")
		tgtNodeName, err := m.selectMigrateTargetNode(migrate, lvg)
		if err != nil {
			logCtx.WithError(err).Error("No valid target node for the migration")
			migrate.Status.Message = err.Error()
			return m.apiClient.Status().Update(ctx, migrate)
		}
		migrate.Status.TargetNodeName = tgtNodeName
		migrate.Status.Message = "Selected target node"
		return m.apiClient.Status().Update(ctx, migrate)
	}

	volList := []*apisv1alpha1.LocalVolume{vol}
	if migrate.Spec.MigrateAllVols {
		vols, err := m.getAllVolumesInGroup(lvg)
		if err != nil {
			logCtx.WithField("LocalVolumeGroup", lvg.Name).WithError(err).Error("Failed to get all volumes in the group")
			migrate.Status.Message = err.Error()
			return m.apiClient.Status().Update(ctx, migrate)
		}
		volList = vols
	}
	for i := range volList {
		if err := m.checkReplicasForVolume(volList[i]); err != nil {
			logCtx.WithField("LocalVolume", volList[i].Name).WithError(err).Error("Replicas are in problem")
			migrate.Status.Message = err.Error()
			return m.apiClient.Status().Update(ctx, migrate)
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
	if arrays.ContainsString(lvg.Spec.Accessibility.Nodes, migrate.Status.TargetNodeName) == -1 {
		lvg.Spec.Accessibility.Nodes = append(lvg.Spec.Accessibility.Nodes, migrate.Status.TargetNodeName)
		if err := m.apiClient.Update(ctx, lvg); err != nil {
			migrate.Status.Message = fmt.Sprintf("Failed to set the target node: %s", err.Error())
			return m.apiClient.Status().Update(ctx, migrate)
		}
	}

	migrate.Status.State = apisv1alpha1.OperationStateMigrateAddReplica
	return m.apiClient.Status().Update(ctx, migrate)
}

func (m *manager) volumeMigrateAddReplica(migrate *apisv1alpha1.LocalVolumeMigrate, vol *apisv1alpha1.LocalVolume, lvg *apisv1alpha1.LocalVolumeGroup) error {
	logCtx := m.logger.WithFields(log.Fields{"migration": migrate.Name, "spec": migrate.Spec, "status": migrate.Status})
	logCtx.Debug("Start adding replicas")

	ctx := context.TODO()

	volList := []*apisv1alpha1.LocalVolume{vol}
	if migrate.Spec.MigrateAllVols {
		vols, err := m.getAllVolumesInGroup(lvg)
		if err != nil {
			logCtx.WithField("LocalVolumeGroup", lvg.Name).WithError(err).Error("Failed to get all volumes in the group")
			migrate.Status.Message = err.Error()
			return m.apiClient.Status().Update(ctx, migrate)
		}
		volList = vols
	}

	for i := range volList {
		if volList[i].Spec.ReplicaNumber > migrate.Status.OriginalReplicaNumber {
			if err := m.checkReplicasForVolume(volList[i]); err != nil {
				logCtx.WithField("LocalVolume", volList[i].Name).WithError(err).Debug("Volume migration is still in process")
				migrate.Status.Message = fmt.Sprintf("In progress: %s", err.Error())
				return m.apiClient.Status().Update(ctx, migrate)
			}
			logCtx.WithField("LocalVolume", volList[i].Name).Debug("Volume has added a new replica successfully")
			continue
		}
		volList[i].Spec.ReplicaNumber++
		if err := m.apiClient.Update(ctx, volList[i]); err != nil {
			logCtx.WithField("LocalVolume", volList[i].Name).WithError(err).Error("Failed to add a new replica to the volume")
			migrate.Status.Message = fmt.Sprintf("Failed to migrate volume %s", volList[i].Name)
		} else {
			migrate.Status.Message = fmt.Sprintf("Adding a new replica to volume %s", volList[i].Name)
		}
		return m.apiClient.Status().Update(ctx, migrate)
	}

	logCtx.Debug("Successfully added the new replicas")
	migrate.Status.Message = "Successfully added the new replica"
	migrate.Status.State = apisv1alpha1.OperationStateMigrateSyncReplica
	return m.apiClient.Status().Update(ctx, migrate)
}

func (m *manager) volumeMigratePruneReplica(migrate *apisv1alpha1.LocalVolumeMigrate, vol *apisv1alpha1.LocalVolume, lvg *apisv1alpha1.LocalVolumeGroup) error {
	logCtx := m.logger.WithFields(log.Fields{"migration": migrate.Name, "spec": migrate.Spec, "status": migrate.Status})
	logCtx.Debug("Start a volumeMigrateInProgress")

	ctx := context.TODO()

	if arrays.ContainsString(lvg.Spec.Accessibility.Nodes, migrate.Spec.SourceNodesName) != -1 {
		lvg.Spec.Accessibility.Nodes = utils.RemoveStringItem(lvg.Spec.Accessibility.Nodes, migrate.Spec.SourceNodesName)
		return m.apiClient.Update(ctx, lvg)
	}

	volList := []*apisv1alpha1.LocalVolume{vol}
	if migrate.Spec.MigrateAllVols {
		vols, err := m.getAllVolumesInGroup(lvg)
		if err != nil {
			logCtx.WithField("LocalVolumeGroup", lvg.Name).WithError(err).Error("Failed to get all volumes in the group")
			migrate.Status.Message = err.Error()
			return m.apiClient.Status().Update(ctx, migrate)
		}
		volList = vols
	}

	for i := range volList {
		// New replica is added and synced successfully, will remove the to-be-migrated replica from Volume's config
		if volList[i].Spec.ReplicaNumber > migrate.Status.OriginalReplicaNumber {
			// prune the to-be-migrated replica
			volList[i].Spec.ReplicaNumber = migrate.Status.OriginalReplicaNumber
			if err := m.apiClient.Update(ctx, volList[i]); err != nil {
				logCtx.WithField("LocalVolume", volList[i].Name).WithError(err).Error("Failed to prune a replica")
				migrate.Status.Message = fmt.Sprintf("Failed to prune a replica of volume %s", volList[i].Name)
			} else {
				migrate.Status.Message = fmt.Sprintf("Pruning a replica of volume %s", volList[i].Name)
			}
			return m.apiClient.Status().Update(ctx, migrate)
		}
	}

	logCtx.Debug("Successfully prune the replicas to be migrated, and completed the volume migration task")
	migrate.Status.State = apisv1alpha1.OperationStateCompleted
	return m.apiClient.Status().Update(ctx, migrate)

}

func (m *manager) volumeMigrateSyncReplica(migrate *apisv1alpha1.LocalVolumeMigrate, vol *apisv1alpha1.LocalVolume, lvg *apisv1alpha1.LocalVolumeGroup) error {
	logCtx := m.logger.WithFields(log.Fields{"migration": migrate.Name, "spec": migrate.Spec, "status": migrate.Status})
	logCtx.Debug("Start syncing replicas for convertable volumes")

	logCtx.Debug("Ignore the replica syncing for convertable volumes")
	migrate.Status.State = apisv1alpha1.OperationStateMigratePruneReplica
	return m.apiClient.Status().Update(context.TODO(), migrate)

}

func (m *manager) nonConvertibleVolumeMigrateSyncReplica(migrate *apisv1alpha1.LocalVolumeMigrate, vol *apisv1alpha1.LocalVolume, lvg *apisv1alpha1.LocalVolumeGroup) error {
	logCtx := m.logger.WithFields(log.Fields{"migration": migrate.Name, "spec": migrate.Spec, "status": migrate.Status})
	logCtx.Debug("Start syncing replicas for non-convertable volumes")

	ctx := context.TODO()

	// prepare the resources for rclone to execute
	rcl := m.dataCopyManager.UseRclone(rcloneImageName, rcloneConfigMapName, rcloneKeyConfigMapName, m.namespace, rcloneConfigMapKey, rcloneCertKey)
	if err := rcl.EnsureRcloneConfigMapToTargetNamespace(m.namespace); err != nil {
		migrate.Status.Message = err.Error()
		return m.apiClient.Status().Update(ctx, migrate)
	}
	tmpRcloneKeyConfigMap := &corev1.ConfigMap{}
	if err := m.apiClient.Get(ctx, types.NamespacedName{Namespace: m.namespace, Name: rcloneKeyConfigMapName}, tmpRcloneKeyConfigMap); err == nil {
		if delErr := m.apiClient.Delete(ctx, tmpRcloneKeyConfigMap); delErr != nil {
			m.logger.WithError(err).Error("Failed to delete Configmap")
			return delErr
		}
	}
	rcloneKeyCM := rcl.GenerateRcloneKeyConfigMap(m.namespace)
	if err := m.apiClient.Create(ctx, rcloneKeyCM); err != nil {
		if errors.IsAlreadyExists(err) {
			logCtx.WithError(err).Warning("RcloneKeyConfigmap already exists")
		} else {
			logCtx.WithError(err).Error("Failed to create RcloneKeyConfigmap")
			migrate.Status.Message = err.Error()
			return m.apiClient.Status().Update(ctx, migrate)
		}
	}
	if err := m.makeMigrateRcloneConfigmap(migrate.Status.TargetNodeName, migrate.Spec.SourceNodesName, m.namespace, migrate.Spec.VolumeName); err != nil {
		if errors.IsAlreadyExists(err) {
			logCtx.WithError(err).Warning("MigrateRcloneConfigmap already exists")
		} else {
			logCtx.WithError(err).Error("Failed to makeMigrateRcloneConfigmap")
			migrate.Status.Message = err.Error()
			return m.apiClient.Status().Update(ctx, migrate)
		}
	}

	volList := []*apisv1alpha1.LocalVolume{vol}
	if migrate.Spec.MigrateAllVols {
		vols, err := m.getAllVolumesInGroup(lvg)
		if err != nil {
			logCtx.WithField("LocalVolumeGroup", lvg.Name).WithError(err).Error("Failed to get all volumes in the group")
			migrate.Status.Message = err.Error()
			return m.apiClient.Status().Update(ctx, migrate)
		}
		volList = vols
	}

	for i := range volList {
		jobName := migrate.Name + "-job-" + volList[i].Spec.PersistentVolumeClaimName
		syncJob := &batchv1.Job{}
		if err := m.apiClient.Get(ctx, types.NamespacedName{Namespace: m.namespace, Name: jobName}, syncJob); err != nil {
			if errors.IsNotFound(err) {
				logCtx.WithField("Job", jobName).Info("No job is created to sync replicas, create one ...")
				if err := rcl.SrcMountPointToRemoteMountPoint(
					jobName,
					m.namespace,
					volList[i].Spec.PoolName,
					volList[i].Name, migrate.Spec.SourceNodesName,
					migrate.Status.TargetNodeName,
					true,
					0,
				); err != nil {
					logCtx.WithField("LocalVolume", volList[i].Name).WithError(err).Error("Failed to start a job to sync replicas")
					migrate.Status.Message = fmt.Sprintf("Failed to start a job to sync replicas for volume %s", volList[i].Name)
				} else {
					migrate.Status.Message = fmt.Sprintf("Started a job to sync replicas for volume %s", volList[i].Name)
				}
				return m.apiClient.Status().Update(ctx, migrate)
			}
			logCtx.WithError(err).Error("Failed to get MigrateJob from cache")
			migrate.Status.Message = "Failed to get the job from cache"
			return m.apiClient.Status().Update(ctx, migrate)
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
				if err := m.apiClient.Delete(ctx, syncJob); err != nil {
					migrate.Status.Message = "Failed to cleanup the sync job"
					return m.apiClient.Status().Update(ctx, migrate)
				}
				isJobCompleted = true
			}
			break
		}
		if !isJobCompleted {
			migrate.Status.Message = fmt.Sprintf("Waiting for the sync job to complete: %s", syncJob.Name)
			return m.apiClient.Status().Update(ctx, migrate)
		}
	}

	logCtx.Debug("Successfully synchronized the replicas")
	migrate.Status.State = apisv1alpha1.OperationStateMigratePruneReplica
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
	if len(vol.Status.PublishedNodeName) > 0 {
		return fmt.Errorf("volume still in use")
	}

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
		if migrate.Spec.SourceNodesName == validNodes[i].Name {
			// skip the source node
			continue
		}
		if arrays.ContainsString(migrate.Spec.TargetNodesNames, validNodes[i].Name) != -1 {
			// return the target node immediately if it's qualified
			return validNodes[i].Name, nil
		}
		qualifiedNodes = append(qualifiedNodes, validNodes[i].Name)
	}

	if len(qualifiedNodes) == 0 {
		return "", fmt.Errorf("no qualified target node")
	}

	if len(migrate.Spec.TargetNodesNames) == 0 {
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

func (m *manager) makeMigrateRcloneConfigmap(targetNodeName, sourceNodeName, ns, lvname string) error {
	ctx := context.TODO()

	tmpConfigMap := &corev1.ConfigMap{}
	err := m.apiClient.Get(context.TODO(), types.NamespacedName{Namespace: ns, Name: rcloneConfigMapName}, tmpConfigMap)
	if err == nil {
		if delErr := m.apiClient.Delete(context.TODO(), tmpConfigMap); delErr != nil {
			m.logger.WithError(err).Error("Failed to delete Configmap")
			return delErr
		}
	}

	remoteNameData := "[remote]" + "\n"
	sourceNameData := "[source]" + "\n"
	typeData := "type = sftp" + "\n"
	remoteHostData := "host = " + targetNodeName + "\n"
	sourceHostData := "host = " + sourceNodeName + "\n"
	keyFileData := "key_file = /config/rclone/rclonemerged" + "\n"
	shellTypeData := "shell_type = unix" + "\n"
	md5sumCommandData := "md5sum_command = md5sum" + "\n"
	sha1sumCommandData := "sha1sum_command = sha1sum" + "\n"

	data := map[string]string{
		rcloneConfigMapKey: remoteNameData + typeData + remoteHostData + keyFileData + shellTypeData + md5sumCommandData + sha1sumCommandData +
			sourceNameData + typeData + sourceHostData + keyFileData + shellTypeData + md5sumCommandData + sha1sumCommandData,
		"lvname":         lvname,
		"targetNodeName": targetNodeName,
		"sourceNodeName": sourceNodeName,
	}
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rcloneConfigMapName,
			Namespace: ns,
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
