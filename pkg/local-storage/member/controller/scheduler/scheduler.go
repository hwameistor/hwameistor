package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	runtimecache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

const (
	AFFINITY         = "hwameistor.io/affinity-annotations"
	TOLERATION       = "hwameistor.io/tolerations-annotations"
	REPLICA_AFFINITY = "hwameistor.io/replica-affinity"
	SkipAffinity     = "hwameistor.io/skip-affinity-annotations"
)

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
func New(apiClient client.Client, informerCache runtimecache.Cache, maxHAVolumeCount int) apisv1alpha1.VolumeScheduler {
	return &scheduler{
		apiClient:           apiClient,
		informerCache:       informerCache,
		resourceCollections: newResources(maxHAVolumeCount, apiClient),
		logger:              log.WithField("Module", "Scheduler"),
	}
}

func (s *scheduler) Init() {

	s.resourceCollections.init(s.apiClient, s.informerCache)

}

// GetNodeCandidates gets available nodes for the volume, used by K8s scheduler
func (s *scheduler) GetNodeCandidates(vols []*apisv1alpha1.LocalVolume) (qualifiedNodes []*apisv1alpha1.LocalStorageNode) {
	logCtx := s.logger.WithFields(log.Fields{"vols": lvString(vols)})

	// show available node candidates for debug
	defer func() {
		logCtx.WithField("candidates", func() (ns string) {
			for _, node := range qualifiedNodes {
				ns = ns + "," + node.GetName()
			}
			return strings.TrimPrefix(ns, ",")
		}()).Debugf("matchable node candidates")
	}()

	// init all available nodes resources
	s.resourceCollections.syncTotalStorage()

	for _, vol := range vols {
		if nodes, err := s.resourceCollections.getNodeCandidates(vol); err != nil {
			logCtx.WithError(err).WithField("volumes", vol).Debugf("fail to getNodeCandidates")
			return qualifiedNodes
		} else {
			if len(qualifiedNodes) == 0 {
				qualifiedNodes = nodes
			} else {
				qualifiedNodes = unionSet(qualifiedNodes, nodes)
			}
		}
	}

	//Affinity and taint verification are enabled by default
	//The first creation of a single copy is still the default logic
	pvc := corev1.PersistentVolumeClaim{}
	if err := s.apiClient.Get(context.Background(), client.ObjectKey{Name: vols[0].Spec.PersistentVolumeClaimName, Namespace: vols[0].Spec.PersistentVolumeClaimNamespace}, &pvc); err != nil {
		log.Debugf("qualifiedNodes len is %d", len(qualifiedNodes))
		return qualifiedNodes
	}
	if pvc.Annotations == nil {
		pvc.Annotations = make(map[string]string)
	}
	if pvc.Annotations[SkipAffinity] == "true" {
		log.Debugf("Skip Affinity ,qualifiedNodes len is %d", len(qualifiedNodes))
		return qualifiedNodes
	}
	if vols[0].Annotations == nil {
		vols[0].Annotations = make(map[string]string)
	}

	if (vols[0].Spec.ReplicaNumber > 1 && vols[0].Annotations[REPLICA_AFFINITY] != "forbid") || vols[0].Annotations[REPLICA_AFFINITY] == "need" {
		nodes, err := s.filterNodeByTaint(vols[0], qualifiedNodes)
		if err != nil {
			logCtx.WithError(err).WithField("volumes", vols[0]).Debugf("fail to filterNodeByTaint")
		}
		nodes, err = s.filterNodeByAffinity(vols[0], nodes)
		if err != nil {
			logCtx.WithError(err).WithField("volumes", vols[0]).Debugf("fail to filterNodeByAffinity")
		}
		qualifiedNodes = nodes
	} else {
		log.Debugf("Ignore affinity and taint")
	}

	log.Debugf("qualifiedNodes len is %d", len(qualifiedNodes))
	return qualifiedNodes
}

func (s *scheduler) filterNodeByTaint(vol *apisv1alpha1.LocalVolume, qualifiedNodes []*apisv1alpha1.LocalStorageNode) ([]*apisv1alpha1.LocalStorageNode, error) {
	log.Debugf("filterNodeByTain start")
	pvc := corev1.PersistentVolumeClaim{}
	if err := s.apiClient.Get(context.Background(), client.ObjectKey{Name: vol.Spec.PersistentVolumeClaimName, Namespace: vol.Spec.PersistentVolumeClaimNamespace}, &pvc); err != nil {
		return qualifiedNodes, err
	}

	if pvc.Annotations == nil || pvc.Annotations[SkipAffinity] == "true" {
		log.Debugf("skip filterNodeByTaint, qualifiedNodes len is %d", len(qualifiedNodes))
		return qualifiedNodes, nil
	}

	var tolerations []corev1.Toleration
	if tolerationsStr, ok := pvc.Annotations[TOLERATION]; !ok {
		log.Debugf("no tolerations found in pvc annotations")
	} else {
		if err := json.Unmarshal([]byte(tolerationsStr), &tolerations); err != nil {
			log.WithError(err).Errorf("failed to parse tolerations")
			return qualifiedNodes, err
		}
	}

	// filter each qualified node and judge if taints exist on node but not tolerated by pvc
	filteredNodes := make([]*apisv1alpha1.LocalStorageNode, 0, len(qualifiedNodes))
	for _, qNode := range qualifiedNodes {
		node := corev1.Node{}
		if err := s.apiClient.Get(context.Background(), client.ObjectKey{Name: qNode.Name}, &node); err != nil {
			return qualifiedNodes, err
		}
		if match := canTaintBeTolerated(&node, tolerations); match {
			filteredNodes = append(filteredNodes, qNode)
		}
	}

	log.Debugf("filterNodeByTaint end, filteredNodes len is %d", len(filteredNodes))
	return filteredNodes, nil
}

func canTaintBeTolerated(node *corev1.Node, tolerations []corev1.Toleration) bool {
	// only consider "NoSchedule or NoExecute" taints
	nodeTaints := filterNodeTaints(node, corev1.TaintEffectNoSchedule, corev1.TaintEffectNoExecute)
	if len(nodeTaints) == 0 {
		return true
	}

	// compare each taint(node) and tolerations(pvc)
	for _, nodeTaint := range nodeTaints {
		taintTolerated := false
		for _, toleration := range tolerations {
			if doesTolerationAllowTaint(&nodeTaint, &toleration) {
				taintTolerated = true
				break
			}
		}

		if !taintTolerated {
			log.Debugf("Node %s has untolerated taint: %s", node.Name, nodeTaint.Key)
			return false
		}
	}

	return true
}

func filterNodeTaints(node *corev1.Node, effects ...corev1.TaintEffect) []corev1.Taint {
	var taints []corev1.Taint
	for _, taint := range node.Spec.Taints {
		for _, effect := range effects {
			if taint.Effect == effect {
				taints = append(taints, taint)
			}
		}
	}

	return taints
}

func hasNoScheduleTaint(node corev1.Node) bool {
	for _, taint := range node.Spec.Taints {
		if taint.Effect == corev1.TaintEffectNoSchedule || taint.Effect == corev1.TaintEffectNoExecute {
			return true
		}
	}
	return false
}

func doesTolerationAllowTaint(taint *corev1.Taint, toleration *corev1.Toleration) bool {
	if taint.Key == toleration.Key && taint.Effect == toleration.Effect {
		switch toleration.Operator {
		case corev1.TolerationOpEqual:
			return taint.Value == toleration.Value
		case corev1.TolerationOpExists:
			return true
		default:
			return false
		}
	}

	return false
}

func (s *scheduler) filterNodeByAffinity(vol *apisv1alpha1.LocalVolume, qualifiedNodes []*apisv1alpha1.LocalStorageNode) ([]*apisv1alpha1.LocalStorageNode, error) {
	log.Debugf("filterNodeByAffinityAndTain start")
	pvc := corev1.PersistentVolumeClaim{}
	if err := s.apiClient.Get(context.Background(), client.ObjectKey{Name: vol.Spec.PersistentVolumeClaimName, Namespace: vol.Spec.PersistentVolumeClaimNamespace}, &pvc); err != nil {
		return qualifiedNodes, err
	}

	if pvc.Annotations == nil || pvc.Annotations[SkipAffinity] == "true" {
		log.Debugf("skip filterNodeByAffinity, qualifiedNodes len is %d", len(qualifiedNodes))
		return qualifiedNodes, nil
	}

	var podAffinity corev1.Affinity
	if affinityStr, ok := pvc.Annotations[AFFINITY]; !ok {
		log.Debugf("no affinity found in pvc annotations")
	} else {
		if err := json.Unmarshal([]byte(affinityStr), &podAffinity); err != nil {
			log.WithError(err).Errorf("failed to parse affinity")
			return qualifiedNodes, err
		}
	}

	filteredNodes := make([]*apisv1alpha1.LocalStorageNode, 0, len(qualifiedNodes))
	for _, qNode := range qualifiedNodes {
		node := corev1.Node{}
		if err := s.apiClient.Get(context.Background(), client.ObjectKey{Name: qNode.Name}, &node); err != nil {
			return qualifiedNodes, err
		}
		if match := isAffinityMatch(&node, &podAffinity, s.apiClient); match {
			filteredNodes = append(filteredNodes, qNode)
		}
	}

	log.Debugf("filterNodeByAffinity end, filteredNodes len is %d", len(filteredNodes))
	return filteredNodes, nil
}

func isAffinityMatch(node *corev1.Node, affinity *corev1.Affinity, apiClient client.Client) bool {

	if affinity != nil {
		nodeLabels := node.GetLabels()
		if affinity.NodeAffinity != nil && affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
			requiredNodeSelector := affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution
			if requiredNodeSelector.NodeSelectorTerms != nil {
				for _, term := range requiredNodeSelector.NodeSelectorTerms {
					match := matchNodeSelectorTerm(term, nodeLabels)
					if !match {
						log.Debugf("Node affinity does not match schedule :%s ", node.Name)
						return false
					}
				}
			}
		}

		if affinity.PodAffinity != nil && affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
			podAffinityTerms := affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution
			if podAffinityTerms != nil {
				for _, term := range podAffinityTerms {
					match := matchPodAffinityTerm(term, node.Name, apiClient)
					if !match {
						log.Debugf("Node affinity does not match schedule :%s ", node.Name)
						return false
					}
				}
			}
		}

		if affinity.PodAntiAffinity != nil && affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
			podAntiAffinityTerms := affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution
			for _, term := range podAntiAffinityTerms {
				match := matchPodAntiAffinityTerm(term, node.Name, apiClient)
				if !match {
					log.Debugf("Node affinity does not match schedule :%s ", node.Name)
					return false
				}
			}
		}
	}

	return true
}

func matchNodeSelectorTerm(term corev1.NodeSelectorTerm, nodeLabels map[string]string) bool {
	if nodeLabels == nil {
		return len(term.MatchExpressions) == 0 && len(term.MatchFields) == 0
	}

	for _, expr := range term.MatchExpressions {
		nodeValue, ok := nodeLabels[expr.Key]
		if !ok {
			if expr.Operator == corev1.NodeSelectorOpDoesNotExist {
				continue
			} else {
				return false
			}
		}

		if !matchValue(expr.Operator, nodeValue, expr.Values) {
			return false
		}
	}

	for _, field := range term.MatchFields {
		nodeValue, ok := getField(nodeLabels, field.Key)
		if !ok {
			if field.Operator == corev1.NodeSelectorOpDoesNotExist {
				continue
			} else {
				return false
			}
		}

		if !matchValue(field.Operator, nodeValue, field.Values) {
			return false
		}
	}

	return true
}

func getField(labels map[string]string, key string) (string, bool) {
	return labels[key], len(labels[key]) > 0
}

func matchValue(operator corev1.NodeSelectorOperator, value string, values []string) bool {
	flag := false
	switch operator {
	case corev1.NodeSelectorOpIn:
		for _, val := range values {
			if value == val {
				return true
			}
		}
	case corev1.NodeSelectorOpNotIn:
		flag = true
		for _, val := range values {
			if value == val {
				return false
			}
		}
	case corev1.NodeSelectorOpExists:
		return true
	default:
		return false
	}

	return flag
}

func matchPodAffinityTerm(term corev1.PodAffinityTerm, nodeName string, apiClient client.Client) bool {
	if term.LabelSelector != nil {
		selector := metav1.LabelSelector{
			MatchLabels:      term.LabelSelector.MatchLabels,
			MatchExpressions: term.LabelSelector.MatchExpressions,
		}
		selectorSet, err := metav1.LabelSelectorAsSelector(&selector)
		if err != nil {
			log.Errorf("LabelSelectorAsSelector error: %v", err)
			return true
		}
		listOption := client.ListOptions{
			LabelSelector: selectorSet,
		}
		pods := &corev1.PodList{}
		err = apiClient.List(context.TODO(), pods, &listOption)
		if err != nil {
			log.Errorf("list pods error: %v", err)
			return true
		}
		return MatchAffinityByPods(nodeName, pods)
	}
	return true
}

func MatchAffinityByPods(nodeName string, pods *corev1.PodList) bool {
	for _, p := range pods.Items {
		if p.Spec.NodeName == nodeName {
			return true
		}
	}
	return false
}

func matchPodAntiAffinityTerm(term corev1.PodAffinityTerm, nodeName string, apiClient client.Client) bool {
	if term.LabelSelector != nil {
		selector := metav1.LabelSelector{
			MatchLabels:      term.LabelSelector.MatchLabels,
			MatchExpressions: term.LabelSelector.MatchExpressions,
		}
		selectorSet, err := metav1.LabelSelectorAsSelector(&selector)
		if err != nil {
			log.Errorf("LabelSelectorAsSelector error: %v", err)
			return true
		}

		listOption := client.ListOptions{
			LabelSelector: selectorSet,
		}

		pods := &corev1.PodList{}
		err = apiClient.List(context.TODO(), pods, &listOption)
		if err != nil {
			log.Errorf("list pods error: %v", err)
			return true
		}

		return MatchAntiAffinityByPods(nodeName, pods)
	}

	return true
}

func MatchAntiAffinityByPods(nodeName string, pods *corev1.PodList) bool {
	for _, p := range pods.Items {
		if p.Spec.NodeName == nodeName {
			return false
		}
	}
	return true
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
			logCtx.WithError(err).Error("Failed to get list of available sorted LocalStorageNodes")
			return nil, err
		}

		// when the volume is HA or migration happens, we need to filter nodes by taint and affinity
		if vol.Spec.ReplicaNumber > 1 {
			nodes, err = s.filterNodeByTaint(vol, nodes)
			if err != nil {
				logCtx.WithError(err).Error("Failed to filterNodeByTaint")
				return nil, err
			}

			nodes, err = s.filterNodeByAffinity(vol, nodes)
			if err != nil {
				logCtx.WithError(err).Error("Failed to filterNodeByAffinity")
				return nil, err
			}
		}

		logCtx.WithFields(log.Fields{"needs": neededNodeNumber, "candidates": len(nodes)}).Debug("try to allocate more replica")

		if len(nodes) < neededNodeNumber {
			logCtx.Error("No enough LocalStorageNodes available for LocalVolume")
			return nil, fmt.Errorf("no enough available node")
		}
		selectedNodes = nodes
	}

	return s.ConfigureVolumeOnAdditionalNodes(vol, selectedNodes)
}

func (s *scheduler) ConfigureVolumeOnAdditionalNodes(vol *apisv1alpha1.LocalVolume, nodes []*apisv1alpha1.LocalStorageNode) (*apisv1alpha1.VolumeConfig, error) {
	if len(nodes) == 0 && vol.Spec.Config != nil {
		if vol.Spec.Config.RequiredCapacityBytes < vol.Spec.RequiredCapacityBytes {
			newConfig := vol.Spec.Config.DeepCopy()
			newConfig.RequiredCapacityBytes = vol.Spec.RequiredCapacityBytes
			return newConfig, nil
		}
		return vol.Spec.Config, nil
	}

	// for the same volume, will always get the same ID
	resID, err := s.resourceCollections.getResourceIDForVolume(vol)
	if err != nil {
		return nil, err
	}

	conf := &apisv1alpha1.VolumeConfig{
		Version:     1,
		VolumeName:  vol.Name,
		Initialized: false,
		Replicas:    []apisv1alpha1.VolumeReplica{},
	}
	if vol.Spec.Config != nil {
		conf = vol.Spec.Config.DeepCopy()
		conf.Version++
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
	if !vol.Spec.Convertible {
		// always set to false for non-HA volume
		conf.Initialized = false
	}

	return conf, nil
}

func isLocalVolumeSameClass(lv1 *apisv1alpha1.LocalVolume, lv2 *apisv1alpha1.LocalVolume) bool {
	if lv1 == nil || lv2 == nil {
		return true
	}
	if lv1.Spec.PoolName != lv2.Spec.PoolName {
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

func lvString(vols []*apisv1alpha1.LocalVolume) (vs string) {
	for _, vol := range vols {
		vs = vs + "," + vol.GetName()
	}
	return strings.TrimPrefix(vs, ",")
}
