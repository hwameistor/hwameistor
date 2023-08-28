package auditor

import (
	"time"

	localstorageinformers "github.com/hwameistor/hwameistor/pkg/apis/client/informers/externalversions"
	localstorageinformersv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/client/informers/externalversions/hwameistor/v1alpha1"
	localstorageapis "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	})
	go ad.informer.Informer().Run(stopCh)

}

func (ad *auditorForLocalVolumeConvert) onAdd(obj interface{}) {
	instance, _ := obj.(*localstorageapis.LocalVolumeConvert)

	if len(instance.Status.State) != 0 {
		return
	}

	record := &localstorageapis.EventRecord{
		Time:   metav1.Time{Time: time.Now()},
		ID:     instance.Name,
		Action: ActionVolumeConvert,
		State:  ActionStateSubmit,
	}

	ad.events.AddRecordForResource(ResourceTypeVolume, instance.Spec.VolumeName, record)
}

func (ad *auditorForLocalVolumeConvert) onUpdate(oldObj, newObj interface{}) {
	instance, _ := newObj.(*localstorageapis.LocalVolumeConvert)

	record := &localstorageapis.EventRecord{
		Time:   metav1.Time{Time: time.Now()},
		ID:     instance.Name,
		Action: ActionVolumeConvert,
	}
	if instance.Status.State == localstorageapis.OperationStateSubmitted {
		record.State = ActionStateStart
	} else if instance.Status.State == localstorageapis.OperationStateCompleted {
		record.State = ActionStateComplete
	} else if instance.Status.State == localstorageapis.OperationStateToBeAborted {
		record.State = ActionStateAbort
		record.StateContent = contentString(instance.Status)
	} else {
		return
	}

	ad.events.AddRecordForResource(ResourceTypeVolume, instance.Spec.VolumeName, record)
}
