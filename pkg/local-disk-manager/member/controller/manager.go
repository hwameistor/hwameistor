package controller

import (
	"context"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils/kubernetes"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
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
}

// Options are the arguments for creating a new Manager
type Options struct {
	// NodeName represents where the Manager is running
	NodeName string

	// Namespace
	Namespace string

	// K8sClient is used to perform CRUD operations on Kubernetes objects
	K8sClient client.Client

	// Cache is used to load Kubernetes objects
	Cache cache.Cache

	// Logger  is the logger that should be used by this manager.
	// If none is set, it defaults to log.Log global logger.
	Logger *log.Entry

	Recorder record.EventRecorder
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
		nodeName:  options.NodeName,
		namespace: options.Namespace,
		k8sClient: options.K8sClient,
		cache:     options.Cache,
		logger:    options.Logger,
		lock:      sync.RWMutex{},
		recorder:  options.Recorder,
	}, nil
}

// nodeManager implements Manager interface
type nodeManager struct {
	nodeName string

	namespace string

	// k8sClient knows how to perform CRUD operations on Kubernetes objects.
	k8sClient client.Client

	// cache knows how to load Kubernetes objects
	cache cache.Cache

	logger *log.Entry

	lock sync.RWMutex

	recorder record.EventRecorder
}

// setOptionsDefaults set default values for Options fields
func setDefaultOptions(options Options) Options {
	if options.Logger == nil {
		options.Logger = log.WithField("Module", "NodeControllerManager")
	}

	return options
}

func (m *nodeManager) GetClient() client.Client {
	return m.k8sClient
}

func (m *nodeManager) GetCache() cache.Cache {
	return m.cache
}

func (m *nodeManager) Start(c context.Context) error {
	// start node heartbeats
	go m.startHeartBeatsDetection(c)

	<-c.Done()
	return nil
}
