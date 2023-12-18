package faultmanagement

import (
	"fmt"
	hwameistorclient "github.com/hwameistor/hwameistor/pkg/apis/client/clientset/versioned"
	"github.com/hwameistor/hwameistor/pkg/apis/client/informers/externalversions/hwameistor/v1alpha1"
	listers "github.com/hwameistor/hwameistor/pkg/apis/client/listers/hwameistor/v1alpha1"
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/common"
	"github.com/hwameistor/hwameistor/pkg/fault-management/graph"
	log "github.com/sirupsen/logrus"
	informercorev1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type manager struct {
	name      string
	namespace string
	logger    *log.Entry
	kclient   client.Client
	graph     *graph.Topology[string, string]

	topologyGraph       graph.TopologyGraphManager
	hmClient            hwameistorclient.Interface
	faultTicketInformer v1alpha1.FaultTicketInformer
	faultTicketLister   listers.FaultTicketLister
	faultTicketSynced   cache.InformerSynced

	faultTicketTaskQueue *common.TaskQueue
}

func New(name, namespace string,
	kclient client.Client,
	hmClient hwameistorclient.Interface,
	podInformer informercorev1.PodInformer,
	pvcInformer informercorev1.PersistentVolumeClaimInformer,
	pvInformer informercorev1.PersistentVolumeInformer,
	lvInformer v1alpha1.LocalVolumeInformer,
	lsnInformer v1alpha1.LocalStorageNodeInformer,
	faultTickerInformer v1alpha1.FaultTicketInformer,
) *manager {
	m := &manager{
		name:      name,
		namespace: namespace,
		kclient:   kclient,
		hmClient:  hmClient,
		// don't set maxRetries, 0 means no limit, and events won't be dropped
		faultTicketTaskQueue: common.NewTaskQueue("FaultTicketTaskQueue", 0),
		logger:               log.WithField("Module", "FaultManagement"),
		faultTicketInformer:  faultTickerInformer,
		faultTicketLister:    faultTickerInformer.Lister(),
		faultTicketSynced:    faultTickerInformer.Informer().HasSynced,
		topologyGraph:        graph.New(name, namespace, kclient, hmClient, podInformer, pvcInformer, pvInformer, lsnInformer, lvInformer),
	}

	return m
}

func (m *manager) Run(stopCh <-chan struct{}) error {
	// Wait for the caches to be synced before starting processors
	m.logger.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, m.faultTicketSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	// setup informer for FaultTicket
	m.faultTicketInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    m.handleTicketAdd,
		UpdateFunc: m.handleTicketUpdate,
	})

	m.logger.Info("Start FaultTicket worker")
	go m.startFaultTicketTaskWorker(stopCh)
	return nil
}

func (m *manager) handleTicketAdd(obj interface{}) {
	if _, ok := obj.(*apisv1alpha1.FaultTicket); !ok {
		return
	}
	m.faultTicketTaskQueue.Add(obj.(*apisv1alpha1.FaultTicket).Name)
}

func (m *manager) handleTicketUpdate(_, newObj interface{}) {
	if _, ok := newObj.(*apisv1alpha1.FaultTicket); !ok {
		return
	}
	m.faultTicketTaskQueue.Add(newObj.(*apisv1alpha1.FaultTicket).Name)
}
