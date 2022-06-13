package scheduler

import (
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"
	runtimecache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apisv1alpha1 "github.com/hwameistor/local-storage/pkg/apis/hwameistor/v1alpha1"
)

// todo: design a better plugin register/enable
type scheduler struct {
	apiClient client.Client
	logger    *log.Entry

	// collections of the resources to be allocated
	resourceCollections *resources

	lock sync.Mutex

	informerCache runtimecache.Cache

	once sync.Once
}

// New a scheduler instance
func New(apiClient client.Client, informerCache runtimecache.Cache, maxHAVolumeCount int) apisv1alpha1.VolumeScheduler {
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

func (s *scheduler) initResources() {

	s.once.Do(func() {
		s.resourceCollections.apiClient = s.apiClient
		s.resourceCollections.initilizeResources()
	})

}

// GetNodeCandidates gets available nodes for the volume, used by K8s scheduler
func (s *scheduler) GetNodeCandidates(vols []*apisv1alpha1.LocalVolume) []*apisv1alpha1.LocalStorageNode {

	s.initResources()

	qualifiedNodes := []*apisv1alpha1.LocalStorageNode{}

	bigLVs := map[string]*apisv1alpha1.LocalVolume{}
	for _, lv := range vols {
		if !isLocalVolumeSameClass(bigLVs[lv.Spec.PoolName], lv) {
			return qualifiedNodes
		}
		bigLVs[lv.Spec.PoolName] = appendLocalVolume(bigLVs[lv.Spec.PoolName], lv)
	}

	for _, lv := range bigLVs {
		if nodes, err := s.resourceCollections.getNodeCandidates(lv); err != nil {
			return qualifiedNodes
		} else {
			if len(qualifiedNodes) == 0 {
				qualifiedNodes = nodes
			} else {
				qualifiedNodes = unionSet(qualifiedNodes, nodes)
			}
		}
	}

	return qualifiedNodes
}

func isLocalVolumeSameClass(lv1 *apisv1alpha1.LocalVolume, lv2 *apisv1alpha1.LocalVolume) bool {
	if lv1 == nil || lv2 == nil {
		return true
	}
	if lv1.Spec.PoolName != lv2.Spec.PoolName {
		return false
	}
	if lv1.Spec.ReplicaNumber != lv2.Spec.ReplicaNumber {
		return false
	}
	if lv1.Spec.Convertible != lv2.Spec.Convertible {
		return false
	}
	return true
}

func appendLocalVolume(bigLv *apisv1alpha1.LocalVolume, lv *apisv1alpha1.LocalVolume) *apisv1alpha1.LocalVolume {
	if bigLv == nil {
		return lv
	}
	if lv == nil {
		return bigLv
	}
	bigLv.Spec.RequiredCapacityBytes += lv.Spec.RequiredCapacityBytes
	return bigLv
}

func unionSet(strs1 []*apisv1alpha1.LocalStorageNode, strs2 []*apisv1alpha1.LocalStorageNode) []*apisv1alpha1.LocalStorageNode {
	strs := []*apisv1alpha1.LocalStorageNode{}
	for _, s1 := range strs1 {
		for _, s2 := range strs2 {
			if s1.Name == s2.Name {
				strs = append(strs, s1)
			}
		}
	}
	return strs
}

// Allocate schedule right nodes and generate volume config
func (s *scheduler) Allocate(vol *apisv1alpha1.LocalVolume) (*apisv1alpha1.VolumeConfig, error) {
	logCtx := s.logger.WithFields(log.Fields{"volume": vol.Name, "spec": vol.Spec})
	logCtx.Debug("Allocating resources for LocalVolume")

	// will allocate resources for volumes one by one
	s.lock.Lock()
	defer s.lock.Unlock()

	neededNodeNumber := int(vol.Spec.ReplicaNumber)
	if vol.Spec.Config != nil {
		neededNodeNumber -= len(vol.Spec.Config.Replicas)
	}

	var selectedNodes []*apisv1alpha1.LocalStorageNode
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

func (s *scheduler) generateConfig(vol *apisv1alpha1.LocalVolume, nodes []*apisv1alpha1.LocalStorageNode, resID int) *apisv1alpha1.VolumeConfig {
	conf := &apisv1alpha1.VolumeConfig{
		Version:     1,
		VolumeName:  vol.Name,
		Initialized: false,
		Replicas:    []apisv1alpha1.VolumeReplica{},
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
		replica := apisv1alpha1.VolumeReplica{
			ID:       freeIDs[nodeIDIndex],
			Hostname: nodes[nodeIndex].Spec.HostName,
			IP:       nodes[nodeIndex].Spec.StorageIP,
			Primary:  false,
		}
		if len(vol.Spec.Accessibility.Nodes) > 0 && replica.Hostname == vol.Spec.Accessibility.Nodes[0] {
			replica.Primary = true
		}
		conf.Replicas = append(conf.Replicas, replica)
		nodeIDIndex++
		nodeIndex++
	}
	if len(vol.Spec.Accessibility.Nodes) == 0 && len(conf.Replicas) > 0 {
		conf.Replicas[0].Primary = true
	}
	if len(conf.Replicas) < 2 {
		// always set to false for non-HA volume
		conf.Initialized = false
	}

	return conf
}
