package evictor

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	localstorageclientset "github.com/hwameistor/hwameistor/pkg/apis/client/clientset/versioned"
	localstorageinformers "github.com/hwameistor/hwameistor/pkg/apis/client/informers/externalversions"
	localstorageinformersv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/client/informers/externalversions/hwameistor/v1alpha1"
	localstorageapis "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-storage/common"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	informercorev1 "k8s.io/client-go/informers/core/v1"
	informerstoragev1 "k8s.io/client-go/informers/storage/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

const (
	lvmCSIDriverName  = "lvm.hwameistor.io"
	diskCSIDriverName = "disk.hwameistor.io"

	volumeNameIndex = "volumename"
	nodeNameIndex   = "nodename"

	labelKeyForVolumeEviction            = "hwameistor.io/eviction"
	labelValueForVolumeEvictionStart     = "start"
	labelValueForVolumeEvictionCompleted = "completed"
	labelValueForVolumeEvictionDisable   = "disable"
)

// Evictor interface
type Evictor interface {
	Run(stopCh <-chan struct{}) error
}

type evictor struct {
	clientset *kubernetes.Clientset

	nodeInformer informercorev1.NodeInformer
	podInformer  informercorev1.PodInformer
	pvcInformer  informercorev1.PersistentVolumeClaimInformer
	scInformer   informerstoragev1.StorageClassInformer

	lsClientset       *localstorageclientset.Clientset
	lsnInformer       localstorageinformersv1alpha1.LocalStorageNodeInformer
	lvInformer        localstorageinformersv1alpha1.LocalVolumeInformer
	lvrInformer       localstorageinformersv1alpha1.LocalVolumeReplicaInformer
	lvMigrateInformer localstorageinformersv1alpha1.LocalVolumeMigrateInformer

	evictNodeQueue   *common.TaskQueue
	evictPodQueue    *common.TaskQueue
	evictVolumeQueue *common.TaskQueue
}

/* steps:
1. watch for Pod update event, insert the Pod with Evicted status into evictedPodQueue;
2. pick up a Pod from evictedPodQueue, check if it is using HwameiStor volume. If yes, insert it into the migrateVolumeQueue; if not, ignore
3. pick up a volume form migrateVolumeQueue, and migrate it. Make sure there is no replica located at the node where the pod is evicted;
*/

// New an assistant instance
func New(clientset *kubernetes.Clientset) Evictor {
	return &evictor{
		clientset:        clientset,
		evictNodeQueue:   common.NewTaskQueue("EvictNodes", 0),
		evictPodQueue:    common.NewTaskQueue("EvictPods", 0),
		evictVolumeQueue: common.NewTaskQueue("EvictVolumes", 0),
	}
}

func (ev *evictor) Run(stopCh <-chan struct{}) error {

	log.Debug("start informer factory")
	factory := informers.NewSharedInformerFactory(ev.clientset, 0)
	factory.Start(stopCh)

	ev.nodeInformer = factory.Core().V1().Nodes()
	ev.nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    ev.onNodeAdd,
		UpdateFunc: ev.onNodeUpdate,
	})
	go ev.nodeInformer.Informer().Run(stopCh)

	ev.podInformer = factory.Core().V1().Pods()
	ev.podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: ev.onPodUpdate,
	})
	go ev.podInformer.Informer().Run(stopCh)

	ev.pvcInformer = factory.Core().V1().PersistentVolumeClaims()
	go ev.pvcInformer.Informer().Run(stopCh)
	ev.scInformer = factory.Storage().V1().StorageClasses()
	go ev.scInformer.Informer().Run(stopCh)

	// Initialize HwameiStor LocalStorage resources
	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		log.WithError(err).Fatal("Failed to get kubernetes cluster config")
	}

	log.Debug("start local storage informer factory")
	ev.lsClientset = localstorageclientset.NewForConfigOrDie(cfg)
	lsFactory := localstorageinformers.NewSharedInformerFactory(ev.lsClientset, 0)
	lsFactory.Start(stopCh)

	ev.lvInformer = lsFactory.Hwameistor().V1alpha1().LocalVolumes()
	go ev.lvInformer.Informer().Run(stopCh)

	ev.lsnInformer = lsFactory.Hwameistor().V1alpha1().LocalStorageNodes()
	go ev.lsnInformer.Informer().Run(stopCh)

	// index: lvr.spec.nodename
	lvrNodeNameIndexFunc := func(obj interface{}) ([]string, error) {
		lvr, ok := obj.(*localstorageapis.LocalVolumeReplica)
		if !ok || lvr == nil {
			return []string{}, fmt.Errorf("wrong LocalStorageReplica resource")
		}
		return []string{lvr.Spec.NodeName}, nil
	}
	ev.lvrInformer = lsFactory.Hwameistor().V1alpha1().LocalVolumeReplicas()
	ev.lvrInformer.Informer().AddIndexers(cache.Indexers{nodeNameIndex: lvrNodeNameIndexFunc})
	go ev.lvrInformer.Informer().Run(stopCh)

	// index: lvmigrate.spec.volumename
	lvMigrateVolumeNameIndexFunc := func(obj interface{}) ([]string, error) {
		lvm, ok := obj.(*localstorageapis.LocalVolumeMigrate)
		if !ok || lvm == nil {
			return []string{}, fmt.Errorf("wrong LocalStorageMigrate resource")
		}
		return []string{lvm.Spec.VolumeName}, nil
	}
	ev.lvMigrateInformer = lsFactory.Hwameistor().V1alpha1().LocalVolumeMigrates()
	ev.lvMigrateInformer.Informer().AddIndexers(cache.Indexers{volumeNameIndex: lvMigrateVolumeNameIndexFunc})
	go ev.lvMigrateInformer.Informer().Run(stopCh)

	go ev.startVolumeWorker(stopCh)
	go ev.startNodeWorker(stopCh)
	go ev.startPodWorker(stopCh)

	<-stopCh
	return nil
}

func (ev *evictor) onNodeAdd(obj interface{}) {
	node, _ := obj.(*corev1.Node)
	for _, taint := range node.Spec.Taints {
		if taint.Key == corev1.TaintNodeUnschedulable && taint.Effect == corev1.TaintEffectNoSchedule {
			if node.Labels[labelKeyForVolumeEviction] != labelValueForVolumeEvictionDisable {
				ev.addEvictNode(node.Name)
				return
			}
		}
	}
}

func (ev *evictor) onNodeUpdate(oldObj, newObj interface{}) {
	ev.onNodeAdd(newObj)
}

func (ev *evictor) onPodAdd(obj interface{}) {
	pod, _ := obj.(*corev1.Pod)
	if isPodEvicted(pod) || pod.Labels[labelKeyForVolumeEviction] == labelValueForVolumeEvictionStart {
		ev.addEvictPod(pod.Namespace, pod.Name)
	}
}

func (ev *evictor) onPodUpdate(oldObj, newObj interface{}) {
	ev.onPodAdd(newObj)
}

func isPodEvicted(pod *corev1.Pod) bool {
	podFailed := pod.Status.Phase == corev1.PodFailed
	podEvicted := pod.Status.Reason == "Evicted"
	return podFailed && podEvicted
}
