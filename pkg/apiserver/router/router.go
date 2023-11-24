package api

import (
	"context"
	"fmt"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	mgrpkg "sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/apiserver/api"
	"github.com/hwameistor/hwameistor/pkg/apiserver/controller"
	"github.com/hwameistor/hwameistor/pkg/apiserver/manager"
)

var (
	RetryCounts         = 5
	RetryInterval       = 100 * time.Millisecond
	MetricsHost         = "0.0.0.0"
	MetricsPort   int32 = 8384
)

func CollectRoute(r *gin.Engine) *gin.Engine {
	log.Info("CollectRoute start ...")

	sm, _ := BuildServerMgr()
	v1 := r.Group("/apis/hwameistor.io/v1alpha1")

	authController := controller.NewAuthController(sm)
	v1.POST("/cluster/auth/auth", authController.Auth)
	v1.POST("/cluster/auth/logout", authController.Logout)
	v1.GET("/cluster/auth/info", authController.Info)
	// middleware should be first be register to router, the previous route will not pass the middleware
	v1.Use(authController.GetAuthMiddleWare())

	metricsController := controller.NewMetricsController(sm)
	v1.GET("/cluster/status", metricsController.ModuleStatus)
	v1.GET("/cluster/operations", metricsController.OperationList)
	v1.GET("/cluster/operations/:operationName", metricsController.OperationGet)
	v1.GET("/cluster/events", metricsController.EventList)
	v1.GET("/cluster/events/:eventName", metricsController.EventGet)

	volumeController := controller.NewVolumeController(sm)

	v1.GET("/cluster/volumes", volumeController.VolumeList)
	v1.GET("/cluster/volumes/:volumeName", volumeController.VolumeGet)

	v1.GET("/cluster/volumes/:volumeName/replicas", volumeController.VolumeReplicasGet)

	v1.GET("/cluster/volumes/:volumeName/migrate", volumeController.GetVolumeMigrateOperation)
	v1.POST("/cluster/volumes/:volumeName/migrate", volumeController.VolumeMigrateOperation)

	v1.GET("/cluster/volumes/:volumeName/convert", volumeController.GetVolumeConvertOperation)
	v1.POST("/cluster/volumes/:volumeName/convert", volumeController.VolumeConvertOperation)

	//volumes expand
	v1.GET("/cluster/volumes/:volumeName/expand", volumeController.GetVolumeExpandOperation)
	v1.POST("/cluster/volumes/:volumeName/expand", volumeController.VolumeExpandOperation)

	v1.GET("/cluster/volumes/:volumeName/operations", volumeController.VolumeOperationGet)
	v1.GET("/cluster/volumes/:volumeName/events", volumeController.VolumeEventList)
	//volumes snapshots
	v1.GET("/cluster/volumes/:volumeName/snapshot", volumeController.VolumeSnapshotList)

	volumeGroupController := controller.NewVolumeGroupController(sm)
	v1.GET("/cluster/volumegroups/:vgName", volumeGroupController.VolumeGroupGet)
	v1.GET("/cluster/volumegroups", volumeGroupController.VolumeGroupList)

	//snapshots
	snapController := controller.NewSnapshotController(sm)
	v1.GET("/cluster/snapshots", snapController.SnapshotList)
	v1.GET("/cluster/snapshot/:snapshotName/snapshot", snapController.SnapshotGet)

	ldController := controller.NewLocalDiskController(sm)
	v1.GET("/cluster/localdisks", ldController.LocalDiskList)

	ldnController := controller.NewLocalDiskNodeController(sm)
	v1.GET("/cluster/localdisknodes", ldnController.LocalDiskNodeList)

	nodeController := controller.NewNodeController(sm)
	v1.GET("/cluster/nodes", nodeController.StorageNodeList)
	v1.GET("/cluster/nodes/:nodeName", nodeController.StorageNodeGet)
	v1.GET("/cluster/nodes/:nodeName/migrates", nodeController.StorageNodeMigrateGet)
	v1.GET("/cluster/nodes/:nodeName/disks", nodeController.StorageNodeDisksList)
	//diskName is devicePath, not really ld-name
	v1.GET("/cluster/nodes/:nodeName/disks/:diskName", nodeController.GetStorageNodeDisk)
	v1.POST("/cluster/nodes/:nodeName/disks/:devicePath", nodeController.UpdateStorageNodeDisk)
	v1.POST("/cluster/nodes/:nodeName/disks/:devicePath/owner", nodeController.SetStorageNodeDiskOwner)
	v1.POST("/cluster/nodes/:nodeName", nodeController.StorageNodeUpdate)

	v1.GET("/cluster/nodes/:nodeName/pools", nodeController.StorageNodePoolsList)
	v1.GET("/cluster/nodes/:nodeName/pools/:poolName", nodeController.StorageNodePoolGet)
	v1.GET("/cluster/nodes/:nodeName/pools/:poolName/disks", nodeController.StorageNodePoolDisksList)
	v1.GET("/cluster/nodes/:nodeName/pools/:poolName/disks/:diskName", nodeController.StorageNodePoolDiskGet)

	poolController := controller.NewPoolController(sm)
	v1.GET("/cluster/pools", poolController.StoragePoolList)
	v1.GET("/cluster/pools/:poolName", poolController.StoragePoolGet)
	v1.GET("/cluster/pools/:poolName/nodes/:nodeName/disks", poolController.StorageNodeDisksGetByPoolName)
	v1.GET("/cluster/pools/:poolName/nodes", poolController.StorageNodesGetByPoolName)
	// Expand operation include add pool and expand pool, so we not add `:poolName` at here
	v1.POST("/cluster/pools/expand", poolController.StoragePoolExpand)

	settingController := controller.NewSettingController(sm)
	v1.POST("/cluster/drbd", settingController.EnableDRBDSetting)
	v1.GET("/cluster/drbd", settingController.DRBDSettingGet)

	log.Info("CollectRoute end ...")

	return r
}

func BuildServerMgr() (*manager.ServerManager, mgrpkg.Manager) {
	log.Info("buildServerMgr start ...")

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Set default manager options
	options := mgrpkg.Options{MetricsBindAddress: fmt.Sprintf("%s:%d", MetricsHost, MetricsPort)}

	// Create a new manager to provide shared dependencies and start components
	mgr, err := mgrpkg.New(cfg, options)
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Setup Scheme for all resources
	if err = api.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	if err = v1alpha1.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "Failed to setup scheme for ldm resources")
		os.Exit(1)
	}

	setIndexField(mgr.GetCache())

	// Setup all Controllers
	if err = controller.AddToManager(mgr); err != nil {
		log.Error(err)
		os.Exit(1)
	}

	stopCh := signals.SetupSignalHandler()
	// Start the resource controllers manager
	go func() {
		log.Info("Starting the manager of all storage resources.")
		if err := mgr.Start(stopCh); err != nil {
			log.WithError(err).Error("Failed to run resources manager")
			os.Exit(1)
		}
	}()

	uiClientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.WithError(err).Error("Failed to create client set")
		os.Exit(1)
	}

	// Create a new manager to provide shared dependencies and start components
	smgr, err := manager.NewServerManager(mgr, uiClientset)
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}
	return smgr, mgr
}

// setIndexField must be called after scheme has been added
func setIndexField(cache cache.Cache) {
	indexes := []struct {
		field string
		Func  func(client.Object) []string
	}{
		{
			field: "spec.nodeName",
			Func: func(obj client.Object) []string {
				return []string{obj.(*v1alpha1.LocalDisk).Spec.NodeName}
			},
		},
		{
			field: "spec.devicePath",
			Func: func(obj client.Object) []string {
				return []string{obj.(*v1alpha1.LocalDisk).Spec.DevicePath}
			},
		},
		{
			field: "spec.nodeName/devicePath",
			Func: func(obj client.Object) []string {
				return []string{obj.(*v1alpha1.LocalDisk).Spec.NodeName + "/" + obj.(*v1alpha1.LocalDisk).Spec.DevicePath}
			},
		},
	}

	for _, index := range indexes {
		if err := cache.IndexField(context.Background(), &v1alpha1.LocalDisk{}, index.field, index.Func); err != nil {
			log.Error(err, "failed to setup index field %s", index.field)
			// indexer is required, exit immediately if it fails, more details see issue: #1209
			os.Exit(1)
		}
		log.Info("setup index field successfully", "field", index.field)
	}
}
