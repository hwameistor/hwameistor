package auditor

import (
	"time"

	localstorageinformers "github.com/hwameistor/hwameistor/pkg/apis/client/informers/externalversions"
	localstorageinformersv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/client/informers/externalversions/hwameistor/v1alpha1"
	localstorageapis "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

type auditorForLocalDisk struct {
	informer localstorageinformersv1alpha1.LocalDiskInformer

	events *EventStore
}

func newAuditorForLocalDisk(events *EventStore) *auditorForLocalDisk {
	return &auditorForLocalDisk{events: events}
}

func (ad *auditorForLocalDisk) Run(lsFactory localstorageinformers.SharedInformerFactory, stopCh <-chan struct{}) {
	ad.informer = lsFactory.Hwameistor().V1alpha1().LocalDisks()
	ad.informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    ad.onAdd,
		UpdateFunc: ad.onUpdate,
	})
	go ad.informer.Informer().Run(stopCh)

}

func (ad *auditorForLocalDisk) onAdd(obj interface{}) {
	instance, _ := obj.(*localstorageapis.LocalDisk)

	if len(instance.Status.State) > 0 {
		return
	}

	record := &localstorageapis.EventRecord{
		Time:          metav1.Time{Time: time.Now()},
		Action:        ActionDiskAdd,
		ActionContent: contentString(instance.Spec),
	}

	ad.events.AddRecordForResource(ResourceTypeDisk, instance.Name, record)
}

func (ad *auditorForLocalDisk) onUpdate(oldObj, newObj interface{}) {
	oldInstance, _ := oldObj.(*localstorageapis.LocalDisk)
	newInstance, _ := newObj.(*localstorageapis.LocalDisk)

	if newInstance.Spec.Owner != oldInstance.Spec.Owner {
		record := &localstorageapis.EventRecord{
			Time:          metav1.Time{Time: time.Now()},
			Action:        ActionDiskAllocate,
			ActionContent: contentString(newInstance.Spec),
		}

		ad.events.AddRecordForResource(ResourceTypeDisk, newInstance.Name, record)
	}

	if newInstance.Spec.NodeName != oldInstance.Spec.NodeName {
		record := &localstorageapis.EventRecord{
			Time:          metav1.Time{Time: time.Now()},
			Action:        ActionDiskRelocate,
			ActionContent: contentString(newInstance.Spec),
		}

		ad.events.AddRecordForResource(ResourceTypeDisk, newInstance.Name, record)
	}

	if newInstance.Spec.Reserved && !oldInstance.Spec.Reserved {
		record := &localstorageapis.EventRecord{
			Time:          metav1.Time{Time: time.Now()},
			Action:        ActionDiskReserve,
			ActionContent: contentString(newInstance.Spec),
		}

		ad.events.AddRecordForResource(ResourceTypeDisk, newInstance.Name, record)
	}

	if !newInstance.Spec.Reserved && oldInstance.Spec.Reserved {
		record := &localstorageapis.EventRecord{
			Time:          metav1.Time{Time: time.Now()},
			Action:        ActionDiskRelease,
			ActionContent: contentString(newInstance.Spec),
		}

		ad.events.AddRecordForResource(ResourceTypeDisk, newInstance.Name, record)
	}

	if newInstance.Status.State != oldInstance.Spec.State {
		record := &localstorageapis.EventRecord{
			Time:          metav1.Time{Time: time.Now()},
			Action:        ActionDiskChange,
			ActionContent: contentString(newInstance.Status),
		}

		ad.events.AddRecordForResource(ResourceTypeDisk, newInstance.Name, record)
	}
}
