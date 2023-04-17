package node

import (
	"context"
	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/member/types"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"k8s.io/apimachinery/pkg/api/errors"
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
func (m *nodeManager) processLocalDiskClaimBound(diskClaim *v1alpha1.LocalDiskClaim) error {
	logCtx := m.logger.WithFields(log.Fields{"diskClaim": diskClaim.Name, "status": diskClaim.Status.Status})
	logCtx.Debugf("Start processing Bound LocalDiskClaim")

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

	// disks backing the diskClaim must be the same class
	poolName := types.GetLocalDiskPoolName(diskClaim.Spec.Description.DiskType)
	for _, disk := range tobeExtendedDisks {
		ok, err := m.poolManager.ExtendPool(poolName, disk.Spec.DevicePath)
		// don't block pool expand process for one disk error
		if ok {
			logCtx.WithFields(log.Fields{"poolName": poolName, "extendDisk": disk.Spec.DevicePath}).Infof("Succeed to expand DiskPool")
		} else if err != nil {
			logCtx.WithError(err).WithFields(log.Fields{"poolName": poolName, "extendDisk": disk.Spec.DevicePath}).Errorf("Failed to expand DiskPool")
		}
	}

	// sync pool registry to ApiServer
	if err = m.syncNodeResources(); err != nil {
		return err
	}

	// finally confirm disks consumed
	return confirmLocalDisksConsumed(m.k8sClient, diskClaim)
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
