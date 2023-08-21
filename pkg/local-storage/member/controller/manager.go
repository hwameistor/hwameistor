package controller

import (
	"context"
	"reflect"
	"sync"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
	runtimecache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apis "github.com/hwameistor/hwameistor/pkg/apis/hwameistor"
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-storage/common"
	"github.com/hwameistor/hwameistor/pkg/local-storage/member/controller/scheduler"
	"github.com/hwameistor/hwameistor/pkg/local-storage/member/controller/volumegroup"
	"github.com/hwameistor/hwameistor/pkg/local-storage/utils"
	datacopyutil "github.com/hwameistor/hwameistor/pkg/local-storage/utils/datacopy"
)

// maxRetries is the number of times a task will be retried before it is dropped out of the queue.
// With the current rate-limiter in use math.Max(16s, (1s*2^(maxRetries-1))) the following numbers represent the times
// a task is going to be requeued:
//
// Infinitely retry
const maxRetries = 0

type manager struct {
	name string

	namespace string

	apiClient client.Client

	informersCache runtimecache.Cache

	scheme *runtime.Scheme

	volumeScheduler apisv1alpha1.VolumeScheduler

	volumeGroupManager apisv1alpha1.VolumeGroupManager

	nodeTaskQueue *common.TaskQueue

	k8sNodeTaskQueue *common.TaskQueue

	volumeTaskQueue *common.TaskQueue

	volumeExpandTaskQueue *common.TaskQueue

	volumeMigrateTaskQueue *common.TaskQueue

	volumeSnapshotTaskQueue *common.TaskQueue

	volumeSnapshotRecoverTaskQueue *common.TaskQueue

	volumeGroupMigrateTaskQueue *common.TaskQueue

	volumeConvertTaskQueue *common.TaskQueue

	volumeGroupConvertTaskQueue *common.TaskQueue

	localNodes map[string]apisv1alpha1.State // nodeName -> status

	logger *log.Entry

	lock sync.Mutex

	dataCopyManager *datacopyutil.DataCopyManager
}

// New cluster manager
func New(name string, namespace string, cli client.Client, scheme *runtime.Scheme, informersCache runtimecache.Cache, systemConfig apisv1alpha1.SystemConfig) (apis.ControllerManager, error) {
	dataCopyStatusCh := make(chan *datacopyutil.DataCopyStatus, 100)
	dcm, _ := datacopyutil.NewDataCopyManager(context.TODO(), "", cli, dataCopyStatusCh, namespace)
	return &manager{
		name:               name,
		namespace:          namespace,
		apiClient:          cli,
		informersCache:     informersCache,
		scheme:             scheme,
		volumeScheduler:    scheduler.New(cli, informersCache, systemConfig.MaxHAVolumeCount),
		volumeGroupManager: volumegroup.NewManager(cli, informersCache),

		nodeTaskQueue:    common.NewTaskQueue("NodeTask", maxRetries),
		k8sNodeTaskQueue: common.NewTaskQueue("K8sNodeTask", maxRetries),

		volumeTaskQueue:        common.NewTaskQueue("VolumeTask", maxRetries),
		volumeExpandTaskQueue:  common.NewTaskQueue("VolumeExpandTask", maxRetries),
		volumeMigrateTaskQueue: common.NewTaskQueue("VolumeMigrateTask", maxRetries),
		volumeConvertTaskQueue: common.NewTaskQueue("VolumeConvertTask", maxRetries),

		volumeGroupMigrateTaskQueue:    common.NewTaskQueue("VolumeGroupMigrateTask", maxRetries),
		volumeGroupConvertTaskQueue:    common.NewTaskQueue("VolumeGroupConvertTask", maxRetries),
		volumeSnapshotTaskQueue:        common.NewTaskQueue("VolumeSnapshotTask", maxRetries),
		volumeSnapshotRecoverTaskQueue: common.NewTaskQueue("VolumeSnapshotRecoverTask", maxRetries),
		localNodes:                     map[string]apisv1alpha1.State{},
		logger:                         log.WithField("Module", "ControllerManager"),
		dataCopyManager:                dcm,
	}, nil
}

func (m *manager) Run(stopCh <-chan struct{}) {
	m.volumeGroupManager.Init(stopCh)

	m.dataCopyManager.Run()

	go m.start(stopCh)
}

func (m *manager) start(stopCh <-chan struct{}) {
	runFunc := func(ctx context.Context) {
		m.logger.Info("Successfully became the cluster controller")

		m.volumeScheduler.Init()

		go m.syncNodesStatusForever(stopCh)
		go m.startNodeTaskWorker(stopCh)
		go m.startK8sNodeTaskWorker(stopCh)

		go m.startVolumeTaskWorker(stopCh)
		go m.startVolumeExpandTaskWorker(stopCh)
		go m.startVolumeMigrateTaskWorker(stopCh)
		go m.startVolumeConvertTaskWorker(stopCh)
		go m.startVolumeSnapshotTaskWorker(stopCh)
		go m.startVolumeSnapshotRecoverTaskWorker(stopCh)

		m.setupInformers()

		<-stopCh
		m.logger.Info("Stopped cluster controller")
	}

	m.logger.Debug("Trying to run as the cluster controller")
	if err := utils.RunWithLease(m.namespace, m.name, apis.ControllerLeaseName, runFunc); err != nil {
		m.logger.WithError(err).Fatal("failed to initialize leader election")
	}
}

func (m *manager) setupInformers() {
	volumeInformer, err := m.informersCache.GetInformer(context.TODO(), &apisv1alpha1.LocalVolume{})
	if err != nil {
		// error happens, crash the node
		m.logger.WithError(err).Fatal("Failed to get informer for Volume")
	}
	volumeInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: m.handleVolumeCRDDeletedEvent,
		UpdateFunc: m.handleVolumeCRDUpdateEvent,
	})

	expansionInformer, err := m.informersCache.GetInformer(context.TODO(), &apisv1alpha1.LocalVolumeExpand{})
	if err != nil {
		// error happens, crash the node
		m.logger.WithError(err).Fatal("Failed to get informer for VolumeExpand")
	}
	expansionInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: m.handleVolumeExpandCRDDeletedEvent,
	})

	convertInformer, err := m.informersCache.GetInformer(context.TODO(), &apisv1alpha1.LocalVolumeConvert{})
	if err != nil {
		// error happens, crash the node
		m.logger.WithError(err).Fatal("Failed to get informer for VolumeConvert")
	}
	convertInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: m.handleVolumeConvertCRDDeletedEvent,
	})

	migrateInformer, err := m.informersCache.GetInformer(context.TODO(), &apisv1alpha1.LocalVolumeMigrate{})
	if err != nil {
		// error happens, crash the node
		m.logger.WithError(err).Fatal("Failed to get informer for VolumeMigrate")
	}
	migrateInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: m.handleVolumeMigrateCRDDeletedEvent,
	})

	k8sNodeInformer, err := m.informersCache.GetInformer(context.TODO(), &corev1.Node{})
	if err != nil {
		// error happens, crash the node
		m.logger.WithError(err).Fatal("Failed to get informer for k8s Node")
	}
	k8sNodeInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: m.handleK8sNodeUpdatedEvent,
	})

	podInformer, err := m.informersCache.GetInformer(context.TODO(), &corev1.Pod{})
	if err != nil {
		m.logger.WithError(err).Fatal("Failed to get informer for k8s Pod")
	}
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    m.handlePodAddEvent,
		UpdateFunc: m.handlePodUpdateEvent,
	})

	// setup LocalVolumeSnapshot informer
	volumeSnapshotInformer, err := m.informersCache.GetInformer(context.TODO(), &apisv1alpha1.LocalVolumeSnapshot{})
	if err != nil {
		// error happens, crash the node
		m.logger.WithError(err).Fatal("Failed to get informer for LocalVolumeSnapshot")
	}
	volumeSnapshotInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    m.handleVolumeSnapshotAddEvent,
		UpdateFunc: m.handleVolumeSnapshotUpdateEvent,
	})
	// setup LocalVolumeSnapshotRecover informer
	volumeSnapshotRecoverInformer, err := m.informersCache.GetInformer(context.TODO(), &apisv1alpha1.LocalVolumeSnapshotRecover{})
	if err != nil {
		// error happens, crash the node
		m.logger.WithError(err).Fatal("Failed to get informer for LocalVolumeSnapshotRecover")
	}
	volumeSnapshotRecoverInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    m.handleVolumeSnapshotRecoverAddEvent,
		UpdateFunc: m.handleVolumeSnapshotRecoverUpdateEvent,
	})

	// setup LocalVolumeReplicaSnapshot informer
	volumeReplicaSnapshotInformer, err := m.informersCache.GetInformer(context.TODO(), &apisv1alpha1.LocalVolumeReplicaSnapshot{})
	if err != nil {
		// error happens, crash the node
		m.logger.WithError(err).Fatal("Failed to get informer for LocalVolumeReplicaSnapshot")
	}
	volumeReplicaSnapshotInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    m.handleVolumeSnapshotAddEvent,
		UpdateFunc: m.handleVolumeSnapshotUpdateEvent,
		DeleteFunc: m.handleVolumeSnapshotDeleteEvent,
	})

	// setup LocalVolumeReplicaSnapshotRecover informer
	volumeReplicaSnapshotRecoverInformer, err := m.informersCache.GetInformer(context.TODO(), &apisv1alpha1.LocalVolumeReplicaSnapshotRecover{})
	if err != nil {
		// error happens, crash the node
		m.logger.WithError(err).Fatal("Failed to get informer for LocalVolumeReplicaSnapshotRecover")
	}
	volumeReplicaSnapshotRecoverInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    m.handleVolumeSnapshotRecoverAddEvent,
		UpdateFunc: m.handleVolumeSnapshotRecoverUpdateEvent,
		DeleteFunc: m.handleVolumeSnapshotRecoverDeleteEvent,
	})
}

func (m *manager) handleVolumeSnapshotDeleteEvent(newObj interface{}) {
	m.handleVolumeSnapshotAddEvent(newObj)
}

func (m *manager) handleVolumeSnapshotUpdateEvent(oldObj, newObj interface{}) {
	m.handleVolumeSnapshotAddEvent(newObj)
}

func (m *manager) handleVolumeSnapshotAddEvent(newObject interface{}) {
	volumeSnapshot, ok := newObject.(*apisv1alpha1.LocalVolumeSnapshot)
	if ok {
		m.volumeSnapshotTaskQueue.Add(volumeSnapshot.Name)
		return
	}
	volumeReplicaSnapshot, ok := newObject.(*apisv1alpha1.LocalVolumeReplicaSnapshot)
	if ok {
		m.volumeSnapshotTaskQueue.Add(volumeReplicaSnapshot.Spec.VolumeSnapshotName)
		return
	}
	return
}

func (m *manager) handleVolumeSnapshotRecoverAddEvent(newObject interface{}) {
	volumeSnapshotRecover, ok := newObject.(*apisv1alpha1.LocalVolumeSnapshotRecover)
	if ok {
		m.volumeSnapshotRecoverTaskQueue.Add(volumeSnapshotRecover.Name)
		return
	}
	volumeReplicaSnapshotRecover, ok := newObject.(*apisv1alpha1.LocalVolumeReplicaSnapshotRecover)
	if ok {
		m.volumeSnapshotRecoverTaskQueue.Add(volumeReplicaSnapshotRecover.Spec.VolumeSnapshotRecover)
		return
	}
	return
}

func (m *manager) handleVolumeSnapshotRecoverUpdateEvent(oldObj, newObj interface{}) {
	m.handleVolumeSnapshotRecoverAddEvent(newObj)
}

func (m *manager) handleVolumeSnapshotRecoverDeleteEvent(newObj interface{}) {
	m.handleVolumeSnapshotRecoverAddEvent(newObj)
}

// VolumeScheduler retrieve the volume scheduler instance
func (m *manager) VolumeScheduler() apisv1alpha1.VolumeScheduler {
	return m.volumeScheduler
}

// VolumeGroupManager retrieves the volume group manager instance
func (m *manager) VolumeGroupManager() apisv1alpha1.VolumeGroupManager {
	return m.volumeGroupManager
}

// ReconcileNode reconciles Node CRD for any node resource change
func (m *manager) ReconcileNode(node *apisv1alpha1.LocalStorageNode) {
	m.nodeTaskQueue.Add(node.Name)
}

// ReconcileVolume reconciles Volume CRD for any volume resource change
func (m *manager) ReconcileVolume(vol *apisv1alpha1.LocalVolume) {
	m.volumeTaskQueue.Add(vol.Name)
}

// ReconcileVolumeGroup reconciles VolumeGroup CRD for any volume resource change
func (m *manager) ReconcileVolumeGroup(volGroup *apisv1alpha1.LocalVolumeGroup) {
	m.volumeGroupManager.ReconcileVolumeGroup(volGroup)
}

// ReconcileVolumeExpand reconciles VolumeExpand CRD for any volume resource change
func (m *manager) ReconcileVolumeExpand(expand *apisv1alpha1.LocalVolumeExpand) {
	m.volumeExpandTaskQueue.Add(expand.Name)
}

// ReconcileVolumeMigrate reconciles VolumeMigrate CRD for any volume resource change
func (m *manager) ReconcileVolumeMigrate(migrate *apisv1alpha1.LocalVolumeMigrate) {
	m.volumeMigrateTaskQueue.Add(migrate.Name)
}

// ReconcileVolumeConvert reconciles VolumeConvert CRD for any volume resource change
func (m *manager) ReconcileVolumeConvert(convert *apisv1alpha1.LocalVolumeConvert) {
	m.volumeConvertTaskQueue.Add(convert.Name)
}

func (m *manager) handleK8sNodeUpdatedEvent(_, newObj interface{}) {
	newNode, _ := newObj.(*corev1.Node)
	if _, ok := m.localNodes[newNode.Name]; !ok {
		// ignore not-interested node
		return
	}
	newConds := map[corev1.NodeConditionType]corev1.ConditionStatus{}
	for _, cond := range newNode.Status.Conditions {
		newConds[cond.Type] = cond.Status
	}
	if newConds[corev1.NodeReady] == corev1.ConditionUnknown {
		m.k8sNodeTaskQueue.Add(newNode.Name)
	}
}

func (m *manager) handleVolumeCRDUpdateEvent(oldObj, newObj interface{}) {
	oldVol := oldObj.(*apisv1alpha1.LocalVolume)
	newVol := newObj.(*apisv1alpha1.LocalVolume)

	// if volume's replica update, we should notify its group
	if !reflect.DeepEqual(oldVol.Spec.Accessibility.Nodes, newVol.Spec.Accessibility.Nodes) {
		lvg, err := m.queryLocalVolumeGroup(context.TODO(), newVol.Spec.VolumeGroup)
		if err != nil {
			m.logger.WithError(err).Error("Failed to query local volume group")
		}
		m.ReconcileVolumeGroup(lvg)
	}
}

func (m *manager) handleVolumeCRDDeletedEvent(obj interface{}) {
	instance, _ := obj.(*apisv1alpha1.LocalVolume)
	m.logger.WithFields(log.Fields{"volume": instance.Name, "spec": instance.Spec, "status": instance.Status}).Info("Observed a Volume CRD deletion...")
	if instance.Status.State != apisv1alpha1.VolumeStateDeleted {
		// must be deleted by a mistake, rebuild it
		// TODO: need retry considering the case of creating failure
		newInstance := &apisv1alpha1.LocalVolume{}
		newInstance.Name = instance.Name
		newInstance.Spec = instance.Spec
		newInstance.Status = instance.Status

		m.logger.WithFields(log.Fields{"volume": instance.Name, "spec": instance.Spec, "status": instance.Status}).Info("Rebuilding a Volume CRD ...")
		if err := m.apiClient.Create(context.TODO(), newInstance); err != nil {
			m.logger.WithError(err).Error("Failed to rebuild Volume")
		}
		if err := m.apiClient.Status().Update(context.TODO(), newInstance); err != nil {
			m.logger.WithError(err).Error("Failed to rebuild Volume's status")
		}
	}
}

func (m *manager) handleVolumeExpandCRDDeletedEvent(obj interface{}) {
	instance, _ := obj.(*apisv1alpha1.LocalVolumeExpand)
	m.logger.WithFields(log.Fields{"expand": instance.Name, "spec": instance.Spec, "status": instance.Status}).Info("Observed a VolumeExpand CRD deletion...")
	if instance.Status.State != apisv1alpha1.OperationStateCompleted && instance.Status.State != apisv1alpha1.OperationStateAborted {
		// must be deleted by a mistake, rebuild it
		// TODO: need retry considering the case of creating failure
		newInstance := &apisv1alpha1.LocalVolumeExpand{}
		newInstance.Name = instance.Name
		newInstance.Spec = instance.Spec
		newInstance.Status = instance.Status

		m.logger.WithFields(log.Fields{"expand": instance.Name, "spec": instance.Spec, "status": instance.Status}).Info("Rebuilding a VolumeExpand CRD ...")
		if err := m.apiClient.Create(context.TODO(), newInstance); err != nil {
			m.logger.WithError(err).Error("Failed to rebuild VolumeExpand")
		}
		if err := m.apiClient.Status().Update(context.TODO(), newInstance); err != nil {
			m.logger.WithError(err).Error("Failed to rebuild VolumeExpand's status")
		}
	}
}

func (m *manager) handleVolumeConvertCRDDeletedEvent(obj interface{}) {
	instance, _ := obj.(*apisv1alpha1.LocalVolumeConvert)
	m.logger.WithFields(log.Fields{"expand": instance.Name, "spec": instance.Spec, "status": instance.Status}).Info("Observed a VolumeConvert CRD deletion...")
	if instance.Status.State != apisv1alpha1.OperationStateCompleted && instance.Status.State != apisv1alpha1.OperationStateAborted {
		// must be deleted by a mistake, rebuild it
		// TODO: need retry considering the case of creating failure
		newInstance := &apisv1alpha1.LocalVolumeConvert{}
		newInstance.Name = instance.Name
		newInstance.Spec = instance.Spec
		newInstance.Status = instance.Status

		m.logger.WithFields(log.Fields{"expand": instance.Name, "spec": instance.Spec, "status": instance.Status}).Info("Rebuilding a VolumeConvert CRD ...")
		if err := m.apiClient.Create(context.TODO(), newInstance); err != nil {
			m.logger.WithError(err).Error("Failed to rebuild VolumeConvert")
		}
		if err := m.apiClient.Status().Update(context.TODO(), newInstance); err != nil {
			m.logger.WithError(err).Error("Failed to rebuild VolumeConvert's status")
		}
	}
}

func (m *manager) handleVolumeMigrateCRDDeletedEvent(obj interface{}) {
	instance, _ := obj.(*apisv1alpha1.LocalVolumeMigrate)
	m.logger.WithFields(log.Fields{"migrate": instance.Name, "spec": instance.Spec, "status": instance.Status}).Info("Observed a VolumeMigrate CRD deletion...")
	if !instance.Spec.Abort && instance.Status.State != apisv1alpha1.OperationStateCompleted && instance.Status.State != apisv1alpha1.OperationStateAborted {
		// must be deleted by a mistake, rebuild it
		// TODO: need retry considering the case of creating failure
		newInstance := &apisv1alpha1.LocalVolumeMigrate{}
		newInstance.Name = instance.Name
		newInstance.Spec = instance.Spec
		newInstance.Status = instance.Status

		m.logger.WithFields(log.Fields{"volume": instance.Name, "spec": instance.Spec, "status": instance.Status}).Info("Rebuilding a VolumeMigrate CRD ...")
		if err := m.apiClient.Create(context.TODO(), newInstance); err != nil {
			m.logger.WithError(err).Error("Failed to rebuild VolumeMigrate")
		}
		if err := m.apiClient.Status().Update(context.TODO(), newInstance); err != nil {
			m.logger.WithError(err).Error("Failed to rebuild VolumeMigrate's status")
		}
	}
}

func (m *manager) handlePodUpdateEvent(_, nObj interface{}) {
	pod, _ := nObj.(*corev1.Pod)

	// this is for the pod orphan pod which is abandoned by migration rclone job
	m.rclonePodGC(pod)
}

func (m *manager) handlePodAddEvent(obj interface{}) {
	pod, _ := obj.(*corev1.Pod)
	m.rclonePodGC(pod)
}
