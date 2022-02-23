package storage

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"

	log "github.com/sirupsen/logrus"

	localstoragev1alpha1 "github.com/hwameistor/local-storage/pkg/apis/localstorage/v1alpha1"
	"github.com/hwameistor/local-storage/pkg/exechelper"
	"github.com/hwameistor/local-storage/pkg/exechelper/nsexecutor"
	"github.com/hwameistor/local-storage/pkg/member/node/healths"
)

const (
	ddParamBs    = "bs=1024"
	ddParamCount = "count=20"
)

type diskExecutor struct {
	lm      *LocalManager
	cmdExec exechelper.Executor
	logger  *log.Entry
}

var diskExecutorInstance *diskExecutor

func newDiskExecutor(lm *LocalManager) *diskExecutor {
	if diskExecutorInstance == nil {
		diskExecutorInstance = &diskExecutor{
			lm:      lm,
			cmdExec: nsexecutor.New(),
			logger:  log.WithField("Module", "NodeManager/DiskExecutor"),
		}
	}
	return diskExecutorInstance
}

func (rb *diskExecutor) CreateVolumeReplica(replica *localstoragev1alpha1.LocalVolumeReplica) (*localstoragev1alpha1.LocalVolumeReplica, error) {
	pool := rb.lm.registry.Pools()[replica.Spec.PoolName]
	for _, disk := range pool.Disks {
		if disk.State == localstoragev1alpha1.DiskStateAvailable {
			linkPath := rb.genLinkPath(replica.Spec.PoolName, replica.Spec.VolumeName)
			if len(linkPath) == 0 {
				return nil, fmt.Errorf("invalid link path. Pool name: %s, replica name: %s", replica.Spec.PoolName, replica.Spec.VolumeName)
			}

			if err := rb.hardLink(disk.DevPath, linkPath); err != nil {
				rb.logger.WithError(err).Error("Failed to exec replica create.")
				return nil, err
			}

			newReplica := replica.DeepCopy()
			newReplica.Status.StoragePath = disk.DevPath
			newReplica.Status.DevicePath = linkPath
			newReplica.Status.AllocatedCapacityBytes = disk.CapacityBytes
			newReplica.Status.Disks = []string{disk.DevPath}

			return newReplica, nil
		}
	}
	return nil, fmt.Errorf("not found disk")
}

func (rb *diskExecutor) DeleteVolumeReplica(replica *localstoragev1alpha1.LocalVolumeReplica) error {
	if err := rb.eraseThenRemoveDisk(replica.Status.DevicePath); err != nil {
		rb.logger.WithError(err).Error("Failed to erase or remove disk.")
		return err
	}
	return nil
}

func (rb *diskExecutor) ExpandVolumeReplica(replica *localstoragev1alpha1.LocalVolumeReplica, newCapacityBytes int64) (*localstoragev1alpha1.LocalVolumeReplica, error) {
	return nil, fmt.Errorf("not supported")
}

func (rb *diskExecutor) TestVolumeReplica(replica *localstoragev1alpha1.LocalVolumeReplica) (*localstoragev1alpha1.LocalVolumeReplica, error) {
	newReplica := replica.DeepCopy()
	isHealthy, err := healths.NewSmartCtl().IsDiskHealthy(replica.Status.StoragePath)
	if err != nil {
		return newReplica, err
	}

	if isHealthy {
		newReplica.Status.State = localstoragev1alpha1.VolumeReplicaStateReady
		newReplica.Status.Synced = true
	} else {
		newReplica.Status.State = localstoragev1alpha1.VolumeReplicaStateNotReady
		newReplica.Status.Synced = false
	}

	return newReplica, nil
}

func (rb *diskExecutor) ConsistencyCheck(crdReplicas map[string]*localstoragev1alpha1.LocalVolumeReplica) {

	rb.logger.Debug("Consistency Checking for disk volume ...")

	replicas, err := rb.GetReplicas()
	if err != nil {
		rb.logger.Error("Failed to collect volume replicas info from OS")
		return
	}

	for volName, crd := range crdReplicas {
		rb.logger.WithField("volume", volName).Debug("Checking VolumeReplica CRD")
		replica, exists := replicas[volName]
		if !exists {
			rb.logger.WithField("volume", volName).WithError(fmt.Errorf("not found on Host")).Warning("Volume replica consistency check failed")
			continue
		}
		if crd.Status.StoragePath != replica.Status.StoragePath {
			rb.logger.WithFields(log.Fields{
				"volume":   volName,
				"crd.path": crd.Status.StoragePath,
				"rep.path": replica.Status.StoragePath,
			}).WithError(fmt.Errorf("mismatched storage path")).Warning("Volume replica consistency check failed")
		}
		if crd.Status.DevicePath != replica.Status.DevicePath {
			rb.logger.WithFields(log.Fields{
				"volume":   volName,
				"crd.path": crd.Status.DevicePath,
				"rep.path": replica.Status.DevicePath,
			}).WithError(fmt.Errorf("mismatched device path")).Warning("Volume replica consistency check failed")
		}
		if crd.Status.AllocatedCapacityBytes != replica.Status.AllocatedCapacityBytes {
			rb.logger.WithFields(log.Fields{
				"volume":       volName,
				"crd.capacity": crd.Status.AllocatedCapacityBytes,
				"rep.capacity": replica.Status.AllocatedCapacityBytes,
			}).WithError(fmt.Errorf("mismatched allocated capacity")).Warning("Volume replica consistency check failed")
		}
	}

	for volName, replica := range replicas {
		rb.logger.WithField("volume", volName).Debug("Checking volume replica on Host")
		crd, exists := crdReplicas[volName]
		if !exists {
			rb.logger.WithField("volume", volName).WithError(fmt.Errorf("not found the CRD")).Warning("Volume replica consistency check failed")
			continue
		}
		if crd.Status.StoragePath != replica.Status.StoragePath {
			rb.logger.WithFields(log.Fields{
				"volume":   volName,
				"crd.path": crd.Status.StoragePath,
				"rep.path": replica.Status.StoragePath,
			}).WithError(fmt.Errorf("mismatched storage path")).Warning("Volume replica consistency check failed")
		}
		if crd.Status.DevicePath != replica.Status.DevicePath {
			rb.logger.WithFields(log.Fields{
				"volume":   volName,
				"crd.path": crd.Status.DevicePath,
				"rep.path": replica.Status.DevicePath,
			}).WithError(fmt.Errorf("mismatched device path")).Warning("Volume replica consistency check failed")
		}
		if crd.Status.AllocatedCapacityBytes != replica.Status.AllocatedCapacityBytes {
			rb.logger.WithFields(log.Fields{
				"volume":       volName,
				"crd.capacity": crd.Status.AllocatedCapacityBytes,
				"rep.capacity": replica.Status.AllocatedCapacityBytes,
			}).WithError(fmt.Errorf("mismatched allocated capacity")).Warning("Volume replica consistency check failed")
		}
	}

	rb.logger.Debug("Consistency check completed")
}

func (rb *diskExecutor) genLinkPath(poolName, volumeName string) string {
	switch poolName {
	case localstoragev1alpha1.PoolNameForHDD:
		return filepath.Join(localstoragev1alpha1.AssigedDiskPoolHDD, volumeName)
	case localstoragev1alpha1.PoolNameForSSD:
		return filepath.Join(localstoragev1alpha1.AssigedDiskPoolSSD, volumeName)
	case localstoragev1alpha1.PoolNameForNVMe:
		return filepath.Join(localstoragev1alpha1.AssigedDiskPoolNVMe, volumeName)
	}
	return ""
}

type localDiskWithLinkPath struct {
	LocalDisk *localstoragev1alpha1.LocalDisk
	LinkPath  string
}

type poolInfo struct {
	TotalCapacityBytes       int64
	UsedCapacityBytes        int64
	TotalVolumeCount         int64
	UsedVolumeCount          int64
	VolumeCapacityBytesLimit int64
	Disks                    []localstoragev1alpha1.LocalDisk
	Volumes                  []string
	NotEmpty                 bool
	DisksMap                 map[string]*localstoragev1alpha1.LocalDisk
	replicaMap               map[string]*localstoragev1alpha1.LocalVolumeReplica
}

func (rb *diskExecutor) ExtendPools(availableLocalDisks []*localstoragev1alpha1.LocalDisk) error {
	return nil
}

func (rb *diskExecutor) getPoolInfo(disks []*localDiskWithLinkPath) *poolInfo {
	poolInfoToReturn := &poolInfo{
		Disks:      make([]localstoragev1alpha1.LocalDisk, 0, len(disks)),
		Volumes:    make([]string, 0),
		DisksMap:   make(map[string]*localstoragev1alpha1.LocalDisk),
		replicaMap: make(map[string]*localstoragev1alpha1.LocalVolumeReplica),
	}
	if len(disks) == 0 {
		poolInfoToReturn.NotEmpty = false
		return poolInfoToReturn

	}
	poolInfoToReturn.DisksMap = make(map[string]*localstoragev1alpha1.LocalDisk)
	poolInfoToReturn.replicaMap = make(map[string]*localstoragev1alpha1.LocalVolumeReplica)
	poolInfoToReturn.NotEmpty = true
	poolInfoToReturn.TotalVolumeCount = int64(len(disks))

	for _, disk := range disks {
		poolInfoToReturn.TotalCapacityBytes += disk.LocalDisk.CapacityBytes
		poolInfoToReturn.Disks = append(poolInfoToReturn.Disks, *disk.LocalDisk)
		poolInfoToReturn.DisksMap[disk.LocalDisk.DevPath] = disk.LocalDisk
		if disk.LocalDisk.State == localstoragev1alpha1.DiskStateInUse {
			poolInfoToReturn.UsedCapacityBytes += disk.LocalDisk.CapacityBytes
			poolInfoToReturn.UsedVolumeCount++
			poolInfoToReturn.Volumes = append(poolInfoToReturn.Volumes, filepath.Base(disk.LinkPath))
		}
	}

	if len(poolInfoToReturn.Disks) > 0 {
		poolInfoToReturn.VolumeCapacityBytesLimit = poolInfoToReturn.Disks[0].CapacityBytes
	}

	return poolInfoToReturn
}

// getClassifiedPoolList
func (rb *diskExecutor) getClassifiedPoolList() (map[string][]*localDiskWithLinkPath, error) {

	localDisks := rb.lm.registry.Disks()

	allBlockDevices := rb.getDevicesInfo(localstoragev1alpha1.DiskDevRootPath)
	assigedHDDDevices := rb.getDevicesInfo(localstoragev1alpha1.AssigedDiskPoolHDD)
	assigedSSDDevices := rb.getDevicesInfo(localstoragev1alpha1.AssigedDiskPoolSSD)
	assigedNVMeDevices := rb.getDevicesInfo(localstoragev1alpha1.AssigedDiskPoolNVMe)

	classifiedLocalDisk := make(map[string][]*localDiskWithLinkPath)

	for majMin, blockDeviceInfo := range allBlockDevices {
		localDisk, diskExists := localDisks[blockDeviceInfo.Path]
		if !diskExists {
			continue
		}

		localDiskLn := &localDiskWithLinkPath{
			LocalDisk: localDisk,
		}

		if diskInfo, has := assigedHDDDevices[majMin]; has {
			localDiskLn.LinkPath = diskInfo.Path
			localDiskLn.LocalDisk.State = localstoragev1alpha1.DiskStateInUse
			classifiedLocalDisk[localstoragev1alpha1.PoolNameForHDD] = append(classifiedLocalDisk[localstoragev1alpha1.PoolNameForHDD], localDiskLn)
			continue
		} else if diskInfo, has := assigedSSDDevices[majMin]; has {
			localDiskLn.LinkPath = diskInfo.Path
			localDiskLn.LocalDisk.State = localstoragev1alpha1.DiskStateInUse
			classifiedLocalDisk[localstoragev1alpha1.PoolNameForSSD] = append(classifiedLocalDisk[localstoragev1alpha1.PoolNameForSSD], localDiskLn)
			continue
		} else if diskInfo, has := assigedNVMeDevices[majMin]; has {
			localDiskLn.LinkPath = diskInfo.Path
			localDiskLn.LocalDisk.State = localstoragev1alpha1.DiskStateInUse
			classifiedLocalDisk[localstoragev1alpha1.PoolNameForNVMe] = append(classifiedLocalDisk[localstoragev1alpha1.PoolNameForNVMe], localDiskLn)
		} else if localDisks[blockDeviceInfo.Path].State == localstoragev1alpha1.DiskStateAvailable {
			// Available disks
			switch localDisk.Class {
			case localstoragev1alpha1.DiskClassNameHDD:
				classifiedLocalDisk[localstoragev1alpha1.PoolNameForHDD] = append(classifiedLocalDisk[localstoragev1alpha1.PoolNameForHDD], localDiskLn)
			case localstoragev1alpha1.DiskClassNameSSD:
				classifiedLocalDisk[localstoragev1alpha1.PoolNameForSSD] = append(classifiedLocalDisk[localstoragev1alpha1.PoolNameForSSD], localDiskLn)
			case localstoragev1alpha1.DiskClassNameNVMe:
				classifiedLocalDisk[localstoragev1alpha1.PoolNameForNVMe] = append(classifiedLocalDisk[localstoragev1alpha1.PoolNameForNVMe], localDiskLn)
			}
		}
	}

	return classifiedLocalDisk, nil
}

func (rb *diskExecutor) mergeRegistryDiskMap(localDiskMap ...map[string]*localstoragev1alpha1.LocalDisk) map[string]*localstoragev1alpha1.LocalDisk {
	newLocalDiskMap := map[string]*localstoragev1alpha1.LocalDisk{}
	for _, m := range localDiskMap {
		for k, v := range m {
			newLocalDiskMap[k] = v
		}
	}
	return newLocalDiskMap
}

func (rb *diskExecutor) extendClassifiedPoolList(disks map[string]*localstoragev1alpha1.LocalDisk) (map[string][]*localDiskWithLinkPath, error) {

	oldRegistryDisks := rb.lm.registry.Disks()
	localDisks := mergeRegistryDiskMap(oldRegistryDisks, disks)

	allBlockDevices := rb.getDevicesInfo(localstoragev1alpha1.DiskDevRootPath)
	assigedHDDDevices := rb.getDevicesInfo(localstoragev1alpha1.AssigedDiskPoolHDD)
	assigedSSDDevices := rb.getDevicesInfo(localstoragev1alpha1.AssigedDiskPoolSSD)
	assigedNVMeDevices := rb.getDevicesInfo(localstoragev1alpha1.AssigedDiskPoolNVMe)

	classifiedLocalDisk := make(map[string][]*localDiskWithLinkPath)

	for majMin, blockDeviceInfo := range allBlockDevices {
		localDisk, diskExists := localDisks[blockDeviceInfo.Path]
		if !diskExists {
			continue
		}

		localDiskLn := &localDiskWithLinkPath{
			LocalDisk: localDisk,
		}

		if diskInfo, has := assigedHDDDevices[majMin]; has {
			localDiskLn.LinkPath = diskInfo.Path
			localDiskLn.LocalDisk.State = localstoragev1alpha1.DiskStateInUse
			classifiedLocalDisk[localstoragev1alpha1.PoolNameForHDD] = append(classifiedLocalDisk[localstoragev1alpha1.PoolNameForHDD], localDiskLn)
			continue
		} else if diskInfo, has := assigedSSDDevices[majMin]; has {
			localDiskLn.LinkPath = diskInfo.Path
			localDiskLn.LocalDisk.State = localstoragev1alpha1.DiskStateInUse
			classifiedLocalDisk[localstoragev1alpha1.PoolNameForSSD] = append(classifiedLocalDisk[localstoragev1alpha1.PoolNameForSSD], localDiskLn)
			continue
		} else if diskInfo, has := assigedNVMeDevices[majMin]; has {
			localDiskLn.LinkPath = diskInfo.Path
			localDiskLn.LocalDisk.State = localstoragev1alpha1.DiskStateInUse
			classifiedLocalDisk[localstoragev1alpha1.PoolNameForNVMe] = append(classifiedLocalDisk[localstoragev1alpha1.PoolNameForNVMe], localDiskLn)
		} else if localDisks[blockDeviceInfo.Path].State == localstoragev1alpha1.DiskStateAvailable {
			// Available disks
			switch localDisk.Class {
			case localstoragev1alpha1.DiskClassNameHDD:
				classifiedLocalDisk[localstoragev1alpha1.PoolNameForHDD] = append(classifiedLocalDisk[localstoragev1alpha1.PoolNameForHDD], localDiskLn)
			case localstoragev1alpha1.DiskClassNameSSD:
				classifiedLocalDisk[localstoragev1alpha1.PoolNameForSSD] = append(classifiedLocalDisk[localstoragev1alpha1.PoolNameForSSD], localDiskLn)
			case localstoragev1alpha1.DiskClassNameNVMe:
				classifiedLocalDisk[localstoragev1alpha1.PoolNameForNVMe] = append(classifiedLocalDisk[localstoragev1alpha1.PoolNameForNVMe], localDiskLn)
			}
		}
	}

	return classifiedLocalDisk, nil
}

func (rb *diskExecutor) ExtendPoolsInfo(localDisks map[string]*localstoragev1alpha1.LocalDisk) (map[string]*localstoragev1alpha1.LocalPool, error) {
	classifiedDisks, err := rb.extendClassifiedPoolList(localDisks)
	if err != nil {
		return nil, err
	}

	pools := make(map[string]*localstoragev1alpha1.LocalPool)

	// HDD pool rebuild
	poolInfo := rb.getPoolInfo(classifiedDisks[localstoragev1alpha1.PoolNameForHDD])
	if poolInfo.NotEmpty {
		poolHDD := &localstoragev1alpha1.LocalPool{
			Name:                     localstoragev1alpha1.PoolNameForHDD,
			Class:                    localstoragev1alpha1.DiskClassNameHDD,
			Type:                     localstoragev1alpha1.PoolTypeRegular,
			VolumeKind:               localstoragev1alpha1.VolumeKindDisk,
			Path:                     localstoragev1alpha1.AssigedDiskPoolHDD,
			TotalCapacityBytes:       poolInfo.TotalCapacityBytes,
			UsedCapacityBytes:        poolInfo.UsedCapacityBytes,
			FreeCapacityBytes:        poolInfo.TotalCapacityBytes - poolInfo.UsedCapacityBytes,
			TotalVolumeCount:         poolInfo.TotalVolumeCount,
			UsedVolumeCount:          poolInfo.UsedVolumeCount,
			FreeVolumeCount:          poolInfo.TotalVolumeCount - poolInfo.UsedVolumeCount,
			VolumeCapacityBytesLimit: poolInfo.VolumeCapacityBytesLimit,
			Disks:                    poolInfo.Disks,
			Volumes:                  poolInfo.Volumes,
		}
		pools[localstoragev1alpha1.PoolNameForHDD] = poolHDD
	}

	// SSD pool rebuild
	poolInfo = rb.getPoolInfo(classifiedDisks[localstoragev1alpha1.PoolNameForSSD])
	if poolInfo.NotEmpty {
		poolSSD := &localstoragev1alpha1.LocalPool{
			Name:                     localstoragev1alpha1.PoolNameForSSD,
			Class:                    localstoragev1alpha1.DiskClassNameSSD,
			Type:                     localstoragev1alpha1.PoolTypeRegular,
			VolumeKind:               localstoragev1alpha1.VolumeKindDisk,
			Path:                     localstoragev1alpha1.AssigedDiskPoolSSD,
			TotalCapacityBytes:       poolInfo.TotalCapacityBytes,
			UsedCapacityBytes:        poolInfo.UsedCapacityBytes,
			FreeCapacityBytes:        poolInfo.TotalCapacityBytes - poolInfo.UsedCapacityBytes,
			TotalVolumeCount:         poolInfo.TotalVolumeCount,
			UsedVolumeCount:          poolInfo.UsedVolumeCount,
			FreeVolumeCount:          poolInfo.TotalVolumeCount - poolInfo.UsedVolumeCount,
			VolumeCapacityBytesLimit: poolInfo.VolumeCapacityBytesLimit,
			Disks:                    poolInfo.Disks,
			Volumes:                  poolInfo.Volumes,
		}
		pools[localstoragev1alpha1.PoolNameForSSD] = poolSSD
	}

	// NVMe pool rebuild
	poolInfo = rb.getPoolInfo(classifiedDisks[localstoragev1alpha1.PoolNameForNVMe])
	if poolInfo.NotEmpty {
		poolNVMe := &localstoragev1alpha1.LocalPool{
			Name:                     localstoragev1alpha1.PoolNameForNVMe,
			Class:                    localstoragev1alpha1.DiskClassNameNVMe,
			Type:                     localstoragev1alpha1.PoolTypeRegular,
			VolumeKind:               localstoragev1alpha1.VolumeKindDisk,
			Path:                     localstoragev1alpha1.AssigedDiskPoolNVMe,
			TotalCapacityBytes:       poolInfo.TotalCapacityBytes,
			UsedCapacityBytes:        poolInfo.UsedCapacityBytes,
			FreeCapacityBytes:        poolInfo.TotalCapacityBytes - poolInfo.UsedCapacityBytes,
			TotalVolumeCount:         poolInfo.TotalVolumeCount,
			UsedVolumeCount:          poolInfo.UsedVolumeCount,
			FreeVolumeCount:          poolInfo.TotalVolumeCount - poolInfo.UsedVolumeCount,
			VolumeCapacityBytesLimit: poolInfo.VolumeCapacityBytesLimit,
			Disks:                    poolInfo.Disks,
			Volumes:                  poolInfo.Volumes,
		}
		pools[localstoragev1alpha1.PoolNameForNVMe] = poolNVMe
	}

	return pools, nil
}

func (rb *diskExecutor) GetReplicas() (map[string]*localstoragev1alpha1.LocalVolumeReplica, error) {
	classifiedDisks, err := rb.getClassifiedPoolList()
	if err != nil {
		return nil, err
	}

	replicas := make(map[string]*localstoragev1alpha1.LocalVolumeReplica)

	for _, diskWithLinkPath := range classifiedDisks[localstoragev1alpha1.PoolNameForHDD] {
		if diskWithLinkPath.LinkPath == "" {
			continue
		}
		volumeName := filepath.Base(diskWithLinkPath.LinkPath)
		replicaToTest := &localstoragev1alpha1.LocalVolumeReplica{
			Spec: localstoragev1alpha1.LocalVolumeReplicaSpec{
				VolumeName: volumeName,
				PoolName:   localstoragev1alpha1.PoolNameForHDD,
			},
			Status: localstoragev1alpha1.LocalVolumeReplicaStatus{
				StoragePath:            diskWithLinkPath.LocalDisk.DevPath,
				DevicePath:             diskWithLinkPath.LinkPath,
				AllocatedCapacityBytes: diskWithLinkPath.LocalDisk.CapacityBytes,
				Disks:                  []string{diskWithLinkPath.LocalDisk.DevPath},
			},
		}
		replica, _ := rb.TestVolumeReplica(replicaToTest)
		replicas[volumeName] = replica
	}
	for _, diskWithLinkPath := range classifiedDisks[localstoragev1alpha1.PoolNameForSSD] {
		if diskWithLinkPath.LinkPath == "" {
			continue
		}
		volumeName := filepath.Base(diskWithLinkPath.LinkPath)
		replicaToTest := &localstoragev1alpha1.LocalVolumeReplica{
			Spec: localstoragev1alpha1.LocalVolumeReplicaSpec{
				VolumeName: volumeName,
				PoolName:   localstoragev1alpha1.PoolNameForSSD,
			},
			Status: localstoragev1alpha1.LocalVolumeReplicaStatus{
				StoragePath:            diskWithLinkPath.LocalDisk.DevPath,
				DevicePath:             diskWithLinkPath.LinkPath,
				AllocatedCapacityBytes: diskWithLinkPath.LocalDisk.CapacityBytes,
				Disks:                  []string{diskWithLinkPath.LocalDisk.DevPath},
			},
		}
		replica, _ := rb.TestVolumeReplica(replicaToTest)
		replicas[volumeName] = replica
	}
	for _, diskWithLinkPath := range classifiedDisks[localstoragev1alpha1.PoolNameForNVMe] {
		if diskWithLinkPath.LinkPath == "" {
			continue
		}
		volumeName := filepath.Base(diskWithLinkPath.LinkPath)
		replicaToTest := &localstoragev1alpha1.LocalVolumeReplica{
			Spec: localstoragev1alpha1.LocalVolumeReplicaSpec{
				VolumeName: volumeName,
				PoolName:   localstoragev1alpha1.PoolNameForNVMe,
			},
			Status: localstoragev1alpha1.LocalVolumeReplicaStatus{
				StoragePath:            diskWithLinkPath.LocalDisk.DevPath,
				DevicePath:             diskWithLinkPath.LinkPath,
				AllocatedCapacityBytes: diskWithLinkPath.LocalDisk.CapacityBytes,
				Disks:                  []string{diskWithLinkPath.LocalDisk.DevPath},
			},
		}
		replica, _ := rb.TestVolumeReplica(replicaToTest)
		replicas[volumeName] = replica
	}

	rb.logger.WithField("volumes", len(replicas)).Debug("Finshed disk volume detection")
	return replicas, nil
}

// ======== Helpers ==========
func (rb *diskExecutor) hardLink(sourcePath, targetPath string) error {
	file, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	file.Close()

	file, err = os.Open(filepath.Dir(targetPath))
	if err != nil {
		if os.IsNotExist(err) {
			rb.mkdir(filepath.Dir(targetPath))
		} else {
			return err
		}
	}
	file.Close()

	params := exechelper.ExecParams{
		CmdName: "ln",
		CmdArgs: []string{sourcePath, targetPath},
	}
	res := rb.cmdExec.RunCommand(params)
	if res.ExitCode == 0 {
		return nil
	}
	return res.Error

}

func (rb *diskExecutor) mkdir(targetPath string) error {
	params := exechelper.ExecParams{
		CmdName: "mkdir",
		CmdArgs: []string{"-p", targetPath},
	}
	res := rb.cmdExec.RunCommand(params)
	if res.ExitCode == 0 {
		return nil
	}
	return res.Error

}

func (rb *diskExecutor) eraseThenRemoveDisk(path string) error {
	params := exechelper.ExecParams{
		CmdName: "dd",
		CmdArgs: []string{"if=/dev/zero", fmt.Sprintf("of=%s", path), ddParamBs, ddParamCount},
	}
	res := rb.cmdExec.RunCommand(params)
	if res.ExitCode != 0 {
		rb.logger.WithError(res.Error).Error("Failed to erase disk")
		return res.Error
	}

	params = exechelper.ExecParams{
		CmdName: "rm",
		CmdArgs: []string{"-f", path},
	}
	res = rb.cmdExec.RunCommand(params)
	if res.ExitCode != 0 {
		rb.logger.WithError(res.Error).Error("Failed to remove disk")
		return res.Error
	}

	return nil
}

// DeviceInfo struct
type DeviceInfo struct {
	OSFileInfo   os.FileInfo
	SysTStat     *syscall.Stat_t
	Path         string
	Name         string
	Major        uint32
	Minor        uint32
	MajMinString string
}

// GetDevicesInfo list devices info
func (rb *diskExecutor) getDevicesInfo(root string) map[string]*DeviceInfo {
	deviceInfoList := make(map[string]*DeviceInfo)

	contents, err := ioutil.ReadDir(root)
	if err != nil {
		rb.logger.WithField("dir", root).WithError(err).Error("Failed to read directory")
		return deviceInfoList
	}

	for i, f := range contents {
		if f.Mode()&os.ModeDevice != 0 {
			devpath := filepath.Join(root, f.Name())
			stat := &syscall.Stat_t{}
			err = syscall.Stat(devpath, stat)
			if err != nil {
				rb.logger.Errorf("Syscall stat() error: %s", err.Error())
				return nil
			}
			major, minor := rb.getMajorMinor(stat)
			deviceInfo := &DeviceInfo{
				OSFileInfo:   contents[i],
				SysTStat:     stat,
				Path:         devpath,
				Name:         f.Name(),
				Major:        major,
				Minor:        minor,
				MajMinString: rb.combineMajorMinor(major, minor),
			}
			deviceInfoList[deviceInfo.MajMinString] = deviceInfo
		}
	}

	return deviceInfoList
}

func (rb *diskExecutor) combineMajorMinor(maj, min uint32) string {
	return strconv.Itoa(int(maj)) + ":" + strconv.Itoa(int(min))
}

func (rb *diskExecutor) getMajorMinor(stat *syscall.Stat_t) (uint32, uint32) {
	dev := uint64(stat.Rdev)
	var major, minor uint32
	switch runtime.GOOS {
	case "aix":
		major = uint32((dev & 0x3fffffff00000000) >> 32)
		minor = uint32((dev & 0x00000000ffffffff) >> 0)
	case "linux":
		major = uint32((dev & 0x00000000000fff00) >> 8)
		major |= uint32((dev & 0xfffff00000000000) >> 32)
		minor = uint32((dev & 0x00000000000000ff) >> 0)
		minor |= uint32((dev & 0x00000ffffff00000) >> 12)
	case "darwin":
		major = uint32((dev >> 24) & 0xff)
		minor = uint32(dev & 0xffffff)
	case "dragonfly":
		major = uint32((dev >> 8) & 0xff)
		minor = uint32(dev & 0xffff00ff)
	case "freebsd":
		major = uint32((dev >> 8) & 0xff)
		minor = uint32(dev & 0xffff00ff)
	case "netbsd":
		major = uint32((dev & 0x000fff00) >> 8)
		minor = uint32((dev & 0x000000ff) >> 0)
		minor |= uint32((dev & 0xfff00000) >> 12)
	case "openbsd":
		major = uint32((dev & 0x0000ff00) >> 8)
		minor = uint32((dev & 0x000000ff) >> 0)
		minor |= uint32((dev & 0xffff0000) >> 8)
	}

	return major, minor
}
