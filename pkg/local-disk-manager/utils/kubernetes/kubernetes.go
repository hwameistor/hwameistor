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
			continue
		}
		log.Info("setup index field successfully", "field", index.field)
	}
}
