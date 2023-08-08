package auditor

import (
	"time"

	localstorageinformers "github.com/hwameistor/hwameistor/pkg/apis/client/informers/externalversions"
	localstorageinformersv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/client/informers/externalversions/hwameistor/v1alpha1"
	localstorageapis "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

type auditorForLocalStorageNode struct {
	informer localstorageinformersv1alpha1.LocalStorageNodeInformer

	events *EventStore
}

func newAuditorForLocalStorageNode(events *EventStore) *auditorForLocalStorageNode {
	return &auditorForLocalStorageNode{events: events}
}

func (ad *auditorForLocalStorageNode) Run(lsFactory localstorageinformers.SharedInformerFactory, stopCh <-chan struct{}) {
	ad.informer = lsFactory.Hwameistor().V1alpha1().LocalStorageNodes()
	ad.informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    ad.onAdd,
		UpdateFunc: ad.onUpdate,
		DeleteFunc: ad.onDelete,
	})
	go ad.informer.Informer().Run(stopCh)

}

func (ad *auditorForLocalStorageNode) onAdd(obj interface{}) {
	instance, _ := obj.(*localstorageapis.LocalStorageNode)

	if len(instance.Status.State) != 0 {
		return
	}

	record := &localstorageapis.EventRecord{
		Time:          metav1.Time{Time: time.Now()},
		Action:        ActionNodeAdd,
		ActionContent: contentString(instance.Spec),
	}

	ad.events.AddRecordForResource(ResourceTypeStorageNode, instance.Name, record)
}

func (ad *auditorForLocalStorageNode) onDelete(obj interface{}) {
	instance, _ := obj.(*localstorageapis.LocalStorageNode)
	record := &localstorageapis.EventRecord{
		Time:   metav1.Time{Time: time.Now()},
		Action: ActionNodeRemove,
	}

	ad.events.AddRecordForResource(ResourceTypeStorageNode, instance.Name, record)
}

func (ad *auditorForLocalStorageNode) onUpdate(oldObj, newObj interface{}) {
	oldInstance, _ := oldObj.(*localstorageapis.LocalStorageNode)
	newInstance, _ := newObj.(*localstorageapis.LocalStorageNode)

	if oldInstance.Status.State != newInstance.Status.State {
		record := &localstorageapis.EventRecord{Time: metav1.Time{Time: time.Now()}}

		record.Action = ActionNodeStateChange
		record.ActionContent = contentString(oldInstance.Status)

		ad.events.AddRecordForResource(ResourceTypeStorageNode, newInstance.Name, record)
	}

	for poolName, newPool := range newInstance.Status.Pools {
		oldPool, existed := oldInstance.Status.Pools[poolName]
		if !existed {
			// found an event for pool extend
			record := &localstorageapis.EventRecord{
				Time:          metav1.NewTime(time.Now()),
				Action:        ActionNodeStoragePoolCapacityExpand,
				ActionContent: contentString(newPool.Disks),
			}
			ad.events.AddRecordForResource(ResourceTypeStorageNode, newInstance.Name, record)
		} else {
			expandDisks := []localstorageapis.LocalDevice{}
			flags := map[string]bool{}
			for _, d := range oldPool.Disks {
				flags[d.DevPath] = true
			}
			for i, nd := range newPool.Disks {
				if _, exists := flags[nd.DevPath]; exists {
					continue
				}
				expandDisks = append(expandDisks, newPool.Disks[i])
			}
			if len(expandDisks) > 0 {
				// found an event for pool extend
				record := &localstorageapis.EventRecord{
					Time:          metav1.NewTime(time.Now()),
					Action:        ActionNodeStoragePoolCapacityExpand,
					ActionContent: contentString(expandDisks),
				}
				ad.events.AddRecordForResource(ResourceTypeStorageNode, newInstance.Name, record)
			}
		}

	}

}
