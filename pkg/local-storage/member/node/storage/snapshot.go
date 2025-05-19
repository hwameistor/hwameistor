package storage

import (
	"errors"

	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	log "github.com/sirupsen/logrus"
)

type localVolumeReplicaSnapshotManager struct {
	cmdExec         LocalVolumeReplicaSnapshotExecutor
	ddExec          LocalVolumeReplicaSnapshotRestoreManager
	registry        LocalRegistry
	volumeValidator *validator
	logger          *log.Entry
}

func newLocalVolumeReplicaSnapshotManager(lm *LocalManager) LocalVolumeReplicaSnapshotManager {
	return &localVolumeReplicaSnapshotManager{
		cmdExec:         newLVMExecutor(lm),
		ddExec:          newDDExecutor(lm.snapshotRestoreTimeout),
		volumeValidator: newValidator(),
		registry:        lm.Registry(),
		logger:          log.WithField("Module", "NodeManager/LocalVolumeReplicaSnapshotManager"),
	}
}

func (snapMgr *localVolumeReplicaSnapshotManager) CreateVolumeReplicaSnapshot(replicaSnapshot *v1alpha1.LocalVolumeReplicaSnapshot) error {
	snapMgr.logger.Debugf("Creating VolumeReplicaSnapshot. name:%s, pool:%s, size:%d", replicaSnapshot.Name, replicaSnapshot.Spec.PoolName, replicaSnapshot.Spec.RequiredCapacityBytes)
	if err := snapMgr.volumeValidator.canCreateVolumeReplicaSnapshot(replicaSnapshot, snapMgr.registry); err == nil {
		// case 1: create snap if not exists
		err = snapMgr.cmdExec.CreateVolumeReplicaSnapshot(replicaSnapshot)
		if err != nil {
			snapMgr.logger.WithError(err).Error("Failed to exec replica snap create")
			return err
		}
	} else if errors.Is(err, ErrorReplicaSnapshotExists) {
		// case 2: snap already exists
		snapMgr.logger.Infof("Snap %s has already exists", replicaSnapshot.Spec.VolumeSnapshotName)
	} else {
		// case 3: failed to validate snap - shouldn't happen in normal case
		snapMgr.logger.WithError(err).Errorf("Failed to validate volume replica snap %s", replicaSnapshot.Name)
		return err
	}

	return nil
}

func (snapMgr *localVolumeReplicaSnapshotManager) DeleteVolumeReplicaSnapshot(replicaSnapshot *v1alpha1.LocalVolumeReplicaSnapshot) error {
	return snapMgr.cmdExec.DeleteVolumeReplicaSnapshot(replicaSnapshot)
}

func (snapMgr *localVolumeReplicaSnapshotManager) UpdateVolumeReplicaSnapshot(replicaSnapshot *v1alpha1.LocalVolumeReplicaSnapshot) (*v1alpha1.LocalVolumeReplicaSnapshotStatus, error) {
	//TODO implement me
	panic("implement me")
}

func (snapMgr *localVolumeReplicaSnapshotManager) GetVolumeReplicaSnapshot(replicaSnapshot *v1alpha1.LocalVolumeReplicaSnapshot) (*v1alpha1.LocalVolumeReplicaSnapshotStatus, error) {
	return snapMgr.cmdExec.GetVolumeReplicaSnapshot(replicaSnapshot)
}

func (snapMgr *localVolumeReplicaSnapshotManager) RollbackVolumeReplicaSnapshot(snapshotRestore *v1alpha1.LocalVolumeReplicaSnapshotRestore) error {
	return snapMgr.cmdExec.RollbackVolumeReplicaSnapshot(snapshotRestore)
}

func (snapMgr *localVolumeReplicaSnapshotManager) RestoreVolumeReplicaSnapshot(snapshotRestore *v1alpha1.LocalVolumeReplicaSnapshotRestore) error {
	return snapMgr.ddExec.RestoreVolumeReplicaSnapshot(snapshotRestore)
}
