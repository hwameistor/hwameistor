package controller

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

const (
	volumeGroupFinalizer = "hwameistor.io/localvolumegroup-protection"
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

func (m *manager) processVolumeConvert(vcName string) error {
	logCtx := m.logger.WithFields(log.Fields{"VolumeConvert": vcName})
	logCtx.Debug("Working on a VolumeConvert task")

	convert := &apisv1alpha1.LocalVolumeConvert{}
	if err := m.apiClient.Get(context.TODO(), types.NamespacedName{Name: vcName}, convert); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("Failed to get VolumeConvert from cache")
			return err
		}
		logCtx.Info("Not found the VolumeConvert from cache, should be deleted already.")
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
	logCtx.Debug("Starting to process a VolumeConvert task")
	switch convert.Status.State {
	case "":
		return m.volumeConvertSubmit(convert)
	case apisv1alpha1.OperationStateSubmitted:
		return m.volumeConvertStart(convert)
	case apisv1alpha1.OperationStateInProgress:
		return m.volumeConvertInProgress(convert)
	case apisv1alpha1.OperationStateCompleted:
		return m.volumeConvertCleanup(convert)
	case apisv1alpha1.OperationStateToBeAborted:
		return m.volumeConvertAbort(convert)
	case apisv1alpha1.OperationStateAborted:
		return m.volumeConvertCleanup(convert)
	default:
		logCtx.Error("Invalid state")
	}
	return fmt.Errorf("invalid state")
}

func (m *manager) volumeConvertSubmit(convert *apisv1alpha1.LocalVolumeConvert) error {
	logCtx := m.logger.WithFields(log.Fields{"convert": convert.Name, "spec": convert.Spec})
	logCtx.Debug("Submit a VolumeConvert")

	ctx := context.TODO()
	if len(convert.Spec.VolumeName) == 0 {
		convert.Status.Message = "Invalid volume name"
		return m.apiClient.Status().Update(ctx, convert)
	}

	vol := &apisv1alpha1.LocalVolume{}
	if err := m.apiClient.Get(ctx, types.NamespacedName{Name: convert.Spec.VolumeName}, vol); err != nil {
		logCtx.WithError(err).Error("Failed to query volume")
		convert.Status.Message = err.Error()
		m.apiClient.Status().Update(ctx, convert)
		return err
	}

	localVolumeGroupName := vol.Spec.VolumeGroup

	lvg := &apisv1alpha1.LocalVolumeGroup{}
	if err := m.apiClient.Get(ctx, types.NamespacedName{Name: localVolumeGroupName}, lvg); err != nil {
		logCtx.WithError(err).Error("Failed to query lvg")
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
				m.logger.WithError(err).Error("Failed to list LocalVolumes")
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
						// currently, only support convertible non-HA volume to HA conversion
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

func (m *manager) volumeConvertStart(convert *apisv1alpha1.LocalVolumeConvert) error {
	logCtx := m.logger.WithFields(log.Fields{"convert": convert.Name, "spec": convert.Spec, "status": convert.Status})
	logCtx.Debug("Start a VolumeConvert")

	ctx := context.TODO()

	vol, err := m.queryLocalVolume(ctx, convert.Spec.VolumeName)
	if err != nil {
		convert.Status.Message = err.Error()
		m.apiClient.Status().Update(ctx, convert)
		return err
	}

	lvg, err := m.queryLocalVolumeGroup(ctx, vol.Spec.VolumeGroup)
	if err != nil {
		convert.Status.Message = err.Error()
		m.apiClient.Status().Update(ctx, convert)
		return err
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	for _, fnlr := range lvg.Finalizers {
		if fnlr == volumeGroupFinalizer {
			volList := &apisv1alpha1.LocalVolumeList{}
			if err := m.apiClient.List(context.TODO(), volList); err != nil {
				m.logger.WithError(err).Error("Failed to list LocalVolumes")
			}

			for _, vol := range volList.Items {
				if vol.Spec.VolumeGroup == lvg.Name {
					if vol.Spec.ReplicaNumber == convert.Spec.ReplicaNumber {
						continue
					}
					vol.Spec.ReplicaNumber = convert.Spec.ReplicaNumber
					if err := m.apiClient.Update(ctx, &vol); err != nil {
						logCtx.WithField("volName", vol.Name).WithError(err).Error("Volume failed to start the volume convert")
						convert.Status.Message = err.Error()
						m.apiClient.Status().Update(ctx, convert)
						return err
					}
				}
			}
		}
	}

	convert.Status.State = apisv1alpha1.OperationStateInProgress
	logCtx.WithField("status", convert.Status).Debug("Started volume convert")
	return m.apiClient.Status().Update(ctx, convert)
}

func (m *manager) volumeConvertInProgress(convert *apisv1alpha1.LocalVolumeConvert) error {
	logCtx := m.logger.WithFields(log.Fields{"convert": convert.Name, "spec": convert.Spec, "status": convert.Status})
	logCtx.Debug("Check the status of a VolumeConvert in progress")

	ctx := context.TODO()

	vol, err := m.queryLocalVolume(ctx, convert.Spec.VolumeName)
	if err != nil {
		convert.Status.Message = err.Error()
		m.apiClient.Status().Update(ctx, convert)
		return err
	}

	lvg, err := m.queryLocalVolumeGroup(ctx, vol.Spec.VolumeGroup)
	if err != nil {
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
				m.logger.WithError(err).Error("Failed to list LocalVolumes")
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

func (m *manager) volumeConvertAbort(convert *apisv1alpha1.LocalVolumeConvert) error {
	logCtx := m.logger.WithFields(log.Fields{"convert": convert.Name, "spec": convert.Spec, "status": convert.Status})
	logCtx.Debug("Abort a VolumeConvert")

	convert.Status.State = apisv1alpha1.OperationStateAborted
	return m.apiClient.Status().Update(context.TODO(), convert)
}

func (m *manager) volumeConvertCleanup(convert *apisv1alpha1.LocalVolumeConvert) error {
	logCtx := m.logger.WithFields(log.Fields{"convert": convert.Name, "spec": convert.Spec, "status": convert.Status})
	logCtx.Debug("Cleanup a VolumeConvert")

	return m.apiClient.Delete(context.TODO(), convert)
}

func (m *manager) queryLocalVolume(ctx context.Context, volumeName string) (*apisv1alpha1.LocalVolume, error) {
	vol := &apisv1alpha1.LocalVolume{}
	if err := m.apiClient.Get(ctx, types.NamespacedName{Name: volumeName}, vol); err != nil {
		m.logger.WithError(err).WithField("LocalVolume", volumeName).Error("Failed to query LocalVolume")
		return nil, err
	}
	return vol, nil
}

func (m *manager) queryLocalVolumeGroup(ctx context.Context, volumeGroupName string) (*apisv1alpha1.LocalVolumeGroup, error) {
	lvg := &apisv1alpha1.LocalVolumeGroup{}
	if err := m.apiClient.Get(ctx, types.NamespacedName{Name: volumeGroupName}, lvg); err != nil {
		m.logger.WithError(err).WithField("LocalVolumeGroup", volumeGroupName).Error("Failed to query LocalVolumeGroup")
		return nil, err
	}
	return lvg, nil
}
