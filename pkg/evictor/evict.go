package evictor

import (
	"context"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	localstorageclientset "github.com/hwameistor/hwameistor/pkg/apis/client/clientset/versioned"
	localstorageinformers "github.com/hwameistor/hwameistor/pkg/apis/client/informers/externalversions"
	localstorageinformersv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/client/informers/externalversions/hwameistor/v1alpha1"
	localstorageapis "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-storage/common"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	informercorev1 "k8s.io/client-go/informers/core/v1"
	informerstoragev1 "k8s.io/client-go/informers/storage/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

const (
	lvmCSIDriverName  = "lvm.hwameistor.io"
	diskCSIDriverName = "disk.hwameistor.io"

	localVolumeNameIndex = "volumename"
)

// Evictor interface
type Evictor interface {
	Run(stopCh <-chan struct{}) error
}

type evictor struct {
	clientset *kubernetes.Clientset

	podInformer informercorev1.PodInformer
	pvcInformer informercorev1.PersistentVolumeClaimInformer
	scInformer  informerstoragev1.StorageClassInformer

	lsClientset       *localstorageclientset.Clientset
	lvInformer        localstorageinformersv1alpha1.LocalVolumeInformer
	lvMigrateInformer localstorageinformersv1alpha1.LocalVolumeMigrateInformer

	evictedPodQueue    *common.TaskQueue
	migrateVolumeQueue *common.TaskQueue
}

/* steps:
1. watch for Pod update event, insert the Pod with Evicted status into evictedPodQueue;
2. pick up a Pod from evictedPodQueue, check if it is using HwameiStor volume. If yes, insert it into the migrateVolumeQueue; if not, ignore
3. pick up a volume form migrateVolumeQueue, and migrate it. Make sure there is no replica located at the node where the pod is evicted;
*/

// New an assistant instance
func New(clientset *kubernetes.Clientset) Evictor {
	return &evictor{
		clientset:          clientset,
		evictedPodQueue:    common.NewTaskQueue("EvictedPods", 0),
		migrateVolumeQueue: common.NewTaskQueue("MigrateVolumes", 0),
	}
}

func (ev *evictor) Run(stopCh <-chan struct{}) error {
	log.Debug("start informer factory")
	factory := informers.NewSharedInformerFactory(ev.clientset, 0)
	factory.Start(stopCh)

	ev.podInformer = factory.Core().V1().Pods()
	ev.podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: ev.watchForEvictedPodOnUpdate,
		DeleteFunc: ev.watchForEvictedPodOnDelete,
	})
	go ev.podInformer.Informer().Run(stopCh)

	ev.pvcInformer = factory.Core().V1().PersistentVolumeClaims()
	go ev.pvcInformer.Informer().Run(stopCh)
	ev.scInformer = factory.Storage().V1().StorageClasses()
	go ev.scInformer.Informer().Run(stopCh)

	// Initialize HwameiStor LocalStorage resources
	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		log.WithError(err).Fatal("Failed to get kubernetes cluster config")
	}

	log.Debug("start local storage informer factory")
	ev.lsClientset = localstorageclientset.NewForConfigOrDie(cfg)
	lsFactory := localstorageinformers.NewSharedInformerFactory(ev.lsClientset, 0)
	lsFactory.Start(stopCh)

	ev.lvInformer = lsFactory.Hwameistor().V1alpha1().LocalVolumes()
	go ev.lvInformer.Informer().Run(stopCh)

	// index: lvmigrate.spec.volumename
	lvMigrateVolumeNameIndexFunc := func(obj interface{}) ([]string, error) {
		lvm, ok := obj.(*localstorageapis.LocalVolumeMigrate)
		if !ok || lvm == nil {
			return []string{}, fmt.Errorf("wrong LocalStorageMigrate resource")
		}
		return []string{lvm.Spec.VolumeName}, nil
	}
	ev.lvMigrateInformer = lsFactory.Hwameistor().V1alpha1().LocalVolumeMigrates()
	ev.lvMigrateInformer.Informer().AddIndexers(cache.Indexers{localVolumeNameIndex: lvMigrateVolumeNameIndexFunc})
	go ev.lvMigrateInformer.Informer().Run(stopCh)

	log.Debug("starting migrate volume worker")
	go ev.startMigrateVolumeWorker(stopCh)

	<-stopCh
	return nil
}

func (ev *evictor) watchForEvictedPodOnUpdate(oObj, nObj interface{}) {
	pod, _ := nObj.(*corev1.Pod)
	log.WithFields(log.Fields{
		"namespace": pod.Namespace,
		"pod":       pod.Name,
		"message":   pod.Status.Conditions[0].Type,
		"reason":    pod.Status.Reason,
		"phase":     pod.Status.Phase,
	}).Debug("Watching for a Pod update event ...")
	if isPodEvicted(pod) {
		log.WithFields(log.Fields{
			"namespace": pod.Namespace,
			"pod":       pod.Name,
			"phase":     pod.Status.Phase,
			"reason":    pod.Status.Reason,
		}).Debug("Got an evicted Pod to process")
		go ev.filterForHwameiVolume(pod)
	}
}

func (ev *evictor) watchForEvictedPodOnDelete(obj interface{}) {
	pod, _ := obj.(*corev1.Pod)
	log.WithFields(log.Fields{
		"namespace": pod.Namespace,
		"pod":       pod.Name,
		"message":   pod.Status.Conditions[0].Type,
		"reason":    pod.Status.Reason,
		"phase":     pod.Status.Phase,
	}).Debug("Watching for a Pod delete event ...")
	if isPodEvicted(pod) {
		log.WithFields(log.Fields{
			"namespace": pod.Namespace,
			"pod":       pod.Name,
			"phase":     pod.Status.Phase,
			"reason":    pod.Status.Reason,
		}).Debug("Got an evicted Pod to process")
		go ev.filterForHwameiVolume(pod)
	}
}

func isPodEvicted(pod *corev1.Pod) bool {
	return true
	// podFailed := pod.Status.Phase == corev1.PodFailed
	// podEvicted := pod.Status.Reason == "Evicted"
	// return podFailed && podEvicted
}

func (ev *evictor) filterForHwameiVolume(pod *corev1.Pod) error {
	logCtx := log.WithFields(log.Fields{"pod": pod.Name, "namespace": pod.Namespace})
	logCtx.Debug("Start to process an evicted pod")
	for _, vol := range pod.Spec.Volumes {
		if vol.PersistentVolumeClaim == nil {
			continue
		}
		pvc, err := ev.pvcInformer.Lister().PersistentVolumeClaims(pod.Namespace).Get(vol.PersistentVolumeClaim.ClaimName)
		if err != nil {
			// if pvc can't be found in the cluster, the pod should not be able to be scheduled
			logCtx.WithFields(log.Fields{
				"namespace": pod.Namespace,
				"pvc":       vol.PersistentVolumeClaim.ClaimName,
			}).WithError(err).Error("Failed to get the pvc from the cluster")
			return err
		}
		if pvc.Spec.StorageClassName == nil {
			// should not be the CSI pvc, ignore
			continue
		}
		sc, err := ev.scInformer.Lister().Get(*pvc.Spec.StorageClassName)
		if err != nil {
			// can't found storageclass in the cluster, the pod should not be able to be scheduled
			logCtx.WithFields(log.Fields{
				"pvc": pvc.Name,
				"sc":  *pvc.Spec.StorageClassName,
			}).WithError(err).Error("Failed to get the pvc from the cluster")
			return err
		}
		if sc.Provisioner == lvmCSIDriverName || sc.Provisioner == diskCSIDriverName {
			logCtx.WithFields(log.Fields{
				"pvc":    pvc.Name,
				"sc":     sc.Name,
				"volume": pvc.Spec.VolumeName,
				"node":   pod.Spec.NodeName,
			}).Debug("Got a LocalVolume to migrate")

			ev.migrateVolumeQueue.Add(constructMigrateVolumeTask(pvc.Spec.VolumeName, pod.Spec.NodeName))
		}
	}

	return nil
}

func (ev *evictor) startMigrateVolumeWorker(stopCh <-chan struct{}) {
	log.Debug("Migrate Volume Worker is working now")
	go func() {
		for {
			task, shutdown := ev.migrateVolumeQueue.Get()
			if shutdown {
				log.WithFields(log.Fields{"task": task}).Debug("Stop the Migrate Volume worker")
				break
			}
			if err := ev.evictVolume(task); err != nil {
				log.WithFields(log.Fields{"task": task, "error": err.Error()}).Error("Failed to process Migrate Volume, retry later")
				ev.migrateVolumeQueue.AddRateLimited(task)
			} else {
				log.WithFields(log.Fields{"task": task}).Debug("Completed a task for Migrating Volume.")
				ev.migrateVolumeQueue.Forget(task)
			}
			ev.migrateVolumeQueue.Done(task)
		}
	}()

	<-stopCh
	ev.migrateVolumeQueue.Shutdown()
}

func (ev *evictor) evictVolume(migrateTask string) error {
	logCtx := log.WithField("task", migrateTask)
	logCtx.Debug("Start to process an local volume migrate task")

	lvName, nodeName := parseMigrateVolumeTask(migrateTask)
	lv, err := ev.lvInformer.Lister().Get(lvName)
	if err != nil {
		if errors.IsNotFound(err) {
			logCtx.Debug("Not found the LocalVolume in the system, ignore it")
			return nil
		}
		logCtx.WithError(err).Error("Failed to get the LocalVolume from the system, try it later")
		return err
	}
	// if LV is still in use, waiting for it to be released
	if len(lv.Status.PublishedNodeName) > 0 {
		logCtx.WithField("PublishedNode", lv.Status.PublishedNodeName).Warning("LocalVolume is still in use, try it later")
		return fmt.Errorf("not released")
	}

	for _, replica := range lv.Spec.Config.Replicas {
		if replica.Hostname == nodeName {
			// check if there is a Migrate CR submitted for it
			lvMigrates, err := ev.lvMigrateInformer.Informer().GetIndexer().ByIndex(localVolumeNameIndex, lvName)
			if err != nil {
				logCtx.WithError(err).Error("Failed to get the migration task for the LocalVolume")
				return err
			}
			if len(lvMigrates) == 0 {
				// no migrate task, submit a new one
				logCtx.Debug("There is no Migrate job running against the LocalVolume, submit a new one")
				lvm := &localstorageapis.LocalVolumeMigrate{
					ObjectMeta: metav1.ObjectMeta{
						Name: lvName,
						//Namespace: "hwameistor",
					},
					Spec: localstorageapis.LocalVolumeMigrateSpec{
						VolumeName:       lvName,
						SourceNodesNames: []string{nodeName},
						// don't specify the target nodes, so the scheduler will select from the avaliables
						TargetNodesNames: []string{},
						MigrateAllVols:   true,
					},
				}
				if _, err := ev.lsClientset.HwameistorV1alpha1().LocalVolumeMigrates().Create(context.Background(), lvm, metav1.CreateOptions{}); err != nil {
					log.WithField("LocalVolumeMigrate", lvm).WithError(err).Error("Failed to submit a migrate job")
					return err
				}
				logCtx.WithField("LocalVolumeMigrate", lvm).Debug("Waiting for the migration job to complete ...")
				return fmt.Errorf("waiting for complete")
			}

			migrateJobs := []string{}
			for i := range lvMigrates {
				lvm, _ := lvMigrates[i].(*localstorageapis.LocalVolumeMigrate)
				migrateJobs = append(migrateJobs, lvm.Name)
			}

			logCtx.WithField("jobs", migrateJobs).Debug("Still waiting for the migration job to complete ...")
			return fmt.Errorf("not completed")
		}
	}

	logCtx.Debug("The migration job has already completed")
	return nil
}

// func getNamespacedName(namespace string, name string) string {
// 	return fmt.Sprintf("%s/%s", namespace, name)
// }

// // output: namespace, name
// func parseNamespacedName(nn string) (string, string) {
// 	items := strings.Split(nn, "/")
// 	if len(items) < 2 {
// 		return items[0], ""
// 	}
// 	return items[0], items[1]
// }

func constructMigrateVolumeTask(lvName string, nodeName string) string {
	return fmt.Sprintf("%s/%s", lvName, nodeName)
}

// output: lvName, nodeName
func parseMigrateVolumeTask(task string) (string, string) {
	items := strings.Split(task, "/")
	if len(items) < 2 {
		return items[0], ""
	}
	return items[0], items[1]
}
