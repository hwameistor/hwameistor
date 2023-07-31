package auditor

type auditorForCluster struct {
	events *EventStore
}

func newAuditorForCluster(events *EventStore) *auditorForCluster {
	return &auditorForCluster{events: events}
}

func (ad *auditorForCluster) Run(stopCh <-chan struct{}) {
	// informer :=
	// 	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
	// 		AddFunc:    ad.onAdd,
	// 		UpdateFunc: ad.onUpdate,
	// 		DeleteFunc: ad.onDelete,
	// 	})
	// go informer.Informer().Run(stopCh)

}

// func (ad *auditorForCluster) onAdd(obj interface{}) {
// 	//instance, _ := obj.(*localstorageapis.LocalVolume)
// }

// func (ad *auditorForCluster) onDelete(obj interface{}) {
// }

// func (ad *auditorForCluster) onUpdate(oldObj, newObj interface{}) {
// }
