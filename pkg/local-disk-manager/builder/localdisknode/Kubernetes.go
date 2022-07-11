package localdisknode

import (
	"context"

	clientset "github.com/hwameistor/hwameistor/pkg/apis/generated/local-disk-manager/clientset/versioned"
	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/local-disk-manager/v1alpha1"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type Kubeclient struct {
	// clientset
	clientset *clientset.Clientset

	// kubeConfigPath
	kubeConfigPath string
}

// NewKubeclient
func NewKubeclient() (*Kubeclient, error) {
	client := &Kubeclient{}
	if cli, err := buildInClusterClientset(); err != nil {
		return nil, err
	} else {
		client.clientset = cli
	}

	return client, nil
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

func (k *Kubeclient) Create(ldn *v1alpha1.LocalDiskNode) (*v1alpha1.LocalDiskNode, error) {
	return k.clientset.HwameistorV1alpha1().LocalDiskNodes().Create(context.Background(), ldn, v1.CreateOptions{})
}

func (k *Kubeclient) Get(name string) (*v1alpha1.LocalDiskNode, error) {
	return k.clientset.HwameistorV1alpha1().LocalDiskNodes().Get(context.Background(), name, v1.GetOptions{})
}

func (k *Kubeclient) List() (*v1alpha1.LocalDiskNodeList, error) {
	return k.clientset.HwameistorV1alpha1().LocalDiskNodes().List(context.Background(), v1.ListOptions{})
}

func (k *Kubeclient) Update(ldn *v1alpha1.LocalDiskNode) (*v1alpha1.LocalDiskNode, error) {
	return k.clientset.HwameistorV1alpha1().LocalDiskNodes().Update(context.Background(), ldn, v1.UpdateOptions{})
}

//
//func (k *Kubeclient) updateDiskStatus(node, devPath, status string) (*v1alpha1.LocalDiskNode, error) {
//	ldn, err := k.Get(node)
//	if err != nil {
//		return nil, err
//	}
//
//	ldnNew := ldn.DeepCopy()
//	for i, disk := range ldnNew.Spec.Disks {
//		if disk.DevPath == devPath {
//			ldnNew.Spec.Disks[i].Status = status
//			break
//		}
//	}
//
//	return k.clientset.HwameistorV1alpha1().LocalDiskNodes().Update(context.Background(), ldnNew, v1.UpdateOptions{})
//}

//func (k *Kubeclient) UpdateDiskStatusInUse(node, devPath string) (*v1alpha1.LocalDiskNode, error) {
//	return k.updateDiskStatus(node, devPath, "InUse")
//}
