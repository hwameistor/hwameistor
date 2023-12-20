package graph

import (
	"errors"
	"fmt"
	"github.com/dominikbraun/graph"
	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"strconv"
	"strings"
)

const separator = "_"

var notFoundError = fmt.Errorf("not found")

type ResourceType = string

const (
	Disk   ResourceType = "disk"
	Volume ResourceType = "volume"
	Pod    ResourceType = "pod"
	Pool   ResourceType = "pool"
	Node   ResourceType = "node"
	Pvc    ResourceType = "pvc"
	PV     ResourceType = "pv"
)

func (t *Topology[K, T]) GetPoolUnderLocalDisk(nodeName, diskPath string) (string, error) {
	diskKey := GenerateDiskKey(nodeName, diskPath)
	var poolKey string
	if err := graph.BFS(t.Graph, diskKey, func(vertex string) bool {
		if isPoolResource(vertex) {
			poolKey = vertex
			return true
		}
		return false
	}); err != nil {
		return "", err
	}

	_, poolInfo, err := t.Graph.VertexWithProperties(graph.StringHash(poolKey))
	if err != nil {
		return "", err
	}

	if poolInfo.Attributes == nil {
		return "", notFoundError
	}
	if _, ok := poolInfo.Attributes["name"]; !ok {
		return "", notFoundError
	}

	return poolInfo.Attributes["name"], nil
}

func (t *Topology[K, T]) GetVolumesUnderStoragePool(nodeName, poolName string) ([]string, error) {
	poolKey := GeneratePoolKey(nodeName, poolName)
	var volumeKeys []string
	if err := graph.BFS(t.Graph, poolKey, func(vertex string) bool {
		if isVolumeResource(vertex) {
			volumeKeys = append(volumeKeys, vertex)
		}
		return false
	}); err != nil {
		return nil, err
	}

	getVolumeName := func(volumeKey string) (string, error) {
		_, volumeInfo, err := t.Graph.VertexWithProperties(graph.StringHash(volumeKey))
		if err != nil {
			return "", err
		}

		if volumeInfo.Attributes == nil {
			return "", notFoundError
		}
		if _, ok := volumeInfo.Attributes["name"]; !ok {
			return "", notFoundError
		}
		return volumeInfo.Attributes["name"], nil
	}

	var volumes []string
	for _, volumeKey := range volumeKeys {
		if volume, err := getVolumeName(volumeKey); err != nil {
			return nil, err
		} else {
			volumes = append(volumes, volume)
		}
	}

	return volumes, nil
}

func (t *Topology[K, T]) GetPodsUnderLocalVolume(nodeName, volumeName string) ([]string, error) {
	volumeKey := GenerateLVKey(volumeName)
	var podKeys []string
	if err := graph.BFS(t.Graph, volumeKey, func(vertex string) bool {
		if isPodResource(vertex) {
			podKeys = append(podKeys, vertex)
		}
		return false
	}); err != nil {
		return nil, err
	}

	getPodName := func(podKey string) (string, error) {
		_, podInfo, err := t.Graph.VertexWithProperties(graph.StringHash(podKey))
		if err != nil {
			return "", err
		}

		if podInfo.Attributes == nil {
			return "", notFoundError
		}
		if _, ok := podInfo.Attributes["name"]; !ok {
			return "", notFoundError
		}
		return podInfo.Attributes["name"], nil
	}

	var pods []string
	for _, podKey := range podKeys {
		if podName, err := getPodName(podKey); err != nil {
			return nil, err
		} else {
			pods = append(pods, podName)
		}
	}

	return pods, nil
}

type Topology[K string, T string] struct {
	logger *log.Entry
	graph.Graph[string, string]
}

func NewTopologyStore() Topology[string, string] {
	return Topology[string, string]{
		logger: log.WithField("Module", "TopologyStore"),
		Graph:  graph.New(graph.StringHash, graph.Directed(), graph.Acyclic()),
	}
}

// AddStorageNode inserts a Vertex as the roots of the node
func (t *Topology[K, T]) AddStorageNode(node *v1alpha1.LocalStorageNode) error {
	nodeKey := GenerateNodeKey(node.Name)
	if t.IsVertexExist(nodeKey) {
		return nil
	}

	err := t.Graph.AddVertex(nodeKey, graph.VertexAttributes(map[string]string{
		"type":      Node,
		"name":      node.Name,
		"hostName":  node.Spec.HostName,
		"storageIP": node.Spec.StorageIP,
	}))
	if err != nil {
		return err
	}

	t.logger.WithField("nodeName", nodeKey).Debug("added node vertex")
	return nil
}

func (t *Topology[K, T]) IsVertexExist(key string) bool {
	_, err := t.Graph.Vertex(graph.StringHash(key))
	return err == nil
}

// AddStoragePool inserts Vertex(s) under the root node according to the pool exist in StorageNode
func (t *Topology[K, T]) AddStoragePool(node *v1alpha1.LocalStorageNode) error {
	err := t.AddLocalDisk(node)
	if err != nil {
		return err
	}

	// key format: storageNodeName/poolName
	for poolName, pool := range node.Status.Pools {
		// insert pool
		poolKey := GeneratePoolKey(node.Name, poolName)
		if !t.IsVertexExist(poolKey) {
			if err = t.Graph.AddVertex(poolKey, graph.VertexAttributes(map[string]string{
				"type":  Pool,
				"class": pool.Class,
				"name":  pool.Name,
			})); err != nil {
				return err
			}
			t.logger.WithField("poolName", poolKey).Debug("added pool vertex")
		}

		for _, disk := range pool.Disks {
			diskKey := GenerateDiskKey(node.Name, disk.DevPath)
			// construct edge: disk -> pool
			if err = t.Graph.AddEdge(diskKey, poolKey); err != nil {
				if errors.Is(err, graph.ErrEdgeAlreadyExists) {
					continue
				}
				return err
			}
			t.logger.Debugf("draw edge between disk and pool: %s", fmt.Sprintf("%s -> %s", diskKey, poolKey))
		}
	}
	return nil
}

// AddLocalVolume inserts Vertex(s) under the pool according to the pool exist in StoragePool
func (t *Topology[K, T]) AddLocalVolume(volume *v1alpha1.LocalVolume) error {
	volumeKey := GenerateLVKey(volume.Name)
	if !t.IsVertexExist(volume.Name) {
		if err := t.Graph.AddVertex(volumeKey, graph.VertexAttributes(map[string]string{
			"type":          Volume,
			"name":          volume.Name,
			"poolName":      volume.Spec.PoolName,
			"replicaNumber": strconv.FormatInt(volume.Spec.ReplicaNumber, 10),
		})); err != nil && !errors.Is(err, graph.ErrVertexAlreadyExists) {
			return err
		}
		t.logger.WithField("volumeName", volumeKey).Debug("added volume vertex")
	}

	// construct edge: pool -> volume
	for _, replica := range volume.Spec.Config.Replicas {
		poolKey := GeneratePoolKey(replica.Hostname, volume.Spec.PoolName)
		if err := t.Graph.AddEdge(poolKey, volumeKey); err != nil {
			if errors.Is(err, graph.ErrEdgeAlreadyExists) {
				continue
			}
			return err
		}
		t.logger.Debugf("draw edge between pool and volume: %s", fmt.Sprintf("%s -> %s", poolKey, volumeKey))
	}
	return nil
}

// AddLocalDisk inserts Vertex(s) under the pool according to the disk exist in StoragePool
func (t *Topology[K, T]) AddLocalDisk(node *v1alpha1.LocalStorageNode) error {
	nodeKey := GenerateNodeKey(node.Name)
	if !t.IsVertexExist(nodeKey) {
		if err := t.AddStorageNode(node); err != nil {
			return err
		}
	}

	for _, pool := range node.Status.Pools {
		for _, poolDisk := range pool.Disks {
			// key format: storageNodeName/devPath
			diskKey := GenerateDiskKey(node.Name, poolDisk.DevPath)
			if !t.IsVertexExist(diskKey) {
				err := t.Graph.AddVertex(diskKey, graph.VertexAttributes(map[string]string{
					"type":    Disk,
					"devPath": poolDisk.DevPath,
					"class":   poolDisk.Class,
				}))
				if err != nil {
					return err
				}
				t.logger.WithField("diskPath", poolDisk.DevPath).Debug("added poolDisk vertex")
			}

			// construct edge: node -> poolDisk
			if err := t.Graph.AddEdge(nodeKey, diskKey); err != nil {
				if errors.Is(err, graph.ErrEdgeAlreadyExists) {
					continue
				}
				return err
			}
			t.logger.Debugf("draw edge between node and poolDisk: %s", fmt.Sprintf("%s -> %s", nodeKey, diskKey))
		}
	}
	return nil
}

// AddPod inserts Vertex(s) under the PersistentVolumeClaim
func (t *Topology[K, T]) AddPod(pod *v1.Pod, volumes ...string) error {
	podKey := GeneratePodKey(pod.Namespace, pod.Name)
	if !t.IsVertexExist(podKey) {
		err := t.Graph.AddVertex(podKey, graph.VertexAttributes(map[string]string{
			"type":      Pod,
			"namespace": pod.Namespace,
			"name":      pod.Name,
		}))
		if err != nil {
			return err
		}
	}

	// construct edge: pvc -> pod
	for _, volumeName := range volumes {
		if err := t.Graph.AddEdge(volumeName, podKey); err != nil {
			if errors.Is(err, graph.ErrEdgeAlreadyExists) {
				continue
			}
			return err
		}
		t.logger.Debugf("draw edge between pod and pvc: %s", fmt.Sprintf("%s -> %s", volumeName, podKey))
	}
	return nil
}

// AddPVC inserts Vertex(s) under the volume
func (t *Topology[K, T]) AddPVC(pvc *v1.PersistentVolumeClaim) error {
	// todo: relation between pvc and pv can be changed, maybe updated edge or remove vertex for this pvc is better
	pvcKey := GeneratePVCKey(pvc.Namespace, pvc.Name)
	if t.IsVertexExist(pvcKey) {
		return nil
	}

	err := t.Graph.AddVertex(pvcKey, graph.VertexAttributes(map[string]string{
		"type":             Pvc,
		"namespace":        pvc.Namespace,
		"name":             pvc.Name,
		"storageClassName": *pvc.Spec.StorageClassName,
	}))
	if err != nil {
		return err
	}

	t.logger.WithField("namespacedName", pvcKey).Debug("added pvc vertex")
	return nil
}

// AddPV draw edge between pvc and localvolume
func (t *Topology[K, T]) AddPV(pv *v1.PersistentVolume) error {
	// localvolume use the same key with pv, don't insert pv again
	lvKey := GenerateLVKey(pv.Name)
	if !t.IsVertexExist(lvKey) {
		return fmt.Errorf("not found vertex, waiting for localvolume to create it")
	}

	// construct edge: pv(lv) -> pvc
	pvcKey := GeneratePVCKey(pv.Spec.ClaimRef.Namespace, pv.Spec.ClaimRef.Name)
	if err := t.Graph.AddEdge(lvKey, pvcKey); err != nil {
		if errors.Is(err, graph.ErrEdgeAlreadyExists) {
			return nil
		}
		return err
	}
	t.logger.Debugf("draw edge between pv(lv) and pvc: %s", fmt.Sprintf("%s -> %s", lvKey, pvcKey))
	return nil
}

func (t *Topology[K, T]) Draw() {
	//file, _ := os.Create("./k_string_graph.gv")
	//_ = draw.DOT(t.Graph, file)
}

func GeneratePodKey(namespace, name string) string {
	return Pod + separator + types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}.String()
}

func GeneratePVCKey(namespace, name string) string {
	return Pvc + separator + types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}.String()
}

func GeneratePVKey(name string) string {
	return PV + separator + name
}

func GenerateLVKey(name string) string {
	return Volume + separator + name
}

func GenerateDiskKey(nodeName, devPath string) string {
	return Disk + separator + nodeName + separator + devPath
}

func GeneratePoolKey(nodeName, name string) string {
	return Pool + separator + nodeName + separator + name
}

func GenerateNodeKey(nodeName string) string {
	return Node + separator + nodeName
}

func isHwameiStorVolume(provisioner string) bool {
	return strings.HasSuffix(provisioner, hwameistorDomain)
}

func isPoolResource(key string) bool {
	return strings.HasPrefix(key, Pool+separator)
}

func isVolumeResource(key string) bool {
	return strings.HasPrefix(key, Volume+separator)
}

func isPodResource(key string) bool {
	return strings.HasPrefix(key, Pod+separator)
}
