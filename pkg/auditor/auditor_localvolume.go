package auditor

import (
	"time"

	localstorageinformers "github.com/hwameistor/hwameistor/pkg/apis/client/informers/externalversions"
	localstorageinformersv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/client/informers/externalversions/hwameistor/v1alpha1"
	localstorageapis "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	instance, _ := obj.(*localstorageapis.LocalVolume)

	if len(instance.Status.State) != 0 {
		return
	}

	record := &localstorageapis.EventRecord{
		Time:          metav1.Time{Time: time.Now()},
		Action:        ActionVolumeCreate,
		ActionContent: contentString(instance.Spec),
	}

	ad.events.AddRecordForResource(ResourceTypeVolume, instance.Name, record)
}

func (ad *auditorForLocalVolume) onDelete(obj interface{}) {
	instance, _ := obj.(*localstorageapis.LocalVolume)
	record := &localstorageapis.EventRecord{
		Time:          metav1.Time{Time: time.Now()},
		Action:        ActionVolumeDelete,
		ActionContent: contentString(instance.Spec),
	}

	ad.events.AddRecordForResource(ResourceTypeVolume, instance.Name, record)
}

func (ad *auditorForLocalVolume) onUpdate(oldObj, newObj interface{}) {
	oldInstance, _ := oldObj.(*localstorageapis.LocalVolume)
	newInstance, _ := newObj.(*localstorageapis.LocalVolume)

	//check for mount or umount
	if oldInstance.Status.PublishedNodeName != newInstance.Status.PublishedNodeName {
		record := &localstorageapis.EventRecord{Time: metav1.Time{Time: time.Now()}}

		if len(newInstance.Status.PublishedNodeName) == 0 {
			// unmount
			record.Action = ActionVolumeUnmount
			record.ActionContent = contentString(oldInstance.Status)

		} else {
			// mount
			record.Action = ActionVolumeMount
			record.ActionContent = contentString(newInstance.Status)
		}

		ad.events.AddRecordForResource(ResourceTypeVolume, newInstance.Name, record)
	}
}
