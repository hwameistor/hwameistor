package node

import (
	"context"
	"math"
	"strings"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-storage/member/node/diskmonitor"
)

func (m *manager) startLocalDiskTaskWorker(stopCh <-chan struct{}) {

	m.logger.Debug("localDisk Worker is working now")
	go func() {
		for {
			task, shutdown := m.localDiskTaskQueue.Get()
			if shutdown {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Stop the localDisk worker")
				break
			}
			if err := m.processLocalDisk(task); err != nil {
				m.logger.WithFields(log.Fields{"task": task, "error": err.Error()}).Error("Failed to process localDisk task, retry later")
				m.localDiskTaskQueue.AddRateLimited(task)
			} else {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Completed a localDisk task.")
				m.localDiskTaskQueue.Forget(task)
			}
			m.localDiskTaskQueue.Done(task)
		}
	}()

	<-stopCh
	m.volumeReplicaTaskQueue.Shutdown()
}

func (m *manager) processLocalDisk(localDiskNameSpacedName string) error {
	m.logger.Debug("processLocalDisk start ...")
	logCtx := m.logger.WithFields(log.Fields{"localDisk": localDiskNameSpacedName})
	logCtx.Debug("Working on a localDisk task")
	splitRes := strings.Split(localDiskNameSpacedName, "/")
	var diskName string
	if len(splitRes) >= 2 {
		// nameSpace = splitRes[0]
		diskName = splitRes[1]
	}

	localDisk := &apisv1alpha1.LocalDisk{}
	if err := m.apiClient.Get(context.TODO(), types.NamespacedName{Name: diskName}, localDisk); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("Failed to get localDisk from cache, retry it later ...")
			return err
		}
		logCtx.Info("Not found the localDisk from cache, should be deleted already.")
		return nil
	}

	m.logger.Debugf("Required node name %s, current node name %s.", localDisk.Spec.NodeName, m.name)
	if localDisk.Spec.NodeName != m.name {
		return nil
	}

	var err error
	switch localDisk.Status.State {
	case apisv1alpha1.LocalDiskAvailable:
		err = m.processLocalDiskAvailable(localDisk)

	case apisv1alpha1.LocalDiskBound:
		err = m.processLocalDiskBound(localDisk)

	default:
		logCtx.Error("Invalid localDisk state")
	}

	return err
}

func (m *manager) processLocalDiskBound(localDisk *apisv1alpha1.LocalDisk) error {
	logCtx := m.logger.WithFields(log.Fields{"localDisk": localDisk.GetName()})
	logCtx.Info("Start processing Bound localdisk")

	// when disk is bound, means disk has been used in StoragePool, so we need to
	// make notification if disk state has changed(i.g Active -> InActive -> UnKnown)
	m.handleDiskStateChange(localDisk)

	err := m.resizeStoragePoolCapacity(localDisk)
	if err != nil {
		logCtx.WithError(err).Error("Failed to resize storagepool capacity")
		return err
	}

	return nil
}

func (m *manager) processLocalDiskAvailable(_ *apisv1alpha1.LocalDisk) error {
	return nil
}

// resize StoragePool capacity according disk's capacity
func (m *manager) resizeStoragePoolCapacity(localDisk *apisv1alpha1.LocalDisk) error {
	// find out if disk has used in StoragePool and compare recorded capacity and current capacity
	// if capacity has been resized, rebuild localstorage registry and update node resource
	registeredDisks := m.Storage().Registry().Disks()
	registeredDisk, exist := registeredDisks[localDisk.Spec.DevicePath]
	if !exist {
		return nil
	}

	if !m.needResizePVCapacity(localDisk.Spec.Capacity, registeredDisk.CapacityBytes) {
		return nil
	}

	var resizeDisks = map[string]*apisv1alpha1.LocalDevice{
		localDisk.Spec.DevicePath: {
			DevPath:       localDisk.Spec.DevicePath,
			Class:         localDisk.Spec.DiskAttributes.Type,
			CapacityBytes: localDisk.Spec.Capacity,
			State:         apisv1alpha1.DiskStateInUse,
		},
	}
	// resize pv first
	err := m.Storage().PoolManager().ResizePhysicalVolumes(resizeDisks)
	if err != nil {
		return err
	}

	return m.Storage().Registry().SyncNodeResources()
}

// reduce disk from StoragePool or do migrate volume which located on the disk according to disk state
func (m *manager) handleDiskStateChange(localDisk *apisv1alpha1.LocalDisk) {
	logCtx := m.logger.WithFields(log.Fields{"localDisk": localDisk.GetName()})
	switch localDisk.Spec.State {
	case apisv1alpha1.LocalDiskInactive:
		event := &diskmonitor.DiskEvent{}
		m.diskEventQueue.Add(event)
	case apisv1alpha1.LocalDiskActive:
	case apisv1alpha1.LocalDiskUnknown:
	default:
		logCtx.Error("Invalid localDisk state")
	}
	return
}

func (m *manager) needResizePVCapacity(currentCapacity, recordCapacity int64) bool {
	// consider metadata size, the capacity in pv may smaller than actual disk capacity
	var pe = 4 * 1024 * 1024
	return math.Abs(float64(currentCapacity-recordCapacity)) > float64(pe)
}
