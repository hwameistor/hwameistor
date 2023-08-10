package auditor

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	localstorageclientset "github.com/hwameistor/hwameistor/pkg/apis/client/clientset/versioned"
	localstorageinformers "github.com/hwameistor/hwameistor/pkg/apis/client/informers/externalversions"
	localstorageinformersv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/client/informers/externalversions/hwameistor/v1alpha1"
	localstorageapis "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceTypeCluster     = "Cluster"
	ResourceTypeStorageNode = "StorageNode"
	ResourceTypeVolume      = "Volume"
	ResourceTypeDisk        = "Disk"

	ActionVolumeCreate  = "Create"
	ActionVolumeDelete  = "Delete"
	ActionVolumeMount   = "Mount"
	ActionVolumeUnmount = "Unmount"
	ActionVolumeConvert = "Convert"
	ActionVolumeMigrate = "Migrate"
	ActionVolumeExpand  = "Expand"

	ActionStateSubmit   = "Submit"
	ActionStateStart    = "Start"
	ActionStateComplete = "Complete"
	ActionStateAbort    = "Abort"

	ActionNodeAdd                       = "Add"
	ActionNodeRemove                    = "Remove"
	ActionNodeStateChange               = "StateChange"
	ActionNodeStoragePoolCapacityExpand = "CapacityExpand"

	ActionClusterInstall = "Install"
	ActionClusterChange  = "Change"
	ActionClusterUpgrade = "Upgrade"

	ActionDiskAdd      = "Add"
	ActionDiskRelocate = "Relocate"
	ActionDiskAllocate = "Allocate"
	ActionDiskReserve  = "Reserve"
	ActionDiskRelease  = "Release"
	ActionDiskChange   = "Change"

	ErrMsgSuccess = "Added a record"
	ErrMsgFailure = "Failed to add a record"
)

func contentString(v any) string {
	bytes, _ := json.Marshal(v)
	return string(bytes)
}

type EventStore struct {
	clientSet *localstorageclientset.Clientset
	informer  localstorageinformersv1alpha1.EventInformer
}

func NewEventStore() *EventStore {
	return &EventStore{}
}

func (es *EventStore) Run(clientSet *localstorageclientset.Clientset, lsFactory localstorageinformers.SharedInformerFactory, stopCh <-chan struct{}) {
	es.clientSet = clientSet

	es.informer = lsFactory.Hwameistor().V1alpha1().Events()
	go es.informer.Informer().Run(stopCh)

}

func (es *EventStore) AddRecordForResource(resType string, resName string, record *localstorageapis.EventRecord) {

	logCtx := log.WithFields(log.Fields{
		"ResourceName": resName,
		"ResourceType": resType,
		"Record":       record,
	})

	event, _ := es.informer.Lister().Get(es.eventKey(resType, resName))
	if event == nil {
		var err error
		event, err = es.createEvent(resType, resName)
		if err != nil {
			logCtx.WithError(err).Error(ErrMsgFailure)
			return
		}
	}
	event.Spec.Records = append(event.Spec.Records, *record)
	if _, err := es.clientSet.HwameistorV1alpha1().Events().Update(context.TODO(), event, metav1.UpdateOptions{}); err != nil {
		logCtx.WithError(err).Error(ErrMsgSuccess)
	} else {
		logCtx.Debug(ErrMsgSuccess)
	}

}

func (es *EventStore) createEvent(resType string, resName string) (*localstorageapis.Event, error) {
	event := &localstorageapis.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name: es.eventKey(resType, resName),
		},
		Spec: localstorageapis.EventSpec{
			ResourceType: resType,
			ResourceName: resName,
			Records:      []localstorageapis.EventRecord{},
		},
	}
	return es.clientSet.HwameistorV1alpha1().Events().Create(context.TODO(), event, metav1.CreateOptions{})
}

func (es *EventStore) eventKey(resType string, resName string) string {
	return strings.ToLower(fmt.Sprintf("%s-%s", resType, resName))
}
