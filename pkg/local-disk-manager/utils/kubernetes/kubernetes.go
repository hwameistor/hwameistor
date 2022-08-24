package kubernetes

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var mgr manager.Manager

func NewClient() (client.Client, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	return client.New(cfg, client.Options{})
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

func init() {
	cfg, err := config.GetConfig()
	if err != nil {
		log.WithError(err).Error("Failed to init manager")
		return
	}

	mgr, err = manager.New(cfg, manager.Options{
		MetricsBindAddress: "0",
	})
	if err != nil {
		log.WithError(err).Error("Failed to init manager")
		return
	}
}
