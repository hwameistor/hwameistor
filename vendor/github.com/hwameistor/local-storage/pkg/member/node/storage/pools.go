package storage

import (
	"fmt"
	"strings"

	apisv1alpha1 "github.com/hwameistor/local-storage/pkg/apis/hwameistor/v1alpha1"
	log "github.com/sirupsen/logrus"
)

type localPoolManager struct {
	cmdExec LocalPoolExecutor
	logger  *log.Entry
	lm      *LocalManager
}

func (mgr *localPoolManager) ExtendPools(localDisks []*apisv1alpha1.LocalDisk) error {

	return mgr.cmdExec.ExtendPools(localDisks)
}

func (mgr *localPoolManager) ExtendPoolsInfo(localDisks map[string]*apisv1alpha1.LocalDisk) (map[string]*apisv1alpha1.LocalPool, error) {
	return mgr.cmdExec.ExtendPoolsInfo(localDisks)
}

func (mgr *localPoolManager) GetReplicas() (map[string]*apisv1alpha1.LocalVolumeReplica, error) {
	return mgr.cmdExec.GetReplicas()
}

func newLocalPoolManager(lm *LocalManager) LocalPoolManager {
	return &localPoolManager{
		cmdExec: newLVMExecutor(lm),
		lm:      lm,
		logger:  log.WithField("Module", "NodeManager/LocalPoolManager"),
	}
}

func getPoolClassTypeByName(poolName string) (poolClass, poolType string) {
	switch strings.TrimSpace(poolName) {
	case apisv1alpha1.PoolNameForHDD:
		return apisv1alpha1.DiskClassNameHDD, apisv1alpha1.PoolTypeRegular
	case apisv1alpha1.PoolNameForSSD:
		return apisv1alpha1.DiskClassNameSSD, apisv1alpha1.PoolTypeRegular
	case apisv1alpha1.PoolNameForNVMe:
		return apisv1alpha1.DiskClassNameNVMe, apisv1alpha1.PoolTypeRegular
	}
	return "", ""
}

func getPoolNameAccordingDisk(disk *apisv1alpha1.LocalDisk) (string, error) {
	switch disk.Class {
	case apisv1alpha1.DiskClassNameHDD:
		return apisv1alpha1.PoolNameForHDD, nil
	case apisv1alpha1.DiskClassNameSSD:
		return apisv1alpha1.PoolNameForSSD, nil
	case apisv1alpha1.DiskClassNameNVMe:
		return apisv1alpha1.PoolNameForNVMe, nil
	}
	return "", fmt.Errorf("not supported pool type %s", disk.Class)
}
