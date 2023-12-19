package graph

import (
	"fmt"
	hwameistorclient "github.com/hwameistor/hwameistor/pkg/apis/client/clientset/versioned"
	"github.com/hwameistor/hwameistor/pkg/apis/client/informers/externalversions/hwameistor/v1alpha1"
	listers "github.com/hwameistor/hwameistor/pkg/apis/client/listers/hwameistor/v1alpha1"
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/common"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	informercorev1 "k8s.io/client-go/informers/core/v1"
	v12 "k8s.io/client-go/listers/core/v1"
	storagev1lister "k8s.io/client-go/listers/storage/v1"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const hwameistorDomain = "hwameistor.io"

type TopologyGraphManager interface {
	GetPoolUnderLocalDisk(nodeName, diskPath string) (string, error)
	GetVolumesUnderStoragePool(nodeName, poolName string) ([]string, error)
	GetPodsUnderLocalVolume(nodeName, volumeName string) ([]string, error)
	Draw()
	Run(stopCh <-chan struct{}) error
}

var _ TopologyGraphManager = &manager{}

type manager struct {
	name      string
	namespace string
	logger    *log.Entry
	kclient   client.Client

	hmClient          hwameistorclient.Interface
	localVolumeLister listers.LocalVolumeLister
	storageNodeLister listers.LocalStorageNodeLister
	pvLister          v12.PersistentVolumeLister
	pvcLister         v12.PersistentVolumeClaimLister
	podLister         v12.PodLister
	scLister          storagev1lister.StorageClassLister

	storageNodeInformer v1alpha1.LocalStorageNodeInformer
	storageNodeSynced   cache.InformerSynced
	localVolumeInformer v1alpha1.LocalVolumeInformer
	localVolumeSynced   cache.InformerSynced
	podInformer         informercorev1.PodInformer
	podSynced           cache.InformerSynced
	pvcInformer         informercorev1.PersistentVolumeClaimInformer
	pvcSynced           cache.InformerSynced
	pvInformer          informercorev1.PersistentVolumeInformer
	pvSynced            cache.InformerSynced

	podTaskQueue         *common.TaskQueue
	pvcTaskQueue         *common.TaskQueue
	pvTaskQueue          *common.TaskQueue
	localVolumeTaskQueue *common.TaskQueue
	storageNodeTaskQueue *common.TaskQueue

	Topology[string, string]
}

func New(name, namespace string, kclient client.Client, hmClient hwameistorclient.Interface,
	podInformer informercorev1.PodInformer,
	pvcInformer informercorev1.PersistentVolumeClaimInformer,
	pvInformer informercorev1.PersistentVolumeInformer,
	storageNodeInformer v1alpha1.LocalStorageNodeInformer,
	localVolumeInformer v1alpha1.LocalVolumeInformer,
	scLister storagev1lister.StorageClassLister,
) TopologyGraphManager {
	m := &manager{
		name:                 name,
		namespace:            namespace,
		hmClient:             hmClient,
		kclient:              kclient,
		Topology:             NewTopologyStore(),
		podInformer:          podInformer,
		podLister:            podInformer.Lister(),
		podSynced:            podInformer.Informer().HasSynced,
		pvcInformer:          pvcInformer,
		pvcLister:            pvcInformer.Lister(),
		pvcSynced:            pvcInformer.Informer().HasSynced,
		pvInformer:           pvInformer,
		pvLister:             pvInformer.Lister(),
		pvSynced:             pvInformer.Informer().HasSynced,
		scLister:             scLister,
		localVolumeInformer:  localVolumeInformer,
		localVolumeLister:    localVolumeInformer.Lister(),
		localVolumeSynced:    localVolumeInformer.Informer().HasSynced,
		storageNodeInformer:  storageNodeInformer,
		storageNodeLister:    storageNodeInformer.Lister(),
		storageNodeSynced:    storageNodeInformer.Informer().HasSynced,
		podTaskQueue:         common.NewTaskQueue("GraphManagerPodTaskQueue", 0),
		pvcTaskQueue:         common.NewTaskQueue("GraphManagerPVCTaskQueue", 0),
		pvTaskQueue:          common.NewTaskQueue("GraphManagerPVTaskQueue", 0),
		localVolumeTaskQueue: common.NewTaskQueue("GraphManagerLocalVolumeTaskQueue", 0),
		storageNodeTaskQueue: common.NewTaskQueue("GraphManagerStorageNodeTaskQueue", 0),
		logger:               log.WithField("Module", "GraphManager"),
	}
	m.setupInformers()
	return m
}

func (m *manager) Run(stopCh <-chan struct{}) error {
	// Wait for the caches to be synced before starting processors
	m.logger.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, m.podSynced, m.localVolumeSynced, m.storageNodeSynced); !ok {
		m.logger.Error("Timeout waiting for caches to sync")
		return fmt.Errorf("timeout waiting caches to sync")
	}

	m.logger.Info("Starting GraphManager worker")
	go m.startGraphManagementTaskWorker(stopCh)

	return nil
}

func (m *manager) setupInformers() {
	// ----------------------------------------
	// PersistentVolumeClaims informer handlers
	m.pvcInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    m.handlePVCAdd,
		UpdateFunc: m.handlePVCUpdate,
	})

	// ----------------------------------------
	// PersistentVolumes informer handlers
	m.pvInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    m.handlePVAdd,
		UpdateFunc: m.handlePVUpdate,
	})

	// ----------------------
	// Pods informer handlers
	m.podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    m.handlePodAdd,
		UpdateFunc: m.handlePodUpdate,
	})

	// ------------------------------
	// LocalVolumes informer handlers
	m.localVolumeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    m.handleLocalVolumeAdd,
		UpdateFunc: m.handleLocalVolumeUpdate,
	})

	// -----------------------------------
	// LocalStorageNodes informer handlers
	m.storageNodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    m.handleLocalStorageNodeAdd,
		UpdateFunc: m.handleLocalStorageNodeUpdate,
	})
}

func (m *manager) handlePVCUpdate(_, obj interface{}) {
	m.handlePVCAdd(obj)
}

func (m *manager) handlePVCAdd(obj interface{}) {
	if _, ok := obj.(*v1.PersistentVolumeClaim); !ok {
		return
	}
	pvc := obj.(*v1.PersistentVolumeClaim)
	if pvc.Spec.StorageClassName == nil {
		m.logger.WithField("pvcNamespacedName", pvc.Namespace+"/"+pvc.Name).Debug("not hwameistor volume, drop it")
		return
	}
	m.pvcTaskQueue.Add(types.NamespacedName{Namespace: pvc.Namespace, Name: pvc.Name}.String())
}

func (m *manager) handlePVUpdate(_, obj interface{}) {
	m.handlePVAdd(obj)
}

func (m *manager) handlePVAdd(obj interface{}) {
	if _, ok := obj.(*v1.PersistentVolume); !ok {
		return
	}
	pv := obj.(*v1.PersistentVolume)
	if pv.Spec.StorageClassName == "" {
		return
	}
	m.pvTaskQueue.Add(pv.Name)
}

func (m *manager) handlePodUpdate(_, obj interface{}) {
	m.handlePodAdd(obj)
}

func (m *manager) handlePodAdd(obj interface{}) {
	if _, ok := obj.(*v1.Pod); !ok {
		return
	}

	p := obj.(*v1.Pod)
	hasVolume := func(pod *v1.Pod) bool {
		for _, volume := range pod.Spec.Volumes {
			if volume.PersistentVolumeClaim != nil {
				return true
			}
		}
		return false
	}
	if !hasVolume(p) {
		return
	}
	m.podTaskQueue.Add(types.NamespacedName{Namespace: p.Namespace, Name: p.Name}.String())
}

func (m *manager) handleLocalVolumeUpdate(_, obj interface{}) {
	m.handleLocalVolumeAdd(obj)
}

func (m *manager) handleLocalVolumeAdd(obj interface{}) {
	if _, ok := obj.(*apisv1alpha1.LocalVolume); !ok {
		return
	}
	m.localVolumeTaskQueue.Add(obj.(*apisv1alpha1.LocalVolume).Name)
}

func (m *manager) handleLocalStorageNodeUpdate(_, obj interface{}) {
	m.handleLocalStorageNodeAdd(obj)
}

func (m *manager) handleLocalStorageNodeAdd(obj interface{}) {
	if _, ok := obj.(*apisv1alpha1.LocalStorageNode); !ok {
		return
	}
	m.storageNodeTaskQueue.Add(obj.(*apisv1alpha1.LocalStorageNode).Name)
}
