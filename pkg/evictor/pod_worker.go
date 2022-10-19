package evictor

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
)

func (ev *evictor) startPodWorker(stopCh <-chan struct{}) {
	log.Debug("Start a worker to process pod eviction")
	go func() {
		for {
			task, shutdown := ev.evictPodQueue.Get()
			if shutdown {
				log.WithFields(log.Fields{"task": task}).Debug("Stop the pod eviction worker")
				break
			}
			if err := ev.evictPod(task); err != nil {
				log.WithFields(log.Fields{"task": task, "error": err.Error()}).Error("Failed to process pod eviction task, retry later ...")
				ev.evictPodQueue.AddRateLimited(task)
			} else {
				log.WithFields(log.Fields{"task": task}).Debug("Completed a pod eviction task.")
				ev.evictPodQueue.Forget(task)
			}
			ev.evictPodQueue.Done(task)
		}
	}()

	<-stopCh
	ev.evictPodQueue.Shutdown()
}

func (ev *evictor) evictPod(task string) error {
	logCtx := log.WithField("pod", task)
	logCtx.Debug("Start to process a node eviction")

	podNamespace, podName := parseEvictPodTask(task)
	pod, err := ev.podInformer.Lister().Pods(podNamespace).Get(podName)
	if err != nil {
		if errors.IsNotFound(err) {
			logCtx.Debug("Pod doesn't exist")
			return nil
		}
		logCtx.WithError(err).Error("Failed to get pod from cache")
		return err
	}

	for _, vol := range pod.Spec.Volumes {
		if vol.PersistentVolumeClaim == nil {
			continue
		}
		pvc, err := ev.pvcInformer.Lister().PersistentVolumeClaims(pod.Namespace).Get(vol.PersistentVolumeClaim.ClaimName)
		if err != nil {
			// if pvc can't be found in the cluster, the pod should not be able to be scheduled
			logCtx.WithFields(log.Fields{
				"namespace": pod.Namespace,
				"pvc":       vol.PersistentVolumeClaim.ClaimName,
			}).WithError(err).Error("Failed to get the pvc from the cluster")
			return err
		}
		if pvc.Spec.StorageClassName == nil {
			// should not be the CSI pvc, ignore
			continue
		}
		sc, err := ev.scInformer.Lister().Get(*pvc.Spec.StorageClassName)
		if err != nil {
			// can't found storageclass in the cluster, the pod should not be able to be scheduled
			logCtx.WithFields(log.Fields{
				"pvc": pvc.Name,
				"sc":  *pvc.Spec.StorageClassName,
			}).WithError(err).Error("Failed to get the pvc from the cluster")
			return err
		}
		if sc.Provisioner == lvmCSIDriverName {
			logCtx.WithFields(log.Fields{
				"pvc":    pvc.Name,
				"sc":     sc.Name,
				"volume": pvc.Spec.VolumeName,
				"node":   pod.Spec.NodeName,
			}).Debug("Got a LocalVolume to migrate")

			ev.addEvictVolume(pvc.Spec.VolumeName, pod.Spec.NodeName)
		}
	}

	return nil
}

func (ev *evictor) addEvictPod(namespace, name string) {
	ev.evictPodQueue.AddRateLimited(fmt.Sprintf("%s/%s", namespace, name))
}

func parseEvictPodTask(task string) (podNamespace string, podName string) {
	items := strings.Split(task, "/")
	podNamespace = items[0]
	podName = items[1]
	return
}
