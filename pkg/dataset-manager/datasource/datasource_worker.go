package datasource

import (
	"context"
	dsclientset "github.com/hwameistor/datastore/pkg/apis/client/clientset/versioned"
	dsinformers "github.com/hwameistor/datastore/pkg/apis/client/informers/externalversions/datastore/v1alpha1"
	dslisters "github.com/hwameistor/datastore/pkg/apis/client/listers/datastore/v1alpha1"
	datastore "github.com/hwameistor/datastore/pkg/apis/datastore/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"strings"

	hmclientset "github.com/hwameistor/hwameistor/pkg/apis/client/clientset/versioned"
	"github.com/hwameistor/hwameistor/pkg/common"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type DSController interface {
	Run(stopCh <-chan struct{})
}

// dsController is a controller to manage datasource
type dsController struct {
	dsClientset *dsclientset.Clientset
	hmClientset *hmclientset.Clientset
	kubeClient  *kubernetes.Clientset

	dsLister       dslisters.DataSourceLister
	dsListerSynced cache.InformerSynced
	dsQueue        *common.TaskQueue
}

func New(kubeClientset *kubernetes.Clientset, dsClientset *dsclientset.Clientset, hmClientset *hmclientset.Clientset, dsInformer dsinformers.DataSourceInformer) DSController {
	ctr := &dsController{
		dsClientset: dsClientset,
		kubeClient:  kubeClientset,
		hmClientset: hmClientset,
		dsQueue:     common.NewTaskQueue("DataSourceTask", 0),
	}

	dsInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    ctr.dsAdded,
		UpdateFunc: ctr.dsUpdated,
		DeleteFunc: ctr.dsDeleted,
	})
	ctr.dsLister = dsInformer.Lister()
	ctr.dsListerSynced = dsInformer.Informer().HasSynced

	return ctr
}

func (ctr *dsController) dsAdded(obj interface{}) {
	ds := obj.(*datastore.DataSource)
	ctr.dsQueue.Add(ds.Namespace + "/" + ds.Name)
}

func (ctr *dsController) dsUpdated(oldObj, newObj interface{}) {
	ctr.dsAdded(newObj)
}

func (ctr *dsController) dsDeleted(obj interface{}) {
	ctr.dsAdded(obj)
}

func (ctr *dsController) Run(stopCh <-chan struct{}) {
	defer ctr.dsQueue.Shutdown()

	klog.V(5).Infof("Starting Datasource controller")
	defer klog.Infof("Shutting Datasource controller")

	if !cache.WaitForCacheSync(stopCh, ctr.dsListerSynced) {
		klog.Fatalf("Cannot sync caches")
	}

	go wait.Until(ctr.syncDataSource, 0, stopCh)
	<-stopCh
}

func (ctr *dsController) syncDataSource() {
	key, quiet := ctr.dsQueue.Get()
	if quiet {
		return
	}
	defer ctr.dsQueue.Done(key)

	klog.V(4).Infof("Started Datasource porcessing %q", key)
	dsNamespace := strings.Split(key, "/")[0]
	dsName := strings.Split(key, "/")[1]

	// get Datasource to process
	ds, err := ctr.dsLister.DataSources(dsNamespace).Get(dsName)
	if err != nil {
		if errors.IsNotFound(err) {
			klog.V(4).Infof("Datasource %q has been deleted, ignoring", key)
		}
		klog.Errorf("Error getting Datasource %q: %v", key, err)
		return
	}
	ctr.SyncNewOrUpdatedDatasource(ds)
}

func (ctr *dsController) SyncNewOrUpdatedDatasource(ds *datastore.DataSource) {
	klog.V(4).Infof("Processing Datasource %s/%s", ds.Namespace, ds.Name)

	// check if PV created for this datasource
	_, err := ctr.kubeClient.CoreV1().PersistentVolumes().Get(context.Background(), ds.Name, metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			klog.Errorf("Error getting PV for datasource %s/%s: %v", ds.Namespace, ds.Name, err)
			ctr.dsQueue.AddRateLimited(ds.Namespace + "/" + ds.Name)
			return
		}
		// PV not found, create it
		if err = ctr.createRelatedPersistentVolume(ds.Name); err == nil {
			klog.V(4).Infof("Created PersistentVolume %s", ds.Name)
		}
	}
	if err != nil {
		klog.V(4).Infof("Error processing Datasource %s/%s: %v", ds.Namespace, ds.Name, err)
		ctr.dsQueue.AddRateLimited(ds.Namespace + "/" + ds.Name)
		return
	}

	ctr.dsQueue.Forget(ds.Namespace + "/" + ds.Name)
	klog.V(4).Infof("Finished processing Datasource %s/%s", ds.Namespace, ds.Name)
}

func (ctr *dsController) createRelatedPersistentVolume(pvName string) (err error) {
	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: pvName,
		},
		Spec: v1.PersistentVolumeSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadOnlyMany},
			Capacity: v1.ResourceList{
				v1.ResourceStorage: resource.MustParse("1Gi"), // FIXME: get capacity from datasource
			},
			PersistentVolumeReclaimPolicy: v1.PersistentVolumeReclaimRetain,
			PersistentVolumeSource: v1.PersistentVolumeSource{
				CSI: &v1.CSIPersistentVolumeSource{
					Driver:       "lvm.hwameistor.io",
					FSType:       "xfs",
					VolumeHandle: pvName,
				},
			},
			StorageClassName: "hwameistor-storage-lvm-hdd-static", // FIXME: get storageclass from datasource
		},
	}
	volumeMode := v1.PersistentVolumeFilesystem
	volumeAttr := make(map[string]string)
	volumeAttr["convertible"] = "false"
	volumeAttr["csi.storage.k8s.io/pv/name"] = pvName
	volumeAttr["volumeKind"] = "LVM"
	volumeAttr["volumeUsage"] = "AccelDataset" // to identify the dataset volume
	volumeAttr["poolClass"] = "HDD"            // FIXME: get poolClass from datasource

	pv.Spec.VolumeMode = &volumeMode
	pv.Spec.CSI.VolumeAttributes = volumeAttr

	_, err = ctr.kubeClient.CoreV1().PersistentVolumes().Create(context.Background(), pv, metav1.CreateOptions{})
	return
}
