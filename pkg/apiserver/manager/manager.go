package manager

import (
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/handler/localdisk"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	mgrpkg "sigs.k8s.io/controller-runtime/pkg/manager"

	hwameistorctr "github.com/hwameistor/hwameistor/pkg/apiserver/manager/hwameistor"
)

type ServerManager struct {
	apiClient client.Client
	clientset *kubernetes.Clientset
	mgr       mgrpkg.Manager
	logger    *log.Entry

	lsnController     *hwameistorctr.LocalStorageNodeController
	lvController      *hwameistorctr.LocalVolumeController
	vgController      *hwameistorctr.VolumeGroupController
	mController       *hwameistorctr.MetricController
	lspController     *hwameistorctr.LocalStoragePoolController
	settingController *hwameistorctr.SettingController
	ldController      *hwameistorctr.LocalDiskController
	ldnController     *hwameistorctr.LocalDiskNodeController
	authController    *hwameistorctr.AuthController
	lvsController     *hwameistorctr.LocalSnapshotController
	scController      *hwameistorctr.StorageClassController
}

func NewServerManager(mgr mgrpkg.Manager, clientset *kubernetes.Clientset) (*ServerManager, error) {
	return &ServerManager{
		apiClient: mgr.GetClient(),
		clientset: clientset,
		mgr:       mgr,
		logger:    log.WithField("Module", "ServerManager"),
	}, nil
}

func (m *ServerManager) StorageNodeController() *hwameistorctr.LocalStorageNodeController {
	var recorder record.EventRecorder
	if m.lsnController == nil {
		m.lsnController = hwameistorctr.NewLocalStorageNodeController(m.mgr.GetClient(), m.clientset, recorder)
		// set localdisk handler
		localDiskRecorder := m.mgr.GetEventRecorderFor("apiserver-localdisk-controller")
		diskHandler := localdisk.NewLocalDiskHandler(m.mgr.GetClient(), localDiskRecorder)
		m.lsnController.SetLdHandler(diskHandler)
	}
	return m.lsnController
}

func (m *ServerManager) VolumeController() *hwameistorctr.LocalVolumeController {
	var recorder record.EventRecorder
	if m.lvController == nil {
		m.lvController = hwameistorctr.NewLocalVolumeController(m.mgr.GetClient(), recorder)
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

func (m *ServerManager) AuthController() *hwameistorctr.AuthController {
	var recorder record.EventRecorder
	if m.authController == nil {
		m.authController = hwameistorctr.NewAuthController(m.mgr.GetClient(), recorder)
	}
	return m.authController
}

func (m *ServerManager) SnapshotController() *hwameistorctr.LocalSnapshotController {
	var recorder record.EventRecorder
	if m.lvsController == nil {
		m.lvsController = hwameistorctr.NewLocalSnapshotController(m.mgr.GetClient(), recorder)
	}
	return m.lvsController
}
