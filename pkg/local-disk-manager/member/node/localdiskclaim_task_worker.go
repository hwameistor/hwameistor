package node

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/member/types"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	types2 "k8s.io/apimachinery/pkg/types"
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
	logCtx := m.logger.WithFields(log.Fields{"diskClaim": diskClaim.Name})
	logCtx.Debugf("Start processing Bound LocalDiskClaim")

	// fetch disks that bounded by the claim
	allocatedDisks, err := fetchBoundedLocalDisks(m.k8sClient, diskClaim)
	if err != nil {
		logCtx.WithError(err).Error("Failed to fetch allocated LocalDisk")
		return err
	}
	logCtx.Infof("Found %d LocalDisk(s) bound by LocalDiskClaim", len(allocatedDisks))

	// find out disks that need to be extended actually
	var tobeExtendedDisks []*v1alpha1.LocalDisk
	poolExtendedSet := make(map[string]struct{})
	for _, localDisk := range allocatedDisks {
		poolName := types.GetLocalDiskPoolName(localDisk.Spec.DiskAttributes.Type)
		poolExtendedSet[poolName] = struct{}{}

		exist := m.registryManager.DiskExist(localDisk.Spec.DevicePath)
		// skip exist disk
		if exist {
			logCtx.Infof("Disk %s already exist in pool, skip it", localDisk.Spec.DevicePath)
			continue
		}
		tobeExtendedDisks = append(tobeExtendedDisks, localDisk.DeepCopy())
	}

	// defer update pool extend record
	defer func() {
		if err == nil {
			poolNames := make([]string, 0, len(poolExtendedSet))
			for poolName := range poolExtendedSet {
				poolNames = append(poolNames, poolName)
			}
			if err = m.updatePoolExtendRecord(poolNames, diskClaim.Spec); err != nil {
				logCtx.WithFields(log.Fields{"poolNames": poolNames}).WithError(err).Error("Failed to update extend record")
			}
		}
	}()

	if len(tobeExtendedDisks) == 0 {
		logCtx.Info("No LocalDisk need to process")
		return confirmLocalDisksConsumed(m.k8sClient, diskClaim)
	}
	logCtx.Infof("Found %d LocalDisk(s) need to process", len(tobeExtendedDisks))

	// record pool extend events
	defer func() {
		if err != nil {
			m.recorder.Event(diskClaim, v1.EventTypeWarning, string(v1alpha1.StorageExpandFailure), fmt.Sprintf("Failed to expand storage capacity, err: %s", err.Error()))
			return
		}
		m.recorder.Event(diskClaim, v1.EventTypeNormal, string(v1alpha1.StorageExpandSuccess), fmt.Sprintf("Succeed to expand storage capacity"))
	}()

	for _, disk := range tobeExtendedDisks {
		poolName := types.GetLocalDiskPoolName(disk.Spec.DiskAttributes.Type)
		ok, err := m.poolManager.ExtendPool(poolName, disk.Spec.DevLinks, disk.Spec.DiskAttributes.SerialNumber)
		if ok {
			logCtx.WithFields(log.Fields{"poolName": poolName, "extendDisk": disk.Spec.DevicePath}).Infof("Succeed to expand DiskPool")
		} else if err != nil {
			logCtx.WithError(err).WithFields(log.Fields{"poolName": poolName, "extendDisk": disk.Spec.DevicePath}).Errorf("Failed to expand DiskPool")
			return err
		}
	}

	// finally confirm disks consumed
	return confirmLocalDisksConsumed(m.k8sClient, diskClaim)
}

func (m *nodeManager) updatePoolExtendRecord(poolNames []string, record v1alpha1.LocalDiskClaimSpec) error {
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

	// append pool records
	for _, poolName := range poolNames {
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
	}

	return m.k8sClient.Status().Patch(context.TODO(), &storageNode, client.MergeFrom(storageNodeOld))
}

func fetchBoundedLocalDisks(cli client.Client, diskClaim *v1alpha1.LocalDiskClaim) ([]*v1alpha1.LocalDisk, error) {
	var allocatedDisks []*v1alpha1.LocalDisk
	eg, _ := errgroup.WithContext(context.TODO())
	for _, localDisk := range diskClaim.Spec.DiskRefs {
		diskName := localDisk.Name
		eg.Go(func() error {
			disk, err := fetchLocalDisk(cli, diskName)
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
