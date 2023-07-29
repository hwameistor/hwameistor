package auditor

import (
	localstorageinformers "github.com/hwameistor/hwameistor/pkg/apis/client/informers/externalversions"
	localstorageinformersv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/client/informers/externalversions/hwameistor/v1alpha1"
	"k8s.io/client-go/tools/cache"
)

type auditorForLocalDiskClaim struct {
	informer localstorageinformersv1alpha1.LocalDiskClaimInformer

	events *EventStore
}

func newAuditorForLocalDiskClaim(events *EventStore) *auditorForLocalDiskClaim {
	return &auditorForLocalDiskClaim{events: events}
}

func (ad *auditorForLocalDiskClaim) Run(lsFactory localstorageinformers.SharedInformerFactory, stopCh <-chan struct{}) {
	ad.informer = lsFactory.Hwameistor().V1alpha1().LocalDiskClaims()
	ad.informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    ad.onAdd,
		UpdateFunc: ad.onUpdate,
		DeleteFunc: ad.onDelete,
	})
	go ad.informer.Informer().Run(stopCh)

}

func (ad *auditorForLocalDiskClaim) onAdd(obj interface{}) {
	//instance, _ := obj.(*localstorageapis.LocalVolume)
}

func (ad *auditorForLocalDiskClaim) onDelete(obj interface{}) {
}

func (ad *auditorForLocalDiskClaim) onUpdate(oldObj, newObj interface{}) {
}
