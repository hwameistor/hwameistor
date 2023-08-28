package auditor

import (
	"time"

	localstorageinformers "github.com/hwameistor/hwameistor/pkg/apis/client/informers/externalversions"
	localstorageinformersv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/client/informers/externalversions/hwameistor/v1alpha1"
	localstorageapis "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	})
	go ad.informer.Informer().Run(stopCh)

}

func (ad *auditorForLocalVolumeMigrate) onAdd(obj interface{}) {
	instance, _ := obj.(*localstorageapis.LocalVolumeMigrate)

	if len(instance.Status.State) != 0 {
		return
	}

	record := &localstorageapis.EventRecord{
		Time:          metav1.Time{Time: time.Now()},
		ID:            instance.Name,
		Action:        ActionVolumeMigrate,
		ActionContent: contentString(instance.Spec),
		State:         ActionStateSubmit,
	}

	ad.events.AddRecordForResource(ResourceTypeVolume, instance.Spec.VolumeName, record)
}

func (ad *auditorForLocalVolumeMigrate) onUpdate(oldObj, newObj interface{}) {
	instance, _ := newObj.(*localstorageapis.LocalVolumeMigrate)

	record := &localstorageapis.EventRecord{
		Time:          metav1.Time{Time: time.Now()},
		ID:            instance.Name,
		Action:        ActionVolumeMigrate,
		ActionContent: contentString(instance.Spec),
	}

	if instance.Status.State == localstorageapis.OperationStateSubmitted {
		record.State = ActionStateStart
	} else if instance.Status.State == localstorageapis.OperationStateCompleted {
		record.State = ActionStateComplete
		record.StateContent = contentString(instance.Status)
	} else if instance.Status.State == localstorageapis.OperationStateToBeAborted {
		record.State = ActionStateAbort
		record.StateContent = contentString(instance.Status)
	} else {
		return
	}

	ad.events.AddRecordForResource(ResourceTypeVolume, instance.Spec.VolumeName, record)
}
