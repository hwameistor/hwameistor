package controller

import (
	"context"
	"sync"

	apis "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/local-storage"
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/local-storage/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-storage/common"
	"github.com/hwameistor/hwameistor/pkg/local-storage/member/controller/scheduler"
	"github.com/hwameistor/hwameistor/pkg/local-storage/member/controller/volumegroup"
	"github.com/hwameistor/hwameistor/pkg/local-storage/utils"
	datacopyutil "github.com/hwameistor/hwameistor/pkg/local-storage/utils/datacopy"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
	runtimecache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

	volumeGroupMigrateTaskQueue *common.TaskQueue

	volumeConvertTaskQueue *common.TaskQueue

	volumeGroupConvertTaskQueue *common.TaskQueue

	localNodes map[string]apisv1alpha1.State // nodeName -> status

	logger *log.Entry

	lock sync.Mutex

	dataCopyManager *datacopyutil.DataCopyManager

	rclone *datacopyutil.Rclone
}

// New cluster manager
func New(name string, namespace string, cli client.Client, scheme *runtime.Scheme, informersCache runtimecache.Cache, systemConfig apisv1alpha1.SystemConfig) (apis.ControllerManager, error) {
	dataCopyStatusCh := make(chan *datacopyutil.DataCopyStatus, 100)
	dcm, _ := datacopyutil.NewDataCopyManager(context.TODO(), "", cli, dataCopyStatusCh)
	return &manager{
		name:                        name,
		namespace:                   namespace,
		apiClient:                   cli,
		informersCache:              informersCache,
		scheme:                      scheme,
		volumeScheduler:             scheduler.New(cli, informersCache, systemConfig.MaxHAVolumeCount),
		volumeGroupManager:          volumegroup.NewManager(cli, informersCache),
		nodeTaskQueue:               common.NewTaskQueue("NodeTask", maxRetries),
		k8sNodeTaskQueue:            common.NewTaskQueue("K8sNodeTask", maxRetries),
		volumeTaskQueue:             common.NewTaskQueue("VolumeTask", maxRetries),
		volumeExpandTaskQueue:       common.NewTaskQueue("VolumeExpandTask", maxRetries),
		volumeMigrateTaskQueue:      common.NewTaskQueue("VolumeMigrateTask", maxRetries),
		volumeGroupMigrateTaskQueue: common.NewTaskQueue("VolumeGroupMigrateTask", maxRetries),
		volumeConvertTaskQueue:      common.NewTaskQueue("VolumeConvertTask", maxRetries),
		volumeGroupConvertTaskQueue: common.NewTaskQueue("VolumeGroupConvertTask", maxRetries),
		localNodes:                  map[string]apisv1alpha1.State{},
		logger:                      log.WithField("Module", "ControllerManager"),
		dataCopyManager:             dcm,
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

		go m.startVolumeTaskWorker(stopCh)

		go m.startVolumeExpandTaskWorker(stopCh)
		go m.startVolumeMigrateTaskWorker(stopCh)
		go m.startVolumeGroupMigrateTaskWorker(stopCh)
		go m.startVolumeConvertTaskWorker(stopCh)
		go m.startVolumeGroupConvertTaskWorker(stopCh)

		go m.startK8sNodeTaskWorker(stopCh)

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
	})

	expansionInformer, err := m.informersCache.GetInformer(context.TODO(), &apisv1alpha1.LocalVolumeExpand{})
	if err != nil {
		// error happens, crash the node
		m.logger.WithError(err).Fatal("Failed to get informer for VolumeExpand")
	}
	expansionInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: m.handleVolumeExpandCRDDeletedEvent,
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
	m.volumeMigrateTaskQueue.Add(migrate.Namespace + "/" + migrate.Name)
}

// ReconcileVolumeGroupMigrate reconciles VolumeGroupMigrate CRD for any localvolumegroup resource change
func (m *manager) ReconcileVolumeGroupMigrate(lvgmigrate *apisv1alpha1.LocalVolumeGroupMigrate) {
	m.volumeGroupMigrateTaskQueue.Add(lvgmigrate.Namespace + "/" + lvgmigrate.Name)
}

// ReconcileVolumeConvert reconciles VolumeConvert CRD for any volume resource change
func (m *manager) ReconcileVolumeConvert(convert *apisv1alpha1.LocalVolumeConvert) {
	m.volumeConvertTaskQueue.Add(convert.Namespace + "/" + convert.Name)
}

// ReconcileVolumeGroupConvert reconciles VolumeGroupConvert CRD for any volumegroup resource change
func (m *manager) ReconcileVolumeGroupConvert(lvgconvert *apisv1alpha1.LocalVolumeGroupConvert) {
	m.volumeGroupConvertTaskQueue.Add(lvgconvert.Namespace + "/" + lvgconvert.Name)
}

func (m *manager) handleK8sNodeUpdatedEvent(oldObj, newObj interface{}) {
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

func (m *manager) handleVolumeCRDDeletedEvent(obj interface{}) {
	instance, _ := obj.(*apisv1alpha1.LocalVolume)
	m.logger.WithFields(log.Fields{"volume": instance.Name, "spec": instance.Spec, "status": instance.Status}).Info("Observed a Volume CRD deletion...")
	if instance.Status.State != apisv1alpha1.VolumeStateDeleted {
		// must be deleted by a mistake, rebuild it
		// TODO: need retry considering the case of creating failure
		newInstance := &apisv1alpha1.LocalVolume{}
		newInstance.Name = instance.Name
		newInstance.Spec = instance.Spec

		m.logger.WithFields(log.Fields{"volume": instance.Name, "spec": instance.Spec, "status": instance.Status}).Info("Rebuilding a Volume CRD ...")
		if err := m.apiClient.Create(context.TODO(), newInstance); err != nil {
			m.logger.WithError(err).Error("Failed to rebuild Volume")
		}
	}
}

// ReconcileVolume reconciles Volume CRD for any volume resource change
func (m *manager) handleVolumeExpandCRDDeletedEvent(obj interface{}) {
	instance, _ := obj.(*apisv1alpha1.LocalVolumeExpand)
	m.logger.WithFields(log.Fields{"expand": instance.Name, "spec": instance.Spec, "status": instance.Status}).Info("Observed a VolumeExpand CRD deletion...")
	if instance.Status.State != apisv1alpha1.OperationStateCompleted && instance.Status.State != apisv1alpha1.OperationStateAborted {
		// must be deleted by a mistake, rebuild it
		// TODO: need retry considering the case of creating failure
		newInstance := &apisv1alpha1.LocalVolumeExpand{}
		newInstance.Name = instance.Name
		newInstance.Spec = instance.Spec

		m.logger.WithFields(log.Fields{"expand": instance.Name, "spec": instance.Spec, "status": instance.Status}).Info("Rebuilding a VolumeExpand CRD ...")
		if err := m.apiClient.Create(context.TODO(), newInstance); err != nil {
			m.logger.WithError(err).Error("Failed to rebuild VolumeExpand")
		}
	}
}

// ReconcileVolume reconciles Volume CRD for any volume resource change
func (m *manager) handleVolumeMigrateCRDDeletedEvent(obj interface{}) {
	instance, _ := obj.(*apisv1alpha1.LocalVolumeMigrate)
	m.logger.WithFields(log.Fields{"migrate": instance.Name, "spec": instance.Spec, "status": instance.Status}).Info("Observed a VolumeMigrate CRD deletion...")
	if instance.Spec.Abort != true && instance.Status.State != apisv1alpha1.OperationStateCompleted && instance.Status.State != apisv1alpha1.OperationStateAborted {
		// must be deleted by a mistake, rebuild it
		// TODO: need retry considering the case of creating failure
		newInstance := &apisv1alpha1.LocalVolumeMigrate{}
		newInstance.Namespace = instance.Namespace
		newInstance.Name = instance.Name
		newInstance.Spec = instance.Spec

		m.logger.WithFields(log.Fields{"volume": instance.Name, "spec": instance.Spec, "status": instance.Status}).Info("Rebuilding a VolumeMigrate CRD ...")
		if err := m.apiClient.Create(context.TODO(), newInstance); err != nil {
			m.logger.WithError(err).Error("Failed to rebuild VolumeMigrate")
		}
	}
}
