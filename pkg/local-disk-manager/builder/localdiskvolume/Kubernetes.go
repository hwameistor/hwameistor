package localdiskvolume

import (
	"context"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"

	clientset "github.com/hwameistor/hwameistor/pkg/apis/client/clientset/versioned"
	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

type Kubeclient struct {
	// clientset
	clientset clientset.Interface

	// kubeConfigPath
	//kubeConfigPath string
}

// NewKubeclient
func NewKubeclient() (*Kubeclient, error) {
	client := &Kubeclient{}
	cli, err := buildInClusterClientset()
	client.clientset = cli

	return client, err
}

// buildInClusterClientset builds a kubernetes in-cluster clientset
func buildInClusterClientset() (*clientset.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.WithError(err).Error("Failed to build kubernetes config")
		return nil, err
	}
	return clientset.NewForConfig(config)
}

func (k *Kubeclient) Create(vol *v1alpha1.LocalDiskVolume) (*v1alpha1.LocalDiskVolume, error) {
	return k.clientset.HwameistorV1alpha1().LocalDiskVolumes().Create(context.Background(), vol, v1.CreateOptions{})
}

func (k *Kubeclient) Get(name string) (*v1alpha1.LocalDiskVolume, error) {
	return k.clientset.HwameistorV1alpha1().LocalDiskVolumes().Get(context.Background(), name, v1.GetOptions{})
}

func (k *Kubeclient) Update(volume *v1alpha1.LocalDiskVolume) (*v1alpha1.LocalDiskVolume, error) {
	return k.clientset.HwameistorV1alpha1().LocalDiskVolumes().Update(context.Background(), volume, v1.UpdateOptions{})
}
func (k *Kubeclient) SetClient(cli clientset.Interface) {
	k.clientset = cli
}
