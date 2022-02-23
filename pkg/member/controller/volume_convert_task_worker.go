package controller

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	localstoragev1alpha1 "github.com/hwameistor/local-storage/pkg/apis/localstorage/v1alpha1"
)

func (m *manager) startVolumeConvertTaskWorker(stopCh <-chan struct{}) {
	m.logger.Debug("VolumeConvert Worker is working now")
	go func() {
		for {
			task, shutdown := m.volumeConvertTaskQueue.Get()
			if shutdown {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Stop the VolumeConvert worker")
				break
			}
			if err := m.processVolumeConvert(task); err != nil {
				m.logger.WithFields(log.Fields{"task": task, "attempts": m.volumeConvertTaskQueue.NumRequeues(task), "error": err.Error()}).Error("Failed to process VolumeConvert task, retry later")
				m.volumeConvertTaskQueue.AddRateLimited(task)
			} else {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Completed a VolumeConvert task.")
				m.volumeConvertTaskQueue.Forget(task)
			}
			m.volumeConvertTaskQueue.Done(task)
		}
	}()

	<-stopCh
	m.volumeConvertTaskQueue.Shutdown()
}

func (m *manager) processVolumeConvert(name string) error {
	logCtx := m.logger.WithFields(log.Fields{"VolumeConvert": name})
	logCtx.Debug("Working on a VolumeConvert task")
	convert := &localstoragev1alpha1.LocalVolumeConvert{}
	if err := m.apiClient.Get(context.TODO(), types.NamespacedName{Name: name}, convert); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("Failed to get VolumeConvert from cache")
			return err
		}
		logCtx.Info("Not found the VolumeConvert from cache, should be deleted already.")
		return nil
	}

	if convert.Spec.Abort &&
		convert.Status.State != localstoragev1alpha1.OperationStateToBeAborted &&
		convert.Status.State != localstoragev1alpha1.OperationStateAborting &&
		convert.Status.State != localstoragev1alpha1.OperationStateAborted &&
		convert.Status.State != localstoragev1alpha1.OperationStateCompleted {

		convert.Status.State = localstoragev1alpha1.OperationStateToBeAborted
		return m.apiClient.Status().Update(context.TODO(), convert)
	}

	// log with namespace/name is enough
	logCtx = m.logger.WithFields(log.Fields{"convert": convert.Name, "spec": convert.Spec, "status": convert.Status})
	logCtx.Debug("Starting to process a VolumeConvert task")
	switch convert.Status.State {
	case "":
		return m.volumeConvertSubmit(convert)
	case localstoragev1alpha1.OperationStateSubmitted:
		return m.volumeConvertStart(convert)
	case localstoragev1alpha1.OperationStateInProgress:
		return m.volumeConvertInProgress(convert)
	case localstoragev1alpha1.OperationStateCompleted:
		return m.volumeConvertCleanup(convert)
	case localstoragev1alpha1.OperationStateToBeAborted:
		return m.volumeConvertAbort(convert)
	case localstoragev1alpha1.OperationStateAborted:
		return m.volumeConvertCleanup(convert)
	default:
		logCtx.Error("Invalid state")
	}
	return fmt.Errorf("invalid state")
}

func (m *manager) volumeConvertSubmit(convert *localstoragev1alpha1.LocalVolumeConvert) error {
	logCtx := m.logger.WithFields(log.Fields{"convert": convert.Name, "spec": convert.Spec})
	logCtx.Debug("Submit a VolumeConvert")

	ctx := context.TODO()
	if len(convert.Spec.VolumeName) == 0 {
		convert.Status.Message = "Invalid volume name"
		return m.apiClient.Status().Update(ctx, convert)
	}

	vol := &localstoragev1alpha1.LocalVolume{}
	if err := m.apiClient.Get(ctx, types.NamespacedName{Name: convert.Spec.VolumeName}, vol); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("Failed to query volume")
		} else {
			logCtx.Error("Not found the volume")
		}
		convert.Status.Message = err.Error()
		m.apiClient.Status().Update(ctx, convert)
		return err
	}

	if vol.Spec.ReplicaNumber == convert.Spec.ReplicaNumber {
		convert.Status.State = localstoragev1alpha1.OperationStateSubmitted
	} else if vol.Spec.ReplicaNumber == 1 && !vol.Spec.Convertible {
		logCtx.WithField("volume", vol.Spec).Error("Can't convert the incovertible volume")
		convert.Status.Message = "Inconvertible volume"
	} else if vol.Spec.ReplicaNumber == 1 && convert.Spec.ReplicaNumber == 2 {
		// currently, only support convertible non-HA volume to HA convertion
		convert.Status.State = localstoragev1alpha1.OperationStateSubmitted
	} else {
		logCtx.WithField("volume", vol.Spec).Error("Too big convert")
		convert.Status.Message = "Not supported"
	}

	return m.apiClient.Status().Update(ctx, convert)
}

func (m *manager) volumeConvertStart(convert *localstoragev1alpha1.LocalVolumeConvert) error {
	logCtx := m.logger.WithFields(log.Fields{"convert": convert.Name, "spec": convert.Spec, "status": convert.Status})
	logCtx.Debug("Start a VolumeConvert")

	ctx := context.TODO()

	vol := &localstoragev1alpha1.LocalVolume{}
	if err := m.apiClient.Get(ctx, types.NamespacedName{Name: convert.Spec.VolumeName}, vol); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("Failed to query volume")
		} else {
			logCtx.Info("Not found the volume")
		}
		convert.Status.Message = err.Error()
		m.apiClient.Status().Update(ctx, convert)
		return err
	}

	if vol.Spec.ReplicaNumber == convert.Spec.ReplicaNumber {
		convert.Status.State = localstoragev1alpha1.OperationStateInProgress
		m.apiClient.Status().Update(ctx, convert)
	}

	vol.Spec.ReplicaNumber = convert.Spec.ReplicaNumber
	if err := m.apiClient.Update(ctx, vol); err != nil {
		logCtx.WithError(err).Error("Failed to start the volume convert")
		convert.Status.Message = err.Error()
		return m.apiClient.Status().Update(ctx, convert)
	}

	convert.Status.State = localstoragev1alpha1.OperationStateInProgress
	logCtx.WithField("status", convert.Status).Debug("Started volume convert")
	return m.apiClient.Status().Update(ctx, convert)
}

func (m *manager) volumeConvertInProgress(convert *localstoragev1alpha1.LocalVolumeConvert) error {
	logCtx := m.logger.WithFields(log.Fields{"convert": convert.Name, "spec": convert.Spec, "status": convert.Status})
	logCtx.Debug("Check the status of a VolumeConvert in progress")

	ctx := context.TODO()

	replicas, err := m.getReplicasForVolume(convert.Spec.VolumeName)
	if err != nil {
		logCtx.WithError(err).Error("Failed to list LocalVolumeReplica")
		return err
	}
	if len(replicas) != int(convert.Spec.ReplicaNumber) {
		logCtx.WithField("replicas_found", len(replicas)).Debug("Not enough volume replicas")
		return fmt.Errorf("not enough replicas")
	}
	for _, replica := range replicas {
		if replica.Status.State != localstoragev1alpha1.VolumeReplicaStateReady {
			logCtx.WithField("replica", replica.Name).Debug("The replica is not ready")
			return fmt.Errorf("replica not ready")
		}
	}
	// update volume's capacity
	vol := &localstoragev1alpha1.LocalVolume{}
	if err := m.apiClient.Get(ctx, types.NamespacedName{Name: convert.Spec.VolumeName}, vol); err != nil {
		logCtx.WithError(err).Error("Failed to query volume")
		return err
	}

	if vol.Status.State != localstoragev1alpha1.VolumeStateReady {
		logCtx.Debug("Volume is not ready")
		convert.Status.Message = "In Progress"
		return fmt.Errorf("volume not ready")
	}

	convert.Status.State = localstoragev1alpha1.OperationStateCompleted
	convert.Status.Message = ""
	logCtx.WithField("status", convert.Status).Debug("Volume convert completed")

	return m.apiClient.Status().Update(ctx, convert)
}

func (m *manager) volumeConvertAbort(convert *localstoragev1alpha1.LocalVolumeConvert) error {
	logCtx := m.logger.WithFields(log.Fields{"convert": convert.Name, "spec": convert.Spec, "status": convert.Status})
	logCtx.Debug("Abort a VolumeConvert")

	convert.Status.State = localstoragev1alpha1.OperationStateAborted
	return m.apiClient.Status().Update(context.TODO(), convert)
}

func (m *manager) volumeConvertCleanup(convert *localstoragev1alpha1.LocalVolumeConvert) error {
	logCtx := m.logger.WithFields(log.Fields{"convert": convert.Name, "spec": convert.Spec, "status": convert.Status})
	logCtx.Debug("Cleanup a VolumeConvert")

	return m.apiClient.Delete(context.TODO(), convert)
}
