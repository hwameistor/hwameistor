package storage

import (
	"fmt"
	"strings"

	udsv1alpha1 "github.com/HwameiStor/local-storage/pkg/apis/uds/v1alpha1"
	"github.com/HwameiStor/local-storage/pkg/utils"
	log "github.com/sirupsen/logrus"
)

type localPoolManager struct {
	cmdExec        LocalPoolExecutor
	ramDiskCmdExec LocalPoolExecutor
	logger         *log.Entry
	lm             *LocalManager
}

func (mgr *localPoolManager) ExtendPools(localDisks []*udsv1alpha1.LocalDisk) error {
	if mgr.cmdExec == nil {
		return nil
	}

	return mgr.cmdExec.ExtendPools(localDisks)
}

func (mgr *localPoolManager) ExtendPoolsInfo(localDisks map[string]*udsv1alpha1.LocalDisk) (map[string]*udsv1alpha1.LocalPool, error) {
	pools := map[string]*udsv1alpha1.LocalPool{}

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
	ramPool, has := ramPools[udsv1alpha1.PoolNameForRAM]
	if !has {
		return nil, fmt.Errorf("wrong ramdisk pool")
	}
	if mgr.lm.nodeConf.LocalStorageConfig != nil {
		totalCapacityBytes, _ := utils.ParseBytes(mgr.lm.nodeConf.LocalStorageConfig.RAMDiskTotalCapacity)
		ramPool.TotalCapacityBytes = totalCapacityBytes
		ramPool.FreeCapacityBytes = totalCapacityBytes - ramPool.UsedCapacityBytes
		ramPool.VolumeCapacityBytesLimit = totalCapacityBytes
	}
	pools[udsv1alpha1.PoolNameForRAM] = ramPool

	return pools, nil
}

func (mgr *localPoolManager) GetReplicas() (map[string]*udsv1alpha1.LocalVolumeReplica, error) {
	replicas := map[string]*udsv1alpha1.LocalVolumeReplica{}

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
		ramDiskCmdExec: newLocalPoolExecutor(lm, udsv1alpha1.VolumeKindRAM),
		lm:             lm,
		logger:         log.WithField("Module", "NodeManager/LocalPoolManager"),
	}
	if lm.nodeConf.LocalStorageConfig.VolumeKind == udsv1alpha1.VolumeKindDisk || lm.nodeConf.LocalStorageConfig.VolumeKind == udsv1alpha1.VolumeKindLVM {
		mgr.cmdExec = newLocalPoolExecutor(lm, lm.nodeConf.LocalStorageConfig.VolumeKind)
	}
	return mgr
}

func getPoolClassTypeByName(poolName string) (poolClass, poolType string) {
	switch strings.TrimSpace(poolName) {
	case udsv1alpha1.PoolNameForHDD:
		return udsv1alpha1.DiskClassNameHDD, udsv1alpha1.PoolTypeRegular
	case udsv1alpha1.PoolNameForSSD:
		return udsv1alpha1.DiskClassNameSSD, udsv1alpha1.PoolTypeRegular
	case udsv1alpha1.PoolNameForNVMe:
		return udsv1alpha1.DiskClassNameNVMe, udsv1alpha1.PoolTypeRegular
	case udsv1alpha1.PoolNameForRAM:
		return udsv1alpha1.DiskClassNameRAM, udsv1alpha1.PoolTypeRegular
	}
	return "", ""
}

func getPoolNameAccordingDisk(disk *udsv1alpha1.LocalDisk) (string, error) {
	switch disk.Class {
	case udsv1alpha1.DiskClassNameHDD:
		return udsv1alpha1.PoolNameForHDD, nil
	case udsv1alpha1.DiskClassNameSSD:
		return udsv1alpha1.PoolNameForSSD, nil
	case udsv1alpha1.DiskClassNameNVMe:
		return udsv1alpha1.PoolNameForNVMe, nil
	case udsv1alpha1.DiskClassNameRAM:
		return udsv1alpha1.PoolNameForRAM, nil
	}
	return "", fmt.Errorf("not supported pool type %s", disk.Class)
}
