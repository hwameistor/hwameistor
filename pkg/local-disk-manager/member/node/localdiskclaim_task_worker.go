package node

import (
	"context"
	"fmt"
	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/member/types"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	types2 "k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (m *nodeManager) startDiskClaimTaskWorker(ctx context.Context) {
	m.logger.Info("Start LocalDiskClaim worker now")
	go func() {
		for {
			task, shutdown := m.diskClaimTaskQueue.Get()
			if shutdown {
				m.logger.Info("Stop the LocalDiskClaim worker")
				break
			}
			if err := m.processLocalDiskClaim(task); err != nil {
				m.logger.WithFields(log.Fields{"task": task, "error": err.Error()}).Error("Failed to process LocalDiskClaim task, retry later")
				m.diskClaimTaskQueue.AddRateLimited(task)
			} else {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Completed a LocalDiskClaim task")
				m.diskClaimTaskQueue.Forget(task)
			}
			m.diskClaimTaskQueue.Done(task)
		}
	}()

	// We are done, Stop Node Manager
	<-ctx.Done()
	m.diskClaimTaskQueue.Shutdown()
	return
}

// processLocalDiskClaim handle the diskClaim which owned by LocalDiskManager and setup DiskPool with the disks backing the claim
func (m *nodeManager) processLocalDiskClaim(diskClaim string) error {
	logCtx := m.logger.WithField("diskClaim", diskClaim)
	logCtx.Info("Start processing LocalDiskClaim")

	localDiskClaim := &v1alpha1.LocalDiskClaim{}
	err := m.k8sClient.Get(context.TODO(), client.ObjectKey{Name: diskClaim}, localDiskClaim)
	if err != nil {
		if errors.IsNotFound(err) {
			logCtx.Info("LocalDiskClaim has been deleted already, skip processing this object")
			return nil
		}
		return err
	}

	// handle diskClaim according to diskClaim status
	switch localDiskClaim.Status.Status {
	case v1alpha1.LocalDiskClaimStatusBound:
		err = m.processLocalDiskClaimBound(localDiskClaim)
	default:
		logCtx.WithField("Status", localDiskClaim.Status.Status).Info("Skip processing LocalDiskClaim")
		return nil
	}

	return err
}

// processLocalDiskClaimBound use all disks backing the claim to set up DiskPool
func (m *nodeManager) processLocalDiskClaimBound(diskClaim *v1alpha1.LocalDiskClaim) (err error) {
	logCtx := m.logger.WithFields(log.Fields{"diskClaim": diskClaim.Name, "status": diskClaim.Status.Status})
	logCtx.Debugf("Start processing Bound LocalDiskClaim")

	// disks backing the diskClaim must be the same class
	poolName := types.GetLocalDiskPoolName(diskClaim.Spec.Description.DiskType)
	defer func() {
		if err == nil {
			if err = m.updatePoolExtendRecord(poolName, diskClaim.Spec); err != nil {
				m.logger.WithFields(log.Fields{"poolName": poolName, "diskClaim": diskClaim.Name}).WithError(err).Error("Failed to update extend record")
			}
		}
	}()

	// fetch allocated disks
	allocatedDisks, err := fetchAllocatedLocalDisks(m.k8sClient, diskClaim)
	if err != nil {
		logCtx.WithError(err).Error("Failed to fetch allocated LocalDisk")
		return err
	}

	// remove disks that already exist in DiskPool
	var tobeExtendedDisks []*v1alpha1.LocalDisk
	for _, localDisk := range allocatedDisks {
		exist := m.registryManager.DiskExist(localDisk.Spec.DevicePath)
		// skip exist disk
		if exist {
			logCtx.Infof("Disk %s already exist in DiskPool, skip it", localDisk.Spec.DevicePath)
			continue
		}
		tobeExtendedDisks = append(tobeExtendedDisks, localDisk.DeepCopy())
	}

	if len(tobeExtendedDisks) == 0 {
		logCtx.Infof("No LocalDisk found from LocalDiskClaim %s", diskClaim.Name)
		return confirmLocalDisksConsumed(m.k8sClient, diskClaim)
	}
	logCtx.Infof("Found %d LocalDisk(s) need to process from LocalDiskClaim %s", len(tobeExtendedDisks), diskClaim.Name)

	// record pool extend events
	defer func() {
		if err != nil {
			m.recorder.Event(diskClaim, v1.EventTypeWarning, string(v1alpha1.StorageExpandFailure), fmt.Sprintf("Failed to expand storage capacity, err: %s", err.Error()))
			return
		}
		m.recorder.Event(diskClaim, v1.EventTypeNormal, string(v1alpha1.StorageExpandSuccess), fmt.Sprintf("Succeed to expand storage capacity"))
	}()

	for _, disk := range tobeExtendedDisks {
		ok, err := m.poolManager.ExtendPool(poolName, disk.Spec.DevicePath)
		// don't block pool expand process for one disk error
		if ok {
			logCtx.WithFields(log.Fields{"poolName": poolName, "extendDisk": disk.Spec.DevicePath}).Infof("Succeed to expand DiskPool")
		} else if err != nil {
			logCtx.WithError(err).WithFields(log.Fields{"poolName": poolName, "extendDisk": disk.Spec.DevicePath}).Errorf("Failed to expand DiskPool")
		}
	}

	// finally confirm disks consumed
	return confirmLocalDisksConsumed(m.k8sClient, diskClaim)
}

func (m *nodeManager) updatePoolExtendRecord(poolName string, record v1alpha1.LocalDiskClaimSpec) error {
	var storageNode v1alpha1.LocalDiskNode
	err := m.k8sClient.Get(context.TODO(), types2.NamespacedName{Name: m.nodeName}, &storageNode)
	if err != nil {
		return err
	}
	storageNodeOld := storageNode.DeepCopy()

	// init records map
	if storageNode.Status.PoolExtendRecords == nil {
		storageNode.Status.PoolExtendRecords = make(map[string]v1alpha1.LocalDiskClaimSpecArray)
	}

	// init pool records
	if _, ok := storageNode.Status.PoolExtendRecords[poolName]; !ok {
		storageNode.Status.PoolExtendRecords[poolName] = make(v1alpha1.LocalDiskClaimSpecArray, 0)
	}

	// append this record if not exist
	exist := false
	for _, poolRecord := range storageNode.Status.PoolExtendRecords[poolName] {
		if reflect.DeepEqual(poolRecord, record) {
			exist = true
		}
	}
	if !exist {
		storageNode.Status.PoolExtendRecords[poolName] = append(storageNode.Status.PoolExtendRecords[poolName], record)
	}
	return m.k8sClient.Status().Patch(context.TODO(), &storageNode, client.MergeFrom(storageNodeOld))
}

func fetchAllocatedLocalDisks(cli client.Client, diskClaim *v1alpha1.LocalDiskClaim) ([]*v1alpha1.LocalDisk, error) {
	var allocatedDisks []*v1alpha1.LocalDisk
	eg, _ := errgroup.WithContext(context.TODO())
	for _, localDisk := range diskClaim.Spec.DiskRefs {
		eg.Go(func() error {
			disk, err := fetchLocalDisk(cli, localDisk.Name)
			if err != nil {
				return err
			}
			allocatedDisks = append(allocatedDisks, disk)
			return nil
		})
	}
	return allocatedDisks, eg.Wait()
}

func fetchLocalDisk(cli client.Client, localDisk string) (*v1alpha1.LocalDisk, error) {
	localDiskObject := &v1alpha1.LocalDisk{}
	err := cli.Get(context.TODO(), client.ObjectKey{Name: localDisk}, localDiskObject)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Infof("LocalDisk %s has been deleted already", localDisk)
			return nil, err
		}
	}
	return localDiskObject, err
}

func confirmLocalDisksConsumed(cli client.Client, diskClaim *v1alpha1.LocalDiskClaim) error {
	diskClaim.Status.Status = v1alpha1.LocalDiskClaimStatusConsumed
	return cli.Status().Update(context.TODO(), diskClaim)
}
