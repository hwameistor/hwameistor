package scheduler

import (
	"container/heap"
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	runtimecache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-storage/utils"
)

const (
	HwameiStorSchedulerName = "hwameistor-scheduler"
)

type resources struct {
	apiClient client.Client

	// resourceID is for HA volumes only. Each HA volume must have a unique resourceID.
	// For DRBD, resourceID means the network port.
	// For all non-HA volumes, resourceID is set to '-1'
	allocatedResourceIDs map[string]int
	freeResourceIDList   []int
	maxHAVolumeCount     int

	allocatedStorages *storageCollection
	totalStorages     *storageCollection

	storageNodes map[string]*apisv1alpha1.LocalStorageNode

	podToPVCs map[string][]string
	pvcToPods map[string][]string

	pvcsMap map[string]*corev1.PersistentVolumeClaim
	scsMap  map[string]*storagev1.StorageClass

	lock sync.Mutex

	logger *log.Entry
}

func newResources(maxHAVolumeCount int, apiClient client.Client) *resources {
	return &resources{
		apiClient:            apiClient,
		logger:               log.WithField("Module", "Scheduler/Resources"),
		allocatedResourceIDs: make(map[string]int),
		freeResourceIDList:   make([]int, 0, maxHAVolumeCount),
		maxHAVolumeCount:     maxHAVolumeCount,
		allocatedStorages:    newStorageCollection(),
		totalStorages:        newStorageCollection(),
		storageNodes:         map[string]*apisv1alpha1.LocalStorageNode{},
		podToPVCs:            map[string][]string{},
		pvcToPods:            map[string][]string{},
		pvcsMap:              map[string]*corev1.PersistentVolumeClaim{},
		scsMap:               map[string]*storagev1.StorageClass{},
	}
}

func (r *resources) init(apiClient client.Client, informerCache runtimecache.Cache) {
	r.apiClient = apiClient

	// initialize the resources, e.g. resource IDs
	r.initilizeResources()

	nodeInformer, err := informerCache.GetInformer(context.TODO(), &apisv1alpha1.LocalStorageNode{})
	if err != nil {
		r.logger.WithError(err).Fatal("Failed to initiate informer for LocalStorageNode")
	}
	nodeInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    r.handleNodeAdd,
		UpdateFunc: r.handleNodeUpdate,
		DeleteFunc: r.handleNodeDelete,
	})

	volumeInformer, err := informerCache.GetInformer(context.TODO(), &apisv1alpha1.LocalVolume{})
	if err != nil {
		r.logger.WithError(err).Fatal("Failed to initiate informer for LocalVolume")
	}
	volumeInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: r.handleVolumeUpdate,
	})

	scInformer, err := informerCache.GetInformer(context.TODO(), &storagev1.StorageClass{})
	if err != nil {
		r.logger.WithError(err).Fatal("Failed to get informer for k8s StorageClass")
	}
	scInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    r.handleStorageClassAdd,
		UpdateFunc: r.handleStorageClassUpdate,
		DeleteFunc: r.handleStorageClassDelete,
	})

	pvcInformer, err := informerCache.GetInformer(context.TODO(), &corev1.PersistentVolumeClaim{})
	if err != nil {
		r.logger.WithError(err).Fatal("Failed to get informer for k8s PVC")
	}
	pvcInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    r.handlePVCAdd,
		UpdateFunc: r.handlePVCUpdate,
	})

	podInformer, err := informerCache.GetInformer(context.TODO(), &corev1.Pod{})
	if err != nil {
		r.logger.WithError(err).Fatal("Failed to get informer for k8s Pod")
	}
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    r.handlePodAdd,
		UpdateFunc: r.handlePodUpdate,
	})

}

// syncTotalStorage sync available LocalStorageNodes to storageNodes at now
func (r *resources) syncTotalStorage() {
	nodeList := &apisv1alpha1.LocalStorageNodeList{}
	if err := r.apiClient.List(context.TODO(), nodeList); err != nil {
		r.logger.WithError(err).Fatal("Failed to list LocalStorageNodes")
	}
	// initialize total capacity
	for i := range nodeList.Items {
		if nodeList.Items[i].Status.State == apisv1alpha1.NodeStateReady {
			r.addTotalStorage(&nodeList.Items[i])
		} else {
			r.logger.WithField("node", nodeList.Items[i].Name).
				WithField("state", nodeList.Items[i].Status.State).
				Debugf("delete node from totalStorage")
			r.delTotalStorage(&nodeList.Items[i])
		}
	}
}

func (r *resources) initilizeResources() {
	r.logger.Debug("Initializing resources ...")

	// show available nodes resources for debug
	defer func(nodes map[string]*apisv1alpha1.LocalStorageNode) {
		r.logger.Debugf("%d available resource: %v", len(nodes), func() (ns []string) {
			for _, node := range nodes {
				ns = append(ns, node.Name)
			}
			return
		}())
	}(r.storageNodes)

	volList := &apisv1alpha1.LocalVolumeList{}
	if err := r.apiClient.List(context.TODO(), volList); err != nil {
		r.logger.WithError(err).Fatal("Failed to list LocalVolumes")
	}

	// initialize resource IDs
	usedResourceIDMap := make(map[int]bool)
	for _, vol := range volList.Items {
		if vol.Spec.Config == nil || vol.Spec.Config.ResourceID == -1 || vol.Status.State == apisv1alpha1.VolumeStateDeleted {
			continue
		}
		if !vol.Spec.Config.Convertible && len(vol.Spec.Config.Replicas) < 2 {
			continue
		}
		r.allocatedResourceIDs[vol.Name] = vol.Spec.Config.ResourceID
		usedResourceIDMap[vol.Spec.Config.ResourceID] = true
	}
	for i := 0; i < r.maxHAVolumeCount; i++ {
		if !usedResourceIDMap[i] {
			r.freeResourceIDList = append(r.freeResourceIDList, i)
		}
	}

	// initialize total capacity
	r.syncTotalStorage()

	// initialize allocated capacity
	for i := range volList.Items {
		r.addAllocatedStorage(&volList.Items[i])
	}
}

// poolname -> volumes
func (r *resources) getAssociatedVolumes(vol *apisv1alpha1.LocalVolume) map[string][]*apisv1alpha1.LocalVolume {
	lvs := map[string][]*apisv1alpha1.LocalVolume{}
	pvcs := []string{}

	r.logger.WithFields(log.Fields{"pvcToPods": r.pvcToPods, "namespace": vol.Spec.PersistentVolumeClaimNamespace, "name": vol.Spec.PersistentVolumeClaimName}).Debug("Getting associated volumes")
	pods, exists := r.pvcToPods[NamespacedName(vol.Spec.PersistentVolumeClaimNamespace, vol.Spec.PersistentVolumeClaimName)]
	if !exists && len(vol.Spec.VolumeGroup) == 0 {
		lvs[vol.Spec.PoolName] = []*apisv1alpha1.LocalVolume{vol}
		return lvs
	}

	r.logger.Debugf("get %d pod(s) from cache", len(pods))

	if len(pods) > 0 {
		marks := map[string]bool{}
		for _, podNamespacedName := range pods {
			if ps, has := r.podToPVCs[podNamespacedName]; has {
				for _, pvcNamespacedName := range ps {
					if _, in := marks[pvcNamespacedName]; !in {
						marks[pvcNamespacedName] = true
						r.logger.Debugf("found associated PVC %s from pod %s", pvcNamespacedName, podNamespacedName)
						pvcs = append(pvcs, pvcNamespacedName)
					}
				}
			}
		}
	} else {
		lvg := &apisv1alpha1.LocalVolumeGroup{}
		if err := r.apiClient.Get(context.TODO(), types.NamespacedName{Name: vol.Spec.VolumeGroup}, lvg); err != nil {
			return lvs
		}
		for _, v := range lvg.Spec.Volumes {
			r.logger.Debugf("found associated PVC %s from VolumeGroup %s", NamespacedName(lvg.Spec.Namespace, v.PersistentVolumeClaimName), lvg.Name)
			pvcs = append(pvcs, NamespacedName(lvg.Spec.Namespace, v.PersistentVolumeClaimName))
		}
	}

	r.logger.Debugf("found %d associated PVC(s)", len(pvcs))

	for _, pvcNamespacedName := range pvcs {
		pvc, exists := r.pvcsMap[pvcNamespacedName]
		if !exists || pvc == nil {
			pvcNamespace, pvcName := GetNamespaceAndName(pvcNamespacedName)
			r.logger.WithFields(log.Fields{"pvc": pvcName, "namespace": pvcNamespace}).Debugf("not found in the map")
			pvcInCluster := &corev1.PersistentVolumeClaim{}
			if err := r.apiClient.Get(context.TODO(), types.NamespacedName{Namespace: pvcNamespace, Name: pvcName}, pvcInCluster); err != nil {
				r.logger.WithError(err).WithFields(log.Fields{"pvc": pvcName, "namespace": pvcNamespace}).Errorf("get pvc in cluster err")
				continue
			}
			r.cachePVC(pvcInCluster)
			pvc = pvcInCluster
		}
		sc, exists := r.scsMap[*pvc.Spec.StorageClassName]
		if !exists || sc == nil {
			r.logger.WithField("sc", sc.Name).Debugf("not found in the map")
			continue
		}

		poolName, err := utils.BuildStoragePoolName(
			sc.Parameters[apisv1alpha1.VolumeParameterPoolClassKey],
			)
		if err != nil {
			r.logger.WithError(err).Errorf("build storagepoolname err")
			return lvs
		}
		if _, exists := lvs[poolName]; !exists {
			lvs[poolName] = []*apisv1alpha1.LocalVolume{}
		}

		lv := &apisv1alpha1.LocalVolume{}
		lv.Spec.PoolName = poolName
		storage := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
		lv.Spec.RequiredCapacityBytes = storage.Value()
		replica, _ := strconv.Atoi(sc.Parameters[apisv1alpha1.VolumeParameterReplicaNumberKey])
		lv.Spec.ReplicaNumber = int64(replica)
		lvs[poolName] = append(lvs[poolName], lv)
		r.logger.Debugf("adding associated LV(capacity: %d) to pool %s, current %d volume(s)", lv.Spec.RequiredCapacityBytes, poolName, len(lvs[poolName]))
	}

	return lvs
}

func (r *resources) predicate(vol *apisv1alpha1.LocalVolume, nodeName string) error {
	r.logger.WithFields(log.Fields{"namespace": vol.Spec.PersistentVolumeClaimNamespace, "pvc": vol.Spec.PersistentVolumeClaimName, "node": nodeName}).Debug("Predicting a volume against a node")
	if _, ok := r.storageNodes[nodeName]; !ok {
		r.logger.WithField("node", nodeName).Error("Storage node doesn't exist")
		return fmt.Errorf("storage node %s not exists", nodeName)
	}

	vols := r.getAssociatedVolumes(vol)
	if len(vols) == 0 {
		r.logger.Error("Not found associated volumes")
		return fmt.Errorf("not found associated volumes")
	}
	for poolName, lvs := range vols {
		volumeMaxCapacityBytes := int64(0)
		requiredVolumeCount := len(lvs)
		requiredCapacityBytes := int64(0)

		r.logger.Debugf("found %d volume(s) in pool %s", len(lvs), poolName)

		for _, lv := range lvs {
			requiredCapacityBytes += lv.Spec.RequiredCapacityBytes
			r.logger.Debugf("adding requiredCapacity %d to pool %s, current requiredCapacity %d", lv.Spec.RequiredCapacityBytes, poolName, requiredCapacityBytes)
			if lv.Spec.RequiredCapacityBytes > volumeMaxCapacityBytes {
				volumeMaxCapacityBytes = lv.Spec.RequiredCapacityBytes
			}
		}

		totalPool := r.totalStorages.pools[poolName]
		allocatedPool := r.allocatedStorages.pools[poolName]

		if requiredCapacityBytes > totalPool.capacities[nodeName]-allocatedPool.capacities[nodeName] {
			r.logger.WithFields(log.Fields{"pool": poolName,
				"requireCapacityBytes":   requiredCapacityBytes,
				"totalPoolCapacityBytes": totalPool.capacities[nodeName],
				"allocatedCapacityBytes": allocatedPool.capacities[nodeName]}).Error("No enough capacity")
			return fmt.Errorf("not enough capacity in pool %s", poolName)
		}
		if totalPool.volumeCount[nodeName] < allocatedPool.volumeCount[nodeName]+int64(requiredVolumeCount) {
			r.logger.WithField("pool", poolName).Error("No enough volume count")
			return fmt.Errorf("not enough free volume count in pool %s", poolName)
		}
	}

	return nil
}

// Score calculate node socre for this volume
func (r *resources) Score(vol *apisv1alpha1.LocalVolume, nodeName string) (score int64, err error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	return r.score(vol, nodeName)
}

func (r *resources) score(vol *apisv1alpha1.LocalVolume, nodeName string) (int64, error) {
	r.logger.WithFields(log.Fields{"namespace": vol.Spec.PersistentVolumeClaimNamespace, "pvc": vol.Spec.PersistentVolumeClaimName, "node": nodeName}).Debug("Scoring a volume against a node")
	if _, ok := r.storageNodes[nodeName]; !ok {
		r.logger.WithField("node", nodeName).Error("Storage node doesn't exist")
		return 0, fmt.Errorf("storage node %s not exists", nodeName)
	}

	var score int64 = 0
	vols := r.getAssociatedVolumes(vol)
	if len(vols) == 0 {
		r.logger.Error("Not found associated volumes")
		return 0, fmt.Errorf("not found associated volumes")
	}
	for poolName, lvs := range vols {
		requiredCapacityBytes := int64(0)
		for _, lv := range lvs {
			requiredCapacityBytes += lv.Spec.RequiredCapacityBytes
		}
		totalPool := r.totalStorages.pools[poolName]
		allocatedPool := r.allocatedStorages.pools[poolName]
		score += int64(1-float64(requiredCapacityBytes)/float64(totalPool.capacities[nodeName]-allocatedPool.capacities[nodeName])) * 100
	}

	return score / int64(len(vols)), nil
}

func (r *resources) getNodeCandidates(vol *apisv1alpha1.LocalVolume) ([]*apisv1alpha1.LocalStorageNode, error) {
	logCtx := r.logger.WithFields(log.Fields{"volume": fmt.Sprintf("%s/%s[%s]", vol.Name, vol.Spec.PersistentVolumeClaimName, vol.Spec.PersistentVolumeClaimNamespace)})
	logCtx.Debug("getting available nodes for LocalVolumes")

	r.lock.Lock()
	defer r.lock.Unlock()

	// step 1. filter out the nodes which have already been allocated
	excludedNodes := map[string]bool{}
	if vol.Spec.Config != nil {
		for _, rep := range vol.Spec.Config.Replicas {
			excludedNodes[rep.Hostname] = true
			// show excludedNodes for debug
			logCtx.WithField("node", rep.Hostname).Debug("node will not be added to candidates, because of founding a exist volume replica allocated on this node")
		}
	}

	candidates := []*apisv1alpha1.LocalStorageNode{}

	ctx := context.TODO()
	lvg := &apisv1alpha1.LocalVolumeGroup{}
	if err := r.apiClient.Get(ctx, types.NamespacedName{Name: vol.Spec.VolumeGroup}, lvg); err == nil {
		// found LVG, check if firstly, if nodes are specified in LVG, return them
		volReplicaNumber := 0
		if vol.Spec.Config != nil {
			volReplicaNumber = len(vol.Spec.Config.Replicas)
		}
		if len(lvg.Spec.Accessibility.Nodes) == int(vol.Spec.ReplicaNumber) && len(lvg.Spec.Accessibility.Nodes) > volReplicaNumber {
			// get candidate nodes from LVG
			for _, nn := range lvg.Spec.Accessibility.Nodes {
				if !excludedNodes[nn] {
					candidates = append(candidates, r.storageNodes[nn])
				}
			}
			return candidates, nil
		}
	} else if !errors.IsNotFound(err) {
		logCtx.WithError(err).Error("Failed to check LVG")
		return candidates, err
	}

	// not found LVG or new node is not specified in LVG

	// step 2. check the required nodes firstly, if not satisfied, return error immediately
	for _, nn := range vol.Spec.Accessibility.Nodes {
		if len(nn) > 0 && !excludedNodes[nn] {
			if err := r.predicate(vol, nn); err != nil {
				logCtx.WithField("node", nn).WithError(err).Error("predicate accessibility node fail")
				return nil, err
			}
			candidates = append(candidates, r.storageNodes[nn])
			logCtx.WithField("node", nn).Debug("Adding a candidate")
			excludedNodes[nn] = true
		}
	}

	// step 3. check the rest of all nodes for the volume replica, and queue the qualified by the available storage space
	pq := make(PriorityQueue, 0)
	for _, node := range r.storageNodes {
		r.logger.WithField("node", node.Name).Debug("filtering a node")
		if excludedNodes[node.Name] {
			continue
		}

		if err := r.predicate(vol, node.Name); err != nil {
			logCtx.WithError(err).WithField("node", node.Name).Debug("filter out a candidate node for predicate fail")
			continue
		}
		priority, err := r.score(vol, node.Name)
		if err != nil {
			logCtx.WithError(err).WithField("node", node.Name).Debug("filter out a candidate node for score fail")
			continue
		}
		heap.Push(
			&pq,
			&PriorityItem{
				name:     node.Name,
				priority: priority,
				index:    pq.Len(),
			},
		)
	}

	for pq.Len() > 0 {
		item := heap.Pop(&pq).(*PriorityItem)
		candidates = append(candidates, r.storageNodes[item.name])
		r.logger.WithFields(log.Fields{"node": item.name, "total": pq.Len()}).Debug("Adding a candidate")
	}

	return candidates, nil
}

func (r *resources) getResourceIDForVolume(vol *apisv1alpha1.LocalVolume) (int, error) {
	if vol.Spec.ReplicaNumber <= 2 && !vol.Spec.Convertible {
		// try to recycle the resource ID in case of this volume is HA before
		r.recycleResourceID(vol)
		// for non-HA volume, resource ID is -1
		return -1, nil
	}

	return r.allocateResourceID(vol.Name)
}

func (r *resources) allocateResourceID(volName string) (int, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	// check if the volume already got resource ID allocated before
	if resID, exists := r.allocatedResourceIDs[volName]; exists {
		return resID, nil
	}

	if len(r.allocatedResourceIDs) >= r.maxHAVolumeCount {
		return -1, fmt.Errorf("can't allocate reourceID, exceeds max volume count")
	}

	if len(r.freeResourceIDList) > 0 {
		resID := r.freeResourceIDList[0]
		r.freeResourceIDList = r.freeResourceIDList[1:]
		r.allocatedResourceIDs[volName] = resID
		return resID, nil
	}

	return -1, fmt.Errorf("can't allocate resource ID")
}

func (r *resources) recycleResourceID(vol *apisv1alpha1.LocalVolume) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if id, exists := r.allocatedResourceIDs[vol.Name]; exists {
		delete(r.allocatedResourceIDs, vol.Name)
		r.freeResourceIDList = append(r.freeResourceIDList, id)
	}
}

func (r *resources) addAllocatedStorage(vol *apisv1alpha1.LocalVolume) {
	if vol.Spec.Config == nil || len(vol.Spec.Config.Replicas) == 0 {
		return
	}

	r.logger.Infof("add allocated storage for %v", vol.Name)

	r.lock.Lock()
	defer r.lock.Unlock()

	for _, replica := range vol.Spec.Config.Replicas {
		// for capacity
		if _, exists := r.allocatedStorages.pools[vol.Spec.PoolName].capacities[replica.Hostname]; !exists {
			r.allocatedStorages.pools[vol.Spec.PoolName].capacities[replica.Hostname] = 0
		}
		r.allocatedStorages.pools[vol.Spec.PoolName].capacities[replica.Hostname] += vol.Spec.Config.RequiredCapacityBytes

		// for volume count
		if _, exists := r.allocatedStorages.pools[vol.Spec.PoolName].volumeCount[replica.Hostname]; !exists {
			r.allocatedStorages.pools[vol.Spec.PoolName].volumeCount[replica.Hostname] = 0
		}
		r.allocatedStorages.pools[vol.Spec.PoolName].volumeCount[replica.Hostname]++
	}
}

func (r *resources) recycleAllocatedStorage(vol *apisv1alpha1.LocalVolume) {
	if vol.Spec.Config == nil || len(vol.Spec.Config.Replicas) == 0 {
		return
	}

	r.logger.Infof("recycle allocated storage for %v", vol.Name)

	r.lock.Lock()
	defer r.lock.Unlock()

	for _, replica := range vol.Spec.Config.Replicas {
		// for capacity
		if _, exists := r.allocatedStorages.pools[vol.Spec.PoolName].capacities[replica.Hostname]; !exists {
			r.allocatedStorages.pools[vol.Spec.PoolName].capacities[replica.Hostname] = 0
		}
		r.allocatedStorages.pools[vol.Spec.PoolName].capacities[replica.Hostname] -= vol.Spec.Config.RequiredCapacityBytes

		// for volume count
		if _, exists := r.allocatedStorages.pools[vol.Spec.PoolName].volumeCount[replica.Hostname]; !exists {
			r.allocatedStorages.pools[vol.Spec.PoolName].volumeCount[replica.Hostname] = 0
		}
		r.allocatedStorages.pools[vol.Spec.PoolName].volumeCount[replica.Hostname]--
	}

}

func (r *resources) addTotalStorage(node *apisv1alpha1.LocalStorageNode) {
	r.lock.Lock()
	defer r.lock.Unlock()

	for _, pool := range node.Status.Pools {
		r.totalStorages.pools[pool.Name].capacities[node.Name] = pool.TotalCapacityBytes
		r.totalStorages.pools[pool.Name].volumeCount[node.Name] = pool.TotalVolumeCount
	}
	r.storageNodes[node.Name] = node
}

func (r *resources) delTotalStorage(node *apisv1alpha1.LocalStorageNode) {
	r.lock.Lock()
	defer r.lock.Unlock()

	for _, pool := range node.Status.Pools {
		delete(r.totalStorages.pools[pool.Name].capacities, node.Name)
		delete(r.totalStorages.pools[pool.Name].volumeCount, node.Name)
	}
	delete(r.storageNodes, node.Name)
}

func (r *resources) handleNodeAdd(obj interface{}) {
	node := obj.(*apisv1alpha1.LocalStorageNode)
	r.addTotalStorage(node)
}

func (r *resources) handleNodeUpdate(_, newObj interface{}) {
	node := newObj.(*apisv1alpha1.LocalStorageNode)
	r.addTotalStorage(node)
}

func (r *resources) handleNodeDelete(obj interface{}) {
	node := obj.(*apisv1alpha1.LocalStorageNode)
	r.delTotalStorage(node)

}

func (r *resources) handleVolumeUpdate(oldObj, newObj interface{}) {
	oVol := oldObj.(*apisv1alpha1.LocalVolume)
	nVol := newObj.(*apisv1alpha1.LocalVolume)

	// 1. calculate allocated capacity according to LocalVolume.Spec.Config
	// recycle old volume
	r.recycleAllocatedStorage(oVol)

	// 2. recycle resource ID when LocalVolume is deleted
	if nVol.Status.State == apisv1alpha1.VolumeStateDeleted {
		r.recycleResourceID(nVol)
	} else {
		r.addAllocatedStorage(nVol)
	}
	if nVol.Spec.Config == nil {
		r.recycleResourceID(nVol)
	} else if !nVol.Spec.Config.Convertible && len(nVol.Spec.Config.Replicas) < 2 {
		r.recycleResourceID(nVol)
	}
}

func (r *resources) handlePodUpdate(oldObj, newObj interface{}) {
	r.handlePodAdd(newObj)
}

func (r *resources) handlePodAdd(obj interface{}) {

	pod := obj.(*corev1.Pod)
	if pod.Spec.SchedulerName != HwameiStorSchedulerName {
		return
	}

	r.logger.WithFields(log.Fields{"namespace": pod.Namespace, "name": pod.Name}).Debug("Added a Pod")

	r.lock.Lock()
	defer r.lock.Unlock()

	r.cleanupPod(pod.Namespace, pod.Name)

	if pod.DeletionTimestamp != nil {
		return
	}

	pvcNames := []string{}
	for _, vol := range pod.Spec.Volumes {
		if vol.PersistentVolumeClaim == nil {
			continue
		}
		pvcNamespacedName := NamespacedName(pod.Namespace, vol.PersistentVolumeClaim.ClaimName)
		if _, exists := r.pvcsMap[pvcNamespacedName]; exists {
			pvcNames = append(pvcNames, pvcNamespacedName)
		}
	}

	r.addPodAndPVCs(pod.Namespace, pod.Name, pvcNames)
}

func (r *resources) addPodAndPVCs(namespace string, podName string, pvcNamespacedNames []string) {
	podNamespacedName := NamespacedName(namespace, podName)
	r.podToPVCs[podNamespacedName] = []string{}
	for _, pvcNamespacedName := range pvcNamespacedNames {
		r.podToPVCs[podNamespacedName] = append(r.podToPVCs[podNamespacedName], pvcNamespacedName)
		if _, exists := r.pvcToPods[pvcNamespacedName]; !exists {
			r.pvcToPods[pvcNamespacedName] = []string{}
		}
		r.pvcToPods[pvcNamespacedName] = append(r.pvcToPods[pvcNamespacedName], podNamespacedName)
	}
}

func (r *resources) cleanupPod(namespace string, podName string) {
	// pod is to be deleted, clean it up
	pName := NamespacedName(namespace, podName)
	if pvcs, exists := r.podToPVCs[pName]; exists {
		for _, pvc := range pvcs {
			if pNames, has := r.pvcToPods[pvc]; has {
				items := utils.RemoveStringItem(pNames, pName)
				if len(items) > 0 {
					r.pvcToPods[pvc] = items
				} else {
					delete(r.pvcToPods, pvc)
				}
			}
		}
	}
	delete(r.podToPVCs, pName)
}

func (r *resources) handlePVCAdd(obj interface{}) {
	pvc := obj.(*corev1.PersistentVolumeClaim)

	r.cachePVC(pvc)
}

func (r *resources) cachePVC(pvc *corev1.PersistentVolumeClaim) {
	if pvc.Spec.StorageClassName == nil {
		return
	}
	if _, exists := r.scsMap[*pvc.Spec.StorageClassName]; !exists {
		r.logger.WithFields(log.Fields{"pvc": pvc.Name, "namespace": pvc.Namespace, "storageclassname": pvc.Spec.StorageClassName}).Errorf("storageclass not found in map")
		return
	}

	pvcNamespacedName := NamespacedName(pvc.Namespace, pvc.Name)
	if pvc.DeletionTimestamp != nil {
		delete(r.pvcsMap, pvcNamespacedName)
		return
	}

	r.pvcsMap[pvcNamespacedName] = pvc
	r.logger.WithField("pvc", pvcNamespacedName).Debugf("added into cache map")
}

func (r *resources) handlePVCUpdate(oldObj, newObj interface{}) {
	r.handlePVCAdd(newObj)
}

func (r *resources) handleStorageClassAdd(obj interface{}) {
	sc := obj.(*storagev1.StorageClass)

	if sc.Provisioner == apisv1alpha1.CSIDriverName {
		r.scsMap[sc.Name] = sc
	}
}

func (r *resources) handleStorageClassUpdate(oldObj, newObj interface{}) {
	r.handleStorageClassAdd(newObj)
}

func (r *resources) handleStorageClassDelete(obj interface{}) {
	sc := obj.(*storagev1.StorageClass)

	if sc.Provisioner == apisv1alpha1.CSIDriverName {
		delete(r.scsMap, sc.Name)
	}
}

func NamespacedName(namespace, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}

func GetNamespaceAndName(namespacedName string) (string, string) {
	items := strings.Split(namespacedName, "/")
	if len(items) == 0 {
		return "", ""
	}
	if len(items) == 1 {
		return "", items[0]
	}
	return items[0], items[1]
}
