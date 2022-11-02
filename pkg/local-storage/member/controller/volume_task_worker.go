package controller

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

func (m *manager) startVolumeTaskWorker(stopCh <-chan struct{}) {
	m.logger.Debug("Volume Worker is working now")
	go func() {
		for {
			task, shutdown := m.volumeTaskQueue.Get()
			if shutdown {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Stop the Volume worker")
				break
			}
			if err := m.processVolume(task); err != nil {
				m.logger.WithFields(log.Fields{"task": task, "attempts": m.volumeTaskQueue.NumRequeues(task), "error": err.Error()}).Error("Failed to process Volume task, retry later")
				m.volumeTaskQueue.AddRateLimited(task)
			} else {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Completed a Volume task.")
				m.volumeTaskQueue.Forget(task)
			}
			m.volumeTaskQueue.Done(task)
		}
	}()

	<-stopCh
	m.volumeTaskQueue.Shutdown()
}

func (m *manager) processVolume(volName string) error {
	logCtx := m.logger.WithFields(log.Fields{"Volume": volName})
	logCtx.Debug("Working on a Volume task")
	vol := &apisv1alpha1.LocalVolume{}
	if err := m.apiClient.Get(context.TODO(), types.NamespacedName{Name: volName}, vol); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("Failed to get Volume from cache")
			return err
		}
		logCtx.Info("Not found the Volume from cache, should be deleted already.")
		return nil
	}

	if vol.Spec.Delete && vol.Status.State != apisv1alpha1.VolumeStateToBeDeleted && vol.Status.State != apisv1alpha1.VolumeStateDeleted {
		vol.Status.State = apisv1alpha1.VolumeStateToBeDeleted
		return m.apiClient.Status().Update(context.TODO(), vol)
	}

	// log with namespace/name is enough
	logCtx = m.logger.WithFields(log.Fields{"volume": vol.Name, "spec": vol.Spec, "status": vol.Status})
	logCtx.Debug("Starting to process a Volume")
	switch vol.Status.State {
	case "":
		return m.processVolumeSubmit(vol)
	case apisv1alpha1.VolumeStateCreating:
		return m.processVolumeCreate(vol)
	case apisv1alpha1.VolumeStateReady, apisv1alpha1.VolumeStateNotReady:
		return m.processVolumeReadyAndNotReady(vol)
	case apisv1alpha1.VolumeStateToBeDeleted:
		return m.processVolumeDelete(vol)
	case apisv1alpha1.VolumeStateDeleted:
		return m.processVolumeCleanup(vol)
	default:
		logCtx.Error("Invalid state")
	}
	return fmt.Errorf("invalid state")
}

func (m *manager) processVolumeSubmit(vol *apisv1alpha1.LocalVolume) error {
	logCtx := m.logger.WithFields(log.Fields{"volume": vol.Name, "spec": vol.Spec})
	logCtx.Debug("Submit a Volume")

	vol.Status.State = apisv1alpha1.VolumeStateCreating
	return m.apiClient.Status().Update(context.TODO(), vol)
}

func (m *manager) processVolumeCreate(vol *apisv1alpha1.LocalVolume) error {
	logCtx := m.logger.WithFields(log.Fields{"volume": vol.Name})
	logCtx.Debug("Configuring a LocalVolume")

	// scheduler will consider all the cases as following:
	/* 1. fresh new volume: 1) no config, 2) no replica
	   2. old volume with additional replica: 1) has config, 2) has replica, 3) replicaNumber > replicas
	   3. old volume with bigger capacity: 1) has config, 2) has replicas, 3) requiredCapacity > allocated
	*/
	config, err := m.volumeScheduler.Allocate(vol)
	if err != nil {
		logCtx.WithError(err).Error("Failed to schedule the LocalVolume")
		return err
	}
	logCtx.WithFields(log.Fields{"oldConfig": vol.Spec.Config, "newConfig": config}).Debug("Allocated config")

	if !config.DeepEqual(vol.Spec.Config) {
		logCtx.WithFields(log.Fields{"oldConfig": vol.Spec.Config, "newConfig": config}).Debug("Generated a different config")
		if vol.Spec.Config != nil {
			config.Version = vol.Spec.Config.Version + 1
		}
		vol.Spec.Config = config
		return m.apiClient.Update(context.TODO(), vol)
	}

	vol.Status.State = apisv1alpha1.VolumeStateNotReady
	return m.apiClient.Status().Update(context.TODO(), vol)
}

func (m *manager) processVolumeReadyAndNotReady(vol *apisv1alpha1.LocalVolume) error {
	logCtx := m.logger.WithFields(log.Fields{"volume": vol.Name, "state": vol.Status.State})

	if vol.Spec.Config == nil {
		logCtx.Debug("No config generated, create it firstly")
		vol.Status.State = apisv1alpha1.VolumeStateCreating
		return m.apiClient.Status().Update(context.TODO(), vol)
	}

	// check for case of adding replica, especially for replica migration
	if len(vol.Spec.Config.Replicas) < int(vol.Spec.ReplicaNumber) {
		logCtx.Debug("Allocated replicas can't meet requirement, try it once more")
		vol.Status.State = apisv1alpha1.VolumeStateCreating
		return m.apiClient.Status().Update(context.TODO(), vol)
	}

	if vol.Spec.ReplicaNumber == 1 && !vol.Spec.Config.Convertible && vol.Spec.Config.ResourceID > -1 {
		logCtx.Debug("Incorrected resource ID for non-HA volume, try to correct it")
		vol.Status.State = apisv1alpha1.VolumeStateCreating
		return m.apiClient.Status().Update(context.TODO(), vol)
	}

	// check for case of capacity expansion
	if vol.Spec.RequiredCapacityBytes > vol.Spec.Config.RequiredCapacityBytes {
		logCtx.WithFields(log.Fields{"current": vol.Spec.Config.RequiredCapacityBytes, "require": vol.Spec.RequiredCapacityBytes}).Debug("Requiring more capacity, create it")
		vol.Status.State = apisv1alpha1.VolumeStateCreating
		return m.apiClient.Status().Update(context.TODO(), vol)
	}

	// list all the LocalVolumeReplicas by Volume
	replicas, err := m.getReplicasForVolume(vol.Name)
	if err != nil {
		logCtx.WithError(err).Error("Failed to list LocalVolumeReplica")
		return err
	}
	vol.SetReplicas(replicas)

	// check if there is any LocalVolumeReplica not created yet
	if len(replicas) < int(vol.Spec.ReplicaNumber) {
		logCtx.Debug("Not all replicas are generated, waiting for")
		vol.Status.State = apisv1alpha1.VolumeStateNotReady
		return m.apiClient.Status().Update(context.TODO(), vol)
	}

	healthyReplicaCount := 0
	upReplicaCount := 0
	allocatedCapacityBytes := int64(0)
	for _, replica := range replicas {
		if replica.Status.State == apisv1alpha1.VolumeReplicaStateReady {
			healthyReplicaCount++
			allocatedCapacityBytes = replica.Status.AllocatedCapacityBytes
		}
		if isVolumeReplicaUp(replica) {
			upReplicaCount++
		}
	}

	// check if there are enough VolumeReplicas in ready.
	// The new added replica should not impact the Volume health, e.g. replica migration
	if healthyReplicaCount >= int(vol.Spec.ReplicaNumber) {
		vol.Status.State = apisv1alpha1.VolumeStateReady
		vol.Status.AllocatedCapacityBytes = allocatedCapacityBytes
	} else {
		vol.Status.State = apisv1alpha1.VolumeStateNotReady
	}

	if !vol.Spec.Config.ReadyToInitialize && upReplicaCount >= int(vol.Spec.ReplicaNumber) {
		oldVol := vol.DeepCopy()
		vol.Spec.Config.ReadyToInitialize = true
		patch := client.MergeFrom(oldVol)
		if err = m.apiClient.Patch(context.TODO(), vol, patch); err != nil {
			return err
		}
	}

	return m.apiClient.Status().Update(context.TODO(), vol)
}

func (m *manager) processVolumeDelete(vol *apisv1alpha1.LocalVolume) error {
	logCtx := m.logger.WithFields(log.Fields{"volume": vol.Name, "spec": vol.Spec, "state": vol.Status.State})
	logCtx.Debug("Deleting a Volume")

	// when volume's delete is set to true, nodeManager should watch for this event, and delete volume replica automatically
	// controller will not call replicas' nodes by setting VolumeReplica's delete.

	// check if all the replicas are deleted
	isDeleted := true
	for _, replicaName := range vol.Status.Replicas {
		replica := &apisv1alpha1.LocalVolumeReplica{}
		if err := m.apiClient.Get(context.TODO(), types.NamespacedName{Name: replicaName}, replica); err != nil {
			if !errors.IsNotFound(err) {
				m.logger.WithFields(log.Fields{"replica": replicaName, "error": err.Error()}).Error("Failed to query volume replica")
				return err
			}
			// not found replica, should be deleted already
			m.logger.WithFields(log.Fields{"replica": replicaName, "error": err.Error()}).Warning("Not found volume replica")
			continue
		}
		if replica.Status.State == apisv1alpha1.VolumeReplicaStateDeleted {
			continue
		}
		isDeleted = false
	}

	if isDeleted {
		vol.Status.State = apisv1alpha1.VolumeStateDeleted
		return m.apiClient.Status().Update(context.TODO(), vol)
	}
	return fmt.Errorf("volume deletion not completed")
}

func (m *manager) processVolumeCleanup(vol *apisv1alpha1.LocalVolume) error {
	logCtx := m.logger.WithFields(log.Fields{"volume": vol.Name, "spec": vol.Spec, "state": vol.Status.State})
	logCtx.Debug("Cleanup a Volume")

	return m.apiClient.Delete(context.TODO(), vol)
}

func (m *manager) getReplicasForVolume(volName string) ([]*apisv1alpha1.LocalVolumeReplica, error) {
	// todo
	replicaList := &apisv1alpha1.LocalVolumeReplicaList{}
	if err := m.apiClient.List(context.TODO(), replicaList); err != nil {
		return nil, err
	}

	var replicas []*apisv1alpha1.LocalVolumeReplica
	for i := range replicaList.Items {
		if replicaList.Items[i].Spec.VolumeName == volName {
			replicas = append(replicas, &replicaList.Items[i])
		}
	}
	return replicas, nil
}

// isVolumeReplicaUp check if HA volume replica already up
func isVolumeReplicaUp(replica *apisv1alpha1.LocalVolumeReplica) bool {
	if replica.Status.HAState == nil {
		return false
	}

	switch replica.Status.HAState.State {
	case apisv1alpha1.HAVolumeReplicaStateUp, apisv1alpha1.HAVolumeReplicaStateConsistent, apisv1alpha1.HAVolumeReplicaStateInconsistent:
		return true
	}

	return false
}
