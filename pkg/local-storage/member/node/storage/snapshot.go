package storage

import (
	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	log "github.com/sirupsen/logrus"
)

type localVolumeReplicaSnapshotManager struct {
	cmdExec  LocalVolumeExecutor
	registry LocalRegistry
	logger   *log.Entry
	lm       *LocalManager
}

func newLocalVolumeReplicaSnapshotManager(lm *LocalManager) LocalVolumeReplicaSnapshotManager {
	return &localVolumeReplicaSnapshotManager{
		cmdExec:  newLVMExecutor(lm),
		registry: lm.Registry(),
		lm:       lm,
		logger:   log.WithField("Module", "NodeManager/LocalVolumeReplicaSnapshotManager"),
	}
}

func (l localVolumeReplicaSnapshotManager) CreateVolumeReplicaSnapshot(replicaSnapshot *v1alpha1.LocalVolumeReplicaSnapshot) (*v1alpha1.LocalVolumeReplicaSnapshot, error) {
	return nil, nil
}

func (l localVolumeReplicaSnapshotManager) DeleteVolumeReplicaSnapshot(replicaSnapshot *v1alpha1.LocalVolumeReplicaSnapshot) error {
	//TODO implement me
	panic("implement me")
}

func (l localVolumeReplicaSnapshotManager) ExpandVolumeReplicaSnapshot(replicaSnapshot *v1alpha1.LocalVolumeReplicaSnapshot) (*v1alpha1.LocalVolumeReplica, error) {
	//TODO implement me
	panic("implement me")
}

func (l localVolumeReplicaSnapshotManager) GetVolumeReplicaSnapshot(replicaSnapshot *v1alpha1.LocalVolumeReplicaSnapshot) (*v1alpha1.LocalVolumeReplica, error) {
	//TODO implement me
	panic("implement me")
}
