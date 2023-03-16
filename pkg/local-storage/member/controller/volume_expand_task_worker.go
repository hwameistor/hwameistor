package controller

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-storage/utils"
)

func (m *manager) startVolumeExpandTaskWorker(stopCh <-chan struct{}) {
	m.logger.Debug("VolumeExpand Worker is working now")
	go func() {
		for {
			task, shutdown := m.volumeExpandTaskQueue.Get()
			if shutdown {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Stop the VolumeExpand worker")
				break
			}
			if err := m.processVolumeExpand(task); err != nil {
				m.logger.WithFields(log.Fields{"task": task, "attempts": m.volumeExpandTaskQueue.NumRequeues(task), "error": err.Error()}).Error("Failed to process VolumeExpand task, retry later")
				m.volumeExpandTaskQueue.AddRateLimited(task)
			} else {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Completed a VolumeExpand task.")
				m.volumeExpandTaskQueue.Forget(task)
			}
			m.volumeExpandTaskQueue.Done(task)
		}
	}()

	<-stopCh
	m.volumeExpandTaskQueue.Shutdown()
}

func (m *manager) processVolumeExpand(name string) error {
	logCtx := m.logger.WithFields(log.Fields{"VolumeExpand": name})
	logCtx.Debug("Working on a VolumeExpand task")
	expand := &apisv1alpha1.LocalVolumeExpand{}
	if err := m.apiClient.Get(context.TODO(), types.NamespacedName{Name: name}, expand); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("Failed to get VolumeExpand from cache")
			return err
		}
		logCtx.Info("Not found the VolumeExpand from cache, should be deleted already.")
		return nil
	}

	if expand.Spec.Abort &&
		expand.Status.State != apisv1alpha1.OperationStateToBeAborted &&
		expand.Status.State != apisv1alpha1.OperationStateAborting &&
		expand.Status.State != apisv1alpha1.OperationStateAborted &&
		expand.Status.State != apisv1alpha1.OperationStateCompleted {

		expand.Status.State = apisv1alpha1.OperationStateToBeAborted
		return m.apiClient.Status().Update(context.TODO(), expand)
	}

	// log with namespace/name is enough
	logCtx = m.logger.WithFields(log.Fields{"expansion": expand.Name, "spec": expand.Spec, "status": expand.Status})
	logCtx.Debug("Starting to process a VolumeExpand task")
	switch expand.Status.State {
	case "":
		return m.volumeExpandSubmit(expand)
	case apisv1alpha1.OperationStateSubmitted:
		return m.volumeExpandStart(expand)
	case apisv1alpha1.OperationStateInProgress:
		return m.volumeExpandInProgress(expand)
	case apisv1alpha1.OperationStateCompleted:
		return m.volumeExpandCleanup(expand)
	case apisv1alpha1.OperationStateToBeAborted:
		return m.volumeExpandAbort(expand)
	case apisv1alpha1.OperationStateAborted:
		return m.volumeExpandCleanup(expand)
	default:
		logCtx.Error("Invalid state")
	}
	return fmt.Errorf("invalid state")
}

func (m *manager) volumeExpandSubmit(expand *apisv1alpha1.LocalVolumeExpand) error {
	logCtx := m.logger.WithFields(log.Fields{"expansion": expand.Name, "spec": expand.Spec})
	logCtx.Debug("Submit a VolumeExpand")

	expand.Status.State = apisv1alpha1.OperationStateSubmitted
	return m.apiClient.Status().Update(context.TODO(), expand)
}

func (m *manager) volumeExpandStart(expand *apisv1alpha1.LocalVolumeExpand) error {
	logCtx := m.logger.WithFields(log.Fields{"expansion": expand.Name, "spec": expand.Spec, "status": expand.Status})
	logCtx.Debug("Start a VolumeExpand")

	ctx := context.TODO()

	vol := &apisv1alpha1.LocalVolume{}
	if err := m.apiClient.Get(ctx, types.NamespacedName{Name: expand.Spec.VolumeName}, vol); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("Failed to query volume")
		} else {
			logCtx.Info("Not found the volume")
		}
		expand.Status.Message = err.Error()
		m.apiClient.Status().Update(ctx, expand)
		return err
	}

	if utils.NumericToLVMBytes(expand.Spec.RequiredCapacityBytes) > utils.NumericToLVMBytes(vol.Spec.RequiredCapacityBytes) {
		vol.Spec.RequiredCapacityBytes = utils.NumericToLVMBytes(expand.Spec.RequiredCapacityBytes)
		if err := m.apiClient.Update(ctx, vol); err != nil {
			logCtx.WithError(err).Error("Failed to update Volume with new capacity")
			return err
		}
	}

	expand.Status.State = apisv1alpha1.OperationStateInProgress
	logCtx.WithField("status", expand.Status).Debug("Started volume expansion")

	return m.apiClient.Status().Update(ctx, expand)
}

func (m *manager) volumeExpandInProgress(expand *apisv1alpha1.LocalVolumeExpand) error {
	logCtx := m.logger.WithFields(log.Fields{"expansion": expand.Name, "spec": expand.Spec, "status": expand.Status})
	logCtx.Debug("Check the status of a VolumeExpand in progress")

	ctx := context.TODO()

	replicas, err := m.getReplicasForVolume(expand.Spec.VolumeName)
	if err != nil {
		logCtx.WithError(err).Error("Failed to list LocalVolumeReplica")
		return err
	}
	for _, replica := range replicas {
		if replica.Status.State != apisv1alpha1.VolumeReplicaStateReady || replica.Status.AllocatedCapacityBytes != utils.NumericToLVMBytes(expand.Spec.RequiredCapacityBytes) {
			logCtx.WithField("replica", replica.Name).Debug("The replica is not ready")
			return fmt.Errorf("replica not ready")
		}
	}

	// update volume's capacity
	vol := &apisv1alpha1.LocalVolume{}
	if err := m.apiClient.Get(ctx, types.NamespacedName{Name: expand.Spec.VolumeName}, vol); err != nil {
		logCtx.WithError(err).Error("Failed to query volume")
		return err
	}

	if vol.Status.State != apisv1alpha1.VolumeStateReady || vol.Status.AllocatedCapacityBytes != utils.NumericToLVMBytes(expand.Spec.RequiredCapacityBytes) {
		logCtx.Debug("Volume is not ready")
		expand.Status.Message = "In Progress"
		return fmt.Errorf("not ready")
	}

	expand.Status.State = apisv1alpha1.OperationStateCompleted
	expand.Status.AllocatedCapacityBytes = vol.Status.AllocatedCapacityBytes
	expand.Status.Message = ""
	logCtx.WithField("status", expand.Status).Debug("Volume expansion completed")

	return m.apiClient.Status().Update(ctx, expand)
}

func (m *manager) volumeExpandAbort(expand *apisv1alpha1.LocalVolumeExpand) error {
	logCtx := m.logger.WithFields(log.Fields{"expansion": expand.Name, "spec": expand.Spec, "status": expand.Status})
	logCtx.Debug("Abort a VolumeExpand")

	expand.Status.State = apisv1alpha1.OperationStateAborted
	return m.apiClient.Status().Update(context.TODO(), expand)
}

func (m *manager) volumeExpandCleanup(expand *apisv1alpha1.LocalVolumeExpand) error {
	logCtx := m.logger.WithFields(log.Fields{"expansion": expand.Name, "spec": expand.Spec, "status": expand.Status})
	logCtx.Debug("Cleanup a VolumeExpand")

	return m.apiClient.Delete(context.TODO(), expand)
}
