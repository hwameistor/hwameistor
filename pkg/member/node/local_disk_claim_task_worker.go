package node

import (
	"context"
	"fmt"
	"strings"
	"sync"

	ldmv1alpha1 "github.com/hwameistor/local-disk-manager/pkg/apis/hwameistor/v1alpha1"
	localstoragev1alpha1 "github.com/hwameistor/local-storage/pkg/apis/localstorage/v1alpha1"
	log "github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func (m *manager) startLocalDiskClaimTaskWorker(stopCh <-chan struct{}) {

	m.logger.Debug("LocalDiskClaim Worker is working now")
	go func() {
		for {
			task, shutdown := m.localDiskClaimTaskQueue.Get()
			if shutdown {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Stop the LocalDiskClaim worker")
				break
			}
			if err := m.processLocalDiskClaim(task); err != nil {
				m.logger.WithFields(log.Fields{"task": task, "error": err.Error()}).Error("Failed to process LocalDiskClaim task, retry later")
				m.localDiskClaimTaskQueue.AddRateLimited(task)
			} else {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Completed a LocalDiskClaim task.")
				m.localDiskClaimTaskQueue.Forget(task)
			}
			m.localDiskClaimTaskQueue.Done(task)
		}
	}()

	<-stopCh
	m.volumeReplicaTaskQueue.Shutdown()
}

func (m *manager) processLocalDiskClaim(localDiskNameSpacedName string) error {
	m.logger.Debug("processLocalDiskClaim start ...")
	logCtx := m.logger.WithFields(log.Fields{"LocalDiskClaim": localDiskNameSpacedName})

	logCtx.Debug("Working on a LocalDiskClaim task")
	splitRes := strings.Split(localDiskNameSpacedName, "/")
	var nameSpace, diskName string
	if len(splitRes) >= 2 {
		nameSpace = splitRes[0]
		diskName = splitRes[1]
	}
	localDiskClaim := &ldmv1alpha1.LocalDiskClaim{}
	if err := m.apiClient.Get(context.TODO(), types.NamespacedName{Namespace: nameSpace, Name: diskName}, localDiskClaim); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("Failed to get LocalDiskClaim from cache, retry it later ...")
			return err
		}
		//logCtx.Info("Not found the LocalDiskClaim from cache, should be deleted already. err = %v", err)
		fmt.Printf("Not found the LocalDiskClaim from cache, should be deleted already. err = %v", err)
		return nil
	}

	m.logger.Debugf("Required node name %s, current node name %s.", localDiskClaim.Spec.NodeName, m.name)
	if localDiskClaim.Spec.NodeName != m.name {
		return nil
	}

	switch localDiskClaim.Status.Status {
	case ldmv1alpha1.DiskClaimStatusEmpty:
		return nil
	case ldmv1alpha1.LocalDiskClaimStatusBound:
		return m.processLocalDiskClaimBound(localDiskClaim)
	case ldmv1alpha1.LocalDiskClaimStatusPending:
		return nil
	default:
		logCtx.Error("Invalid LocalDiskClaim state")
	}

	return fmt.Errorf("invalid LocalDiskClaim state")
}

func (m *manager) processLocalDiskClaimBound(claim *ldmv1alpha1.LocalDiskClaim) error {
	m.logger.Debug("processLocalDiskClaimBound start ...")

	availableLocalDisks, err := m.getLocalDisksByLocalDiskClaim(claim)
	if err != nil {
		log.WithError(err).Error("Failed to getLocalDisksByLocalDiskClaim.")
		return err
	}

	if err := m.storageMgr.PoolManager().ExtendPools(availableLocalDisks); err != nil {
		log.WithError(err).Error("Failed to ExtendPools")
		return err
	}

	// 如果lvm扩容失败，就不会执行如下同步资源流程
	localDisks, err := m.getLocalDisksMapByLocalDiskClaim(claim)
	if err != nil {
		log.WithError(err).Error("Failed to getLocalDisksMapByLocalDiskClaim")
		return err
	}
	//m.logger.Debug("processLocalDiskClaimBound getLocalDisksMapByLocalDiskClaim  localDisks = %v, claim = %v", localDisks, claim)
	fmt.Printf("processLocalDiskClaimBound getLocalDisksMapByLocalDiskClaim  localDisks = %v, claim = %v", localDisks, claim)

	if err := m.storageMgr.Registry().SyncResourcesToNodeCRD(localDisks); err != nil {
		log.WithError(err).Error("Failed to SyncResourcesToNodeCRD")
		return err
	}
	return nil
}

// getLocalDisksByLocalDiskClaim get disks, including HDD, SSD, NVMe triggered by ldc callback
func (m *manager) getLocalDisksByLocalDiskClaim(ldc *ldmv1alpha1.LocalDiskClaim) ([]*localstoragev1alpha1.LocalDisk, error) {
	localDisksMap, err := m.getLocalDisksMapByLocalDiskClaim(ldc)
	if err != nil {
		return nil, err
	}

	localDisks := []*localstoragev1alpha1.LocalDisk{}
	for _, disk := range localDisksMap {
		localDisks = append(localDisks, disk)
	}

	return localDisks, nil
}

func (m *manager) getLocalDisksMapByLocalDiskClaim(ldc *ldmv1alpha1.LocalDiskClaim) (map[string]*localstoragev1alpha1.LocalDisk, error) {
	m.logger.Debug("getLocalDisksMapByLocalDiskClaim ...")
	disks := make(map[string]*localstoragev1alpha1.LocalDisk)
	disksAvailable, err := m.listAllAvailableLocalDisksByLocalClaimDisk(ldc)
	if err != nil {
		m.logger.WithError(err).Error("Failed to listAllAvailableLocalDisks")
		return disks, err
	}
	for _, diskAvailable := range disksAvailable {
		if diskAvailable == nil {
			continue
		}
		devicePath := diskAvailable.Spec.DevicePath
		if devicePath == "" || !strings.HasPrefix(devicePath, "/dev") || strings.Contains(devicePath, "mapper") {
			continue
		}
		disk := &localstoragev1alpha1.LocalDisk{}
		disk.State = localstoragev1alpha1.DiskStateAvailable
		disk.CapacityBytes = diskAvailable.Spec.Capacity
		disk.DevPath = devicePath
		disk.Class = diskAvailable.Spec.DiskAttributes.Type
		disks[devicePath] = disk
	}

	disksInUse, err := m.listAllInUseLocalDisksByLocalClaimDisk(ldc)
	if err != nil {
		m.logger.WithError(err).Error("Failed to listAllInUseLocalDisks")
		return disks, err
	}
	for _, diskInUse := range disksInUse {
		if diskInUse == nil {
			continue
		}
		devicePath := diskInUse.Spec.DevicePath
		if devicePath == "" || !strings.HasPrefix(devicePath, "/dev") || strings.Contains(devicePath, "mapper") {
			continue
		}
		disk := &localstoragev1alpha1.LocalDisk{}
		disk.State = localstoragev1alpha1.DiskStateInUse
		disk.CapacityBytes = diskInUse.Spec.Capacity
		disk.DevPath = devicePath
		disk.Class = diskInUse.Spec.DiskAttributes.Type
		disks[devicePath] = disk
	}
	return disks, nil
}

func (m *manager) listAllAvailableLocalDisksByLocalClaimDisk(ldc *ldmv1alpha1.LocalDiskClaim) ([]*ldmv1alpha1.LocalDisk, error) {
	m.logger.Debug("listAllAvailableLocalDisksByLocalClaimDisk ...")
	localDisks, err := m.listLocalDisksByLocalDiskClaim(ldc)
	if err != nil {
		m.logger.WithError(err).Error("Failed to listLocalDisksByLocalDiskClaim")
		return nil, err
	}
	availableLocalDisks := []*ldmv1alpha1.LocalDisk{}
	for _, ld := range localDisks {
		if ld.Spec.HasPartition == true {
			continue
		}

		for _, partition := range ld.Spec.PartitionInfo {
			if partition.HasFileSystem == true {
				continue
			}
		}

		if ld.Spec.State == ldmv1alpha1.LocalDiskActive {
			availableLocalDisks = append(availableLocalDisks, ld)
		}
	}
	return availableLocalDisks, nil
}

func (m *manager) listAllInUseLocalDisksByLocalClaimDisk(ldc *ldmv1alpha1.LocalDiskClaim) ([]*ldmv1alpha1.LocalDisk, error) {
	m.logger.Debug("listAllInUseLocalDisksByLocalClaimDisk ...")
	localDisks, err := m.listLocalDisksByLocalDiskClaim(ldc)
	if err != nil {
		m.logger.WithError(err).Error("Failed to listLocalDisksByLocalDiskClaim")
		return nil, err
	}
	inUseLocalDisks := []*ldmv1alpha1.LocalDisk{}
	for _, ld := range localDisks {
		if ld.Spec.HasPartition == false {
			for _, partition := range ld.Spec.PartitionInfo {
				if partition.HasFileSystem == false {
					continue
				}
			}
		}
		if ld.Spec.State == ldmv1alpha1.LocalDiskActive {
			inUseLocalDisks = append(inUseLocalDisks, ld)
		}
	}
	return inUseLocalDisks, nil
}

func (m *manager) listLocalDisksByLocalDiskClaim(ldc *ldmv1alpha1.LocalDiskClaim) ([]*ldmv1alpha1.LocalDisk, error) {
	m.logger.Debug("listLocalDisksByLocalDiskClaim ...")
	if ldc == nil {
		err := errors.NewBadRequest("ldc cannot be nil")
		m.logger.WithError(err).Error("Failed to listLocalDisksByLocalDiskClaim")
		return nil, err
	}
	var diskNames []string
	for _, diskRef := range ldc.Spec.DiskRefs {
		if diskRef == nil {
			continue
		}
		diskNames = append(diskNames, diskRef.Name)
	}
	localDisks, _ := m.getLocalDisksByDiskRefs(diskNames, ldc.Namespace)
	return localDisks, nil
}

func (m *manager) getLocalDisksByDiskRefs(localDiskNames []string, nameSpace string) ([]*ldmv1alpha1.LocalDisk, error) {
	m.logger.Debug("getLocalDisksByDiskRefs ...")
	var wg sync.WaitGroup
	localDiskList := []*ldmv1alpha1.LocalDisk{}
	for _, diskName := range localDiskNames {
		name := diskName
		wg.Add(1)
		go func() {
			defer wg.Done()
			localDisk, err := m.getLocalDiskByName(name, nameSpace)
			if err != nil {
				//m.logger.Error("Failed to getLocalDiskByName name = %v, err = %", name, err)
				fmt.Errorf("Failed to getLocalDiskByName name = %v, err = %v", name, err)
				return
			}
			if localDisk != nil && localDisk.Status.State == ldmv1alpha1.LocalDiskClaimed {
				localDiskList = append(localDiskList, localDisk)
			}
		}()
	}
	wg.Wait()
	return localDiskList, nil
}

func (m *manager) getLocalDiskByName(localDiskName, nameSpace string) (*ldmv1alpha1.LocalDisk, error) {
	logCtx := m.logger.WithFields(log.Fields{"LocalDisk": localDiskName})
	logCtx.Debug("getLocalDiskByName ...")
	localDisk := &ldmv1alpha1.LocalDisk{}
	if err := m.apiClient.Get(context.TODO(), types.NamespacedName{Namespace: nameSpace, Name: localDiskName}, localDisk); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("Failed to get LocalDisk from cache, retry it later ...")
			return nil, err
		}
		logCtx.Info("Not found the LocalDisk from cache, should be deleted already.")
		return nil, err
	}
	return localDisk, nil
}
