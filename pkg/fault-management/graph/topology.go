package graph

import (
	"github.com/dominikbraun/graph"
	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	v1 "k8s.io/api/core/v1"
)

func (t *Topology[string, string]) GetPoolUnderLocalDisk(nodeName, diskPath string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (t *Topology[string, string]) GetVolumesUnderStoragePool(nodeName, poolName string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (t *Topology[string, string]) GetPodsUnderLocalVolume(nodeName, volumeName string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

type Topology[K string, T string] struct {
	graph.Graph[string, string]
}

func NewTopologyStore() Topology[string, string] {
	return Topology[string, string]{
		graph.New(graph.StringHash, graph.Directed(), graph.Acyclic()),
	}
}

// AddStorageNode inserts a Vertex as the roots of the node
func (t *Topology[K, T]) AddStorageNode(node *v1alpha1.LocalStorageNode) error {
	return nil
}

// AddStoragePool inserts Vertex(s) under the root node according to the pool exist in StorageNode
func (t *Topology[K, T]) AddStoragePool(node *v1alpha1.LocalStorageNode) error {
	return nil
}

// AddLocalVolume inserts Vertex(s) under the pool according to the pool exist in StoragePool
func (t *Topology[K, T]) AddLocalVolume(volume *v1alpha1.LocalVolume) error {
	return nil
}

// AddLocalDisk inserts Vertex(s) under the pool according to the disk exist in StoragePool
func (t *Topology[K, T]) AddLocalDisk(node *v1alpha1.LocalStorageNode) error {
	return nil
}

// AddPod inserts Vertex(s) under the volume
func (t *Topology[K, T]) AddPod(pod *v1.Pod) error {
	return nil
}
