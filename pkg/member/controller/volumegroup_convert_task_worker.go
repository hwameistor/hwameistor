package controller

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	apisv1alpha1 "github.com/hwameistor/local-storage/pkg/apis/hwameistor/v1alpha1"
)

func (m *manager) startVolumeGroupConvertTaskWorker(stopCh <-chan struct{}) {
	m.logger.Debug("VolumeGroupConvert Worker is working now")
	go func() {
		for {
			task, shutdown := m.volumeGroupConvertTaskQueue.Get()
			if shutdown {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Stop the VolumeGroupConvert worker")
				break
			}
			if err := m.processVolumeGroupConvert(task); err != nil {
				m.logger.WithFields(log.Fields{"task": task, "attempts": m.volumeGroupConvertTaskQueue.NumRequeues(task), "error": err.Error()}).Error("Failed to process VolumeGroupConvert task, retry later")
				m.volumeGroupConvertTaskQueue.AddRateLimited(task)
			} else {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Completed a VolumeGroupConvert task.")
				m.volumeGroupConvertTaskQueue.Forget(task)
			}
			m.volumeGroupConvertTaskQueue.Done(task)
		}
	}()

	<-stopCh
	m.volumeGroupConvertTaskQueue.Shutdown()
}

func (m *manager) processVolumeGroupConvert(name string) error {
	logCtx := m.logger.WithFields(log.Fields{"VolumeGroupConvert": name})
	logCtx.Debug("Working on a VolumeGroupConvert task")
	convert := &apisv1alpha1.LocalVolumeGroupConvert{}
	if err := m.apiClient.Get(context.TODO(), types.NamespacedName{Namespace: m.namespace, Name: name}, convert); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("Failed to get VolumeGroupConvert from cache")
			return err
		}
		logCtx.Info("Not found the VolumeGroupConvert from cache, should be deleted already.")
		return nil
	}

	if convert.Spec.Abort &&
		convert.Status.State != apisv1alpha1.OperationStateToBeAborted &&
		convert.Status.State != apisv1alpha1.OperationStateAborting &&
		convert.Status.State != apisv1alpha1.OperationStateAborted &&
		convert.Status.State != apisv1alpha1.OperationStateCompleted {

		convert.Status.State = apisv1alpha1.OperationStateToBeAborted
		return m.apiClient.Status().Update(context.TODO(), convert)
	}

	// log with namespace/name is enough
	logCtx = m.logger.WithFields(log.Fields{"convert": convert.Name, "spec": convert.Spec, "status": convert.Status})
	logCtx.Debug("Starting to process a VolumeGroupConvert task")
	switch convert.Status.State {
	case "":
		return m.VolumeGroupConvertSubmit(convert)
	case apisv1alpha1.OperationStateSubmitted:
		return m.VolumeGroupConvertStart(convert)
	case apisv1alpha1.OperationStateInProgress:
		return m.VolumeGroupConvertInProgress(convert)
	case apisv1alpha1.OperationStateCompleted:
		return m.VolumeGroupConvertCleanup(convert)
	case apisv1alpha1.OperationStateToBeAborted:
		return m.VolumeGroupConvertAbort(convert)
	case apisv1alpha1.OperationStateAborted:
		return m.VolumeGroupConvertCleanup(convert)
	case apisv1alpha1.OperationStateFailed:
		return m.VolumeGroupConvertFailed(convert)
	default:
		logCtx.Error("Invalid state")
	}
	return fmt.Errorf("invalid state")
}

func (m *manager) VolumeGroupConvertSubmit(convert *apisv1alpha1.LocalVolumeGroupConvert) error {
	logCtx := m.logger.WithFields(log.Fields{"convert": convert.Name, "spec": convert.Spec})
	logCtx.Debug("Submit a VolumeGroupConvert")

	ctx := context.TODO()
	lvg := &apisv1alpha1.LocalVolumeGroup{}
	if err := m.apiClient.Get(ctx, types.NamespacedName{Namespace: convert.Namespace, Name: convert.Spec.LocalVolumeGroupName}, lvg); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("VolumeGroupMigrateStart: Failed to query lvg")
		} else {
			logCtx.Info("VolumeGroupMigrateStart: Not found the lvg")
		}
		convert.Status.Message = err.Error()
		m.apiClient.Status().Update(ctx, convert)
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
					if len(vol.Name) == 0 {
						convert.Status.Message = "Invalid volume name"
						return m.apiClient.Status().Update(ctx, convert)
					}
					if vol.Spec.ReplicaNumber == convert.Spec.ReplicaNumber {
						convert.Status.State = apisv1alpha1.OperationStateSubmitted
					} else if vol.Spec.ReplicaNumber == 1 && !vol.Spec.Convertible {
						logCtx.WithField("volume", vol.Spec).Error("Can't convert the inconvertible volume")
						msg := fmt.Sprintf("Inconvertible volume: %s", vol.Name)
						convert.Status.Message = msg
						convert.Status.State = apisv1alpha1.OperationStateFailed
						break
					} else if vol.Spec.ReplicaNumber == 1 && convert.Spec.ReplicaNumber == 2 {
						// currently, only support convertible non-HA volume to HA convertion
						convert.Status.State = apisv1alpha1.OperationStateSubmitted
					} else {
						logCtx.WithField("volume", vol.Spec).Error("Too big convert")
						convert.Status.Message = "Not supported"
					}
				}
			}
		}
	}

	return m.apiClient.Status().Update(ctx, convert)
}

func (m *manager) VolumeGroupConvertStart(convert *apisv1alpha1.LocalVolumeGroupConvert) error {
	logCtx := m.logger.WithFields(log.Fields{"convert": convert.Name, "spec": convert.Spec, "status": convert.Status})
	logCtx.Debug("Start a VolumeGroupConvert")

	ctx := context.TODO()

	lvg := &apisv1alpha1.LocalVolumeGroup{}
	if err := m.apiClient.Get(ctx, types.NamespacedName{Namespace: convert.Namespace, Name: convert.Spec.LocalVolumeGroupName}, lvg); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("VolumeGroupMigrateStart: Failed to query lvg")
		} else {
			logCtx.Info("VolumeGroupMigrateStart: Not found the lvg")
		}
		convert.Status.Message = err.Error()
		m.apiClient.Status().Update(ctx, convert)
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
					if vol.Spec.ReplicaNumber == convert.Spec.ReplicaNumber {
						convert.Status.State = apisv1alpha1.OperationStateInProgress
						m.apiClient.Status().Update(ctx, convert)
					}

					//if vol.Spec.ReplicaNumber == 1 && !vol.Spec.Convertible {
					//	continue
					//}

					vol.Spec.ReplicaNumber = convert.Spec.ReplicaNumber
					if err := m.apiClient.Update(ctx, &vol); err != nil {
						logCtx.WithField("volName", vol.Name).WithError(err).Error("Volume failed to start the volume convert")
						convert.Status.Message = err.Error()
						m.apiClient.Status().Update(ctx, convert)
					}
				}
			}
		}
	}

	convert.Status.State = apisv1alpha1.OperationStateInProgress
	logCtx.WithField("status", convert.Status).Debug("Started volume convert")
	return m.apiClient.Status().Update(ctx, convert)
}

func (m *manager) VolumeGroupConvertInProgress(convert *apisv1alpha1.LocalVolumeGroupConvert) error {
	logCtx := m.logger.WithFields(log.Fields{"convert": convert.Name, "spec": convert.Spec, "status": convert.Status})
	logCtx.Debug("Check the status of a VolumeGroupConvert in progress")

	ctx := context.TODO()

	lvg := &apisv1alpha1.LocalVolumeGroup{}
	if err := m.apiClient.Get(ctx, types.NamespacedName{Namespace: convert.Namespace, Name: convert.Spec.LocalVolumeGroupName}, lvg); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("VolumeGroupMigrateStart: Failed to query lvg")
		} else {
			logCtx.Info("VolumeGroupMigrateStart: Not found the lvg")
		}
		convert.Status.Message = err.Error()
		m.apiClient.Status().Update(ctx, convert)
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
					replicas, err := m.getReplicasForVolume(vol.Name)
					if err != nil {
						logCtx.WithError(err).Error("Failed to list LocalVolumeReplica")
						return err
					}
					if len(replicas) != int(convert.Spec.ReplicaNumber) {
						logCtx.WithField("replicas_found", len(replicas)).Debug("Not enough volume replicas")
						return fmt.Errorf("not enough replicas")
					}
					for _, replica := range replicas {
						if replica.Status.State != apisv1alpha1.VolumeReplicaStateReady {
							logCtx.WithField("replica", replica.Name).Debug("The replica is not ready")
							return fmt.Errorf("replica not ready")
						}
					}
					if vol.Status.State != apisv1alpha1.VolumeStateReady {
						logCtx.Debug("Volume is not ready")
						convert.Status.Message = "In Progress"
						return fmt.Errorf("volume not ready")
					}
				}
			}
		}
	}

	convert.Status.State = apisv1alpha1.OperationStateCompleted
	convert.Status.Message = ""
	logCtx.WithField("status", convert.Status).Debug("Volume convert completed")

	return m.apiClient.Status().Update(ctx, convert)
}

func (m *manager) VolumeGroupConvertAbort(convert *apisv1alpha1.LocalVolumeGroupConvert) error {
	logCtx := m.logger.WithFields(log.Fields{"convert": convert.Name, "spec": convert.Spec, "status": convert.Status})
	logCtx.Debug("Abort a VolumeGroupConvert")

	convert.Status.State = apisv1alpha1.OperationStateAborted
	return m.apiClient.Status().Update(context.TODO(), convert)
}

func (m *manager) VolumeGroupConvertCleanup(convert *apisv1alpha1.LocalVolumeGroupConvert) error {
	logCtx := m.logger.WithFields(log.Fields{"convert": convert.Name, "spec": convert.Spec, "status": convert.Status})
	logCtx.Debug("Cleanup a VolumeGroupConvert")

	return m.apiClient.Delete(context.TODO(), convert)
}

func (m *manager) VolumeGroupConvertFailed(convert *apisv1alpha1.LocalVolumeGroupConvert) error {
	logCtx := m.logger.WithFields(log.Fields{"convert": convert.Name, "spec": convert.Spec, "status": convert.Status})
	logCtx.Debug("Failed a VolumeGroupConvert")

	convert.Status.State = apisv1alpha1.OperationStateFailed
	return m.apiClient.Status().Update(context.TODO(), convert)
}
