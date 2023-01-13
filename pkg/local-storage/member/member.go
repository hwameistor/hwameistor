package member

import (
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	localapis "github.com/hwameistor/hwameistor/pkg/apis/hwameistor"
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	localctrl "github.com/hwameistor/hwameistor/pkg/local-storage/member/controller"
	localcsi "github.com/hwameistor/hwameistor/pkg/local-storage/member/csi"
	localnode "github.com/hwameistor/hwameistor/pkg/local-storage/member/node"
	localrest "github.com/hwameistor/hwameistor/pkg/local-storage/member/rest"
)

// Node is a member of the cluster.
// It has some data to be shared among all the controllers.
// So, it's a global variable
var nodeInstance localapis.LocalStorageMember

// Member gets member instance
func Member() localapis.LocalStorageMember {
	if nodeInstance == nil {
		nodeInstance = newMember()
	}
	return nodeInstance
}

// New a local storage member
func newMember() localapis.LocalStorageMember {
	return &localStorageMember{}
}

// Member is struct of local storage node
type localStorageMember struct {
	name string

	version string

	namespace string

	csiDriverName string

	apiClient client.Client

	informersCache cache.Cache

	csiDriver localcsi.Driver

	restServer localrest.Server

	controller localapis.ControllerManager

	nodeManager localapis.NodeManager

	systemConfig apisv1alpha1.SystemConfig

	recorder record.EventRecorder
}

func (m *localStorageMember) ConfigureBase(name string, namespace string, systemConfig apisv1alpha1.SystemConfig,
	cli client.Client, informersCache cache.Cache, recorder record.EventRecorder) localapis.LocalStorageMember {
	m.name = name
	m.version = localapis.Version
	m.namespace = namespace
	m.apiClient = cli
	m.informersCache = informersCache
	m.systemConfig = systemConfig
	m.recorder = recorder
	return m
}

func (m *localStorageMember) ConfigureNode(scheme *runtime.Scheme) localapis.LocalStorageMember {
	if m.nodeManager == nil {
		var err error
		m.nodeManager, err = localnode.New(m.name, m.namespace, m.apiClient, m.informersCache, m.systemConfig, scheme, m.recorder)
		if err != nil {
			panic(err)
		}
	}

	return m
}

func (m *localStorageMember) ConfigureController(scheme *runtime.Scheme) localapis.LocalStorageMember {
	if m.controller == nil {
		var err error
		m.controller, err = localctrl.New(m.name, m.namespace, m.apiClient, scheme, m.informersCache, m.systemConfig)
		if err != nil {
			panic(err)
		}
	}

	return m
}

func (m *localStorageMember) ConfigureCSIDriver(driverName string, sockAddr string) localapis.LocalStorageMember {
	if m.csiDriver == nil {
		m.csiDriver = localcsi.New(m.name, m.namespace, driverName, sockAddr, m, m.apiClient)
		m.csiDriverName = driverName
	}

	return m
}

func (m *localStorageMember) ConfigureRESTServer(httpPort int) localapis.LocalStorageMember {
	if m.restServer == nil {
		m.restServer = localrest.New(m.name, m.namespace, httpPort, m, m.apiClient)
	}

	return m
}

func (m *localStorageMember) Run(stopCh <-chan struct{}) {

	log.Debug("Starting node manager")
	m.nodeManager.Run(stopCh)

	log.Debug("Starting cluster controller")
	m.controller.Run(stopCh)

	log.Debug("Starting CSI driver")
	m.csiDriver.Run(stopCh)

	log.Debug("Starting REST server")
	m.restServer.Run(stopCh)

}

func (m *localStorageMember) Controller() localapis.ControllerManager {
	return m.controller
}

func (m *localStorageMember) Node() localapis.NodeManager {
	return m.nodeManager
}

func (m *localStorageMember) Name() string {
	return m.name
}

func (m *localStorageMember) Version() string {
	return m.version
}

func (m *localStorageMember) DriverName() string {
	return m.csiDriverName
}
