package node

import (
	"context"
	"github.com/fsnotify/fsnotify"
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/member/controller/disk"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/member/node/pool"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/member/node/registry"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/member/node/volume"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/member/types"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils/kubernetes"
	"github.com/hwameistor/hwameistor/pkg/local-storage/common"
	utils2 "github.com/hwameistor/hwameistor/pkg/utils"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types2 "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	cache2 "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
	"time"
)

// maxRetries is the number of times a task will be retried before it is dropped out of the queue.
// With the current rate-limiter in use math.Max(16s, (1s*2^(maxRetries-1))) the following numbers represent the times
// a task is going to be requeued:
//
// Infinitely retry
const (
	maxRetries = 0
	duration   = time.Minute * 5
)

type VolumeManagerProvider func() volume.Manager
type DiskManagerProvider func() disk.Manager
type LocalRegistryProvider func() registry.Manager
type PoolManagerProvider func() pool.Manager

var (
	defaultVolumeManagerProvider VolumeManagerProvider = volume.New
	defaultDiskManagerProvider   DiskManagerProvider   = disk.New
	defaultLocalRegistryProvider LocalRegistryProvider = registry.New
	defaultPoolManagerProvider   PoolManagerProvider   = pool.New
)

// Manager  is responsible for managing node resources, including storage pools, disks, and processing-related resources.
type Manager interface {
	// GetClient returns a client.Client
	GetClient() client.Client

	// GetCache returns a cache.Cache
	GetCache() cache.Cache

	// Start all the registered controllers and blocks until the context is cancelled.
	// Returns an error if there is an error starting any controller.
	Start(ctx context.Context) error

	// DiskManager returns a disk.Manager
	DiskManager() disk.Manager

	// VolumeManager returns a volume.Manager
	VolumeManager() volume.Manager

	// LocalRegistry returns a registry.Manager
	LocalRegistry() registry.Manager

	// PoolManager returns a pool.Manager
	PoolManager() pool.Manager
}

// Options are the arguments for creating a new Manager
type Options struct {
	// NodeName represents where the Manager is running
	NodeName string

	// Namespace TBD.
	Namespace string

	// K8sClient is used to perform CRUD operations on Kubernetes objects
	K8sClient client.Client

	// Cache is used to load Kubernetes objects
	Cache cache.Cache

	// DiskTaskQueue is the queue stored LocalDisk objects
	DiskTaskQueue *common.TaskQueue

	// DiskClaimTaskQueue is the queue stored LocalDiskClaim objects
	DiskClaimTaskQueue *common.TaskQueue

	// DiskNodeTaskQueue is the queue stored LocalDiskNode objects
	DiskNodeTaskQueue *common.TaskQueue

	// Logger  is the logger that should be used by this manager.
	// If none is set, it defaults to log.Log global logger.
	Logger *log.Entry

	Recorder record.EventRecorder

	// VolumeManagerProvider provides the manager for Volumes
	VolumeManagerProvider

	// DiskManagerProvider provides the manager for Disks
	DiskManagerProvider

	// LocalRegistryProvider provides the manager for node resources
	LocalRegistryProvider

	// PoolManagerProvider provides the manager for DiskPool
	PoolManagerProvider
}

// NewManager returns a new Manager for creating Controllers.
func NewManager(options Options) (Manager, error) {
	// Set default values for options fields
	options = setDefaultOptions(options)

	if options.K8sClient == nil {
		if cli, err := kubernetes.NewClient(); err != nil {
			return nil, err
		} else {
			options.K8sClient = cli
		}
	}

	return &nodeManager{
		nodeName:           options.NodeName,
		namespace:          options.Namespace,
		k8sClient:          options.K8sClient,
		cache:              options.Cache,
		diskTaskQueue:      options.DiskTaskQueue,
		diskClaimTaskQueue: options.DiskClaimTaskQueue,
		diskNodeTaskQueue:  options.DiskNodeTaskQueue,
		logger:             options.Logger,
		lock:               sync.RWMutex{},
		diskManager:        options.DiskManagerProvider(),
		volumeManager:      options.VolumeManagerProvider(),
		registryManager:    options.LocalRegistryProvider(),
		poolManager:        options.PoolManagerProvider(),
		pools:              make(map[types.DevType]*apisv1alpha1.LocalPool),
		recorder:           options.Recorder,
	}, nil
}

// nodeManager is primarily responsible for creating data volumes, managing disks, monitoring related resources,
// and maintaining storage pools on the current node.
type nodeManager struct {
	nodeName string

	namespace string

	// k8sClient knows how to perform CRUD operations on Kubernetes objects.
	k8sClient client.Client

	// cache knows how to load Kubernetes objects
	cache cache.Cache

	diskTaskQueue *common.TaskQueue

	diskClaimTaskQueue *common.TaskQueue

	diskNodeTaskQueue *common.TaskQueue

	logger *log.Entry

	lock sync.RWMutex

	diskManager disk.Manager

	volumeManager volume.Manager

	poolManager pool.Manager

	registryManager registry.Manager

	pools map[types.DevType]*apisv1alpha1.LocalPool

	recorder record.EventRecorder
}

func (m *nodeManager) PoolManager() pool.Manager {
	return m.poolManager
}

func (m *nodeManager) GetClient() client.Client {
	return m.k8sClient
}

func (m *nodeManager) GetCache() cache.Cache {
	return m.cache
}

func (m *nodeManager) DiskManager() disk.Manager {
	return m.diskManager
}

func (m *nodeManager) VolumeManager() volume.Manager {
	return m.volumeManager
}

func (m *nodeManager) LocalRegistry() registry.Manager {
	return m.registryManager
}

// Start all registered task workers
func (m *nodeManager) Start(c context.Context) error {
	m.setupInformers()

	if err := m.poolManager.Init(); err != nil {
		m.logger.WithError(err).Error("Failed to init pool")
		return err
	}

	if err := m.register(); err != nil {
		m.logger.WithError(err).Error("Failed to register node")
		return err
	}

	go m.startPoolEventsWatcher(c)

	go m.startTimerSyncWorker(c)

	go m.startDiskTaskWorker(c)

	go m.startDiskClaimTaskWorker(c)

	go m.startDiskNodeTaskWorker(c)

	// We are done, Stop Node Manager
	<-c.Done()
	return nil
}

func (m *nodeManager) setupInformers() {
	// LocalDisk Informer
	diskInformer, err := m.cache.GetInformer(context.TODO(), &apisv1alpha1.LocalDisk{})
	if err != nil {
		m.logger.WithError(err).Fatalf("Failed to get informer for LocalDisk")
	}
	diskInformer.AddEventHandler(cache2.ResourceEventHandlerFuncs{
		AddFunc:    m.handleLocalDiskAdd,
		UpdateFunc: m.handleLocalDiskUpdate,
		DeleteFunc: m.handleLocalDiskDelete,
	})

	// LocalDiskClaim Informer
	diskClaimInformer, err := m.cache.GetInformer(context.TODO(), &apisv1alpha1.LocalDiskClaim{})
	if err != nil {
		m.logger.WithError(err).Fatalf("Failed to get informer for LocalDiskClaim")
	}
	diskClaimInformer.AddEventHandler(cache2.ResourceEventHandlerFuncs{
		AddFunc:    m.handleLocalDiskClaimAdd,
		UpdateFunc: m.handleLocalDiskClaimUpdate,
		DeleteFunc: m.handleLocalDiskClaimDelete,
	})

	// LocalDiskNode Informer
	diskNodeInformer, err := m.cache.GetInformer(context.TODO(), &apisv1alpha1.LocalDiskNode{})
	if err != nil {
		m.logger.WithError(err).Fatalf("Failed to get informer fot LocalDiskNode")
	}
	diskNodeInformer.AddEventHandler(cache2.ResourceEventHandlerFuncs{
		DeleteFunc: m.handleLocalDiskNodeDelete,
	})
}

// discoveryNodeResources collect resources on this node and storage to local registryManager
func (m *nodeManager) discoveryNodeResources() {
	// 1. collect disks managed to LocalDiskManager
	// 2. collect volumes managed by LocalDiskManager
	m.registryManager.DiscoveryResources()
}

// findDiskState find if disk inuse according to inuse disk list
var findDiskState = func(devPath string, inuseDisks []string) apisv1alpha1.State {
	if _, ok := utils.StrFind(inuseDisks, devPath); ok {
		return apisv1alpha1.DiskStateInUse
	}
	return apisv1alpha1.DiskStateAvailable
}

// rebuildLocalPools according discovery disks and volumes
func (m *nodeManager) rebuildLocalPools() {
	m.lock.Lock()
	defer m.lock.Unlock()

	for _, devType := range types.DefaultDevTypes {
		poolName := types.GetLocalDiskPoolName(devType)
		// rebuild discovery volumes
		var discoveryVolumes, inuseDisks []string
		var usedCapacity int64
		for _, classVolume := range m.registryManager.ListVolumesByType(devType) {
			discoveryVolumes = append(discoveryVolumes, classVolume.Name)
			inuseDisks = append(inuseDisks, classVolume.AttachPath)
			usedCapacity += classVolume.Capacity
		}

		// rebuild discovery disks
		var discoveryDisks []apisv1alpha1.LocalDevice
		var totalCapacity, maxCapacity int64
		for _, classDisk := range m.registryManager.ListDisksByType(devType) {
			discoveryDisk := apisv1alpha1.LocalDevice{
				DevPath:       classDisk.DevPath,
				Class:         classDisk.DiskType,
				CapacityBytes: classDisk.Capacity,
				State:         findDiskState(classDisk.DevPath, inuseDisks),
			}
			if discoveryDisk.State == apisv1alpha1.DiskStateAvailable && maxCapacity < classDisk.Capacity {
				maxCapacity = classDisk.Capacity
			}
			totalCapacity += classDisk.Capacity
			discoveryDisks = append(discoveryDisks, discoveryDisk)
		}

		if len(discoveryVolumes) == 0 && len(discoveryDisks) == 0 {
			delete(m.pools, poolName)
			continue
		}
		if m.pools[poolName] == nil {
			m.pools[poolName] = &apisv1alpha1.LocalPool{
				Class: devType,
				Type:  apisv1alpha1.PoolTypeRegular,
			}
		}
		m.pools[poolName].Volumes = discoveryVolumes
		m.pools[poolName].Disks = discoveryDisks
		m.pools[poolName].TotalCapacityBytes = totalCapacity
		m.pools[poolName].UsedCapacityBytes = usedCapacity
		m.pools[poolName].FreeCapacityBytes = totalCapacity - usedCapacity
		m.pools[poolName].TotalVolumeCount = int64(len(discoveryDisks))
		m.pools[poolName].UsedVolumeCount = int64(len(discoveryVolumes))
		m.pools[poolName].VolumeCapacityBytesLimit = maxCapacity
		m.pools[poolName].FreeVolumeCount = int64(len(discoveryDisks) - len(discoveryVolumes))
	}
}

// syncNodeResources sync discovery resources to ApiServer
func (m *nodeManager) syncNodeResources() error {
	m.logger.Info("Start to sync node resource")

	// 1. rebuild local registry
	m.discoveryNodeResources()

	// 2. rebuild local pool
	m.rebuildLocalPools()

	// 3. sync resources to ApiServer according to local pool
	diskNode := apisv1alpha1.LocalDiskNode{}
	err := m.k8sClient.Get(context.TODO(), types2.NamespacedName{Name: m.nodeName}, &diskNode)
	if err != nil {
		return err
	}
	diskNodeNew := diskNode.DeepCopy()

	m.lock.RLock()
	defer m.lock.RUnlock()
	if diskNode.Status.Pools == nil {
		diskNode.Status.Pools = make(map[types.DevType]apisv1alpha1.LocalPool)
	}
	var totalDisk, totalCapacity, freeCapacity int64 = 0, 0, 0
	localPools := make(map[types.DevType]apisv1alpha1.LocalPool)
	for pooName, localPool := range m.pools {
		lp := apisv1alpha1.LocalPool{}
		localPool.DeepCopyInto(&lp)
		localPools[pooName] = lp
		totalDisk += int64(len(lp.Disks))
		totalCapacity += lp.TotalCapacityBytes
		freeCapacity += lp.FreeCapacityBytes
	}
	diskNodeNew.Status.Pools = localPools
	diskNodeNew.Status.TotalDisk = totalDisk
	diskNodeNew.Status.TotalCapacity = totalCapacity
	diskNodeNew.Status.FreeCapacity = freeCapacity
	diskNodeNew.Status.State = apisv1alpha1.NodeStateReady
	m.updateStorageNodeCondition(diskNodeNew)
	patch := client.MergeFrom(&diskNode)
	if err = m.k8sClient.Status().Patch(context.TODO(), diskNodeNew, patch); err != nil {
		return err
	}

	m.logger.Info("Succeed to sync node resource")
	return nil
}

func (m *nodeManager) updateStorageNodeCondition(storageNode *apisv1alpha1.LocalDiskNode) {
	condition := apisv1alpha1.StorageNodeCondition{
		Status:             apisv1alpha1.ConditionTrue,
		LastUpdateTime:     metav1.Now(),
		LastTransitionTime: metav1.Now(),
	}

	if storageNode.Status.FreeCapacity > 0 {
		condition.Type = apisv1alpha1.StorageAvailable
		condition.Reason = "Storage" + string(apisv1alpha1.StorageAvailable)
		condition.Message = "Sufficient storage capacity"
	} else {
		condition.Type = apisv1alpha1.StorageUnAvailable
		condition.Reason = "Storage" + string(apisv1alpha1.StorageUnAvailable)
		condition.Message = "Insufficient storage capacity"
	}

	// only record StorageUnavailable events
	switch condition.Type {
	case apisv1alpha1.StorageUnAvailable:
		m.recorder.Event(storageNode, v1.EventTypeWarning, condition.Reason, condition.Message)
	default:
	}

	i, _ := utils2.GetStorageCondition(storageNode.Status.Conditions, condition.Type)
	if i == -1 {
		storageNode.Status.Conditions = append(storageNode.Status.Conditions, condition)
	} else {
		storageNode.Status.Conditions[i] = condition
	}
}

// UpdateCondition append current condition about LocalStorageNode, i.e. StorageExpandSuccess, StorageExpandFail, UnAvailable
func (m *nodeManager) UpdateCondition(condition apisv1alpha1.StorageNodeCondition) error {
	storageNode := &apisv1alpha1.LocalStorageNode{}
	if err := m.k8sClient.Get(context.TODO(), types2.NamespacedName{Name: m.nodeName}, storageNode); err != nil {
		m.logger.WithError(err).WithField("condition", condition).Error("Failed to query Node")
		return err
	}

	oldNode := storageNode.DeepCopy()
	switch condition.Type {
	case apisv1alpha1.StorageExpandFailure, apisv1alpha1.StorageUnAvailable:
		m.recorder.Event(storageNode, v1.EventTypeWarning, string(condition.Type), condition.Message)
	case apisv1alpha1.StorageExpandSuccess, apisv1alpha1.StorageProgressing:
		m.recorder.Event(storageNode, v1.EventTypeNormal, string(condition.Type), condition.Message)
	default:
		m.recorder.Event(storageNode, v1.EventTypeNormal, string(condition.Type), condition.Message)
	}

	i, _ := utils2.GetStorageCondition(storageNode.Status.Conditions, condition.Type)
	if i == -1 {
		storageNode.Status.Conditions = append(storageNode.Status.Conditions, condition)
	} else {
		storageNode.Status.Conditions[i] = condition
	}

	return m.k8sClient.Status().Patch(context.TODO(), storageNode, client.MergeFrom(oldNode))
}

func (m *nodeManager) register() error {
	diskNode := apisv1alpha1.LocalDiskNode{}
	err := m.k8sClient.Get(context.TODO(), types2.NamespacedName{Name: m.nodeName}, &diskNode)
	if err != nil {
		if errors.IsNotFound(err) {
			diskNode.Name = m.nodeName
			diskNode.Spec.NodeName = m.nodeName
			return m.k8sClient.Create(context.TODO(), &diskNode)
		}
		return err
	}
	diskNode.Spec.NodeName = m.nodeName
	return m.k8sClient.Update(context.TODO(), &diskNode)
}

// sync node resource timely
func (m *nodeManager) startTimerSyncWorker(c context.Context) {
	m.logger.WithField("period", duration.String()).Info("Start node resource sync timer worker")

	wait.Until(func() {
		if err := m.syncNodeResources(); err != nil {
			m.logger.WithError(err).Error("Failed to sync node resource")
		}
	}, duration, c.Done())

	m.logger.Info("Stop node resource sync timer worker")
}

func (m *nodeManager) startPoolEventsWatcher(c context.Context) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		m.logger.WithError(err).Fatal("Failed to start pool events watcher")
	}
	defer watcher.Close()

	for _, poolType := range types.DefaultDevTypes {
		volumePath := types.GetPoolVolumePath(poolType)
		if err := watcher.Add(volumePath); err != nil {
			m.logger.WithError(err).WithField("volumePath", volumePath).Fatal("Failed to add pool volume to watch path")
		}
		m.logger.Infof("Succeed to add pool volume %s to watch path ", volumePath)

		diskPath := types.GetPoolDiskPath(poolType)
		if err := watcher.Add(diskPath); err != nil {
			m.logger.WithError(err).WithField("diskPath", volumePath).Fatal("Failed to add pool disk to watch path")
		}
		m.logger.Infof("Succeed to add pool disk %s to watch path ", diskPath)
	}

	// Start listening for events.
	wait.Until(func() {
		m.logger.Info("Start watching pool events")
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				m.logger.WithField("event", event).Debug("Receive events from pool, start sync node resources")
				if err = m.syncNodeResources(); err != nil {
					m.logger.WithError(err).Error("Failed to sync node resource")
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					m.logger.WithError(err).Error("Pool events chan is closed, exiting now")
					return
				}
				m.logger.WithError(err).Error("Error happened when watch pool events, exiting now")
				return
			}
		}
	}, time.Second, c.Done())

	m.logger.Info("Stop watch pool events")
}

// setOptionsDefaults set default values for Options fields
func setDefaultOptions(options Options) Options {
	if options.Logger == nil {
		options.Logger = log.WithField("Module", "NodeManager")
	}

	if options.DiskTaskQueue == nil {
		options.DiskTaskQueue = common.NewTaskQueue("LocalDiskTask", maxRetries)
	}

	if options.DiskClaimTaskQueue == nil {
		options.DiskClaimTaskQueue = common.NewTaskQueue("LocalDiskClaimTask", maxRetries)
	}

	if options.DiskNodeTaskQueue == nil {
		options.DiskNodeTaskQueue = common.NewTaskQueue("LocalDiskNodeTask", maxRetries)
	}

	if options.DiskManagerProvider == nil {
		options.DiskManagerProvider = defaultDiskManagerProvider
	}

	if options.VolumeManagerProvider == nil {
		options.VolumeManagerProvider = defaultVolumeManagerProvider
	}

	if options.LocalRegistryProvider == nil {
		options.LocalRegistryProvider = defaultLocalRegistryProvider
	}

	if options.PoolManagerProvider == nil {
		options.PoolManagerProvider = defaultPoolManagerProvider
	}

	return options
}

func (m *nodeManager) handleLocalDiskAdd(obj interface{}) {
	localDisk := obj.(*apisv1alpha1.LocalDisk)
	if localDisk.Spec.NodeName != m.nodeName || localDisk.Spec.Owner != apisv1alpha1.LocalDiskManager {
		return
	}
	m.diskTaskQueue.Add(localDisk.GetName())
}

func (m *nodeManager) handleLocalDiskUpdate(_, obj interface{}) {
	localDisk := obj.(*apisv1alpha1.LocalDisk)
	if localDisk.Spec.NodeName != m.nodeName || localDisk.Spec.Owner != apisv1alpha1.LocalDiskManager {
		return
	}
	m.diskTaskQueue.Add(localDisk.GetName())
}

func (m *nodeManager) handleLocalDiskDelete(obj interface{}) {
	localDisk := obj.(*apisv1alpha1.LocalDisk)
	if localDisk.Spec.NodeName != m.nodeName || localDisk.Spec.Owner != apisv1alpha1.LocalDiskManager {
		return
	}
	m.diskTaskQueue.Add(localDisk.GetName())
}

func (m *nodeManager) handleLocalDiskClaimAdd(obj interface{}) {
	localDiskClaim := obj.(*apisv1alpha1.LocalDiskClaim)
	if localDiskClaim.Spec.NodeName != m.nodeName || localDiskClaim.Spec.Owner != apisv1alpha1.LocalDiskManager ||
		localDiskClaim.Status.Status != apisv1alpha1.LocalDiskClaimStatusBound {
		return
	}
	m.diskClaimTaskQueue.Add(localDiskClaim.GetName())
}

func (m *nodeManager) handleLocalDiskClaimUpdate(_, obj interface{}) {
	localDiskClaim := obj.(*apisv1alpha1.LocalDiskClaim)
	if localDiskClaim.Spec.NodeName != m.nodeName || localDiskClaim.Spec.Owner != apisv1alpha1.LocalDiskManager ||
		localDiskClaim.Status.Status != apisv1alpha1.LocalDiskClaimStatusBound {
		return
	}
	m.diskClaimTaskQueue.Add(localDiskClaim.GetName())
}

func (m *nodeManager) handleLocalDiskClaimDelete(obj interface{}) {
	localDiskClaim := obj.(*apisv1alpha1.LocalDiskClaim)
	if localDiskClaim.Spec.NodeName != m.nodeName || localDiskClaim.Spec.Owner != apisv1alpha1.LocalDiskManager ||
		localDiskClaim.Status.Status != apisv1alpha1.LocalDiskClaimStatusBound {
		return
	}
	m.diskClaimTaskQueue.Add(localDiskClaim.GetName())
}

func (m *nodeManager) handleLocalDiskNodeDelete(obj interface{}) {
	tobeDeletedNode := obj.(*apisv1alpha1.LocalDiskNode)
	if tobeDeletedNode.GetName() != m.nodeName {
		return
	}
	m.logger.WithFields(log.Fields{"node": tobeDeletedNode.GetName()}).Info("Observed LocalDiskNode resource deletion")

	rebuildNode := apisv1alpha1.LocalDiskNode{}
	rebuildNode.SetName(tobeDeletedNode.Name)
	rebuildNode.Spec = tobeDeletedNode.Spec
	rebuildNode.Status = tobeDeletedNode.Status
	// NOTE: it's useless to push deleted object to queue because reconcile will fail at fetching object, so rebuild here at once
	// try 5 times if occurs timeout error
	err := retry.OnError(retry.DefaultRetry, errors.IsTimeout, func() error {
		m.logger.WithField("node", rebuildNode.GetName()).Info("Rebuilding LocalDiskNode resource")
		err := m.k8sClient.Create(context.TODO(), &rebuildNode)
		if err != nil {
			if !errors.IsAlreadyExists(err) {
				return err
			}
		}
		return m.k8sClient.Status().Update(context.TODO(), &rebuildNode)
	})
	if err != nil {
		m.logger.WithError(err).Error("Failed to rebuild LocalDiskNode resource")
	}
	m.logger.WithField("node", rebuildNode.GetName()).Error("Succeed to rebuild LocalDiskNode resource")
}
