package faultmanagement

import (
	"fmt"
	hwameistorclient "github.com/hwameistor/hwameistor/pkg/apis/client/clientset/versioned"
	"github.com/hwameistor/hwameistor/pkg/apis/client/informers/externalversions/hwameistor/v1alpha1"
	listers "github.com/hwameistor/hwameistor/pkg/apis/client/listers/hwameistor/v1alpha1"
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/common"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type manager struct {
	name      string
	namespace string
	logger    *log.Entry
	kclient   client.Client

	hmClient          hwameistorclient.Interface
	faultTicketLister listers.FaultTicketLister
	faultTicketSynced cache.InformerSynced

	faultTicketTaskQueue *common.TaskQueue
}

func New(name, namespace string, kclient client.Client, hmClient hwameistorclient.Interface, faultTickerInformer v1alpha1.FaultTicketInformer) *manager {
	m := &manager{
		name:      name,
		namespace: namespace,
		kclient:   kclient,
		hmClient:  hmClient,
		// don't set maxRetries, 0 means no limit, and events won't be dropped
		faultTicketTaskQueue: common.NewTaskQueue("FaultTicketTaskQueue", 0),
		logger:               log.WithField("Module", "FaultManagement"),
		faultTicketLister:    faultTickerInformer.Lister(),
		faultTicketSynced:    faultTickerInformer.Informer().HasSynced,
	}

	// setup informer for FaultTicket
	faultTickerInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
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
