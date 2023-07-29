package auditor

import (
	localstorageinformers "github.com/hwameistor/hwameistor/pkg/apis/client/informers/externalversions"
	localstorageinformersv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/client/informers/externalversions/hwameistor/v1alpha1"
	"k8s.io/client-go/tools/cache"
)

type auditorForLocalVolumeExpand struct {
	informer localstorageinformersv1alpha1.LocalVolumeExpandInformer

	events *EventStore
}

func newAuditorForLocalVolumeExpand(events *EventStore) *auditorForLocalVolumeExpand {
	return &auditorForLocalVolumeExpand{events: events}
}

func (ad *auditorForLocalVolumeExpand) Run(lsFactory localstorageinformers.SharedInformerFactory, stopCh <-chan struct{}) {
	ad.informer = lsFactory.Hwameistor().V1alpha1().LocalVolumeExpands()
	ad.informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    ad.onAdd,
		UpdateFunc: ad.onUpdate,
		DeleteFunc: ad.onDelete,
	})
	go ad.informer.Informer().Run(stopCh)

}

func (ad *auditorForLocalVolumeExpand) onAdd(obj interface{}) {
	//instance, _ := obj.(*localstorageapis.LocalVolume)
}

func (ad *auditorForLocalVolumeExpand) onDelete(obj interface{}) {
}

func (ad *auditorForLocalVolumeExpand) onUpdate(oldObj, newObj interface{}) {
}
