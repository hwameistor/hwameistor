package evictor

import (
	"context"
	"fmt"

	localstorageapis "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (ev *evictor) startNodeWorker(stopCh <-chan struct{}) {
	log.Debug("Start a worker to process node eviction")
	go func() {
		for {
			task, shutdown := ev.evictNodeQueue.Get()
			if shutdown {
				log.WithFields(log.Fields{"task": task}).Debug("Stop the node eviction worker")
				break
			}
			if err := ev.evictNode(task); err != nil {
				log.WithFields(log.Fields{"task": task, "error": err.Error()}).Error("Failed to process node eviction task, retry later ...")
				ev.evictNodeQueue.AddRateLimited(task)
			} else {
				log.WithFields(log.Fields{"task": task}).Debug("Completed a node eviction task.")
				ev.evictNodeQueue.Forget(task)
			}
			ev.evictNodeQueue.Done(task)
		}
	}()

	<-stopCh
	ev.evictNodeQueue.Shutdown()
}

func (ev *evictor) evictNode(nodeName string) error {
	logCtx := log.WithField("node", nodeName)
	logCtx.Debug("Start to process a node eviction")

	lvrs, err := ev.lvrInformer.Informer().GetIndexer().ByIndex(nodeNameIndex, nodeName)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get LocalVolumeReplicas on the node")
		return err
	}
	if len(lvrs) == 0 {
		node, err := ev.nodeInformer.Lister().Get(nodeName)
		if err != nil {
			logCtx.WithError(err).Error("Failed to get node from cache")
			return err
		}
		node.Labels[labelKeyForVolumeEviction] = labelValueForVolumeEvictionCompleted
		if _, err := ev.clientset.CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{}); err != nil {
			logCtx.WithError(err).Error("Failed to update node")
			return err
		}
		return nil
	}

	for i := range lvrs {
		lvr, _ := lvrs[i].(*localstorageapis.LocalVolumeReplica)
		logCtx.WithFields(log.Fields{"volume": lvr.Spec.VolumeName, "sourceNode": nodeName}).Debug("Add a volume migrate task")
		ev.addEvictVolume(lvr.Spec.VolumeName, nodeName)
	}
	return fmt.Errorf("waiting for volumes migration complete")
}

func (ev *evictor) addEvictNode(nodeName string) {
	ev.evictNodeQueue.AddRateLimited(nodeName)
}
