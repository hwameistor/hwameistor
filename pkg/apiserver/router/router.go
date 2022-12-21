package api

import (
	"fmt"
	"os"
	"time"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"k8s.io/client-go/kubernetes"

	"github.com/gin-gonic/gin"
	"github.com/hwameistor/hwameistor/pkg/apiserver/api"
	"github.com/hwameistor/hwameistor/pkg/apiserver/controller"
	"github.com/hwameistor/hwameistor/pkg/apiserver/manager"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	mgrpkg "sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

var (
	RetryCounts         = 5
	RetryInterval       = 100 * time.Millisecond
	metricsHost         = "0.0.0.0"
	metricsPort   int32 = 8384
)

func CollectRoute(r *gin.Engine) *gin.Engine {
	fmt.Println("CollectRoute start ...")

	sm, m := BuildServerMgr()

	v1 := r.Group("/apis/v1alpha1")
	metricsController := controller.NewMetricsController(sm)
	metricsRoutes := v1.Group("/metrics")
	metricsRoutes.GET("/basemetric", metricsController.BaseMetric)
	metricsRoutes.GET("/storagepoolusemetric", metricsController.StoragePoolUseMetric)
	metricsRoutes.GET("/nodestorageusemetric/:storagePoolClass", metricsController.NodeStorageUseMetric)
	metricsRoutes.GET("/modulestatusmetric", metricsController.ModuleStatusMetric)
	metricsRoutes.GET("/operations", metricsController.OperationList)

	volumeController := controller.NewVolumeController(sm)

	volumeRoutes := v1.Group("/volumes")
	volumeRoutes.GET("/volumes", volumeController.VolumeList)
	volumeRoutes.GET("/volumes/:name", volumeController.VolumeGet)
	volumeRoutes.GET("/volume/:name/yaml", volumeController.VolumeYamlGet)

	volumeRoutes.GET("/volumereplicas/:volumeName", volumeController.VolumeReplicasGet)
	volumeRoutes.GET("/volumereplica/:volumeReplicaName/yaml", volumeController.VolumeReplicaYamlGet)

	volumeRoutes.GET("/volumeoperations/:volumeName", volumeController.VolumeOperationGet)
	volumeRoutes.GET("/volumeoperation/:volumeOperationName/yaml", volumeController.VolumeOperationYamlGet)

	volumeRoutes.POST("/volumeoperation/:volumeName/migrate", volumeController.VolumeMigrateOperation)
	volumeRoutes.POST("/volumeoperation/:volumeName/convert", volumeController.VolumeConvertOperation)

	volumeRoutes.GET("/volumemigrateoperation/:targetNodeType/targetNodes", volumeController.GetTargetNodesByTargetNodeType)

	volumeGroupController := controller.NewVolumeGroupController(sm)
	volumeGroupRoutes := v1.Group("/volumegroups")
	volumeGroupRoutes.GET("/volumegroups/:name", volumeGroupController.VolumeListByVolumeGroup)

	nodeController := controller.NewNodeController(sm, m)
	nodeRoutes := v1.Group("/nodes")
	nodeRoutes.GET("/storagenodes", nodeController.StorageNodeList)
	nodeRoutes.GET("/storagenodes/:name", nodeController.StorageNodeGet)
	nodeRoutes.GET("/storagenode/:nodeName/migrates", nodeController.StorageNodeMigrateGet)

	nodeRoutes.GET("/storagenode/:nodeName/disks", nodeController.StorageNodeDisksList)
	nodeRoutes.GET("/storagenodeoperations/:migrateOperationName/yaml", nodeController.StorageNodeVolumeOperationYamlGet)

	nodeRoutes.POST("/storagenode/:nodeName/disks/:diskName/reserve", nodeController.ReserveStorageNodeDisk)
	nodeRoutes.POST("/storagenode/:nodeName/disks/:diskName/removereserve", nodeController.RemoveReserveStorageNodeDisk)

	poolController := controller.NewPoolController(sm)
	poolRoutes := v1.Group("/pools")
	poolRoutes.GET("/storagepools", poolController.StoragePoolList)
	poolRoutes.GET("/storagepools/:name", poolController.StoragePoolGet)
	poolRoutes.GET("/storagepool/:storagePoolName/nodes/:nodeName/disks", poolController.StorageNodeDisksGetByPoolName)
	poolRoutes.GET("/storagepool/:storagePoolName/nodes", poolController.StorageNodesGetByPoolName)

	settingController := controller.NewSettingController(sm)
	settingRoutes := v1.Group("/settings")
	settingRoutes.POST("/highavailabilitysetting/:enabledrbd", settingController.EnableDRBDSetting)
	settingRoutes.GET("/highavailabilitysetting", settingController.DRBDSettingGet)

	fmt.Println("CollectRoute end ...")

	return r
}

func BuildServerMgr() (*manager.ServerManager, mgrpkg.Manager) {
	fmt.Println("buildServerMgr start ...")

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Set default manager options
	options := mgrpkg.Options{MetricsBindAddress: fmt.Sprintf("%s:%d", metricsHost, metricsPort)}

	// Create a new manager to provide shared dependencies and start components
	mgr, err := mgrpkg.New(cfg, options)
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Setup Scheme for all resources
	if err := api.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	if err := apisv1alpha1.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "Failed to setup scheme for ldm resources")
		os.Exit(1)
	}

	// Setup all Controllers
	if err := controller.AddToManager(mgr); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	stopCh := signals.SetupSignalHandler()
	// Start the resource controllers manager
	go func() {
		log.Info("Starting the manager of all  storage resources.")
		if err := mgr.Start(stopCh); err != nil {
			log.WithError(err).Error("Failed to run resources manager")
			os.Exit(1)
		}
	}()

	uiClientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.WithError(err).Error("Failed to create client set")
	}

	// Create a new manager to provide shared dependencies and start components
	smgr, err := manager.NewServerManager(mgr, uiClientset)
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}
	return smgr, mgr
}
