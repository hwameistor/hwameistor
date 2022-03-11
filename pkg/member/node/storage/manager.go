package storage

import (
	apisv1alpha1 "github.com/hwameistor/local-storage/pkg/apis/hwameistor/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// LocalManager struct
type LocalManager struct {
	nodeConf    *apisv1alpha1.NodeConfig
	apiClient   client.Client
	poolManager LocalPoolManager
	//diskManager                LocalDiskManager
	volumeReplicaManager       LocalVolumeReplicaManager
	registry                   LocalRegistry
	addEmptyDiskToDefaultPools bool
}

// NewLocalManager creates a local manager
func NewLocalManager(nodeConf *apisv1alpha1.NodeConfig, cli client.Client) *LocalManager {
	lm := &LocalManager{
		nodeConf:                   nodeConf,
		apiClient:                  cli,
		addEmptyDiskToDefaultPools: true,
	}
	//lm.diskManager = newLocalDiskManager(lm)
	lm.registry = newLocalRegistry(lm)
	lm.volumeReplicaManager = newLocalVolumeReplicaManager(lm)
	lm.poolManager = newLocalPoolManager(lm)

	return lm
}

// Register for local storage
func (lm *LocalManager) Register() error {

	lm.volumeReplicaManager.ConsistencyCheck()

	return nil
}

// UpdateNodeForVolumeReplica updates LocalStorageNode for volume replica
func (lm *LocalManager) UpdateNodeForVolumeReplica(replica *apisv1alpha1.LocalVolumeReplica) {
	lm.registry.UpdateNodeForVolumeReplica(replica)
}

// Registry return singleton of local registry
func (lm *LocalManager) Registry() LocalRegistry {
	return lm.registry
}

// DiskManager gets disk manager
//func (lm *LocalManager) DiskManager() LocalDiskManager {
//	return lm.diskManager
//}

// PoolManager gets pool manager
func (lm *LocalManager) PoolManager() LocalPoolManager {
	return lm.poolManager
}

// VolumeReplicaManager gets volume replica manager
func (lm *LocalManager) VolumeReplicaManager() LocalVolumeReplicaManager {
	return lm.volumeReplicaManager
}

// NodeConfig gets node configuration
func (lm *LocalManager) NodeConfig() *apisv1alpha1.NodeConfig {
	return lm.nodeConf
}
