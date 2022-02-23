package controller

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	localstoragev1alpha1 "github.com/hwameistor/local-storage/pkg/apis/localstorage/v1alpha1"
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

func (m *manager) processVolumeMigrate(name string) error {
	logCtx := m.logger.WithFields(log.Fields{"VolumeMigrate": name})
	logCtx.Debug("Working on a VolumeMigrate task")
	migrate := &localstoragev1alpha1.LocalVolumeMigrate{}
	if err := m.apiClient.Get(context.TODO(), types.NamespacedName{Name: name}, migrate); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("Failed to get VolumeMigrate from cache")
			return err
		}
		logCtx.Info("Not found the VolumeMigrate from cache, should be deleted already.")
		return nil
	}

	if migrate.Spec.Abort &&
		migrate.Status.State != localstoragev1alpha1.OperationStateToBeAborted &&
		migrate.Status.State != localstoragev1alpha1.OperationStateAborting &&
		migrate.Status.State != localstoragev1alpha1.OperationStateAborted &&
		migrate.Status.State != localstoragev1alpha1.OperationStateCompleted {

		migrate.Status.State = localstoragev1alpha1.OperationStateToBeAborted
		return m.apiClient.Status().Update(context.TODO(), migrate)
	}

	// state chain: (empty) -> Submitted -> Start -> InProgress -> Completed

	// log with namespace/name is enough
	logCtx = m.logger.WithFields(log.Fields{"migration": migrate.Name, "spec": migrate.Spec, "status": migrate.Status})
	logCtx.Debug("Starting to process a VolumeMigrate task")
	switch migrate.Status.State {
	case "":
		return m.volumeMigrateSubmit(migrate)
	case localstoragev1alpha1.OperationStateSubmitted:
		return m.volumeMigrateStart(migrate)
	case localstoragev1alpha1.OperationStateInProgress:
		return m.volumeMigrateInProgress(migrate)
	case localstoragev1alpha1.OperationStateCompleted:
		return m.volumeMigrateCleanup(migrate)
	case localstoragev1alpha1.OperationStateToBeAborted:
		return m.volumeMigrateAbort(migrate)
	case localstoragev1alpha1.OperationStateAborted:
		return m.volumeMigrateCleanup(migrate)
	default:
		logCtx.Error("Invalid state/phase")
	}
	return fmt.Errorf("invalid state/phase")
}

func (m *manager) volumeMigrateSubmit(migrate *localstoragev1alpha1.LocalVolumeMigrate) error {
	logCtx := m.logger.WithFields(log.Fields{"migration": migrate.Name, "spec": migrate.Spec})
	logCtx.Debug("Submit a VolumeMigrate")

	ctx := context.TODO()

	vol := &localstoragev1alpha1.LocalVolume{}
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

	if vol.Spec.ReplicaNumber == 1 && !vol.Spec.Convertible {
		logCtx.WithFields(log.Fields{"volume": vol.Name, "replicaNumber": vol.Spec.ReplicaNumber}).Error("Can't migrate inconvertible non-HA volume")
		migrate.Status.Message = "Can't migrate inconvertible non-HA volume"
		return m.apiClient.Status().Update(context.TODO(), migrate)
	}

	if vol.Status.State != localstoragev1alpha1.VolumeStateReady {
		logCtx.WithFields(log.Fields{"volume": vol.Name, "state": vol.Status.State}).Error("Volume is not ready")
		migrate.Status.Message = "Volume is not ready"
		return m.apiClient.Status().Update(context.TODO(), migrate)
	}

	migrate.Status.ReplicaNumber = vol.Spec.ReplicaNumber
	migrate.Status.State = localstoragev1alpha1.OperationStateSubmitted
	return m.apiClient.Status().Update(context.TODO(), migrate)
}

func (m *manager) volumeMigrateStart(migrate *localstoragev1alpha1.LocalVolumeMigrate) error {
	logCtx := m.logger.WithFields(log.Fields{"migration": migrate.Name, "spec": migrate.Spec, "status": migrate.Status})
	logCtx.Debug("Start a VolumeMigrate")

	ctx := context.TODO()

	vol := &localstoragev1alpha1.LocalVolume{}
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

	if vol.Spec.Config == nil {
		migrate.Status.Message = "Volume to be migrated is not ready yet"
		m.apiClient.Status().Update(ctx, migrate)
		return fmt.Errorf("volume not ready")
	}

	if vol.Spec.ReplicaNumber > migrate.Status.ReplicaNumber {
		migrate.Status.State = localstoragev1alpha1.OperationStateInProgress
		return m.apiClient.Status().Update(ctx, migrate)
	}

	if vol.Status.State != localstoragev1alpha1.VolumeStateReady {
		logCtx.WithFields(log.Fields{"volume": vol.Name, "state": vol.Status.State}).Error("The volume is not ready")
		migrate.Status.Message = "The volume is not ready"
		return m.apiClient.Status().Update(context.TODO(), migrate)
	}

	// start the migrate by adding a new replica which will be scheduled on a new node
	// trigger the migration
	vol.Spec.ReplicaNumber++
	if err := m.apiClient.Update(ctx, vol); err != nil {
		logCtx.WithFields(log.Fields{"volume": vol.Name}).WithError(err).Error("Failed to add a new replica")
		migrate.Status.Message = err.Error()
	} else {
		migrate.Status.State = localstoragev1alpha1.OperationStateInProgress
	}
	return m.apiClient.Status().Update(ctx, migrate)
}

func (m *manager) volumeMigrateInProgress(migrate *localstoragev1alpha1.LocalVolumeMigrate) error {
	logCtx := m.logger.WithFields(log.Fields{"migration": migrate.Name, "spec": migrate.Spec, "status": migrate.Status})
	logCtx.Debug("Start a VolumeMigrate")

	ctx := context.TODO()

	vol := &localstoragev1alpha1.LocalVolume{}
	if err := m.apiClient.Get(ctx, types.NamespacedName{Name: migrate.Spec.VolumeName}, vol); err != nil {
		logCtx.WithError(err).Error("Failed to query volume")
		migrate.Status.Message = err.Error()
		m.apiClient.Status().Update(ctx, migrate)
		return err
	}

	// firstly, make sure all the replicas are ready
	if int(vol.Spec.ReplicaNumber) != len(vol.Spec.Config.Replicas) {
		logCtx.Debug("Volume is still not configured")
		return fmt.Errorf("volume not ready")

	}
	replicas, err := m.getReplicasForVolume(vol.Name)
	if err != nil {
		logCtx.Error("Failed to list VolumeReplica")
		return err
	}
	if len(replicas) != int(vol.Spec.ReplicaNumber) {
		logCtx.Info("Not all VolumeReplicas are created")
		return fmt.Errorf("volume not ready")
	}
	hasOldReplica := false
	for _, replica := range replicas {
		if replica.Status.State != localstoragev1alpha1.VolumeReplicaStateReady {
			logCtx.Info("Not all VolumeReplicas are ready")
			return fmt.Errorf("volume not ready")
		}
		if replica.Spec.NodeName == migrate.Spec.NodeName {
			hasOldReplica = true
		}
	}

	// New replica is added and synced successfully, will remove the to-be-migrated replica from Volume's config
	if vol.Spec.ReplicaNumber > migrate.Status.ReplicaNumber {
		// prune the to-be-migrated replica
		vol.Spec.ReplicaNumber = migrate.Status.ReplicaNumber
		replicas := []localstoragev1alpha1.VolumeReplica{}
		for i := range vol.Spec.Config.Replicas {
			if vol.Spec.Config.Replicas[i].Hostname != migrate.Spec.NodeName {
				replicas = append(replicas, vol.Spec.Config.Replicas[i])
			}
		}
		vol.Spec.Config.Replicas = replicas
		if err := m.apiClient.Update(ctx, vol); err != nil {
			logCtx.WithError(err).Error("Failed to re-configure Volume")
			migrate.Status.Message = err.Error()
			m.apiClient.Status().Update(ctx, migrate)
			return err
		}

		return fmt.Errorf("wait old replica deleted")
	}

	if hasOldReplica {
		logCtx.Info("The old replica has not been cleanup")
		return fmt.Errorf("not cleanup")
	}

	migrate.Status.State = localstoragev1alpha1.OperationStateCompleted
	return m.apiClient.Status().Update(ctx, migrate)
}

func (m *manager) volumeMigrateAbort(migrate *localstoragev1alpha1.LocalVolumeMigrate) error {
	logCtx := m.logger.WithFields(log.Fields{"migration": migrate.Name, "spec": migrate.Spec, "status": migrate.Status})
	logCtx.Debug("Abort a VolumeMigrate")

	migrate.Status.State = localstoragev1alpha1.OperationStateAborted
	return m.apiClient.Status().Update(context.TODO(), migrate)
}

func (m *manager) volumeMigrateCleanup(migrate *localstoragev1alpha1.LocalVolumeMigrate) error {
	logCtx := m.logger.WithFields(log.Fields{"migration": migrate.Name, "spec": migrate.Spec, "status": migrate.Status})
	logCtx.Debug("Cleanup a VolumeMigrate")

	return m.apiClient.Delete(context.TODO(), migrate)
}
