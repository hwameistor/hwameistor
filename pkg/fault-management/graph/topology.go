package graph

import (
	"errors"
	"fmt"
	"github.com/dominikbraun/graph"
	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"strconv"
)

const separator = "_"

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
			if err = t.Graph.AddEdge(diskKey, poolKey); err != nil && !errors.Is(err, graph.ErrEdgeAlreadyExists) {
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
			"name":          volume.Spec.Config.VolumeName,
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
		if err := t.Graph.AddEdge(poolKey, volume.Name); err != nil && !errors.Is(err, graph.ErrEdgeAlreadyExists) {
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
		for _, disk := range pool.Disks {
			// key format: storageNodeName/devPath
			diskKey := node.Name + separator + disk.DevPath
			if !t.IsVertexExist(diskKey) {
				err := t.Graph.AddVertex(diskKey, graph.VertexAttributes(map[string]string{
					"devPath": disk.DevPath,
					"type":    disk.Class,
				}))
				if err != nil {
					return err
				}
				t.logger.WithField("diskPath", disk.DevPath).Debug("added disk vertex")
			}

			// construct edge: node -> disk
			if err := t.Graph.AddEdge(node.Name, diskKey); err != nil && !errors.Is(err, graph.ErrEdgeAlreadyExists) {
				return err
			}
			t.logger.Debugf("draw edge between node and disk: %s", fmt.Sprintf("%s -> %s", node.Name, diskKey))
		}
	}
	return nil
}

// AddPod inserts Vertex(s) under the volume
func (t *Topology[K, T]) AddPod(pod *v1.Pod) error {
	return nil
}
