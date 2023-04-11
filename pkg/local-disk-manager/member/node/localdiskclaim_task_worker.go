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
		return nil
	}

	return err
}

// processLocalDiskClaimBound use all disks backing the claim to set up DiskPool
func (m *nodeManager) processLocalDiskClaimBound(diskClaim *v1alpha1.LocalDiskClaim) error {
	logCtx := m.logger.WithField("diskClaim", diskClaim)

	// fetch allocated disks
	allocatedDisks, err := fetchAllocatedLocalDisks(m.k8sClient, diskClaim)
	if err != nil {
		logCtx.WithError(err).Error("Failed to fetch allocated LocalDisk")
		return err
	}

	// remove disks that already exist in DiskPool
	var tobeExtendedDisks []*v1alpha1.LocalDisk
	for _, localDisk := range allocatedDisks {
		disk := m.registryManager.GetDiskByPath(localDisk.Spec.DevicePath)
		// skip exist disk
		if disk.Name != "" {
			logCtx.Infof("Disk %s already exist in DiskPool, skip it", disk.DevPath)
			continue
		}
		tobeExtendedDisks = append(tobeExtendedDisks, localDisk.DeepCopy())
	}

	logCtx.Infof("Found %d LocalDisk(s) need to process from LocalDiskClaim %s", len(tobeExtendedDisks), diskClaim.Name)

	for _, disk := range tobeExtendedDisks {
		extendDisk := types.Disk{
			DevPath:  disk.Spec.DevicePath,
			Capacity: disk.Spec.Capacity,
			DiskType: disk.Spec.DiskAttributes.DevType,
		}
		extendPool := types.GetLocalDiskPoolName(extendDisk.DiskType)
		ok, e := m.poolManager.ExtendPool(extendPool, extendDisk)

		// don't block pool expand process for one disk error
		if e != nil {
			err = e
			continue
		}
		if ok {
			logCtx.WithFields(log.Fields{"poolName": extendPool, "extendDisk": extendDisk.DevPath}).Infof("Succeed to expand StoragePool")
		}
	}

	// rebuild local pool
	m.rebuildLocalPools()

	// sync pool registry to ApiServer
	// todo

	return err
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
