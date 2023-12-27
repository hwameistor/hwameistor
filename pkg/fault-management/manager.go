package faultmanagement

import (
	"fmt"
	hwameistorclient "github.com/hwameistor/hwameistor/pkg/apis/client/clientset/versioned"
	"github.com/hwameistor/hwameistor/pkg/apis/client/informers/externalversions/hwameistor/v1alpha1"
	listers "github.com/hwameistor/hwameistor/pkg/apis/client/listers/hwameistor/v1alpha1"
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/common"
	"github.com/hwameistor/hwameistor/pkg/exechelper"
	"github.com/hwameistor/hwameistor/pkg/exechelper/nsexecutor"
	"github.com/hwameistor/hwameistor/pkg/fault-management/graph"
	log "github.com/sirupsen/logrus"
	informercorev1 "k8s.io/client-go/informers/core/v1"
	v1 "k8s.io/client-go/listers/core/v1"
	storagev1lister "k8s.io/client-go/listers/storage/v1"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type manager struct {
	nodeName  string
	namespace string
	logger    *log.Entry
	kclient   client.Client

	topologyGraph       graph.TopologyGraphManager
	hmClient            hwameistorclient.Interface
	faultTicketInformer v1alpha1.FaultTicketInformer
	faultTicketLister   listers.FaultTicketLister
	localVolumeLister   listers.LocalVolumeLister
	storageClassLister  storagev1lister.StorageClassLister
	pvcLister           v1.PersistentVolumeClaimLister
	faultTicketSynced   cache.InformerSynced

	faultTicketTaskQueue *common.TaskQueue
	cmdExec              exechelper.Executor
}

func New(nodeName, namespace string,
	kclient client.Client,
	hmClient hwameistorclient.Interface,
	podInformer informercorev1.PodInformer,
	pvcInformer informercorev1.PersistentVolumeClaimInformer,
	pvInformer informercorev1.PersistentVolumeInformer,
	lvInformer v1alpha1.LocalVolumeInformer,
	lsnInformer v1alpha1.LocalStorageNodeInformer,
	faultTickerInformer v1alpha1.FaultTicketInformer,
	scLister storagev1lister.StorageClassLister,
) *manager {
	m := &manager{
		nodeName:  nodeName,
		namespace: namespace,
		kclient:   kclient,
		hmClient:  hmClient,
		// don't set maxRetries, 0 means no limit, and events won't be dropped
		faultTicketTaskQueue: common.NewTaskQueue("FaultTicketTaskQueue", 0),
		logger:               log.WithField("Module", "FaultManagement"),
		faultTicketInformer:  faultTickerInformer,
		faultTicketLister:    faultTickerInformer.Lister(),
		faultTicketSynced:    faultTickerInformer.Informer().HasSynced,
		topologyGraph:        graph.New(nodeName, namespace, kclient, hmClient, podInformer, pvcInformer, pvInformer, lsnInformer, lvInformer, scLister),
		storageClassLister:   scLister,
		pvcLister:            pvcInformer.Lister(),
		localVolumeLister:    lvInformer.Lister(),
		cmdExec:              nsexecutor.New(),
	}

	// setup informer for FaultTicket
	m.faultTicketInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    m.handleTicketAdd,
		UpdateFunc: m.handleTicketUpdate,
	})

	return m
}

func (m *manager) Run(stopCh <-chan struct{}) error {
	// Wait for the caches to be synced before starting processors
	m.logger.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, m.faultTicketSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	m.logger.Info("Starting FaultTicket worker")
	if err := m.topologyGraph.Run(stopCh); err != nil {
		m.logger.WithError(err).Error("Failed to start topology graph manager")
		return err
	}

	m.logger.Info("Starting FaultTicket worker")
	go m.startFaultTicketTaskWorker(stopCh)

	return nil
}

func (m *manager) handleTicketAdd(obj interface{}) {
	if _, ok := obj.(*apisv1alpha1.FaultTicket); !ok {
		return
	}
	faultTicket := obj.(*apisv1alpha1.FaultTicket)
	if m.nodeName != faultTicket.Spec.NodeName {
		return
	}
	m.faultTicketTaskQueue.Add(faultTicket.Name)
}

func (m *manager) handleTicketUpdate(_, newObj interface{}) {
	if _, ok := newObj.(*apisv1alpha1.FaultTicket); !ok {
		return
	}
	m.faultTicketTaskQueue.Add(newObj.(*apisv1alpha1.FaultTicket).Name)
}
