package storage

import (
	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	log "github.com/sirupsen/logrus"
)

type localVolumeReplicaSnapshotManager struct {
	cmdExec LocalVolumeReplicaSnapshotExecutor
	ddExec  LocalVolumeReplicaSnapshotRestoreManager
	logger  *log.Entry
}

func newLocalVolumeReplicaSnapshotManager(lm *LocalManager) LocalVolumeReplicaSnapshotManager {
	return &localVolumeReplicaSnapshotManager{
		cmdExec: newLVMExecutor(lm),
		ddExec:  newDDExecutor(),
		logger:  log.WithField("Module", "NodeManager/LocalVolumeReplicaSnapshotManager"),
	}
}

func (snapMgr *localVolumeReplicaSnapshotManager) CreateVolumeReplicaSnapshot(replicaSnapshot *v1alpha1.LocalVolumeReplicaSnapshot) error {
	return snapMgr.cmdExec.CreateVolumeReplicaSnapshot(replicaSnapshot)
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
