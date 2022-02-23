package storage

import (
	"fmt"
	"strings"

	localstoragev1alpha1 "github.com/hwameistor/local-storage/pkg/apis/localstorage/v1alpha1"
	"github.com/hwameistor/local-storage/pkg/utils"
	log "github.com/sirupsen/logrus"
)

type localPoolManager struct {
	cmdExec        LocalPoolExecutor
	ramDiskCmdExec LocalPoolExecutor
	logger         *log.Entry
	lm             *LocalManager
}

func (mgr *localPoolManager) ExtendPools(localDisks []*localstoragev1alpha1.LocalDisk) error {
	if mgr.cmdExec == nil {
		return nil
	}

	return mgr.cmdExec.ExtendPools(localDisks)
}

func (mgr *localPoolManager) ExtendPoolsInfo(localDisks map[string]*localstoragev1alpha1.LocalDisk) (map[string]*localstoragev1alpha1.LocalPool, error) {
	pools := map[string]*localstoragev1alpha1.LocalPool{}

	if mgr.cmdExec != nil {
		var err error
		pools, err = mgr.cmdExec.ExtendPoolsInfo(localDisks)
		if err != nil {
			return nil, err
		}
	}
	ramPools, err := mgr.ramDiskCmdExec.ExtendPoolsInfo(localDisks)
	if err != nil {
		return nil, err
	}
	ramPool, has := ramPools[localstoragev1alpha1.PoolNameForRAM]
	if !has {
		return nil, fmt.Errorf("wrong ramdisk pool")
	}
	if mgr.lm.nodeConf.LocalStorageConfig != nil {
		totalCapacityBytes, _ := utils.ParseBytes(mgr.lm.nodeConf.LocalStorageConfig.RAMDiskTotalCapacity)
		ramPool.TotalCapacityBytes = totalCapacityBytes
		ramPool.FreeCapacityBytes = totalCapacityBytes - ramPool.UsedCapacityBytes
		ramPool.VolumeCapacityBytesLimit = totalCapacityBytes
	}
	pools[localstoragev1alpha1.PoolNameForRAM] = ramPool

	return pools, nil
}

func (mgr *localPoolManager) GetReplicas() (map[string]*localstoragev1alpha1.LocalVolumeReplica, error) {
	replicas := map[string]*localstoragev1alpha1.LocalVolumeReplica{}

	if mgr.cmdExec != nil {
		var err error
		replicas, err = mgr.cmdExec.GetReplicas()
		if err != nil {
			return nil, err
		}
	}
	ramReplicas, err := mgr.ramDiskCmdExec.GetReplicas()
	if err != nil {
		return nil, err
	}

	for name := range ramReplicas {
		replicas[name] = ramReplicas[name]
	}

	return replicas, nil
}

func newLocalPoolManager(lm *LocalManager) LocalPoolManager {
	mgr := &localPoolManager{
		ramDiskCmdExec: newLocalPoolExecutor(lm, localstoragev1alpha1.VolumeKindRAM),
		lm:             lm,
		logger:         log.WithField("Module", "NodeManager/LocalPoolManager"),
	}
	if lm.nodeConf.LocalStorageConfig.VolumeKind == localstoragev1alpha1.VolumeKindDisk || lm.nodeConf.LocalStorageConfig.VolumeKind == localstoragev1alpha1.VolumeKindLVM {
		mgr.cmdExec = newLocalPoolExecutor(lm, lm.nodeConf.LocalStorageConfig.VolumeKind)
	}
	return mgr
}

func getPoolClassTypeByName(poolName string) (poolClass, poolType string) {
	switch strings.TrimSpace(poolName) {
	case localstoragev1alpha1.PoolNameForHDD:
		return localstoragev1alpha1.DiskClassNameHDD, localstoragev1alpha1.PoolTypeRegular
	case localstoragev1alpha1.PoolNameForSSD:
		return localstoragev1alpha1.DiskClassNameSSD, localstoragev1alpha1.PoolTypeRegular
	case localstoragev1alpha1.PoolNameForNVMe:
		return localstoragev1alpha1.DiskClassNameNVMe, localstoragev1alpha1.PoolTypeRegular
	case localstoragev1alpha1.PoolNameForRAM:
		return localstoragev1alpha1.DiskClassNameRAM, localstoragev1alpha1.PoolTypeRegular
	}
	return "", ""
}

func getPoolNameAccordingDisk(disk *localstoragev1alpha1.LocalDisk) (string, error) {
	switch disk.Class {
	case localstoragev1alpha1.DiskClassNameHDD:
		return localstoragev1alpha1.PoolNameForHDD, nil
	case localstoragev1alpha1.DiskClassNameSSD:
		return localstoragev1alpha1.PoolNameForSSD, nil
	case localstoragev1alpha1.DiskClassNameNVMe:
		return localstoragev1alpha1.PoolNameForNVMe, nil
	case localstoragev1alpha1.DiskClassNameRAM:
		return localstoragev1alpha1.PoolNameForRAM, nil
	}
	return "", fmt.Errorf("not supported pool type %s", disk.Class)
}
