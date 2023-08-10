package auditor

import (
	"context"
	"time"

	clusterapiv1alpha1 "github.com/hwameistor/hwameistor-operator/api/v1alpha1"
	localstorageapis "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	runtimecache "sigs.k8s.io/controller-runtime/pkg/cache"
)

type auditorForCluster struct {
	events *EventStore
}

func newAuditorForCluster(events *EventStore) *auditorForCluster {
	return &auditorForCluster{events: events}
}

func (ad *auditorForCluster) Run(informersCache runtimecache.Cache, stopCh <-chan struct{}) {
	informer, err := informersCache.GetInformer(context.TODO(), &clusterapiv1alpha1.Cluster{})
	if err != nil {
		// error happens, crash the node
		log.WithError(err).Fatal("Failed to get informer for Node")
	}
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    ad.onAdd,
		UpdateFunc: ad.onUpdate,
	})
}

func (ad *auditorForCluster) onAdd(obj interface{}) {
	instance, _ := obj.(*clusterapiv1alpha1.Cluster)

	if instance.Status.InstalledCRDS {
		return
	}

	record := &localstorageapis.EventRecord{
		Time:          metav1.Time{Time: time.Now()},
		Action:        ActionClusterInstall,
		ActionContent: contentString(instance.Spec),
	}

	ad.events.AddRecordForResource(ResourceTypeCluster, instance.Name, record)
}

func (ad *auditorForCluster) onUpdate(oldObj, newObj interface{}) {
	oldInstance, _ := oldObj.(*clusterapiv1alpha1.Cluster)
	newInstance, _ := newObj.(*clusterapiv1alpha1.Cluster)

	if newInstance.Spec.AdmissionController.Disable != oldInstance.Spec.AdmissionController.Disable {
		record := &localstorageapis.EventRecord{
			Time:          metav1.Time{Time: time.Now()},
			Action:        ActionClusterChange,
			ActionContent: contentString(newInstance.Spec.AdmissionController),
		}

		ad.events.AddRecordForResource(ResourceTypeCluster, newInstance.Name, record)
	}

	if newInstance.Spec.DRBD.Disable != oldInstance.Spec.DRBD.Disable {
		record := &localstorageapis.EventRecord{
			Time:          metav1.Time{Time: time.Now()},
			Action:        ActionClusterChange,
			ActionContent: contentString(newInstance.Spec.DRBD),
		}

		ad.events.AddRecordForResource(ResourceTypeCluster, newInstance.Name, record)
	}
}
