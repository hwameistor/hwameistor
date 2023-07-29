package auditor

import (
	localstorageinformers "github.com/hwameistor/hwameistor/pkg/apis/client/informers/externalversions"
	localstorageinformersv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/client/informers/externalversions/hwameistor/v1alpha1"
	"k8s.io/client-go/tools/cache"
)

type auditorForLocalVolumeConvert struct {
	informer localstorageinformersv1alpha1.LocalVolumeConvertInformer

	events *EventStore
}

func newAuditorForLocalVolumeConvert(events *EventStore) *auditorForLocalVolumeConvert {
	return &auditorForLocalVolumeConvert{events: events}
}

func (ad *auditorForLocalVolumeConvert) Run(lsFactory localstorageinformers.SharedInformerFactory, stopCh <-chan struct{}) {
	ad.informer = lsFactory.Hwameistor().V1alpha1().LocalVolumeConverts()
	ad.informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    ad.onAdd,
		UpdateFunc: ad.onUpdate,
		DeleteFunc: ad.onDelete,
	})
	go ad.informer.Informer().Run(stopCh)

}

func (ad *auditorForLocalVolumeConvert) onAdd(obj interface{}) {
	//instance, _ := obj.(*localstorageapis.LocalVolume)
}

func (ad *auditorForLocalVolumeConvert) onDelete(obj interface{}) {
}

func (ad *auditorForLocalVolumeConvert) onUpdate(oldObj, newObj interface{}) {
}
