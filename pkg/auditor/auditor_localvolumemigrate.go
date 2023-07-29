package auditor

import (
	localstorageinformers "github.com/hwameistor/hwameistor/pkg/apis/client/informers/externalversions"
	localstorageinformersv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/client/informers/externalversions/hwameistor/v1alpha1"
	"k8s.io/client-go/tools/cache"
)

type auditorForLocalVolumeMigrate struct {
	informer localstorageinformersv1alpha1.LocalVolumeMigrateInformer

	events *EventStore
}

func newAuditorForLocalVolumeMigrate(events *EventStore) *auditorForLocalVolumeMigrate {
	return &auditorForLocalVolumeMigrate{events: events}
}

func (ad *auditorForLocalVolumeMigrate) Run(lsFactory localstorageinformers.SharedInformerFactory, stopCh <-chan struct{}) {
	ad.informer = lsFactory.Hwameistor().V1alpha1().LocalVolumeMigrates()
	ad.informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    ad.onAdd,
		UpdateFunc: ad.onUpdate,
		DeleteFunc: ad.onDelete,
	})
	go ad.informer.Informer().Run(stopCh)

}

func (ad *auditorForLocalVolumeMigrate) onAdd(obj interface{}) {
	//instance, _ := obj.(*localstorageapis.LocalVolume)
}

func (ad *auditorForLocalVolumeMigrate) onDelete(obj interface{}) {
}

func (ad *auditorForLocalVolumeMigrate) onUpdate(oldObj, newObj interface{}) {
}
