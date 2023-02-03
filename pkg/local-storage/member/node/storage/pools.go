package storage

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

type localPoolManager struct {
	cmdExec LocalPoolExecutor
	logger  *log.Entry
	lm      *LocalManager
}

func (mgr *localPoolManager) ExtendPools(localDisks []*apisv1alpha1.LocalDevice) (bool, error) {

	return mgr.cmdExec.ExtendPools(localDisks)
}

func (mgr *localPoolManager) GetPools() (map[string]*apisv1alpha1.LocalPool, error) {
	return mgr.cmdExec.GetPools()
}

func (mgr *localPoolManager) GetReplicas() (map[string]*apisv1alpha1.LocalVolumeReplica, error) {
	return mgr.cmdExec.GetReplicas()
}

func (mgr *localPoolManager) ResizePhysicalVolumes(localDisks map[string]*apisv1alpha1.LocalDevice) error {
	return mgr.cmdExec.ResizePhysicalVolumes(localDisks)
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

func getPoolNameAccordingDisk(disk *apisv1alpha1.LocalDevice) (string, error) {
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
