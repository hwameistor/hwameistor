package alerter

import (
	"fmt"
	"strings"

	localstorageinformers "github.com/hwameistor/local-storage/pkg/apis/client/informers/externalversions"
	localstoragev1alpha1 "github.com/hwameistor/local-storage/pkg/apis/localstorage/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	log "github.com/sirupsen/logrus"
)

type storageNodeAlerter struct {
	logger *log.Entry

	moduleName string

	queue workqueue.Interface
}

func newStorageNodeAlerter() Alerter {
	return &storageNodeAlerter{
		logger:     log.WithField("Module", ModuleNode),
		moduleName: ModuleNode,
		queue:      workqueue.New(),
	}
}

func (alt *storageNodeAlerter) Run(informerFactory localstorageinformers.SharedInformerFactory, stopCh <-chan struct{}) {
	informer := informerFactory.Localstorage().V1alpha1().LocalStorageNodes().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    alt.onAdd,
		UpdateFunc: alt.onUpdate,
		DeleteFunc: alt.onDelete,
	})
	go informer.Run(stopCh)
	go alt.process(stopCh)
}

func (alt *storageNodeAlerter) onAdd(obj interface{}) {
	node, _ := obj.(*localstoragev1alpha1.LocalStorageNode)
	alt.logger.WithFields(log.Fields{"node": node.Name}).Debug("Checking for a storage node just added")

	alt.queue.Add(node)
}

func (alt *storageNodeAlerter) onUpdate(oldObj, newObj interface{}) {
	node, _ := newObj.(*localstoragev1alpha1.LocalStorageNode)
	alt.logger.WithFields(log.Fields{"node": node.Name}).Debug("Checking for a storage node just added")

	alt.queue.Add(node)
}

func (alt *storageNodeAlerter) onDelete(obj interface{}) {
	node, _ := obj.(*localstoragev1alpha1.LocalStorageNode)
	alt.logger.WithFields(log.Fields{"node": node.Name}).Debug("Checking for a storage node just deleted")

	alt.queue.Add(node)
}

func (alt *storageNodeAlerter) process(stopCh <-chan struct{}) {
	alt.logger.Debug("Disk Alerter is working now")

	go func() {
		for {
			obj, shutdown := alt.queue.Get()
			if shutdown {
				alt.logger.Debug("Stop the disk alerter worker")
				break
			}
			node, ok := obj.(*localstoragev1alpha1.LocalStorageNode)
			if ok && node != nil {
				alt.predict(node)
			}

			alt.queue.Done(obj)
		}
	}()

	<-stopCh
	alt.queue.ShutDown()
}

func (alt *storageNodeAlerter) predict(node *localstoragev1alpha1.LocalStorageNode) {

	if node == nil {
		return
	}
	if node.Status.State == localstoragev1alpha1.NodeStateOffline {
		alert := &localstoragev1alpha1.LocalStorageAlert{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("%s-%s-%s", strings.ToLower(alt.moduleName), node.Name, genTimeStampString()),
			},
			Spec: localstoragev1alpha1.LocalStorageAlertSpec{
				Severity: SeverityCritical,
				Module:   alt.moduleName,
				Resource: node.Name,
				Event:    NodeEventOffline,
			},
		}

		createAlert(alert)
		return
	}
}
