package dataset_manager

import (
	dsclientset "github.com/hwameistor/datastore/pkg/apis/client/clientset/versioned"
	dsinformers "github.com/hwameistor/datastore/pkg/apis/client/informers/externalversions/datastore/v1alpha1"
	hmclientset "github.com/hwameistor/hwameistor/pkg/apis/client/clientset/versioned"
	"github.com/hwameistor/hwameistor/pkg/dataset-manager/dataset"
	"github.com/hwameistor/hwameistor/pkg/dataset-manager/persistentvolume"
	v12 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
)

// DatasetController interface
type DatasetController interface {
	Run(stopCh <-chan struct{})
}

type datasetController struct {
	dsController dataset.DSController
	pvController persistentvolume.PVController
}

func New(kubeClientset *kubernetes.Clientset, dsClientset *dsclientset.Clientset, hmClientset *hmclientset.Clientset,
	dsInformer dsinformers.DataSetInformer, pvInformer v12.PersistentVolumeInformer) DatasetController {
	return &datasetController{
		dsController: dataset.New(kubeClientset, dsClientset, hmClientset, dsInformer),
		pvController: persistentvolume.New(kubeClientset, dsClientset, hmClientset, pvInformer),
	}
}

func (dc *datasetController) Run(stopCh <-chan struct{}) {
	go dc.dsController.Run(stopCh)
	go dc.pvController.Run(stopCh)

	<-stopCh
}
