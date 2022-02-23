package scheduler

import (
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"
	runtimecache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	localstoragev1alpha1 "github.com/hwameistor/local-storage/pkg/apis/localstorage/v1alpha1"
)

// todo:
// 1. structure/architecture optimize, plugin register, default plugins.
// 		need so much more thinking!!!

// Scheduler interface
type Scheduler interface {
	Init()
	// schedule will schedule all replicas, and generate a valid VolumeConfig
	Allocate(vol *localstoragev1alpha1.LocalVolume) (*localstoragev1alpha1.VolumeConfig, error)

	GetNodeCandidates(vol *localstoragev1alpha1.LocalVolume) ([]*localstoragev1alpha1.LocalStorageNode, error)
}

// todo: design a better plugin register/enable
type scheduler struct {
	apiClient client.Client
	logger    *log.Entry

	// collections of the resources to be allocated
	resourceCollections *resources

	lock sync.Mutex

	informerCache runtimecache.Cache
}

// New a scheduler instance
func New(apiClient client.Client, informerCache runtimecache.Cache, maxHAVolumeCount int) Scheduler {
	return &scheduler{
		apiClient:           apiClient,
		informerCache:       informerCache,
		resourceCollections: newResources(maxHAVolumeCount),
		logger:              log.WithField("Module", "Scheduler"),
	}

}

func (s *scheduler) Init() {

	s.resourceCollections.init(s.apiClient, s.informerCache)

}

// GetNodeCandidates gets available nodes for the volume, used by K8s scheduler
func (s *scheduler) GetNodeCandidates(vol *localstoragev1alpha1.LocalVolume) ([]*localstoragev1alpha1.LocalStorageNode, error) {
	return s.resourceCollections.getNodeCandidates(vol)
}

// Allocate schedule right nodes and generate volume config
func (s *scheduler) Allocate(vol *localstoragev1alpha1.LocalVolume) (*localstoragev1alpha1.VolumeConfig, error) {
	logCtx := s.logger.WithFields(log.Fields{"volume": vol.Name, "spec": vol.Spec})
	logCtx.Debug("Allocating resources for LocalVolume")

	// will allocate resources for volumes one by one
	s.lock.Lock()
	defer s.lock.Unlock()

	neededNodeNumber := int(vol.Spec.ReplicaNumber)
	if vol.Spec.Config != nil {
		neededNodeNumber -= len(vol.Spec.Config.Replicas)
	}

	var selectedNodes []*localstoragev1alpha1.LocalStorageNode
	if neededNodeNumber > 0 {
		nodes, err := s.resourceCollections.getNodeCandidates(vol)
		if err != nil {
			logCtx.WithError(err).Error("Failed to get list of avaliable sorted LocalStorageNodes")
			return nil, err
		}

		logCtx.WithFields(log.Fields{"needs": neededNodeNumber, "candidates": len(nodes)}).Debug("try to allocate more replica")

		if len(nodes) < neededNodeNumber {
			logCtx.Error("No enough LocalStorageNodes available for LocalVolume")
			return nil, fmt.Errorf("no enough avaiable node")
		}
		selectedNodes = nodes
	}

	// for the same volume, will always get the same ID
	resID, err := s.resourceCollections.getResourceIDForVolume(vol)
	if err != nil {
		logCtx.WithError(err).Error("Failed to allocated a resource ID")
		return nil, err
	}

	return s.generateConfig(vol, selectedNodes, resID), nil
}

func (s *scheduler) generateConfig(vol *localstoragev1alpha1.LocalVolume, nodes []*localstoragev1alpha1.LocalStorageNode, resID int) *localstoragev1alpha1.VolumeConfig {
	conf := &localstoragev1alpha1.VolumeConfig{
		Version:     1,
		VolumeName:  vol.Name,
		Initialized: false,
		Replicas:    []localstoragev1alpha1.VolumeReplica{},
	}
	if vol.Spec.Config != nil {
		conf = vol.Spec.Config.DeepCopy()
	}
	conf.ResourceID = resID
	conf.RequiredCapacityBytes = vol.Spec.RequiredCapacityBytes
	conf.Convertible = vol.Spec.Convertible

	// for a volume, the ID of the replica shall not > vol.Spec.ReplicaNumber
	// and always set the first replica to primary
	freeIDs := make([]int, 0, vol.Spec.ReplicaNumber)
	usedIDs := make(map[int]bool)
	for _, replica := range conf.Replicas {
		usedIDs[replica.ID] = true
	}
	for id := 1; id <= int(vol.Spec.ReplicaNumber); id++ {
		if !usedIDs[id] {
			freeIDs = append(freeIDs, id)
		}
	}

	nodeIDIndex := 0
	nodeIndex := 0
	for i := len(conf.Replicas); i < int(vol.Spec.ReplicaNumber); i++ {
		replica := localstoragev1alpha1.VolumeReplica{
			ID:       freeIDs[nodeIDIndex],
			Hostname: nodes[nodeIndex].Spec.HostName,
			IP:       nodes[nodeIndex].Spec.StorageIP,
			Primary:  false,
		}
		if replica.Hostname == vol.Spec.Accessibility.Node {
			replica.Primary = true
		}
		conf.Replicas = append(conf.Replicas, replica)
		nodeIDIndex++
		nodeIndex++
	}
	if len(vol.Spec.Accessibility.Node) == 0 && len(conf.Replicas) > 0 {
		conf.Replicas[0].Primary = true
	}
	if len(conf.Replicas) < 2 {
		// always set to false for non-HA volume
		conf.Initialized = false
	}

	return conf
}
