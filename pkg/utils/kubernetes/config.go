package kubernetes

import (
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func GetConfig() (*rest.Config, error) {
	return config.GetConfig()
}

func NewDiscoveryClient() (*discovery.DiscoveryClient, error) {
	cfg, err := GetConfig()
	if err != nil {
		return nil, err
	}

	return discovery.NewDiscoveryClientForConfig(cfg)
}

func NewClientSet() (*kubernetes.Clientset, error) {
	cfg, err := GetConfig()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(cfg)
}
