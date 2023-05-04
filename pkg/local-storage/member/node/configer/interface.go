package configer

import apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"

//go:generate mockgen -source=interface.go -destination=../../../member/node/configer/drbd_mock.go  -package=configer
type Configer interface {
	Run(stopCh <-chan struct{})
	// check if the config is updated with the new content
	IsConfigUpdated(replica *apisv1alpha1.LocalVolumeReplica, config apisv1alpha1.VolumeConfig) bool
	// create or update config for replica, and use replica.Status.StoragePath
	// create new device at replica.Status.DevicePath
	ApplyConfig(replica *apisv1alpha1.LocalVolumeReplica, config apisv1alpha1.VolumeConfig) error
	// Initialize do the initialization for volume
	Initialize(replica *apisv1alpha1.LocalVolumeReplica, config apisv1alpha1.VolumeConfig) error
	// delete config for replica, will remove the resource
	DeleteConfig(replica *apisv1alpha1.LocalVolumeReplica) error
	// check if there is a config on the replica
	HasConfig(replica *apisv1alpha1.LocalVolumeReplica) bool
	// GetReplicaHAState return replica state, synced, err
	GetReplicaHAState(replica *apisv1alpha1.LocalVolumeReplica) (state apisv1alpha1.HAState, err error)

	ConsistencyCheck(replicas []apisv1alpha1.LocalVolumeReplica)
}
