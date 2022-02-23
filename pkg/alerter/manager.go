package alerter

import (
	localstorageclient "github.com/hwameistor/local-storage/pkg/apis/client/clientset/versioned"
	localstoragealertclient "github.com/hwameistor/local-storage/pkg/apis/client/clientset/versioned/typed/localstorage/v1alpha1"
	localstorageinformers "github.com/hwameistor/local-storage/pkg/apis/client/informers/externalversions"

	"k8s.io/client-go/rest"
)

var (
	gAlertClient localstoragealertclient.LocalStorageAlertInterface
)

// AlertManager struct
type AlertManager struct {
	isVirtualNode bool
}

// NewManager creates an alert manager instance
func NewManager(isVirtualNode bool) *AlertManager {
	return &AlertManager{isVirtualNode: isVirtualNode}
}

// Run alert manager
func (m *AlertManager) Run(stopCh <-chan struct{}) error {

	cfg, _ := rest.InClusterConfig()
	apiClient := localstorageclient.NewForConfigOrDie(cfg)
	factory := localstorageinformers.NewSharedInformerFactory(apiClient, 0)
	factory.Start(stopCh)

	gAlertClient = apiClient.LocalStorageV1alpha1().LocalStorageAlerts()

	if !m.isVirtualNode {
		// not run disk alerter in virtual machine
		newDiskAlerter().Run(factory, stopCh)
	}

	newStorageNodeAlerter().Run(factory, stopCh)

	<-stopCh
	return nil
}
