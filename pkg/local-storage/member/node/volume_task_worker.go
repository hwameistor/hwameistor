package node

import (
	"context"
	"fmt"
	"reflect"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	utilrand "k8s.io/apimachinery/pkg/util/rand"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

func (m *manager) startVolumeTaskWorker(stopCh <-chan struct{}) {

	m.logger.Debug("Replica Assignment Worker is working now")
	go func() {
		for {
			task, shutdown := m.volumeTaskQueue.Get()
			if shutdown {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Stop the Replica Task Assignment worker")
				break
			}
			if err := m.processVolumeReplicaTaskAssignment(task); err != nil {
				m.logger.WithFields(log.Fields{"task": task, "error": err.Error()}).Error("Failed to process assignment, retry later")
				m.volumeTaskQueue.AddRateLimited(task)
			} else {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Completed an assignment.")
				m.volumeTaskQueue.Forget(task)
			}
			m.volumeTaskQueue.Done(task)
		}
	}()

	<-stopCh
	m.volumeTaskQueue.Shutdown()
}

func (m *manager) processVolumeReplicaTaskAssignment(volName string) error {
	logCtx := m.logger.WithFields(log.Fields{"Volume": volName})
	logCtx.Debug("Working on a task assignment")

	vol := &apisv1alpha1.LocalVolume{}
	if err := m.apiClient.Get(context.TODO(), types.NamespacedName{Name: volName}, vol); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("Failed to get Volume from cache, retry it later ...")
			return err
		}
		logCtx.Info("Not found the Volume from cache, should be deleted already.")
		return nil
	}

	if vol.Spec.Config == nil {
		// no assignment, but have to cleanup the useless volume replica
		return m.cleanupVolumeReplica(volName)
	}

	for _, replicaTask := range vol.Spec.Config.Replicas {
		if replicaTask.Hostname != m.name {
			continue
		}
		// found my assignment
		replica, err := m.getMyVolumeReplica(volName)
		if err != nil {
			if !errors.IsNotFound(err) {
				logCtx.WithError(err).Error("Failed to query LocalVolumeReplica")
				return err
			}
			// not found VolumeReplica
			if vol.Spec.Delete {
				// already deleted
				logCtx.Debug("The LocalVolumeReplica has already been deleted")
				return nil
			}
			// create a new VolumeReplica
			return m.createVolumeReplica(vol)
		}
		// found VolumeReplica
		if vol.Spec.Delete {
			// delete it
			return m.deleteVolumeReplica(replica)
		}
		// update it, for expansion, migration ...
		return m.updateVolumeReplica(replica, vol)
	}

	// this node is not in vol.spec.config.replicas, check if it has a replica of this volume
	return m.cleanupVolumeReplica(volName)
}

func (m *manager) createVolumeReplica(vol *apisv1alpha1.LocalVolume) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if replicaName, exists := m.replicaRecords[vol.Name]; exists {
		m.logger.WithField("replica", replicaName).Debug("LocalVolumeReplica exists, ignore the creation")
		return nil
	}

	replica := &apisv1alpha1.LocalVolumeReplica{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s", vol.Name, utilrand.String(6)),
		},
		Spec: apisv1alpha1.LocalVolumeReplicaSpec{
			VolumeName:            vol.Name,
			PoolName:              vol.Spec.PoolName,
			RequiredCapacityBytes: vol.Spec.RequiredCapacityBytes,
			VolumeQoS:             vol.Spec.VolumeQoS,
			NodeName:              m.name,
		},
	}

	err := m.apiClient.Create(context.TODO(), replica)
	if err != nil {
		m.logger.WithField("replica", replica).WithError(err).Error("Failed to create a LocalVolumeReplica")
		return err
	}
	m.logger.WithField("replica", replica).Debug("Created the LocalVolumeReplica")
	m.replicaRecords[vol.Name] = replica.Name

	return nil
}

func (m *manager) deleteVolumeReplica(replica *apisv1alpha1.LocalVolumeReplica) error {
	if replica.Status.State == apisv1alpha1.VolumeReplicaStateToBeDeleted || replica.Status.State == apisv1alpha1.VolumeReplicaStateDeleted {
		m.logger.WithField("replica", replica.Name).Debug("The LocalVolumeReplica is already in process of deleting")
		return nil
	}
	m.logger.WithField("replica", replica.Name).Debug("Deleting the LocalVolumeReplica")
	replica.Status.State = apisv1alpha1.VolumeReplicaStateToBeDeleted
	return m.apiClient.Status().Update(context.TODO(), replica)
}

func (m *manager) updateVolumeReplica(replica *apisv1alpha1.LocalVolumeReplica, vol *apisv1alpha1.LocalVolume) error {
	logCtx := m.logger.WithFields(log.Fields{"replica": replica.Name})
	logCtx.Debug("Updating the LocalVolumeReplica")
	if vol.Spec.RequiredCapacityBytes > replica.Spec.RequiredCapacityBytes {
		// for capacity expansion
		logCtx.Debug("Expand LocalVolumeReplica capacity")
		replica.Spec.RequiredCapacityBytes = vol.Spec.RequiredCapacityBytes
		return m.apiClient.Update(context.TODO(), replica)
	}

	if !reflect.DeepEqual(vol.Spec.VolumeQoS, replica.Spec.VolumeQoS) {
		// for QoS update
		logCtx.Debug("Update LocalVolumeReplica QoS")
		replica.Spec.VolumeQoS = vol.Spec.VolumeQoS
		return m.apiClient.Update(context.TODO(), replica)
	}

	m.volumeReplicaTaskQueue.Add(replica.Name)
	return nil
}

func (m *manager) cleanupVolumeReplica(volName string) error {
	replica, err := m.getMyVolumeReplica(volName)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	return m.deleteVolumeReplica(replica)
}

func (m *manager) getMyVolumeReplica(volName string) (*apisv1alpha1.LocalVolumeReplica, error) {
	replicaName, exists := m.replicaRecords[volName]
	if !exists {
		return nil, errors.NewNotFound(apisv1alpha1.Resource("LocalVolumeReplica"), "LocalVolumeReplica")
	}
	replica := &apisv1alpha1.LocalVolumeReplica{}
	if err := m.apiClient.Get(context.TODO(), types.NamespacedName{Name: replicaName}, replica); err != nil {
		return nil, err
	}
	return replica, nil
}
