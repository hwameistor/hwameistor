package failoverassistant

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hwameistor/hwameistor/pkg/local-storage/common"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	policyv1b1 "k8s.io/api/policy/v1beta1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/informers"
	informercorev1 "k8s.io/client-go/informers/core/v1"
	informerstoragev1 "k8s.io/client-go/informers/storage/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

const (
	failoverLabelKey       = "hwameistor.io/failover"
	failoverLabelStart     = "start"
	failoverLabelCompleted = "completed"

	podScheduledNodeNameIndex   = "assignedNode"
	volumeAttachmentPVNameIndex = "pvName"
	pvClaimNamespacedNameIndex  = "claimNamespacedName"
)

// Assistant interface
type Assistant interface {
	Run(stopCh <-chan struct{}) error
}

type failoverAssistant struct {
	clientset *kubernetes.Clientset

	nodeInformer             informercorev1.NodeInformer
	podInformer              informercorev1.PodInformer
	pvInformer               informercorev1.PersistentVolumeInformer
	volumeAttachmentInformer informerstoragev1.VolumeAttachmentInformer

	failoverNodeQueue *common.TaskQueue
	failoverPodQueue  *common.TaskQueue
}

// New an assistant instance
func New(clientset *kubernetes.Clientset) Assistant {
	return &failoverAssistant{
		clientset:         clientset,
		failoverNodeQueue: common.NewTaskQueue("FailoverNodeTask", 0),
		failoverPodQueue:  common.NewTaskQueue("FailoverPodTask", 0),
	}
}

func (fa *failoverAssistant) Run(stopCh <-chan struct{}) error {
	log.Debug("start informer factory")
	factory := informers.NewSharedInformerFactory(fa.clientset, 0)
	factory.Start(stopCh)
	for _, v := range factory.WaitForCacheSync(stopCh) {
		if !v {
			log.Error("Timed out waiting for cache to sync")
			return fmt.Errorf("timed out waiting for cache to sync")
		}
	}

	log.Debug("setting up informer for Node ...")
	fa.nodeInformer = factory.Core().V1().Nodes()
	fa.nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: fa.onNodeUpdate,
	})
	go fa.nodeInformer.Informer().Run(stopCh)

	log.Debug("setting up informer for Pod ...")
	// index: pod.spec.nodename
	podScheduledNodeNameIndexFunc := func(obj interface{}) ([]string, error) {
		pod, ok := obj.(*corev1.Pod)
		if !ok || pod == nil {
			return []string{}, fmt.Errorf("wrong Pod resource")
		}
		return []string{pod.Spec.NodeName}, nil
	}
	fa.podInformer = factory.Core().V1().Pods()
	fa.podInformer.Informer().AddIndexers(cache.Indexers{podScheduledNodeNameIndex: podScheduledNodeNameIndexFunc})
	fa.podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: fa.onPodUpdate,
	})
	go fa.podInformer.Informer().Run(stopCh)

	log.Debug("setting up informer for PV ...")
	// index: pv.spec.claimref.namespace_name (pvc's namespacedname)
	pvClaimNamespacedNameIndexFunc := func(obj interface{}) ([]string, error) {
		pv, ok := obj.(*corev1.PersistentVolume)
		if !ok || pv == nil {
			return []string{}, fmt.Errorf("wrong PersistantVolume resource")
		}
		if pv.Spec.ClaimRef == nil {
			return []string{}, nil
		}
		return []string{namespacedName(pv.Spec.ClaimRef.Namespace, pv.Spec.ClaimRef.Name)}, nil
	}
	fa.pvInformer = factory.Core().V1().PersistentVolumes()
	fa.pvInformer.Informer().AddIndexers(cache.Indexers{pvClaimNamespacedNameIndex: pvClaimNamespacedNameIndexFunc})
	go fa.pvInformer.Informer().Run(stopCh)

	log.Debug("setting up informer for VolumeAttachment ...")
	// index volumeattachment.spec.source.pvname
	volumeAttachmentPVNameIndexFunc := func(obj interface{}) ([]string, error) {
		va, ok := obj.(*storagev1.VolumeAttachment)
		if !ok || va == nil {
			return []string{}, fmt.Errorf("wrong VolumeAttachment resource")
		}
		if va.Spec.Source.PersistentVolumeName == nil {
			return []string{}, nil
		}
		return []string{*va.Spec.Source.PersistentVolumeName}, nil
	}
	fa.volumeAttachmentInformer = factory.Storage().V1().VolumeAttachments()
	fa.volumeAttachmentInformer.Informer().AddIndexers(cache.Indexers{volumeAttachmentPVNameIndex: volumeAttachmentPVNameIndexFunc})
	go fa.volumeAttachmentInformer.Informer().Run(stopCh)

	log.Debug("start failover worker")
	go fa.startWorkerForNodeFailover(stopCh)

	return nil
}

func (fa *failoverAssistant) onNodeUpdate(oldObj, newObj interface{}) {
	node, _ := newObj.(*corev1.Node)
	fa.failoverNode(node)
}

func (fa *failoverAssistant) onPodUpdate(oldObj, newObj interface{}) {
	pod, _ := newObj.(*corev1.Pod)
	fa.failoverPod(pod)
}

func (fa *failoverAssistant) failoverNode(node *corev1.Node) {
	if fa.shouldFailoverForNode(node) {
		log.WithField("node", node.Name).Debug("Add node into failover queue")
		fa.failoverNodeQueue.Add(node.Name)
	}
}

func (fa *failoverAssistant) failoverPod(pod *corev1.Pod) {
	if fa.shouldFailoverForPod(pod) {
		if err := fa.failoverForPod(context.TODO(), pod.Spec.NodeName, pod); err != nil {
			log.WithFields(log.Fields{"namespace": pod.Namespace, "pod": pod.Name}).WithError(err).Debug("Failed to handle pod failover ")
		} else {
			log.WithFields(log.Fields{"namespace": pod.Namespace, "pod": pod.Name}).Debug("Failed over the pod successfully")
		}
	}
}

func (fa *failoverAssistant) shouldFailoverForNode(node *corev1.Node) bool {
	return node.Labels[failoverLabelKey] == failoverLabelStart
}

func (fa *failoverAssistant) shouldFailoverForPod(pod *corev1.Pod) bool {
	return pod.Labels[failoverLabelKey] == failoverLabelStart
}

func (fa *failoverAssistant) startWorkerForNodeFailover(stopCh <-chan struct{}) {

	log.Debug("Node Failover Worker is working now")
	go func() {
		for {
			time.Sleep(time.Second)
			task, shutdown := fa.failoverNodeQueue.Get()
			if shutdown {
				log.WithFields(log.Fields{"task": task}).Debug("Stop the Node Failover worker")
				break
			}
			if err := fa.processNodeForFailover(task); err != nil {
				log.WithFields(log.Fields{"task": task, "error": err.Error()}).Error("Failed to process Node Failover task, retry later")
				fa.failoverNodeQueue.AddRateLimited(task)
			} else {
				log.WithFields(log.Fields{"task": task}).Debug("Completed a Node Failover task.")
				fa.failoverNodeQueue.Forget(task)
			}
			fa.failoverNodeQueue.Done(task)

		}
	}()

	<-stopCh
	fa.failoverNodeQueue.Shutdown()
}

func (fa *failoverAssistant) processNodeForFailover(nodeName string) error {
	logCtx := log.WithField("node", nodeName)
	logCtx.Debug("Start to failover the stateful pods for the node")
	node, err := fa.nodeInformer.Lister().Get(nodeName)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get node info")
		return err
	}
	if !fa.shouldFailoverForNode(node) {
		// double check in case that nodes'status is changed back to normal shortly
		logCtx.Info("Cancel failover for node")
		return nil
	}

	pods, err := fa.podInformer.Informer().GetIndexer().ByIndex(podScheduledNodeNameIndex, nodeName)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get pods on the node")
		return err
	}

	failedPodList := []string{}
	for i := range pods {
		pod, _ := pods[i].(*corev1.Pod)
		if err := fa.failoverForPod(context.TODO(), pod.Spec.NodeName, pod); err != nil {
			log.WithFields(log.Fields{"namespace": pod.Namespace, "pod": pod.Name}).WithError(err).Debug("Failed to handle pod failover ")
			failedPodList = append(failedPodList, fmt.Sprintf("%s/%s", pod.Namespace, pod.Name))
		} else {
			log.WithFields(log.Fields{"namespace": pod.Namespace, "pod": pod.Name}).Debug("Failed over the pod successfully")
		}
	}

	if len(failedPodList) != 0 {
		return fmt.Errorf("node failover not completed")
	}

	return fa.completeNodeFailover(node)
}

func (fa *failoverAssistant) failoverForPod(ctx context.Context, nodeName string, pod *corev1.Pod) error {
	logCtx := log.WithFields(log.Fields{"node": nodeName, "namespace": pod.Namespace, "pod": pod.Name})

	volFailoverRequests := []*VolumeFailoverRestRequest{}
	for _, vol := range pod.Spec.Volumes {
		if vol.PersistentVolumeClaim == nil {
			continue
		}
		volReqs, err := fa.buildVolumeFailoverRequestForPVC(pod.Namespace, nodeName, vol.PersistentVolumeClaim.ClaimName)
		if err != nil {
			logCtx.WithField("pvc", vol.PersistentVolumeClaim.ClaimName).WithError(err).Error("Failed to build volume failover requests for pvc")
			return err
		}
		volFailoverRequests = append(volFailoverRequests, volReqs...)
	}

	logCtx.Debug("Failover for pod")
	for i := range volFailoverRequests {
		if err := fa.failoverForVolume(ctx, volFailoverRequests[i]); err != nil {
			return err
		}
	}

	return fa.deletePod(ctx, pod, 0)
}

func (fa *failoverAssistant) failoverForVolume(ctx context.Context, req *VolumeFailoverRestRequest) error {
	return fa.deleteVolumeAttachment(ctx, req.VolmeAttachmentID, 0)
}

func (fa *failoverAssistant) deleteVolumeAttachment(ctx context.Context, volAttachID string, gracePeriodSeconds int64) error {
	logCtx := log.WithFields(log.Fields{"VolumeAttachment": volAttachID, "graceperiod": gracePeriodSeconds})
	logCtx.Debug("Cleanning up the volumeattachment")
	deleteOptions := metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
	}
	if gracePeriodSeconds == 0 {
		propagationPolicy := metav1.DeletePropagationBackground
		deleteOptions.PropagationPolicy = &propagationPolicy
	}
	err := fa.clientset.StorageV1().VolumeAttachments().Delete(ctx, volAttachID, deleteOptions)
	if err != nil {
		logCtx.WithError(err).Error("Failed to clean up volumeattachment")
	}
	return err
}

func (fa *failoverAssistant) deletePod(ctx context.Context, pod *corev1.Pod, gracePeriodSeconds int64) error {
	logCtx := log.WithFields(log.Fields{"namespace": pod.Namespace, "pod": pod.Name, "graceperiod": gracePeriodSeconds})
	logCtx.Debug("Deleting Pod")
	err := fa.clientset.CoreV1().Pods(pod.Namespace).Evict(ctx, &policyv1b1.Eviction{
		ObjectMeta:    metav1.ObjectMeta{Name: pod.Name},
		DeleteOptions: &metav1.DeleteOptions{GracePeriodSeconds: &gracePeriodSeconds},
	})
	if err != nil {
		logCtx.WithError(err).Error("Failed to delete Pod")
	}
	return err
}

func (fa *failoverAssistant) buildVolumeFailoverRequestForPVC(ns, nodeName, claimName string) ([]*VolumeFailoverRestRequest, error) {
	failoverRequests := []*VolumeFailoverRestRequest{}
	pvList, err := fa.pvInformer.Informer().GetIndexer().ByIndex(pvClaimNamespacedNameIndex, namespacedName(ns, claimName))
	if err != nil {
		return failoverRequests, err
	}
	if len(pvList) != 1 {
		return failoverRequests, nil
	}
	pv, ok := pvList[0].(*corev1.PersistentVolume)
	if !ok {
		return failoverRequests, fmt.Errorf("wrong pv resource")
	}
	if pv.Spec.CSI == nil {
		// don't handle non-CSI volume
		return failoverRequests, fmt.Errorf("not supported non-CSI volume ")
	}

	vaList, err := fa.volumeAttachmentInformer.Informer().GetIndexer().ByIndex(volumeAttachmentPVNameIndex, pv.Name)
	if err != nil {
		return failoverRequests, err
	}
	if len(vaList) == 0 {
		return failoverRequests, nil
	}
	// it's possible for a volume to be attached multiple times
	for i := range vaList {
		va, ok := vaList[i].(*storagev1.VolumeAttachment)
		if !ok {
			continue
		}
		if va.Spec.NodeName == nodeName {
			failoverRequests = append(failoverRequests,
				&VolumeFailoverRestRequest{
					VolumeName:        pv.Name,
					StorageClassName:  pv.Spec.StorageClassName,
					NodeName:          nodeName,
					VolmeAttachmentID: va.Name,
				})
		}
	}
	return failoverRequests, nil
}

func (fa *failoverAssistant) completeNodeFailover(node *corev1.Node) error {
	old, _ := json.Marshal(node)

	newNode := node.DeepCopy()
	newNode.Labels[failoverLabelKey] = failoverLabelCompleted
	newData, _ := json.Marshal(newNode)

	data, err := strategicpatch.CreateTwoWayMergePatch(old, newData, corev1.Node{})
	if err != nil {
		return fmt.Errorf("create patch data err: %s", err)
	}

	if _, err = fa.clientset.CoreV1().Nodes().Patch(context.TODO(), node.Name, types.StrategicMergePatchType, data, metav1.PatchOptions{}); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	return nil
}

func namespacedName(ns string, name string) string {
	return fmt.Sprintf("%s_%s", ns, name)
}
