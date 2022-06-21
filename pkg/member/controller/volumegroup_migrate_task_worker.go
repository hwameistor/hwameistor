package controller

import (
	"context"
	"fmt"
	apisv1alpha1 "github.com/hwameistor/local-storage/pkg/apis/hwameistor/v1alpha1"
	log "github.com/sirupsen/logrus"
	"github.com/wxnacy/wgo/arrays"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

const (
	volumeGroupFinalizer = "hwameistor.io/localvolumegroup-protection"
)

func (m *manager) startVolumeGroupMigrateTaskWorker(stopCh <-chan struct{}) {
	m.logger.Debug("VolumeGroupMigrate Worker is working now")
	go func() {
		for {
			task, shutdown := m.volumeGroupMigrateTaskQueue.Get()
			if shutdown {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Stop the VolumeGroupMigrate worker")
				break
			}
			if err := m.processVolumeGroupMigrate(task); err != nil {
				m.logger.WithFields(log.Fields{"task": task, "attempts": m.volumeGroupMigrateTaskQueue.NumRequeues(task), "error": err.Error()}).Error("Failed to process VolumeGroupMigrate task, retry later")
				m.volumeGroupMigrateTaskQueue.AddRateLimited(task)
			} else {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Completed a VolumeGroupMigrate task.")
				m.volumeGroupMigrateTaskQueue.Forget(task)
			}
			m.volumeGroupMigrateTaskQueue.Done(task)
		}
	}()

	<-stopCh
	m.volumeGroupMigrateTaskQueue.Shutdown()
}

func (m *manager) processVolumeGroupMigrate(name string) error {
	logCtx := m.logger.WithFields(log.Fields{"VolumeGroupMigrate": name})
	logCtx.Debug("Working on a VolumeGroupMigrate task")

	migrate := &apisv1alpha1.LocalVolumeGroupMigrate{}
	if err := m.apiClient.Get(context.TODO(), types.NamespacedName{Namespace: m.namespace, Name: name}, migrate); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("Failed to get VolumeGroupMigrate from cache")
			return err
		}
		logCtx.Info("Not found the VolumeGroupMigrate from cache, should be deleted already.")
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
	logCtx.Debug("Starting to process a VolumeGroupMigrate task")
	switch migrate.Status.State {
	case "":
		return m.VolumeGroupMigrateSubmit(migrate)
	case apisv1alpha1.OperationStateSubmitted:
		return m.VolumeGroupMigrateStart(migrate)
	case apisv1alpha1.OperationStateInProgress:
		return m.VolumeGroupMigrateInProgress(migrate)
	case apisv1alpha1.OperationStateCompleted:
		return m.VolumeGroupMigrateCleanup(migrate)
	case apisv1alpha1.OperationStateToBeAborted:
		return m.VolumeGroupMigrateAbort(migrate)
	case apisv1alpha1.OperationStateAborted:
		return m.VolumeGroupMigrateCleanup(migrate)
	default:
		logCtx.Error("Invalid state/phase")
	}
	return fmt.Errorf("invalid state/phase")
}

func (m *manager) VolumeGroupMigrateSubmit(migrate *apisv1alpha1.LocalVolumeGroupMigrate) error {
	logCtx := m.logger.WithFields(log.Fields{"migration": migrate.Name, "spec": migrate.Spec})
	logCtx.Debug("VolumeGroupMigrateSubmit: Submit a VolumeGroupMigrate")

	ctx := context.TODO()

	lvg := &apisv1alpha1.LocalVolumeGroup{}
	if err := m.apiClient.Get(ctx, types.NamespacedName{Namespace: migrate.Namespace, Name: migrate.Spec.LocalVolumeGroupName}, lvg); err != nil {
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

func (m *manager) VolumeGroupMigrateStart(migrate *apisv1alpha1.LocalVolumeGroupMigrate) error {
	logCtx := m.logger.WithFields(log.Fields{"migration": migrate.Name, "spec": migrate.Spec, "status": migrate.Status})
	logCtx.Debug("VolumeGroupMigrateStart: Start a VolumeGroupMigrate")

	ctx := context.TODO()

	lvg := &apisv1alpha1.LocalVolumeGroup{}
	if err := m.apiClient.Get(ctx, types.NamespacedName{Namespace: migrate.Namespace, Name: migrate.Spec.LocalVolumeGroupName}, lvg); err != nil {
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
						return m.apiClient.Status().Update(context.TODO(), migrate)
					}

					// start the migrate by adding a new replica which will be scheduled on a new node
					// trigger the migration
					vol.Spec.ReplicaNumber++
					if err := m.apiClient.Update(ctx, &vol); err != nil {
						logCtx.WithFields(log.Fields{"volume": vol.Name}).WithError(err).Error("VolumeGroupMigrateStart: Failed to add a new replica")
						migrate.Status.Message = err.Error()
					} else {
						migrate.Status.State = apisv1alpha1.OperationStateInProgress
					}

					m.apiClient.Status().Update(ctx, migrate)
				}
			}
		}
	}

	return nil
}

func (m *manager) VolumeGroupMigrateInProgress(migrate *apisv1alpha1.LocalVolumeGroupMigrate) error {
	logCtx := m.logger.WithFields(log.Fields{"migration": migrate.Name, "spec": migrate.Spec, "status": migrate.Status})
	logCtx.Debug("VolumeGroupMigrateInProgress: Start a VolumeGroupMigrate")

	ctx := context.TODO()

	lvg := &apisv1alpha1.LocalVolumeGroup{}
	if err := m.apiClient.Get(ctx, types.NamespacedName{Namespace: migrate.Namespace, Name: migrate.Spec.LocalVolumeGroupName}, lvg); err != nil {
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
						for i := range vol.Spec.Config.Replicas {
							for _, nodeName := range migrate.Spec.SourceNodesNames {
								if vol.Spec.Config.Replicas[i].Hostname != nodeName {
									replicas = append(replicas, vol.Spec.Config.Replicas[i])
								}
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

					for _, nodeName := range migrate.Spec.TargetNodesNames {
						if arrays.ContainsString(lvg.Spec.Accessibility.Nodes, nodeName) == -1 {
							lvg.Spec.Accessibility.Nodes = migrate.Spec.TargetNodesNames
							break
						}
					}
					if err := m.apiClient.Update(context.TODO(), lvg); err != nil {
						log.WithError(err).Error("VolumeGroupMigrateInProgress Reconcile : Failed to re-configure Volume")
						return err
					}

					migrate.Status.State = apisv1alpha1.OperationStateCompleted
					m.apiClient.Status().Update(ctx, migrate)
				}
			}
		}
	}

	return nil
}

func (m *manager) VolumeGroupMigrateAbort(migrate *apisv1alpha1.LocalVolumeGroupMigrate) error {
	logCtx := m.logger.WithFields(log.Fields{"migration": migrate.Name, "spec": migrate.Spec, "status": migrate.Status})
	logCtx.Debug("Abort a VolumeGroupMigrate")

	migrate.Status.State = apisv1alpha1.OperationStateAborted
	return m.apiClient.Status().Update(context.TODO(), migrate)
}

func (m *manager) VolumeGroupMigrateCleanup(migrate *apisv1alpha1.LocalVolumeGroupMigrate) error {
	logCtx := m.logger.WithFields(log.Fields{"migration": migrate.Name, "spec": migrate.Spec, "status": migrate.Status})
	logCtx.Debug("Cleanup a VolumeGroupMigrate")

	return m.apiClient.Delete(context.TODO(), migrate)
}
