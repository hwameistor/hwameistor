package node

import (
	"context"
	"fmt"
	"strings"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
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
	localDiskClaim := &apisv1alpha1.LocalDiskClaim{}
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
	case apisv1alpha1.DiskClaimStatusEmpty:
		return nil
	case apisv1alpha1.LocalDiskClaimStatusBound:
		return m.processLocalDiskClaimBound(localDiskClaim)
	case apisv1alpha1.LocalDiskClaimStatusPending:
		return nil
	default:
		logCtx.Error("Invalid LocalDiskClaim state")
	}

	return fmt.Errorf("invalid LocalDiskClaim state")
}

func (m *manager) recordExtendPoolCondition(extend bool, err error) {
	condition := apisv1alpha1.LocalStorageNodeCondition{
		Status:             apisv1alpha1.ConditionTrue,
		LastUpdateTime:     metav1.Now(),
		LastTransitionTime: metav1.Now(),
	}

	if err != nil {
		condition.Type = apisv1alpha1.StorageExpandFailure
		condition.Reason = string(apisv1alpha1.StorageExpandFailure)
		condition.Message = fmt.Sprintf("Failed to expand storage capacity, err: %s", err.Error())
	} else if extend {
		// only record and update condition in extend storage capacity actually
		condition.Type = apisv1alpha1.StorageExpandSuccess
		condition.Reason = string(apisv1alpha1.StorageExpandSuccess)
		condition.Message = "Successfully to expand storage capacity"
	} else {
		// check if any disk has already managed
		if len(m.storageMgr.Registry().Disks()) > 0 {
			condition.Type = apisv1alpha1.StorageAvailable
			condition.Reason = string(apisv1alpha1.StorageAvailable)
			condition.Message = "Sufficient storage capacity"
		} else {
			condition.Type = apisv1alpha1.StorageUnAvailable
			condition.Reason = string(apisv1alpha1.StorageAvailable)
			condition.Message = "Insufficient storage capacity"
		}
	}

	if err := m.storageMgr.Registry().UpdateCondition(condition); err != nil {
		m.logger.WithField("condition", condition).WithError(err).Error("Failed to update condition")
	}
}

func (m *manager) processLocalDiskClaimBound(claim *apisv1alpha1.LocalDiskClaim) (e error) {
	m.logger.Debug("processLocalDiskClaimBound start ...")

	extend := false
	defer func() {
		m.recordExtendPoolCondition(extend, e)
	}()

	availableLocalDisks, err := m.getLocalDisksByLocalDiskClaim(claim)
	if err != nil {
		log.WithError(err).Error("Failed to getLocalDisksByLocalDiskClaim.")
		return err
	}

	if extend, err = m.storageMgr.PoolManager().ExtendPools(availableLocalDisks); err != nil {
		log.WithError(err).Error("Failed to ExtendPools")
		return err
	}

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
func (m *manager) getLocalDisksByLocalDiskClaim(ldc *apisv1alpha1.LocalDiskClaim) ([]*apisv1alpha1.LocalDevice, error) {
	localDisksMap, err := m.getLocalDisksMapByLocalDiskClaim(ldc)
	if err != nil {
		return nil, err
	}

	localDisks := []*apisv1alpha1.LocalDevice{}
	for _, disk := range localDisksMap {
		localDisks = append(localDisks, disk)
	}

	return localDisks, nil
}

func (m *manager) getLocalDisksMapByLocalDiskClaim(ldc *apisv1alpha1.LocalDiskClaim) (map[string]*apisv1alpha1.LocalDevice, error) {
	m.logger.Debug("getLocalDisksMapByLocalDiskClaim ...")
	disks := make(map[string]*apisv1alpha1.LocalDevice)
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
		disk := &apisv1alpha1.LocalDevice{}
		disk.State = apisv1alpha1.DiskStateAvailable
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
		disk := &apisv1alpha1.LocalDevice{}
		disk.State = apisv1alpha1.DiskStateInUse
		disk.CapacityBytes = diskInUse.Spec.Capacity
		disk.DevPath = devicePath
		disk.Class = diskInUse.Spec.DiskAttributes.Type
		disks[devicePath] = disk
	}
	return disks, nil
}

func (m *manager) listAllAvailableLocalDisksByLocalClaimDisk(ldc *apisv1alpha1.LocalDiskClaim) ([]*apisv1alpha1.LocalDisk, error) {
	m.logger.Debug("listAllAvailableLocalDisksByLocalClaimDisk ...")
	localDisks, err := m.listLocalDisksByLocalDiskClaim(ldc)
	if err != nil {
		m.logger.WithError(err).Error("Failed to listLocalDisksByLocalDiskClaim")
		return nil, err
	}
	availableLocalDisks := []*apisv1alpha1.LocalDisk{}
	for _, ld := range localDisks {
		if ld.Spec.HasPartition {
			continue
		}

		if ld.Spec.State == apisv1alpha1.LocalDiskActive {
			availableLocalDisks = append(availableLocalDisks, ld)
		}
	}
	return availableLocalDisks, nil
}

func (m *manager) listAllInUseLocalDisksByLocalClaimDisk(ldc *apisv1alpha1.LocalDiskClaim) ([]*apisv1alpha1.LocalDisk, error) {
	m.logger.Debug("listAllInUseLocalDisksByLocalClaimDisk ...")
	localDisks, err := m.listLocalDisksByLocalDiskClaim(ldc)
	if err != nil {
		m.logger.WithError(err).Error("Failed to listLocalDisksByLocalDiskClaim")
		return nil, err
	}
	inUseLocalDisks := []*apisv1alpha1.LocalDisk{}
	for _, ld := range localDisks {
		if !ld.Spec.HasPartition {
			continue
		}
		if ld.Spec.State == apisv1alpha1.LocalDiskActive {
			inUseLocalDisks = append(inUseLocalDisks, ld)
		}
	}
	return inUseLocalDisks, nil
}

func (m *manager) listLocalDisksByLocalDiskClaim(ldc *apisv1alpha1.LocalDiskClaim) ([]*apisv1alpha1.LocalDisk, error) {
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

func (m *manager) getLocalDisksByDiskRefs(localDiskNames []string, nameSpace string) ([]*apisv1alpha1.LocalDisk, error) {
	m.logger.Debug("getLocalDisksByDiskRefs ...")
	var wg sync.WaitGroup
	localDiskList := []*apisv1alpha1.LocalDisk{}
	for _, diskName := range localDiskNames {
		name := diskName
		wg.Add(1)
		go func() {
			defer wg.Done()
			localDisk, err := m.getLocalDiskByName(name, nameSpace)
			if err != nil {
				//m.logger.Error("Failed to getLocalDiskByName name = %v, err = %", name, err)
				//fmt.Errorf("failed to getLocalDiskByName name = %v, err = %v", name, err)
				return
			}
			if localDisk != nil && localDisk.Status.State == apisv1alpha1.LocalDiskBound {
				localDiskList = append(localDiskList, localDisk)
			}
		}()
	}
	wg.Wait()
	return localDiskList, nil
}

func (m *manager) getLocalDiskByName(localDiskName, nameSpace string) (*apisv1alpha1.LocalDisk, error) {
	logCtx := m.logger.WithFields(log.Fields{"localDisk": localDiskName})
	logCtx.Debug("getLocalDiskByName ...")
	localDisk := &apisv1alpha1.LocalDisk{}
	if err := m.apiClient.Get(context.TODO(), types.NamespacedName{Name: localDiskName}, localDisk); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("Failed to get localDisk from cache, retry it later ...")
			return nil, err
		}
		logCtx.Info("Not found the localDisk from cache, should be deleted already.")
		return nil, err
	}
	return localDisk, nil
}
