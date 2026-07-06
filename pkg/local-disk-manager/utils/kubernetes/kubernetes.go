package kubernetes

import (
	"context"
	"fmt"
	"os"
	"sync"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

var (
	mgr  manager.Manager
	once sync.Once
)

func NewClient() (client.Client, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	return client.New(cfg, client.Options{})
}

func NewClientWithCache() (client.Client, error) {
	once.Do(func() {
		initManager()
	})
	return mgr.GetClient(), nil
}

func NewClientSet() (*kubernetes.Clientset, error) {
	var (
		err error
		c   *rest.Config
	)
	if c, err = rest.InClusterConfig(); err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(c)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

func NewRecorderFor(name string) (record.EventRecorder, error) {
	if mgr == nil {
		return nil, fmt.Errorf("manager object is nil")
	}

	return mgr.GetEventRecorderFor(name), nil
}

func initManager() {
	cfg, err := config.GetConfig()
	if err != nil {
		log.WithError(err).Error("Failed to init manager")
		os.Exit(1)
		return
	}

	mgr, err = manager.New(cfg, manager.Options{
		MetricsBindAddress: "0",
	})

	if err != nil {
		log.WithError(err).Error("Failed to init manager")
		os.Exit(1)
		return
	}

	// Setup Scheme for node resources
	if err := v1alpha1.AddToScheme(mgr.GetScheme()); err != nil {
		log.WithError(err).Error("Failed to add scheme to manager")
		os.Exit(1)
		return
	}

	// Setup Cache for field index
	setIndexField(mgr.GetCache())

	go mgr.GetCache().Start(context.Background())
}

// LocalDiskIndex 描述 LocalDisk CRD 上需要注册的一个 field index。
// 测试代码在构造 controller-runtime fake client 时可以复用同一份定义，
// 保证 production 与单元测试的 index 一致。
type LocalDiskIndex struct {
	Field string
	Func  client.IndexerFunc
}

// LocalDiskFieldIndexes 是所有 production 侧注册在 LocalDisk 上的 field indexes。
var LocalDiskFieldIndexes = []LocalDiskIndex{
	{
		Field: "spec.nodeName",
		Func: func(obj client.Object) []string {
			return []string{obj.(*v1alpha1.LocalDisk).Spec.NodeName}
		},
	},
	{
		Field: "spec.devicePath",
		Func: func(obj client.Object) []string {
			return []string{obj.(*v1alpha1.LocalDisk).Spec.DevicePath}
		},
	},
	{
		Field: "spec.nodeName/devicePath",
		Func: func(obj client.Object) []string {
			return []string{obj.(*v1alpha1.LocalDisk).Spec.NodeName + "/" + obj.(*v1alpha1.LocalDisk).Spec.DevicePath}
		},
	},
}

// LocalDiskClaimIndex 描述 LocalDiskClaim CRD 上需要注册的 field index。
type LocalDiskClaimIndex struct {
	Field string
	Func  client.IndexerFunc
}

// LocalDiskClaimFieldIndexes 是所有 production 侧注册在 LocalDiskClaim 上的 field indexes。
// 目前仅包含 status.status，用于查找未 bound 的 DiskClaim。
var LocalDiskClaimFieldIndexes = []LocalDiskClaimIndex{
	{
		Field: "status.status",
		Func: func(obj client.Object) []string {
			return []string{string(obj.(*v1alpha1.LocalDiskClaim).Status.Status)}
		},
	},
}

// setIndexField must be called after scheme has been added
func setIndexField(cache cache.Cache) {
	for _, index := range LocalDiskFieldIndexes {
		if err := cache.IndexField(context.Background(), &v1alpha1.LocalDisk{}, index.Field, index.Func); err != nil {
			log.Error(err, "failed to setup index field %s", index.Field)
			continue
		}
		log.Info("setup index field successfully", "field", index.Field)
	}
	for _, index := range LocalDiskClaimFieldIndexes {
		if err := cache.IndexField(context.Background(), &v1alpha1.LocalDiskClaim{}, index.Field, index.Func); err != nil {
			log.Error(err, "failed to setup index field %s", index.Field)
			continue
		}
		log.Info("setup index field successfully", "field", index.Field)
	}
}
