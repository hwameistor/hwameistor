package storage

import (
	"context"
	"sync"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-storage/utils"
)

type localRegistry struct {
	apiClient client.Client

	disks    map[string]*apisv1alpha1.LocalDevice
	pools    map[string]*apisv1alpha1.LocalPool
	replicas map[string]*apisv1alpha1.LocalVolumeReplica

	lock   *sync.Mutex
	logger *log.Entry
	lm     *LocalManager
	// recorder is used to record events in the API server
	recorder record.EventRecorder
}

// newLocalRegistry creates a local storage registry
func newLocalRegistry(lm *LocalManager) LocalRegistry {
	return &localRegistry{
		apiClient: lm.apiClient,
		disks:     map[string]*apisv1alpha1.LocalDevice{},
		pools:     map[string]*apisv1alpha1.LocalPool{},
		replicas:  map[string]*apisv1alpha1.LocalVolumeReplica{},
		lock:      &sync.Mutex{},
		logger:    log.WithField("Module", "NodeManager/LocalRegistry"),
		lm:        lm,
		recorder:  lm.recorder,
	}
}

func (lr *localRegistry) Init() {
	_ = lr.SyncNodeResources()
}

// SyncNodeResources rebuild Node resource according to StoragePool info and sync it to LocalStorageNode object
func (lr *localRegistry) SyncNodeResources() error {
	lr.lock.Lock()
	defer lr.lock.Unlock()

	lr.logger.Debugf("Rebuilding Node Resource...")

	// rebuild localRegistry resources according StoragePool
	// 1. rebuild pools
	rebuildPools, err := lr.lm.PoolManager().GetPools()
	if err != nil {
		log.WithError(err).Error("Failed to get StoragePool info")
		return err
	}
	lr.pools = rebuildPools

	// 2. rebuild disks
	rebuildDisks := make(map[string]*apisv1alpha1.LocalDevice)
	for _, pool := range lr.pools {
		for _, disk := range pool.Disks {
			rebuildDisks[disk.DevPath] = disk.DeepCopy()
		}
	}
	lr.disks = rebuildDisks

	// 3. rebuild replicas
	rebuildReplicas, err := lr.lm.PoolManager().GetReplicas()
	if err != nil {
		lr.logger.WithError(err).Fatal("Failed to ConstructReplicas")
		return err
	}
	lr.replicas = rebuildReplicas

	// rebuild LocalStorageNode object
	err = lr.syncToNodeCRD()
	if err != nil {
		lr.logger.WithError(err).Fatal("Failed to sync resource to LocalStorageNode")
		return err
	}

	lr.logger.Debugf("Successed to Rebuild Node Resource")
	return nil
}

// UpdateNodeForVolumeReplica updates LocalStorageNode for volume replica
func (lr *localRegistry) UpdateNodeForVolumeReplica(replica *apisv1alpha1.LocalVolumeReplica) {

	logCtx := lr.logger.WithFields(log.Fields{"replica": replica.Name, "spec": replica.Spec, "status": replica.Status})

	if replica.Status.State == apisv1alpha1.VolumeReplicaStateReady || replica.Status.State == apisv1alpha1.VolumeReplicaStateNotReady {
		if err := lr.registerVolumeReplica(replica); err != nil {
			logCtx.WithError(err).Error("Failed to register VolumeReplica")
		} else {
			logCtx.Debug("Registered VolumeReplica successfully")
		}
	} else if replica.Status.State == apisv1alpha1.VolumeReplicaStateDeleted {
		if err := lr.deregisterVolumeReplica(replica); err != nil {
			logCtx.WithError(err).Error("Failed to deregister VolumeReplica")
		} else {
			logCtx.Debug("Deregistered VolumeReplica successfully")
		}
	}
}

// RegisterVolumeReplica registers a volume replica
func (lr *localRegistry) registerVolumeReplica(replica *apisv1alpha1.LocalVolumeReplica) error {
	logCtx := lr.logger.WithFields(log.Fields{"replica": replica.Name, "pool": replica.Spec.PoolName})
	logCtx.Debug("Registering volume replica into a storage pool")

	lr.lock.Lock()
	defer lr.lock.Unlock()

	pool := lr.pools[replica.Spec.PoolName]
	oldReplica, exists := lr.replicas[replica.Spec.VolumeName]
	if exists {
		if oldReplica.Status.AllocatedCapacityBytes == replica.Status.AllocatedCapacityBytes {
			logCtx.Debug("Skipped the volume replica registration because of no size change")
			return nil
		}
		// update volume replica registration data
		pool.FreeCapacityBytes += oldReplica.Status.AllocatedCapacityBytes
		pool.UsedCapacityBytes -= oldReplica.Status.AllocatedCapacityBytes
		pool.FreeVolumeCount++
		pool.UsedVolumeCount--
	}

	pool.FreeCapacityBytes -= replica.Status.AllocatedCapacityBytes
	pool.UsedCapacityBytes += replica.Status.AllocatedCapacityBytes
	pool.FreeVolumeCount--
	pool.UsedVolumeCount++

	pool.Volumes = utils.AddUniqueStringItem(pool.Volumes, replica.Spec.VolumeName)
	lr.replicas[replica.Spec.VolumeName] = replica.DeepCopy()

	return lr.syncToNodeCRD()
}

// DeregisterVolumeReplica deregisters a volume replica
func (lr *localRegistry) deregisterVolumeReplica(replica *apisv1alpha1.LocalVolumeReplica) error {
	logCtx := lr.logger.WithFields(log.Fields{"replica": replica.Name, "pool": replica.Spec.PoolName})
	logCtx.Debug("Deregistering volume replica from a storage pool")

	lr.lock.Lock()
	defer lr.lock.Unlock()

	pool := lr.pools[replica.Spec.PoolName]
	if _, exists := lr.replicas[replica.Spec.VolumeName]; !exists {
		logCtx.Info("Skipped the deregistration for un-registered volume replica")
		return nil
	}

	pool.FreeCapacityBytes += replica.Status.AllocatedCapacityBytes
	pool.UsedCapacityBytes -= replica.Status.AllocatedCapacityBytes
	pool.FreeVolumeCount++
	pool.UsedVolumeCount--

	pool.Volumes = utils.RemoveStringItem(pool.Volumes, replica.Spec.VolumeName)
	delete(lr.replicas, replica.Spec.VolumeName)

	return lr.syncToNodeCRD()
}

// syncToNodeCRD sync the status into Node CRD
func (lr *localRegistry) syncToNodeCRD() error {
	lr.logger.Debug("Syncing registry info to Node")

	node := &apisv1alpha1.LocalStorageNode{}
	if err := lr.apiClient.Get(context.TODO(), types.NamespacedName{Name: lr.lm.nodeConf.Name}, node); err != nil {
		lr.logger.WithError(err).Error("Failed to query Node")
		return nil
	}
	node.Status.State = apisv1alpha1.NodeStateReady
	node.Status.Pools = make(map[string]apisv1alpha1.LocalPool)
	for poolName, pool := range lr.pools {
		localPool := apisv1alpha1.LocalPool{
			Disks:   []apisv1alpha1.LocalDevice{},
			Volumes: []string{},
		}
		pool.DeepCopyInto(&localPool)
		node.Status.Pools[poolName] = localPool
	}

	return lr.apiClient.Status().Update(context.TODO(), node)
}

func (lr *localRegistry) rebuildRegistryDisks() error {
	lr.logger.Debug("rebuild localRegistry Disks")

	disks := make(map[string]*apisv1alpha1.LocalDevice)
	for _, pool := range lr.pools {
		for _, disk := range pool.Disks {
			disks[disk.DevPath] = disk.DeepCopy()
		}
	}
	if len(disks) > 0 {
		lr.disks = disks
	}

	return nil
}

func (lr *localRegistry) rebuildRegistryReplicas() error {
	lr.logger.Debug("rebuild localRegistry VolumeReplicas")

	replicas, err := lr.lm.PoolManager().GetReplicas()
	if err != nil {
		lr.logger.WithError(err).Fatal("Failed to ConstructReplicas")
		return err
	}
	if len(replicas) > 0 {
		lr.replicas = replicas
	}
	return nil
}

func (lr *localRegistry) Disks() map[string]*apisv1alpha1.LocalDevice {
	return lr.disks
}

func (lr *localRegistry) Pools() map[string]*apisv1alpha1.LocalPool {
	return lr.pools
}

func (lr *localRegistry) VolumeReplicas() map[string]*apisv1alpha1.LocalVolumeReplica {
	lr.showReplicaOnHost()
	return lr.replicas
}

func (lr *localRegistry) HasVolumeReplica(vr *apisv1alpha1.LocalVolumeReplica) bool {
	lr.showReplicaOnHost()
	_, has := lr.replicas[vr.Spec.VolumeName]
	return has
}

// UpdateCondition append current condition about LocalStorageNode, i.e. StorageExpandSuccess, StorageExpandFail, UnAvailable
func (lr *localRegistry) UpdateCondition(condition apisv1alpha1.LocalStorageNodeCondition) error {
	oldNode := &apisv1alpha1.LocalStorageNode{}
	if err := lr.apiClient.Get(context.TODO(), types.NamespacedName{Name: lr.lm.nodeConf.Name}, oldNode); err != nil {
		lr.logger.WithError(err).WithField("condition", condition).Error("Failed to query Node")
		return nil
	}
	switch condition.Type {
	case apisv1alpha1.StorageExpandFailure, apisv1alpha1.StorageUnAvailable:
		lr.recorder.Event(oldNode, v1.EventTypeWarning, string(condition.Type), condition.Message)
	case apisv1alpha1.StorageExpandSuccess, apisv1alpha1.StorageProgressing:
		lr.recorder.Event(oldNode, v1.EventTypeNormal, string(condition.Type), condition.Message)
	default:
		lr.recorder.Event(oldNode, v1.EventTypeNormal, string(condition.Type), condition.Message)
	}

	newNode := oldNode.DeepCopy()
	i, _ := GetStorageCondition(newNode.Status.Conditions, condition.Type)
	if i == -1 {
		newNode.Status.Conditions = append(newNode.Status.Conditions, condition)
	} else {
		newNode.Status.Conditions[i] = condition
	}

	return lr.apiClient.Status().Patch(context.TODO(), newNode, client.MergeFrom(oldNode))
}

// showReplicaOnHost debug func for now
func (lr *localRegistry) showReplicaOnHost() {
	lr.logger.WithFields(log.Fields{"node": lr.lm.NodeConfig().Name, "count": len(lr.replicas)}).Info("Show existing volumes on host")
	for volume := range lr.replicas {
		lr.logger.WithField("volume", volume).Infof("Existing volume replica on host")
	}
}

// GetStorageCondition extracts the provided condition from the given status and returns that.
// Returns nil and -1 if the condition is not present, and the index of the located condition.
func GetStorageCondition(conditions []apisv1alpha1.LocalStorageNodeCondition,
	conditionType apisv1alpha1.LocalStorageNodeConditionType) (int, *apisv1alpha1.LocalStorageNodeCondition) {
	for i, condition := range conditions {
		if condition.Type == conditionType {
			return i, &conditions[i]
		}
	}

	return -1, nil
}
