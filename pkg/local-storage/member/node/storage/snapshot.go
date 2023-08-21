package storage

import (
	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	log "github.com/sirupsen/logrus"
)

type localVolumeReplicaSnapshotManager struct {
	cmdExec LocalVolumeReplicaSnapshotExecutor
	logger  *log.Entry
}

func newLocalVolumeReplicaSnapshotManager(lm *LocalManager) LocalVolumeReplicaSnapshotManager {
	return &localVolumeReplicaSnapshotManager{
		cmdExec: newLVMExecutor(lm),
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

func (snapMgr *localVolumeReplicaSnapshotManager) RollbackVolumeReplicaSnapshot(snapshotRecover *v1alpha1.LocalVolumeReplicaSnapshotRecover) error {
	return snapMgr.cmdExec.RollbackVolumeReplicaSnapshot(snapshotRecover)
}

func (snapMgr *localVolumeReplicaSnapshotManager) RestoreVolumeReplicaSnapshot(snapshotRecover *v1alpha1.LocalVolumeReplicaSnapshotRecover) error {
	return nil
}
