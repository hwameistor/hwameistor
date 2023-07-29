package auditor

import (
	localstorageinformers "github.com/hwameistor/hwameistor/pkg/apis/client/informers/externalversions"
	localstorageinformersv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/client/informers/externalversions/hwameistor/v1alpha1"

	"k8s.io/client-go/tools/cache"
)

type auditorForLocalVolume struct {
	informer localstorageinformersv1alpha1.LocalVolumeInformer

	events *EventStore
}

func newAuditorForLocalVolume(events *EventStore) *auditorForLocalVolume {
	return &auditorForLocalVolume{events: events}
}

func (ad *auditorForLocalVolume) Run(lsFactory localstorageinformers.SharedInformerFactory, stopCh <-chan struct{}) {
	ad.informer = lsFactory.Hwameistor().V1alpha1().LocalVolumes()
	ad.informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    ad.onAdd,
		UpdateFunc: ad.onUpdate,
		DeleteFunc: ad.onDelete,
	})
	go ad.informer.Informer().Run(stopCh)

}

func (ad *auditorForLocalVolume) onAdd(obj interface{}) {
	//instance, _ := obj.(*localstorageapis.LocalVolume)
}

func (ad *auditorForLocalVolume) onDelete(obj interface{}) {
}

func (ad *auditorForLocalVolume) onUpdate(oldObj, newObj interface{}) {
}
