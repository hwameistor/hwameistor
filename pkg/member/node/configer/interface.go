package configer

import localstoragev1alpha1 "github.com/hwameistor/local-storage/pkg/apis/localstorage/v1alpha1"

type Configer interface {
	Run(stopCh <-chan struct{})
	// check if the config is updated with the new content
	IsConfigUpdated(replica *localstoragev1alpha1.LocalVolumeReplica, config localstoragev1alpha1.VolumeConfig) bool
	// create or update config for replica, and use replica.Status.StoragePath
	// create new device at replica.Status.DevicePath
	ApplyConfig(replica *localstoragev1alpha1.LocalVolumeReplica, config localstoragev1alpha1.VolumeConfig) error
	// Initialize do the initalization for volume
	Initialize(replica *localstoragev1alpha1.LocalVolumeReplica, config localstoragev1alpha1.VolumeConfig) error
	// delete config for replica, will remove the resource
	DeleteConfig(replica *localstoragev1alpha1.LocalVolumeReplica) error
	// check if there is a config on the replica
	HasConfig(replica *localstoragev1alpha1.LocalVolumeReplica) bool
	// GetReplicaHAState return replica state, synced, err
	GetReplicaHAState(replica *localstoragev1alpha1.LocalVolumeReplica) (state localstoragev1alpha1.HAState, err error)

	ConsistencyCheck(replicas []localstoragev1alpha1.LocalVolumeReplica)
}
