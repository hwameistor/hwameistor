package manager

import (
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	mgrpkg "sigs.k8s.io/controller-runtime/pkg/manager"

	hwameistorctr "github.com/hwameistor/hwameistor/pkg/apiserver/manager/hwameistor"
	utils "github.com/hwameistor/hwameistor/pkg/apiserver/util"
)

type ServerManager struct {
	nodeName string

	namespace string

	apiClient client.Client

	clientset *kubernetes.Clientset

	lsnController *hwameistorctr.LocalStorageNodeController

	lvController *hwameistorctr.LocalVolumeController

	vgController *hwameistorctr.VolumeGroupController

	mController *hwameistorctr.MetricController

	lspController *hwameistorctr.LocalStoragePoolController

	settingController *hwameistorctr.SettingController

	ldController *hwameistorctr.LocalDiskController

	ldnController *hwameistorctr.LocalDiskNodeController

	mgr mgrpkg.Manager

	logger *log.Entry
}

// NewServerManager
func NewServerManager(mgr mgrpkg.Manager, clientset *kubernetes.Clientset) (*ServerManager, error) {
	var recorder record.EventRecorder
	return &ServerManager{
		nodeName:          utils.GetNodeName(),
		namespace:         utils.GetNamespace(),
		apiClient:         mgr.GetClient(),
		clientset:         clientset,
		lsnController:     hwameistorctr.NewLocalStorageNodeController(mgr.GetClient(), clientset, recorder),
		lvController:      hwameistorctr.NewLocalVolumeController(mgr.GetClient(), clientset, recorder),
		mController:       hwameistorctr.NewMetricController(mgr.GetClient(), clientset, recorder),
		lspController:     hwameistorctr.NewLocalStoragePoolController(mgr.GetClient(), clientset, recorder),
		settingController: hwameistorctr.NewSettingController(mgr.GetClient(), clientset, recorder),
		vgController:      hwameistorctr.NewVolumeGroupController(mgr.GetClient(), clientset, recorder),
		mgr:               mgr,
		logger:            log.WithField("Module", "ServerManager"),
	}, nil
}

func (m *ServerManager) StorageNodeController() *hwameistorctr.LocalStorageNodeController {
	var recorder record.EventRecorder
	if m.lsnController == nil {
		m.lsnController = hwameistorctr.NewLocalStorageNodeController(m.mgr.GetClient(), m.clientset, recorder)
	}
	return m.lsnController
}

func (m *ServerManager) VolumeController() *hwameistorctr.LocalVolumeController {
	var recorder record.EventRecorder
	if m.lvController == nil {
		m.lvController = hwameistorctr.NewLocalVolumeController(m.mgr.GetClient(), m.clientset, recorder)
	}
	return m.lvController
}

func (m *ServerManager) MetricController() *hwameistorctr.MetricController {

	var recorder record.EventRecorder
	if m.mController == nil {
		m.mController = hwameistorctr.NewMetricController(m.mgr.GetClient(), m.clientset, recorder)
	}
	return m.mController
}

func (m *ServerManager) StoragePoolController() *hwameistorctr.LocalStoragePoolController {

	var recorder record.EventRecorder
	if m.lspController == nil {
		m.lspController = hwameistorctr.NewLocalStoragePoolController(m.mgr.GetClient(), m.clientset, recorder)
	}
	return m.lspController
}

func (m *ServerManager) SettingController() *hwameistorctr.SettingController {

	var recorder record.EventRecorder
	if m.settingController == nil {
		m.settingController = hwameistorctr.NewSettingController(m.mgr.GetClient(), m.clientset, recorder)
	}
	return m.settingController
}

func (m *ServerManager) VolumeGroupController() *hwameistorctr.VolumeGroupController {

	var recorder record.EventRecorder
	if m.vgController == nil {
		m.vgController = hwameistorctr.NewVolumeGroupController(m.mgr.GetClient(), m.clientset, recorder)
	}
	return m.vgController
}

func (m *ServerManager) LocalDiskController() *hwameistorctr.LocalDiskController {

	var recorder record.EventRecorder
	if m.ldController == nil {
		m.ldController = hwameistorctr.NewLocalDiskController(m.mgr.GetClient(), recorder)
	}
	return m.ldController
}

func (m *ServerManager) LocalDiskNodeController() *hwameistorctr.LocalDiskNodeController {

	var recorder record.EventRecorder
	if m.ldnController == nil {
		m.ldnController = hwameistorctr.NewLocalDiskNodeController(m.mgr.GetClient(), recorder)
	}
	return m.ldnController
}
