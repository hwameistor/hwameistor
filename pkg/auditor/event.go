package auditor

import (
	localstorageinformers "github.com/hwameistor/hwameistor/pkg/apis/client/informers/externalversions"
	localstorageinformersv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/client/informers/externalversions/hwameistor/v1alpha1"
)

type EventStore struct {
	informer localstorageinformersv1alpha1.EventInformer
}

func NewEventStore() *EventStore {
	return &EventStore{}
}

func (es *EventStore) Run(lsFactory localstorageinformers.SharedInformerFactory, stopCh <-chan struct{}) {
	es.informer = lsFactory.Hwameistor().V1alpha1().Events()

	go es.informer.Informer().Run(stopCh)

}
