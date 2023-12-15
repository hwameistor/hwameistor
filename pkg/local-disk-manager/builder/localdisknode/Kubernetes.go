package localdisknode

import (
	"context"
	clientset "github.com/hwameistor/hwameistor/pkg/apis/client/clientset/versioned"
	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Kubeclient struct {
	// clientset
	clientset clientset.Interface
	// kubeConfigPath
	//	kubeConfigPath string
}

// NewKubeclient
func NewKubeclient() (*Kubeclient, error) {
	c := &Kubeclient{}
	if cli, err := buildInClusterClientset(); err != nil {
		return nil, err
	} else {
		c.clientset = cli
	}

	return c, nil
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

func (k *Kubeclient) Patch(ldnOld, ldnNew *v1alpha1.LocalDiskNode) error {
	patch := client.MergeFrom(ldnOld)
	patchData, err := patch.Data(ldnNew)
	if err != nil {
		return err
	}
	_, err = k.clientset.HwameistorV1alpha1().LocalDiskNodes().Patch(context.Background(), ldnNew.GetName(), patch.Type(), patchData, v1.PatchOptions{})
	return err
}
func (k *Kubeclient) SetClient(cli clientset.Interface) {
	k.clientset = cli
}
