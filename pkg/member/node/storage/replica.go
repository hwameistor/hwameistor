package storage

import (
	"context"
	"fmt"

	localstoragev1alpha1 "github.com/hwameistor/local-storage/pkg/apis/localstorage/v1alpha1"
	log "github.com/sirupsen/logrus"
)

type localVolumeReplicaManager struct {
	cmdExec         LocalVolumeExecutor
	ramDiskCmdExec  LocalVolumeExecutor
	volumeValidator *validator
	registry        LocalRegistry
	logger          *log.Entry
	lm              *LocalManager
}

func newLocalVolumeReplicaManager(lm *LocalManager) LocalVolumeReplicaManager {
	mgr := &localVolumeReplicaManager{
		ramDiskCmdExec:  newLocalVolumeExecutor(lm, localstoragev1alpha1.VolumeKindRAM),
		volumeValidator: newValidator(),
		registry:        lm.Registry(),
		lm:              lm,
		logger:          log.WithField("Module", "NodeManager/LocalVolumeReplicaManager"),
	}
	if lm.nodeConf.LocalStorageConfig.VolumeKind == localstoragev1alpha1.VolumeKindDisk || lm.nodeConf.LocalStorageConfig.VolumeKind == localstoragev1alpha1.VolumeKindLVM {
		mgr.cmdExec = newLocalVolumeExecutor(lm, lm.nodeConf.LocalStorageConfig.VolumeKind)
	}
	return mgr
}

func (mgr *localVolumeReplicaManager) CreateVolumeReplica(replica *localstoragev1alpha1.LocalVolumeReplica) (*localstoragev1alpha1.LocalVolumeReplica, error) {
	mgr.logger.Debugf("Creating VolumeReplica. name:%s, pool:%s, size:%d", replica.Spec.VolumeName, replica.Spec.PoolName, replica.Spec.RequiredCapacityBytes)
	if err := mgr.volumeValidator.canCreateVolumeReplica(replica, mgr.registry); err != nil {
		if err == ErrorReplicaExists {
			mgr.logger.Infof("Replica %s has already exists.", replica.Spec.VolumeName)
			newReplica := replica.DeepCopy()
			currentReplica := mgr.registry.VolumeReplicas()[replica.Spec.VolumeName]
			newReplica.Status.AllocatedCapacityBytes = currentReplica.Status.AllocatedCapacityBytes
			newReplica.Status.StoragePath = currentReplica.Status.StoragePath
			newReplica.Status.DevicePath = currentReplica.Status.DevicePath
			newReplica.Status.Synced = currentReplica.Status.Synced
			return newReplica, nil
		}
		mgr.logger.WithError(err).Errorf("Failed to validate volume replica %s.", replica.Spec.VolumeName)
		return nil, err
	}

	var newReplica *localstoragev1alpha1.LocalVolumeReplica
	var err error

	// FIXME: DISK also needs to be determined (in other words: else if Kind == Disk)
	// cause the DISK is currectly not supported.
	if replica.Spec.Kind == localstoragev1alpha1.VolumeKindRAM {
		newReplica, err = mgr.ramDiskCmdExec.CreateVolumeReplica(replica)
	} else {
		newReplica, err = mgr.cmdExec.CreateVolumeReplica(replica)
	}
	if err != nil {
		mgr.logger.WithError(err).Error("Failed to exec replica create")
		return nil, err
	}

	return newReplica, nil
}

func (mgr *localVolumeReplicaManager) DeleteVolumeReplica(replica *localstoragev1alpha1.LocalVolumeReplica) error {
	mgr.logger.Debugf("Deleting volume replica %s.", replica.Spec.VolumeName)

	if err := mgr.volumeValidator.canDeleteVolumeReplica(replica, mgr.registry); err != nil {
		if err == ErrorReplicaNotFound {
			mgr.logger.Infof("Volume replica %s not found. Skipping.", replica.Spec.VolumeName)
			return nil
		}
		mgr.logger.WithError(err).Errorf("Failed to validate volume replica %s.", replica.Spec.VolumeName)
		return err
	}

	var err error
	if replica.Spec.Kind == localstoragev1alpha1.VolumeKindRAM {
		err = mgr.ramDiskCmdExec.DeleteVolumeReplica(replica)
	} else {
		err = mgr.cmdExec.DeleteVolumeReplica(replica)
	}
	if err != nil {
		mgr.logger.WithError(err).Error("Failed to exec replica delete")
		return err
	}

	return nil
}

func (mgr *localVolumeReplicaManager) ExpandVolumeReplica(replica *localstoragev1alpha1.LocalVolumeReplica, newCapacityBytes int64) (*localstoragev1alpha1.LocalVolumeReplica, error) {
	mgr.logger.Debugf("Extending volume replica %s.", replica.Spec.VolumeName)

	if err := mgr.volumeValidator.canExpandVolumeReplica(replica, newCapacityBytes, mgr.registry); err != nil {
		mgr.logger.WithError(err).Errorf("Failed to validate volume replica %s to extend.", replica.Spec.VolumeName)
		return nil, err
	}

	var newReplica *localstoragev1alpha1.LocalVolumeReplica
	var err error
	if replica.Spec.Kind == localstoragev1alpha1.VolumeKindRAM {
		newReplica, err = mgr.ramDiskCmdExec.ExpandVolumeReplica(replica, newCapacityBytes)
	} else {
		newReplica, err = mgr.cmdExec.ExpandVolumeReplica(replica, newCapacityBytes)
	}
	if err != nil {
		mgr.logger.WithError(err).Error("Failed to exec replica expansion.")
		return nil, err
	}

	return newReplica, nil
}

func (mgr *localVolumeReplicaManager) GetVolumeReplica(replica *localstoragev1alpha1.LocalVolumeReplica) (*localstoragev1alpha1.LocalVolumeReplica, error) {

	currentReplica, exists := mgr.registry.VolumeReplicas()[replica.Spec.VolumeName]
	if !exists {
		return nil, fmt.Errorf("not found")
	}
	return currentReplica, nil
}

func (mgr *localVolumeReplicaManager) TestVolumeReplica(replica *localstoragev1alpha1.LocalVolumeReplica) (*localstoragev1alpha1.LocalVolumeReplica, error) {
	if replica.Spec.Kind == localstoragev1alpha1.VolumeKindRAM {
		return mgr.ramDiskCmdExec.TestVolumeReplica(replica)
	}
	return mgr.cmdExec.TestVolumeReplica(replica)
}

func (mgr *localVolumeReplicaManager) ConsistencyCheck() {

	mgr.logger.Debug("Consistency Checking ...")

	replicaList := &localstoragev1alpha1.LocalVolumeReplicaList{}
	if err := mgr.lm.apiClient.List(context.TODO(), replicaList); err != nil {
		mgr.logger.Error("Failed to list volume replicas info from CRDs")
		return
	}
	crdReplicas := map[string]*localstoragev1alpha1.LocalVolumeReplica{}
	ramCRDReplicas := map[string]*localstoragev1alpha1.LocalVolumeReplica{}
	for i, item := range replicaList.Items {
		if item.Spec.NodeName != mgr.lm.nodeConf.Name {
			continue
		}
		if item.Spec.Kind == localstoragev1alpha1.VolumeKindRAM {
			ramCRDReplicas[item.Spec.VolumeName] = &replicaList.Items[i]
		} else {
			crdReplicas[item.Spec.VolumeName] = &replicaList.Items[i]
		}
	}

	if mgr.cmdExec != nil {
		mgr.cmdExec.ConsistencyCheck(crdReplicas)
	}
	if mgr.ramDiskCmdExec != nil {
		mgr.ramDiskCmdExec.ConsistencyCheck(ramCRDReplicas)
	}

	mgr.logger.Debug("Consistency check completed")
}
