package node

import (
	"context"
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/member/node/disk"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/member/node/pool"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/member/node/registry"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/member/node/volume"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/member/types"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils/kubernetes"
	"github.com/hwameistor/hwameistor/pkg/local-storage/common"
	log "github.com/sirupsen/logrus"
	cache2 "k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
)

// maxRetries is the number of times a task will be retried before it is dropped out of the queue.
// With the current rate-limiter in use math.Max(16s, (1s*2^(maxRetries-1))) the following numbers represent the times
// a task is going to be requeued:
//
// Infinitely retry
const maxRetries = 0

type VolumeManagerProvider func() volume.Manager
type DiskManagerProvider func() disk.Manager
type LocalRegistryProvider func() registry.Manager
type PoolManagerProvider func() pool.Manager

var (
	defaultVolumeManagerProvider VolumeManagerProvider = volume.New
	defaultDiskManagerProvider   DiskManagerProvider   = disk.New
	defaultLocalRegistryProvider LocalRegistryProvider = registry.New
	defaultPoolManagerProvider   PoolManagerProvider   = pool.New
	defaultPoolClasses                                 = []types.DevType{types.DevTypeHDD, types.DevTypeSSD, types.DevTypeNVMe}
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

	m.discoveryNodeResources()

	m.rebuildLocalPools()

	err := m.syncNodeResources()
	if err != nil {
		m.logger.WithError(err).Error("Failed to sync node resources")
		return err
	}

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
	// todo
}

// discoveryNodeResources collect resources on this node and storage to local registryManager
func (m *nodeManager) discoveryNodeResources() {
	// 1. collect disks managed to LocalDiskManager
	// 2. collect volumes managed by LocalDiskManager
	m.registryManager.DiscoveryResources()
}

// rebuildLocalPools according discovery disks and volumes
func (m *nodeManager) rebuildLocalPools() {
	m.lock.Lock()
	defer m.lock.Unlock()

	for _, class := range defaultPoolClasses {
		if m.pools[class] == nil {
			m.pools[class] = &apisv1alpha1.LocalPool{
				Class: class,
				Name:  types.GetLocalDiskPoolName(class),
				Path:  types.GetLocalDiskPoolPath(class),
			}
		}

		// rebuild discovery disks
		var discoveryDisks []apisv1alpha1.LocalDevice
		for _, classDisk := range m.registryManager.ListDisksByType(class) {
			discoveryDisks = append(discoveryDisks, apisv1alpha1.LocalDevice{
				DevPath:       classDisk.DevPath,
				Class:         classDisk.DiskType,
				CapacityBytes: classDisk.Capacity,
			})
		}
		m.pools[class].Disks = discoveryDisks

		// rebuild discovery volumes
		var discoveryVolumes []string
		for _, classVolume := range m.registryManager.ListVolumesByType(class) {
			discoveryVolumes = append(discoveryVolumes, classVolume.Name)
		}
		m.pools[class].Volumes = discoveryVolumes

		// calculate storage capacity including total, used, free, maxAvailable per volume
		// todo
	}
}

// syncNodeResources sync discovery resources to ApiServer
func (m *nodeManager) syncNodeResources() error {
	m.logger.Infof("Succeed to find resources on node %q", m.pools)
	return nil
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
	if localDisk.Spec.NodeName != m.nodeName || localDisk.Spec.Owner == apisv1alpha1.LocalDiskManager {
		return
	}
	m.diskTaskQueue.Add(localDisk.GetName())
}

func (m *nodeManager) handleLocalDiskUpdate(_, obj interface{}) {
	localDisk := obj.(*apisv1alpha1.LocalDisk)
	if localDisk.Spec.NodeName != m.nodeName || localDisk.Spec.Owner == apisv1alpha1.LocalDiskManager {
		return
	}
	m.diskTaskQueue.Add(localDisk.GetName())
}

func (m *nodeManager) handleLocalDiskDelete(obj interface{}) {
	localDisk := obj.(*apisv1alpha1.LocalDisk)
	if localDisk.Spec.NodeName != m.nodeName || localDisk.Spec.Owner == apisv1alpha1.LocalDiskManager {
		return
	}
	m.diskTaskQueue.Add(localDisk.GetName())
}

func (m *nodeManager) handleLocalDiskClaimAdd(obj interface{}) {
	localDiskClaim := obj.(*apisv1alpha1.LocalDiskClaim)
	if localDiskClaim.Spec.NodeName != m.nodeName || localDiskClaim.Spec.Owner != apisv1alpha1.LocalDiskManager {
		return
	}
	m.diskClaimTaskQueue.Add(localDiskClaim.GetName())
}

func (m *nodeManager) handleLocalDiskClaimUpdate(_, obj interface{}) {
	localDiskClaim := obj.(*apisv1alpha1.LocalDiskClaim)
	if localDiskClaim.Spec.NodeName != m.nodeName || localDiskClaim.Spec.Owner == apisv1alpha1.LocalDiskManager {
		return
	}
	m.diskClaimTaskQueue.Add(localDiskClaim.GetName())
}

func (m *nodeManager) handleLocalDiskClaimDelete(obj interface{}) {
	localDiskClaim := obj.(*apisv1alpha1.LocalDiskClaim)
	if localDiskClaim.Spec.NodeName != m.nodeName || localDiskClaim.Spec.Owner == apisv1alpha1.LocalDiskManager {
		return
	}
	m.diskClaimTaskQueue.Add(localDiskClaim.GetName())
}