package storage

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

// LocalManager struct
type LocalManager struct {
	apiClient client.Client
	scheme    *runtime.Scheme
	recorder  record.EventRecorder
	nodeConf  *apisv1alpha1.NodeConfig

	registry                     LocalRegistry
	poolManager                  LocalPoolManager
	volumeReplicaManager         LocalVolumeReplicaManager
	volumeReplicaSnapshotManager LocalVolumeReplicaSnapshotManager
}

// NewLocalManager creates a local manager
func NewLocalManager(nodeConf *apisv1alpha1.NodeConfig, cli client.Client, scheme *runtime.Scheme, recorder record.EventRecorder) *LocalManager {
	lm := &LocalManager{
		nodeConf:  nodeConf,
		apiClient: cli,
		scheme:    scheme,
		recorder:  recorder,
	}
	lm.registry = newLocalRegistry(lm)
	lm.volumeReplicaManager = newLocalVolumeReplicaManager(lm)
	lm.volumeReplicaSnapshotManager = newLocalVolumeReplicaSnapshotManager(lm)
	lm.poolManager = newLocalPoolManager(lm)

	return lm
}

// Register for local storage
func (lm *LocalManager) Register() error {

	lm.volumeReplicaManager.ConsistencyCheck()

	lm.registry.Init()

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

// PoolManager gets pool manager
func (lm *LocalManager) PoolManager() LocalPoolManager {
	return lm.poolManager
}

// VolumeReplicaManager gets volume replica manager
func (lm *LocalManager) VolumeReplicaManager() LocalVolumeReplicaManager {
	return lm.volumeReplicaManager
}

// VolumeReplicaSnapshotManager gets volume replica snapshot manager
func (lm *LocalManager) VolumeReplicaSnapshotManager() LocalVolumeReplicaSnapshotManager {
	return lm.volumeReplicaSnapshotManager
}

// NodeConfig gets node configuration
func (lm *LocalManager) NodeConfig() *apisv1alpha1.NodeConfig {
	return lm.nodeConf
}
