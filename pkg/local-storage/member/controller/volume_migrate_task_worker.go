package controller

import (
	"context"
	"fmt"
	"github.com/wxnacy/wgo/arrays"
	"strings"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/local-storage/v1alpha1"
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

func (m *manager) processVolumeMigrate(vmNamespacedName string) error {
	logCtx := m.logger.WithFields(log.Fields{"VolumeMigrate": vmNamespacedName})
	logCtx.Debug("Working on a VolumeMigrate task")

	splitRes := strings.Split(vmNamespacedName, "/")
	var ns, vmName string
	if len(splitRes) >= 2 {
		ns = splitRes[0]
		vmName = splitRes[1]
	}

	migrate := &apisv1alpha1.LocalVolumeMigrate{}
	if err := m.apiClient.Get(context.TODO(), types.NamespacedName{Namespace: ns, Name: vmName}, migrate); err != nil {
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

	// state chain: (empty) -> Submitted -> Start -> InProgress -> Completed

	// log with namespace/name is enough
	logCtx = m.logger.WithFields(log.Fields{"migration": migrate.Name, "spec": migrate.Spec, "status": migrate.Status})
	logCtx.Debug("Starting to process a VolumeMigrate task")
	switch migrate.Status.State {
	case "":
		return m.volumeMigrateSubmit(migrate)
	case apisv1alpha1.OperationStateSubmitted:
		return m.volumeMigrateStart(migrate)
	case apisv1alpha1.OperationStateInProgress:
		return m.volumeMigrateInProgress(migrate)
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

func (m *manager) volumeMigrateSubmit(migrate *apisv1alpha1.LocalVolumeMigrate) error {
	logCtx := m.logger.WithFields(log.Fields{"migration": migrate.Name, "spec": migrate.Spec})
	logCtx.Debug("Submit a VolumeMigrate")

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

	localVolumeGroupName := vol.Spec.VolumeGroup

	lvg := &apisv1alpha1.LocalVolumeGroup{}
	if err := m.apiClient.Get(ctx, types.NamespacedName{Namespace: migrate.Namespace, Name: localVolumeGroupName}, lvg); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("VolumeGroupMigrateSubmit: Failed to query lvg")
		} else {
			logCtx.Info("VolumeGroupMigrateSubmit: Not found the lvg")
		}
		migrate.Status.Message = err.Error()
		m.apiClient.Status().Update(ctx, migrate)
		return err
	}

	for _, fnlr := range lvg.Finalizers {
		if fnlr == volumeGroupFinalizer {
			m.lock.Lock()
			defer m.lock.Unlock()

			volList := &apisv1alpha1.LocalVolumeList{}
			if err := m.apiClient.List(context.TODO(), volList); err != nil {
				m.logger.WithError(err).Fatal("VolumeGroupMigrateSubmit: Failed to list LocalVolumes")
			}

			for _, vol := range volList.Items {
				if vol.Spec.VolumeGroup == lvg.Name {
					if vol.Name != migrate.Spec.VolumeName && migrate.Spec.MigrateAllVols == false {
						logCtx.WithFields(log.Fields{"volume": vol.Name, "migrateAllVols": migrate.Spec.MigrateAllVols}).Error("VolumeGroupMigrateSubmit: Can't migrate false migrateAllVols flag volume")
						migrate.Status.Message = "VolumeGroupMigrateSubmit: Can't migrate volume whose localVolumeGroup has other volumes, meantime migrate's migrateAllVols flag is false; If you want migrateAllVols, modify migrateAllVols flag into true"
						return m.apiClient.Status().Update(context.TODO(), migrate)
					}
				}
			}

			for _, vol := range volList.Items {
				if vol.Spec.VolumeGroup == lvg.Name {
					if vol.Spec.ReplicaNumber == 1 && !vol.Spec.Convertible {
						logCtx.WithFields(log.Fields{"volume": vol.Name, "replicaNumber": vol.Spec.ReplicaNumber}).Error("VolumeGroupMigrateSubmit: Can't migrate inconvertible non-HA volume")
						migrate.Status.Message = "VolumeGroupMigrateSubmit: Can't migrate inconvertible non-HA volume"
						return m.apiClient.Status().Update(context.TODO(), migrate)
					}

					if vol.Status.State != apisv1alpha1.VolumeStateReady {
						logCtx.WithFields(log.Fields{"volume": vol.Name, "state": vol.Status.State}).Error("Volume is not ready")
						replicas, err := m.getReplicasForVolume(vol.Name)
						if err != nil {
							logCtx.Error("VolumeGroupMigrateSubmit: Failed to list VolumeReplica")
							return err
						}
						if len(replicas) != int(vol.Spec.ReplicaNumber) {
							logCtx.Info("VolumeGroupMigrateSubmit: Not all VolumeReplicas are created")
							return fmt.Errorf("VolumeGroupMigrateSubmit: volume not ready")
						}
						var needMigrate bool
						for _, replica := range replicas {
							if replica.Status.State == apisv1alpha1.VolumeReplicaStateReady {
								needMigrate = true
								break
							}
						}
						if needMigrate == false {
							migrate.Status.Message = "VolumeGroupMigrateSubmit: Volume And VolumeReplica both not ready"
							return m.apiClient.Status().Update(context.TODO(), migrate)
						}
					}

					migrate.Status.ReplicaNumber = vol.Spec.ReplicaNumber
					migrate.Status.State = apisv1alpha1.OperationStateSubmitted
					m.apiClient.Status().Update(context.TODO(), migrate)
				}
			}
		}
	}
	return nil
}

func (m *manager) volumeMigrateStart(migrate *apisv1alpha1.LocalVolumeMigrate) error {
	logCtx := m.logger.WithFields(log.Fields{"migration": migrate.Name, "spec": migrate.Spec, "status": migrate.Status})
	logCtx.Debug("Start a VolumeMigrate")

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

	localVolumeGroupName := vol.Spec.VolumeGroup

	lvg := &apisv1alpha1.LocalVolumeGroup{}
	if err := m.apiClient.Get(ctx, types.NamespacedName{Namespace: migrate.Namespace, Name: localVolumeGroupName}, lvg); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("VolumeGroupMigrateStart: Failed to query lvg")
		} else {
			logCtx.Info("VolumeGroupMigrateStart: Not found the lvg")
		}
		migrate.Status.Message = err.Error()
		m.apiClient.Status().Update(ctx, migrate)
		return err
	}

	for _, fnlr := range lvg.Finalizers {
		if fnlr == volumeGroupFinalizer {
			m.lock.Lock()
			defer m.lock.Unlock()

			volList := &apisv1alpha1.LocalVolumeList{}
			if err := m.apiClient.List(context.TODO(), volList); err != nil {
				m.logger.WithError(err).Fatal("VolumeGroupMigrateStart: Failed to list LocalVolumes")
			}

			for _, vol := range volList.Items {
				if vol.Spec.VolumeGroup == lvg.Name {
					if vol.Spec.Config == nil {
						migrate.Status.Message = "VolumeGroupMigrateStart: Volume to be migrated is not ready yet"
						m.apiClient.Status().Update(ctx, migrate)
						return fmt.Errorf("VolumeGroupMigrateStart: volume not ready")
					}

					if vol.Spec.ReplicaNumber > migrate.Status.ReplicaNumber {
						migrate.Status.State = apisv1alpha1.OperationStateInProgress
						return m.apiClient.Status().Update(ctx, migrate)
					}

					if vol.Status.State != apisv1alpha1.VolumeStateReady {
						logCtx.WithFields(log.Fields{"volume": vol.Name, "state": vol.Status.State}).Error("VolumeGroupMigrateStart: The volume is not ready")
						replicas, err := m.getReplicasForVolume(vol.Name)
						if err != nil {
							logCtx.Error("VolumeGroupMigrateStart: Failed to list VolumeReplica")
							return err
						}
						if len(replicas) != int(vol.Spec.ReplicaNumber) {
							logCtx.Info("VolumeGroupMigrateStart: Not all VolumeReplicas are created")
							return fmt.Errorf("VolumeGroupMigrateStart: volume not ready")
						}
						var needMigrate bool
						for _, replica := range replicas {
							if replica.Status.State == apisv1alpha1.VolumeReplicaStateReady {
								needMigrate = true
								break
							}
						}
						if needMigrate == false {
							migrate.Status.Message = "VolumeGroupMigrateStart: Volume And VolumeReplica both not ready"
							return m.apiClient.Status().Update(context.TODO(), migrate)
						}
					}

					replicas, err := m.getReplicasForVolume(vol.Name)
					if err != nil {
						logCtx.Error("VolumeGroupMigrateStart: Failed to list VolumeReplica")
						return err
					}
					for i := 0; i < len(replicas); i++ {
						// start the migrate by adding a new replica which will be scheduled on a new node
						// trigger the migration
						vol.Spec.ReplicaNumber++
					}
					if err := m.apiClient.Update(ctx, &vol); err != nil {
						logCtx.WithFields(log.Fields{"volume": vol.Name}).WithError(err).Error("VolumeMigrateStart: Failed to add a new replica")
						migrate.Status.Message = err.Error()
						return m.apiClient.Status().Update(ctx, migrate)
					}
					migrate.Status.State = apisv1alpha1.OperationStateInProgress
					m.apiClient.Status().Update(ctx, migrate)
				}
			}
		}
	}

	return nil
}

func (m *manager) volumeMigrateInProgress(migrate *apisv1alpha1.LocalVolumeMigrate) error {
	logCtx := m.logger.WithFields(log.Fields{"migration": migrate.Name, "spec": migrate.Spec, "status": migrate.Status})
	logCtx.Debug("Start a VolumeMigrate")

	ctx := context.TODO()

	vol := &apisv1alpha1.LocalVolume{}
	if err := m.apiClient.Get(ctx, types.NamespacedName{Name: migrate.Spec.VolumeName}, vol); err != nil {
		logCtx.WithError(err).Error("Failed to query volume")
		migrate.Status.Message = err.Error()
		m.apiClient.Status().Update(ctx, migrate)
		return err
	}

	localVolumeGroupName := vol.Spec.VolumeGroup

	lvg := &apisv1alpha1.LocalVolumeGroup{}
	if err := m.apiClient.Get(ctx, types.NamespacedName{Namespace: migrate.Namespace, Name: localVolumeGroupName}, lvg); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("VolumeGroupMigrateInProgress: Failed to query lvg")
		} else {
			logCtx.WithError(err).Error("VolumeGroupMigrateInProgress: Not found the lvg")
		}
		migrate.Status.Message = err.Error()
		m.apiClient.Status().Update(ctx, migrate)
		return err
	}

	for _, fnlr := range lvg.Finalizers {
		if fnlr == volumeGroupFinalizer {
			m.lock.Lock()
			defer m.lock.Unlock()

			volList := &apisv1alpha1.LocalVolumeList{}
			if err := m.apiClient.List(context.TODO(), volList); err != nil {
				m.logger.WithError(err).Fatal("VolumeGroupMigrateInProgress: Failed to list LocalVolumes")
			}

			for _, vol := range volList.Items {
				if vol.Spec.VolumeGroup == lvg.Name {
					// firstly, make sure all the replicas are ready
					if int(vol.Spec.ReplicaNumber) != len(vol.Spec.Config.Replicas) {
						logCtx.Debug("VolumeGroupMigrateInProgress: Volume is still not configured")
						return fmt.Errorf("VolumeGroupMigrateInProgress: volume not ready")

					}
					replicas, err := m.getReplicasForVolume(vol.Name)
					if err != nil {
						logCtx.Error("VolumeGroupMigrateInProgress: Failed to list VolumeReplica")
						return err
					}
					if len(replicas) != int(vol.Spec.ReplicaNumber) {
						logCtx.Info("VolumeGroupMigrateInProgress: Not all VolumeReplicas are created")
						return fmt.Errorf("VolumeGroupMigrateInProgress: volume not ready")
					}
					hasOldReplica := false
					for _, replica := range replicas {
						if replica.Status.State != apisv1alpha1.VolumeReplicaStateReady {
							logCtx.Info("VolumeGroupMigrateInProgress: Not all VolumeReplicas are ready")
							return fmt.Errorf("VolumeGroupMigrateInProgress: volume not ready")
						}
						for _, nodeName := range migrate.Spec.SourceNodesNames {
							if replica.Spec.NodeName == nodeName {
								hasOldReplica = true
								break
							}
						}
					}

					// New replica is added and synced successfully, will remove the to-be-migrated replica from Volume's config
					if vol.Spec.ReplicaNumber > migrate.Status.ReplicaNumber {
						// prune the to-be-migrated replica

						vol.Spec.ReplicaNumber = migrate.Status.ReplicaNumber
						replicas := []apisv1alpha1.VolumeReplica{}
						migrateSrcHostNames := []string{}
						for _, nodeName := range migrate.Spec.SourceNodesNames {
							migrateSrcHostNames = append(migrateSrcHostNames, nodeName)
						}
						for i := range vol.Spec.Config.Replicas {
							if arrays.ContainsString(migrateSrcHostNames, vol.Spec.Config.Replicas[i].Hostname) == -1 {
								replicas = append(replicas, vol.Spec.Config.Replicas[i])
							}
						}
						vol.Spec.Config.Replicas = replicas

						if err := m.apiClient.Update(ctx, &vol); err != nil {
							logCtx.WithError(err).Error("VolumeGroupMigrateInProgress: Failed to re-configure Volume")
							migrate.Status.Message = err.Error()
							m.apiClient.Status().Update(ctx, migrate)
							return err
						}

						return fmt.Errorf("VolumeGroupMigrateInProgress: wait old replica deleted")
					}

					if hasOldReplica {
						logCtx.Info("VolumeGroupMigrateInProgress: The old replica has not been cleanup")
						return fmt.Errorf("VolumeGroupMigrateInProgress: not cleanup")
					}

					migrate.Status.State = apisv1alpha1.OperationStateCompleted
					m.apiClient.Status().Update(ctx, migrate)
				}
			}
		}
	}

	for _, nodeName := range migrate.Spec.TargetNodesNames {
		if arrays.ContainsString(lvg.Spec.Accessibility.Nodes, nodeName) == -1 {
			lvg.Spec.Accessibility.Nodes = append(lvg.Spec.Accessibility.Nodes, nodeName)
		}
	}

	if err := m.apiClient.Update(context.TODO(), lvg); err != nil {
		log.WithError(err).Error("VolumeGroupMigrateInProgress Reconcile : Failed to re-configure Volume")
		return err
	}

	return nil
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
