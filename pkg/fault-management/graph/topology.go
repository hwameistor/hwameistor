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
	//TODO implement me
	panic("implement me")
}

func (t *Topology[K, T]) GetVolumesUnderStoragePool(nodeName, poolName string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (t *Topology[K, T]) GetPodsUnderLocalVolume(nodeName, volumeName string) ([]string, error) {
	//TODO implement me
	panic("implement me")
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
	if t.IsVertexExist(node.Name) {
		return nil
	}

	err := t.Graph.AddVertex(node.Name, graph.VertexAttributes(map[string]string{
		"type":      Node,
		"name":      node.Name,
		"hostName":  node.Spec.HostName,
		"storageIP": node.Spec.StorageIP,
	}))
	if err != nil {
		return err
	}

	t.logger.WithField("nodeName", node.Name).Debug("added node vertex")
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
		poolKey := node.Name + separator + poolName
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
			diskKey := node.Name + separator + disk.DevPath
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
	if !t.IsVertexExist(volume.Name) {
		if err := t.Graph.AddVertex(volume.Name, graph.VertexAttributes(map[string]string{
			"type":          Volume,
			"name":          volume.Name,
			"poolName":      volume.Spec.PoolName,
			"replicaNumber": strconv.FormatInt(volume.Spec.ReplicaNumber, 10),
		})); err != nil && !errors.Is(err, graph.ErrVertexAlreadyExists) {
			return err
		}
		t.logger.WithField("volumeName", volume.Name).Debug("added volume vertex")
	}

	// construct edge: pool -> volume
	for _, replica := range volume.Spec.Config.Replicas {
		poolKey := replica.Hostname + separator + volume.Spec.PoolName
		if err := t.Graph.AddEdge(poolKey, volume.Name); err != nil {
			if errors.Is(err, graph.ErrEdgeAlreadyExists) {
				continue
			}
			return err
		}
		t.logger.Debugf("draw edge between pool and volume: %s", fmt.Sprintf("%s -> %s", poolKey, volume.Name))
	}
	return nil
}

// AddLocalDisk inserts Vertex(s) under the pool according to the disk exist in StoragePool
func (t *Topology[K, T]) AddLocalDisk(node *v1alpha1.LocalStorageNode) error {
	if !t.IsVertexExist(node.Name) {
		if err := t.AddStorageNode(node); err != nil {
			return err
		}
	}

	for _, pool := range node.Status.Pools {
		for _, poolDisk := range pool.Disks {
			// key format: storageNodeName/devPath
			diskKey := node.Name + separator + poolDisk.DevPath
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
			if err := t.Graph.AddEdge(node.Name, diskKey); err != nil {
				if errors.Is(err, graph.ErrEdgeAlreadyExists) {
					continue
				}
				return err
			}
			t.logger.Debugf("draw edge between node and poolDisk: %s", fmt.Sprintf("%s -> %s", node.Name, diskKey))
		}
	}
	return nil
}

// AddPod inserts Vertex(s) under the PersistentVolumeClaim
func (t *Topology[K, T]) AddPod(pod *v1.Pod, volumes ...string) error {
	podKey := types.NamespacedName{Namespace: pod.Namespace, Name: pod.Name}.String()
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
	pvcKey := types.NamespacedName{Namespace: pvc.Namespace, Name: pvc.Name}.String()
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

	t.logger.WithField("namespacedName", pvc.Namespace+"/"+pvc.Name).Debug("added pvc vertex")
	return nil
}

// AddPV draw edge between pvc and localvolume
func (t *Topology[K, T]) AddPV(pv *v1.PersistentVolume) error {
	// localvolume use the same key with pv, don't insert pv again
	if !t.IsVertexExist(pv.Name) {
		return fmt.Errorf("not found vertex, waiting for localvolume to create it")
	}

	// construct edge: pv -> pvc
	pvcKey := types.NamespacedName{Namespace: pv.Spec.ClaimRef.Namespace, Name: pv.Spec.ClaimRef.Name}.String()
	if err := t.Graph.AddEdge(pv.Name, pvcKey); err != nil {
		if errors.Is(err, graph.ErrEdgeAlreadyExists) {
			return nil
		}
		return err
	}
	t.logger.Debugf("draw edge between pv and pvc: %s", fmt.Sprintf("%s -> %s", pv.Name, pvcKey))
	return nil
}

func (t *Topology[K, T]) Draw() {
	//file, _ := os.Create("./k_string_graph.gv")
	//_ = draw.DOT(t.Graph, file)
}

func isHwameiStorVolume(provisioner string) bool {
	return strings.HasSuffix(provisioner, hwameistorDomain)
}
