package node

import (
	"context"
	"fmt"

	localstoragev1alpha1 "github.com/hwameistor/local-storage/pkg/apis/localstorage/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	log "github.com/sirupsen/logrus"
)

func (m *manager) startVolumeReplicaTaskWorker(stopCh <-chan struct{}) {

	m.logger.Debug("VolumeReplica Worker is working now")
	go func() {
		for {
			task, shutdown := m.volumeReplicaTaskQueue.Get()
			if shutdown {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Stop the VolumeReplica worker")
				break
			}
			if err := m.processVolumeReplica(task); err != nil {
				m.logger.WithFields(log.Fields{"task": task, "error": err.Error()}).Error("Failed to process VolumeReplica task, retry later")
				m.volumeReplicaTaskQueue.AddRateLimited(task)
			} else {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Completed a VolumeReplica task.")
				m.volumeReplicaTaskQueue.Forget(task)
			}
			m.volumeReplicaTaskQueue.Done(task)
		}
	}()

	<-stopCh
	m.volumeReplicaTaskQueue.Shutdown()
}

func (m *manager) processVolumeReplica(replicaName string) error {
	logCtx := m.logger.WithFields(log.Fields{"VolumeReplica": replicaName})
	logCtx.Debug("Working on a VolumeReplica task")
	replica := &localstoragev1alpha1.LocalVolumeReplica{}
	if err := m.apiClient.Get(context.TODO(), types.NamespacedName{Name: replicaName}, replica); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("Failed to get VolumeReplica from cache, retry it later ...")
			return err
		}
		logCtx.Info("Not found the VolumeReplica from cache, should be deleted already.")
		return nil
	}

	m.logger.Debugf("Required node name %s, current node name %s.", replica.Spec.NodeName, m.name)
	if replica.Spec.NodeName != m.name {
		return nil
	}

	if replica.Spec.Delete && replica.Status.State != localstoragev1alpha1.VolumeReplicaStateToBeDeleted && replica.Status.State != localstoragev1alpha1.VolumeReplicaStateDeleted {
		replica.Status.State = localstoragev1alpha1.VolumeReplicaStateToBeDeleted
		return m.apiClient.Status().Update(context.TODO(), replica)
	}

	logCtx = m.logger.WithFields(log.Fields{"replica": replica.Name, "spec": replica.Spec, "status": replica.Status})
	logCtx.Debug("Starting to process a VolumeReplica")

	if replica.Spec.Kind != localstoragev1alpha1.VolumeKindRAM && replica.Spec.Kind != m.storageMgr.NodeConfig().LocalStorageConfig.VolumeKind {
		replica.Status.State = localstoragev1alpha1.VolumeReplicaStateInvalid
		return m.apiClient.Status().Update(context.TODO(), replica)
	}

	switch replica.Status.State {
	case "", localstoragev1alpha1.VolumeReplicaStateInvalid:
		return m.processVolumeReplicaSubmit(replica)
	case localstoragev1alpha1.VolumeReplicaStateCreating:
		return m.processVolumeReplicaCreate(replica)
	case localstoragev1alpha1.VolumeReplicaStateReady, localstoragev1alpha1.VolumeReplicaStateNotReady:
		return m.processVolumeReplicaCheck(replica)
	case localstoragev1alpha1.VolumeReplicaStateToBeDeleted:
		return m.processVolumeReplicaDelete(replica)
	case localstoragev1alpha1.VolumeReplicaStateDeleted:
		return m.processVolumeReplicaCleanup(replica)
	default:
		logCtx.Error("Invalid VolumeReplica state")
	}
	return fmt.Errorf("invalid VolumeReplica state")
}

func (m *manager) processVolumeReplicaSubmit(replica *localstoragev1alpha1.LocalVolumeReplica) error {
	logCtx := m.logger.WithFields(log.Fields{"replica": replica.Name, "spec": replica.Spec})
	logCtx.Debug("Submit a VolumeReplica")

	replica.Status.State = localstoragev1alpha1.VolumeReplicaStateCreating
	return m.apiClient.Status().Update(context.TODO(), replica)
}

func (m *manager) processVolumeReplicaCreate(replica *localstoragev1alpha1.LocalVolumeReplica) error {
	logCtx := m.logger.WithFields(log.Fields{"replica": replica.Name, "spec": replica.Spec})
	logCtx.Debug("Creating a VolumeReplica")

	// only create the first layer of the storage volume replica, e.g. LV, Disk, RAM
	// idempotent operation
	newReplica, err := m.Storage().VolumeReplicaManager().CreateVolumeReplica(replica)
	if err != nil {
		m.logger.WithError(err).Error("Failed to process on volume replica creation.")
		return err
	}

	newReplica.Status.State = localstoragev1alpha1.VolumeReplicaStateNotReady
	return m.apiClient.Status().Update(context.TODO(), newReplica)
}

func (m *manager) processVolumeReplicaCheck(replica *localstoragev1alpha1.LocalVolumeReplica) error {
	logCtx := m.logger.WithFields(log.Fields{"replica": replica.Name, "spec": replica.Spec, "status": replica.Status})
	logCtx.Debug("Checking a VolumeReplica")

	// for the case of capacity expansion, only for LVM volume
	if replica.Status.State == localstoragev1alpha1.VolumeReplicaStateReady &&
		replica.Spec.Kind == localstoragev1alpha1.VolumeKindLVM &&
		replica.Spec.RequiredCapacityBytes > replica.Status.AllocatedCapacityBytes+localstoragev1alpha1.VolumeExpansionCapacityBytesMin {
		newReplica, err := m.Storage().VolumeReplicaManager().ExpandVolumeReplica(replica, replica.Spec.RequiredCapacityBytes)
		if err != nil {
			logCtx.WithError(err).Error("Failed to expand volume replica")
			return err
		}
		newReplica.Status.State = localstoragev1alpha1.VolumeReplicaStateNotReady
		return m.apiClient.Status().Update(context.TODO(), newReplica)
	}

	testReplica, err := m.Storage().VolumeReplicaManager().TestVolumeReplica(replica)
	if err != nil {
		m.logger.WithError(err).Error("Failed to test VolumeReplica")
		return err
	}

	// idempotent operation
	// 1. configure for HA volume by replication module like DRBD
	// 2. configure for non-HA volume transit from HA by removing replication module
	if err = m.configManager.EnsureConfig(testReplica); err != nil {
		m.logger.WithError(err).Error("Failed to process on volume replica config.")
		testReplica.Status.State = localstoragev1alpha1.VolumeReplicaStateNotReady
		m.apiClient.Status().Update(context.TODO(), testReplica)
		return err
	}

	if err = m.configManager.TestVolumeReplica(testReplica); err != nil {
		m.logger.WithError(err).Error("Failed to test configed VolumeReplica")
		testReplica.Status.State = localstoragev1alpha1.VolumeReplicaStateNotReady
		m.apiClient.Status().Update(context.TODO(), testReplica)
		return err
	}

	return m.apiClient.Status().Update(context.TODO(), testReplica)
}

func (m *manager) processVolumeReplicaDelete(replica *localstoragev1alpha1.LocalVolumeReplica) error {
	logCtx := m.logger.WithFields(log.Fields{"replica": replica.Name, "spec": replica.Spec, "status": replica.Status})
	logCtx.Debug("Deleting a VolumeReplica")

	if err := m.configManager.DeleteConfig(replica); err != nil {
		logCtx.WithError(err).Debug("Failed to remove the config")
		return err
	}

	if err := m.storageMgr.VolumeReplicaManager().DeleteVolumeReplica(replica); err != nil {
		m.logger.WithError(err).Error("Failed to delete volume replica.")
		return err
	}

	newReplica := replica.DeepCopy()
	patch := client.MergeFrom(replica)
	newReplica.Status.State = localstoragev1alpha1.VolumeReplicaStateDeleted
	return m.apiClient.Status().Patch(context.TODO(), newReplica, patch)
}

func (m *manager) processVolumeReplicaCleanup(replica *localstoragev1alpha1.LocalVolumeReplica) error {
	logCtx := m.logger.WithFields(log.Fields{"replica": replica.Name, "spec": replica.Spec, "status": replica.Status})
	logCtx.Debug("Cleanup a VolumeReplica")

	return m.apiClient.Delete(context.TODO(), replica)
}
