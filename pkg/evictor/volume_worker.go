package evictor

import (
	"context"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	localstorageapis "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

const (
	evictionFinalizer = "hwameistor.io/eviction-protect"
)

func (ev *evictor) startVolumeWorker(stopCh <-chan struct{}) {
	log.Debug("Start a worker to process volume eviction")
	go func() {
		for {
			task, shutdown := ev.evictVolumeQueue.Get()
			if shutdown {
				log.WithFields(log.Fields{"task": task}).Debug("Stop the volume eviction worker")
				break
			}
			if err := ev.evictVolume(task); err != nil {
				log.WithFields(log.Fields{"task": task, "error": err.Error()}).Error("Failed to process volume eviction task, retry later ...")
				ev.evictVolumeQueue.AddRateLimited(task)
			} else {
				log.WithFields(log.Fields{"task": task}).Debug("Completed a volume eviction task.")
				ev.evictVolumeQueue.Forget(task)
			}
			ev.evictVolumeQueue.Done(task)
		}
	}()

	<-stopCh
	ev.evictVolumeQueue.Shutdown()
}

func (ev *evictor) evictVolume(task string) error {
	volName, srcNodeName := parseEvictVolumeTask(task)
	logCtx := log.WithFields(log.Fields{"volume": volName, "sourceNode": srcNodeName})
	logCtx.Debug("Start to process a volume eviction")

	lvmName := fmt.Sprintf("evictor-%s", volName)
	lvm, err := ev.lvMigrateInformer.Lister().Get(lvmName)
	if err == nil {
		// already has a migrate, check the status
		if lvm.Status.State == localstorageapis.OperationStateCompleted {
			logCtx.Debug("Volume migration completed")
			lvm.Finalizers = []string{}
			if _, err := ev.lsClientset.HwameistorV1alpha1().LocalVolumeMigrates().Update(context.TODO(), lvm, metav1.UpdateOptions{}); err != nil {
				logCtx.WithField("migrate", lvm.Name).WithError(err).Error("Failed to cleanup the migration")
				return err
			}
			return nil
		}
		logCtx.Debug("Volume migration still in progress")
		return fmt.Errorf("volume migration in progress")
	}
	if !errors.IsNotFound(err) {
		logCtx.WithField("migrate", lvm.Name).WithError(err).Error("Failed to fetch the migration from cache")
		return err
	}

	vol, err := ev.lvInformer.Lister().Get(volName)
	if err != nil {
		if errors.IsNotFound(err) {
			logCtx.Debug("Not found the LocalVolume, ignore it")
			return nil
		}
		logCtx.WithError(err).Error("Failed to get the LocalVolume from cache, try it later")
		return err
	}

	for _, replica := range vol.Spec.Config.Replicas {
		if replica.Hostname == srcNodeName {
			lvm := &localstorageapis.LocalVolumeMigrate{
				ObjectMeta: metav1.ObjectMeta{
					Name:       lvmName,
					Finalizers: []string{evictionFinalizer},
				},
				Spec: localstorageapis.LocalVolumeMigrateSpec{
					VolumeName: volName,
					SourceNode: srcNodeName,
					// don't specify the target nodes, so the scheduler will select from the avaliables
					TargetNodesSuggested: []string{},
					MigrateAllVols:       true,
				},
			}
			if _, err := ev.lsClientset.HwameistorV1alpha1().LocalVolumeMigrates().Create(context.Background(), lvm, metav1.CreateOptions{}); err != nil {
				log.WithField("migrate", lvm.Name).WithError(err).Error("Failed to submit a migrate job")
				return err
			}
			logCtx.WithField("migrate", lvm.Name).Debug("Submitted a migrate task")
			return fmt.Errorf("volume migration in progress")
		}
	}

	logCtx.Debug("No volume replica to be migrated")
	return nil
}

func (ev *evictor) addEvictVolume(volName string, srcNodeName string) {
	ev.evictVolumeQueue.AddRateLimited(fmt.Sprintf("%s/%s", volName, srcNodeName))
}

func parseEvictVolumeTask(task string) (volName string, srcNodeName string) {
	items := strings.Split(task, "/")
	volName = items[0]
	srcNodeName = items[1]
	return
}
