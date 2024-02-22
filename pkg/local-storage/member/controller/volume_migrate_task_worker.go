package controller

import (
	"context"
	"fmt"
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-storage/utils"
	"github.com/hwameistor/hwameistor/pkg/local-storage/utils/datacopy"
	log "github.com/sirupsen/logrus"
	"github.com/wxnacy/wgo/arrays"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
		return m.volumeMigrateAddReplica(migrate, vol)
	case apisv1alpha1.OperationStateMigrateSyncReplica:
		return m.volumeMigrateSyncReplica(migrate, vol)
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
	//Indicates that lvm is being migrated
	var anno map[string]string
	if anno = vol.GetAnnotations(); anno == nil {
		anno = make(map[string]string)
	}
	anno[apisv1alpha1.VolumeMigrateCompletedAnnoKey] = apisv1alpha1.MigrateStarted
	vol.SetAnnotations(anno)
	err := m.apiClient.Update(ctx, vol)
	if err != nil {
		logCtx.WithField("LocalVolume", vol.Name).WithError(err).Debug("lvm anno set file")
		return err
	}

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

	lvmList := &apisv1alpha1.LocalVolumeMigrateList{}
	if err := m.apiClient.List(context.Background(), lvmList); err != nil {
		logCtx.WithError(err).Error("Failed to list VolumeMigrate from cache")
		return err
	}

	var count int
	for _, lvm := range lvmList.Items {
		if lvm.Name != migrate.Name {
			if lvm.Status.State == apisv1alpha1.OperationStateSubmitted ||
				lvm.Status.State == apisv1alpha1.OperationStateMigrateAddReplica ||
				lvm.Status.State == apisv1alpha1.OperationStateMigrateSyncReplica ||
				lvm.Status.State == apisv1alpha1.OperationStateMigratePruneReplica {
				count++
			}
		}
	}
	if count >= m.migrateConcurrentNumber {
		migrate.Status.Message = "Waiting for other migration tasks to complete"
		m.apiClient.Status().Update(ctx, migrate)
		return fmt.Errorf("the number of concurrent migrations has been reached:%d", m.migrateConcurrentNumber)
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
	}

	logCtx.WithField("volumes", migrate.Status.Volumes).Debug("Checking for volumes to be migrated ...")

	if migrate.Status.Volumes == nil || len(migrate.Status.Volumes) == 0 {
		logCtx.Debug("Looking for the associated volumes to be migrated ...")
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
		migrate.Status.Volumes = []string{}
		for i := range volList {
			if err := m.checkReplicasForVolume(volList[i]); err != nil {
				logCtx.WithField("LocalVolume", volList[i].Name).WithError(err).Error("Replicas are in problem")
				migrate.Status.Message = err.Error()
				m.apiClient.Status().Update(ctx, migrate)
				return err
			}
			migrate.Status.Volumes = append(migrate.Status.Volumes, volList[i].Name)
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

func (m *manager) volumeMigrateAddReplica(migrate *apisv1alpha1.LocalVolumeMigrate, vol *apisv1alpha1.LocalVolume) error {
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

	for _, volName := range migrate.Status.Volumes {
		vol := &apisv1alpha1.LocalVolume{}
		if err := m.apiClient.Get(context.TODO(), types.NamespacedName{Name: volName}, vol); err != nil {
			return err
		}
		if vol.Spec.ReplicaNumber > migrate.Status.OriginalReplicaNumber {
			if err := m.checkReplicasForVolume(vol); err != nil {
				logCtx.WithField("LocalVolume", volName).WithError(err).Debug("Volume migration is still in process")
				migrate.Status.Message = fmt.Sprintf("In progress: %s", err.Error())
				m.apiClient.Status().Update(ctx, migrate)
				return err
			}
			logCtx.WithField("LocalVolume", volName).Debug("Volume has added a new replica successfully")
			continue
		}
		vol.Spec.ReplicaNumber++
		conf, err := m.volumeScheduler.ConfigureVolumeOnAdditionalNodes(vol, []*apisv1alpha1.LocalStorageNode{lsNode})
		if err != nil {
			logCtx.WithField("volume", volName).WithError(err).Error("Failed to configure LocalVolume")
			migrate.Status.Message = fmt.Sprintf("Failed to configure LocalVolume %s", volName)
			m.apiClient.Status().Update(ctx, migrate)
			return err
		}
		vol.Spec.Config = conf
		if err := m.apiClient.Update(ctx, vol); err != nil {
			logCtx.WithField("LocalVolume", volName).WithError(err).Error("Failed to add a new replica to the volume")
			migrate.Status.Message = fmt.Sprintf("Failed to migrate volume %s", volName)
			m.apiClient.Status().Update(ctx, migrate)
			return fmt.Errorf("failed to migrate volume %s", volName)
		}
		migrate.Status.Message = fmt.Sprintf("Adding a new replica to volume %s", volName)
		m.apiClient.Status().Update(ctx, migrate)
		return fmt.Errorf("still in progress")
	}

	logCtx.Debug("Successfully added the new replicas")
	migrate.Status.Message = "Successfully added the new replica"
	migrate.Status.State = apisv1alpha1.OperationStateMigrateSyncReplica
	return m.apiClient.Status().Update(ctx, migrate)
}

func (m *manager) volumeMigrateSyncReplica(migrate *apisv1alpha1.LocalVolumeMigrate, vol *apisv1alpha1.LocalVolume) error {
	logCtx := m.logger.WithFields(log.Fields{"migration": migrate.Name, "spec": migrate.Spec, "status": migrate.Status})
	logCtx.Debug("Start syncing replicas")

	ctx := context.TODO()
	for _, volName := range migrate.Status.Volumes {
		vol := &apisv1alpha1.LocalVolume{}
		if err := m.apiClient.Get(ctx, types.NamespacedName{Name: volName}, vol); err != nil {
			return err
		}

		if vol.Spec.Convertible {
			continue
		}
		// non-convertible volume
		if err := m.syncReplica(migrate, vol); err != nil {
			logCtx.WithField("LocalVolume", volName).WithError(err).Error("Failed to synchronize replicas")
			migrate.Status.Message = err.Error()
			m.apiClient.Status().Update(ctx, migrate)
			return err
		}
	}
	migrate.Status.State = apisv1alpha1.OperationStateMigratePruneReplica
	return m.apiClient.Status().Update(ctx, migrate)
}

func (m *manager) syncReplica(migrate *apisv1alpha1.LocalVolumeMigrate, vol *apisv1alpha1.LocalVolume) (err error) {
	logCtx := m.logger.WithFields(log.Fields{"migration": migrate.Name, "volume": vol.Name})
	logCtx.Debug("Preparing the resources for data sync ...")

	jobName := generateJobName(migrate.Name, vol.Spec.PersistentVolumeClaimName)
	return m.dataCopyManager.Sync(jobName, migrate.Spec.SourceNode, migrate.Status.TargetNode, vol.Name)
}

func (m *manager) volumeMigratePruneReplica(migrate *apisv1alpha1.LocalVolumeMigrate, vol *apisv1alpha1.LocalVolume, lvg *apisv1alpha1.LocalVolumeGroup) error {
	logCtx := m.logger.WithFields(log.Fields{"migration": migrate.Name, "spec": migrate.Spec, "status": migrate.Status})
	logCtx.Debug("Start pruning the replicas")

	ctx := context.TODO()

	// check if source volume can be pruned safely
	// configmap will only be created for localvolumemigrate belongs to nonConvertible localvolume
	if !vol.Spec.Convertible {
		cm := &corev1.ConfigMap{}
		cmName := datacopy.GetConfigMapName(datacopy.SyncConfigMapName, migrate.Spec.VolumeName)
		if err := m.apiClient.Get(ctx, types.NamespacedName{Namespace: m.namespace, Name: cmName}, cm); err != nil {
			logCtx.WithField("configmap", cmName).Error("Not found the data sync configmap")
			return err
		}
		if cm.Data[datacopy.SyncConfigSourceNodeCompleteKey] != datacopy.SyncTrue || cm.Data[datacopy.SyncConfigTargetNodeCompleteKey] != datacopy.SyncTrue {
			logCtx.WithField("configmap", cmName).Error("either source or target node is not completed, can't prune it now")
			return fmt.Errorf("either source or target node is not completed, can't prune it")
		}
		logCtx.WithField("configmap", cmName).Info("source volume is unpublished, start pruning it")
	}

	if arrays.ContainsString(lvg.Spec.Accessibility.Nodes, migrate.Spec.SourceNode) != -1 {
		lvg.Spec.Accessibility.Nodes = utils.RemoveStringItem(lvg.Spec.Accessibility.Nodes, migrate.Spec.SourceNode)
		if err := m.apiClient.Update(ctx, lvg); err != nil {
			logCtx.WithField("LocalVolumeGroup", lvg.Name).WithError(err).Error("Failed to update LocalVolumeGroup")
			return err
		}
	}

	for _, volName := range migrate.Status.Volumes {
		vol := &apisv1alpha1.LocalVolume{}
		if err := m.apiClient.Get(ctx, types.NamespacedName{Name: volName}, vol); err != nil {
			return err
		}

		// New replica is added and synced successfully, will remove the to-be-migrated replica from Volume's config
		if vol.Spec.ReplicaNumber > migrate.Status.OriginalReplicaNumber {
			// prune the to-be-migrated replica
			replicas := []apisv1alpha1.VolumeReplica{}
			for j := range vol.Spec.Config.Replicas {
				if vol.Spec.Config.Replicas[j].Hostname != migrate.Spec.SourceNode {
					replicas = append(replicas, vol.Spec.Config.Replicas[j])
				}
			}
			vol.Spec.Config.Replicas = replicas
			vol.Spec.ReplicaNumber = migrate.Status.OriginalReplicaNumber
			if err := m.apiClient.Update(ctx, vol); err != nil {
				logCtx.WithField("LocalVolume", volName).WithError(err).Error("Failed to prune a replica")
				migrate.Status.Message = fmt.Sprintf("Failed to prune a replica of volume %s", volName)
				m.apiClient.Status().Update(ctx, migrate)
				return fmt.Errorf("failed to prune a replica of volume %s", volName)
			}
			migrate.Status.Message = fmt.Sprintf("Pruning a replica of volume %s", volName)
			m.apiClient.Status().Update(ctx, migrate)
			return fmt.Errorf("pruning replicas in progress")
		}
		if err := m.checkReplicasForVolume(vol); err != nil {
			logCtx.WithField("LocalVolume", volName).WithError(err).Debug("Still pruning the replica")
			return err
		}
	}

	//Indicates lvm migration is complete
	var anno map[string]string
	if anno = vol.GetAnnotations(); anno == nil {
		anno = make(map[string]string)
	}
	anno[apisv1alpha1.VolumeMigrateCompletedAnnoKey] = apisv1alpha1.MigrateCompleted
	vol.SetAnnotations(anno)
	err := m.apiClient.Update(ctx, vol)
	if err != nil {
		logCtx.WithField("LocalVolume", vol.Name).WithError(err).Debug("lvm anno is stall migrateStarted")
		return err
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

	ctx := context.TODO()
	for _, volName := range migrate.Status.Volumes {
		cmName := datacopy.GetConfigMapName(datacopy.SyncConfigMapName, volName)
		cm := &corev1.ConfigMap{}
		if err := m.apiClient.Get(ctx, types.NamespacedName{Namespace: m.namespace, Name: cmName}, cm); err == nil {
			logCtx.WithField("configmap", cmName).Debug("Cleanup the migrate config")
			m.apiClient.Delete(ctx, cm)
		}
	}
	return m.apiClient.Delete(ctx, migrate)
}

func (m *manager) checkReplicasForVolume(vol *apisv1alpha1.LocalVolume) error {
	if vol.Spec.Config == nil {
		return fmt.Errorf("invalid volume configuration")
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

func (m *manager) gcSyncJobPod(pod *corev1.Pod) error {
	if pod.Namespace == m.namespace && pod.Labels["app"] == datacopy.SyncJobLabelApp && len(pod.OwnerReferences) == 0 && pod.Status.Phase == corev1.PodSucceeded {
		return m.apiClient.Delete(context.TODO(), pod)
	}
	return nil
}

// getStorageNodeIP returns the StorageIP configured in the corresponding LocalStorageNode
func (m *manager) getStorageNodeIP(nodeName string) (string, error) {
	storageNode := apisv1alpha1.LocalStorageNode{}
	if err := m.apiClient.Get(context.TODO(), client.ObjectKey{Name: nodeName}, &storageNode); err != nil {
		return "", err
	}
	return storageNode.Spec.StorageIP, nil
}

func generateJobName(mName string, pvcName string) string {
	if len(mName) > 25 {
		mName = mName[:25]
	}
	if len(pvcName) > 25 {
		pvcName = pvcName[:25]
	}
	jobName := fmt.Sprintf("%s-datacopy-%s", mName, pvcName)
	ensuredJobName := ensureNameMatchDNS1123Subdomain(jobName)
	return ensuredJobName
}

func ensureNameMatchDNS1123Subdomain(name string) string {
	for {
		errs := validation.IsDNS1123Subdomain(name)
		if len(errs) != 0 {
			log.Infof("object name %v doesn't match DNS1123Subdomain rule, modify it", name)
			length := len(name)
			name = name[:(length - 1)]
			log.Infof("object name modifie to %v", name)
			continue
		}
		return name
	}
}
