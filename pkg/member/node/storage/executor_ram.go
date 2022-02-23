package storage

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	localstoragev1alpha1 "github.com/hwameistor/local-storage/pkg/apis/localstorage/v1alpha1"
	"github.com/hwameistor/local-storage/pkg/exechelper"
	"github.com/hwameistor/local-storage/pkg/exechelper/nsexecutor"
	"github.com/hwameistor/local-storage/pkg/utils"
)

const (
	ramdiskPrefix = "ramdisk"
)

type ramdiskExecutor struct {
	lm       *LocalManager
	cmdExec  exechelper.Executor
	logger   *log.Entry
	poolName string
	poolPath string
}

func (rd *ramdiskExecutor) ExtendPoolsInfo(localDisks map[string]*localstoragev1alpha1.LocalDisk) (map[string]*localstoragev1alpha1.LocalPool, error) {
	rd.logger.Debug("ConstructPools ...")
	pool := &localstoragev1alpha1.LocalPool{
		Name:                     rd.poolName,
		Class:                    localstoragev1alpha1.DiskClassNameRAM,
		Type:                     localstoragev1alpha1.PoolTypeRegular,
		VolumeKind:               localstoragev1alpha1.VolumeKindRAM,
		Path:                     rd.poolPath,
		TotalVolumeCount:         localstoragev1alpha1.RAMVolumeMaxCount,
		FreeVolumeCount:          localstoragev1alpha1.RAMVolumeMaxCount,
		UsedVolumeCount:          0,
		TotalCapacityBytes:       0, // should be changed according to the configuration
		FreeCapacityBytes:        0, // should be changed according to the configuration
		UsedCapacityBytes:        0,
		VolumeCapacityBytesLimit: 0, // should be changed according to the configuration
		Disks:                    []localstoragev1alpha1.LocalDisk{},
		Volumes:                  []string{},
	}

	replicas, err := rd.GetReplicas()
	if err != nil {
		return nil, err
	}

	for volName, replica := range replicas {
		pool.UsedCapacityBytes += replica.Status.AllocatedCapacityBytes
		pool.Volumes = append(pool.Volumes, volName)
		pool.UsedVolumeCount++
		pool.FreeVolumeCount--
	}

	return map[string]*localstoragev1alpha1.LocalPool{localstoragev1alpha1.PoolNameForRAM: pool}, nil
}

func (rd *ramdiskExecutor) GetReplicas() (map[string]*localstoragev1alpha1.LocalVolumeReplica, error) {
	rd.logger.Debug("ConstructReplicas Probing RAM disk volume replica on Host ...")
	if err := rd.mkdir(rd.poolPath); err != nil {
		rd.logger.WithField("pool", rd.poolPath).WithError(err).Error("Can't determine the pool directory")
		return nil, err
	}
	replicas := make(map[string]*localstoragev1alpha1.LocalVolumeReplica)
	for volName := range rd.getAllDevPaths() {
		replicaToTest := &localstoragev1alpha1.LocalVolumeReplica{}
		replicaToTest.Spec.PoolName = localstoragev1alpha1.PoolNameForRAM
		replicaToTest.Spec.VolumeName = volName
		replicaToTest.Spec.Kind = localstoragev1alpha1.VolumeKindRAM
		replicaToTest.Spec.NodeName = rd.lm.nodeConf.Name
		replica, err := rd.TestVolumeReplica(replicaToTest)
		if err != nil {
			continue
		}
		replicas[volName] = replica
	}

	return replicas, nil
}

var ramdiskExecutorInstance *ramdiskExecutor

func newRAMDiskExecutor(lm *LocalManager) *ramdiskExecutor {
	if ramdiskExecutorInstance == nil {
		ramdiskExecutorInstance = &ramdiskExecutor{
			lm:       lm,
			poolName: localstoragev1alpha1.PoolNameForRAM,
			poolPath: fmt.Sprintf("/dev/%s", localstoragev1alpha1.PoolNameForRAM),
			cmdExec:  nsexecutor.New(),
			logger:   log.WithField("Module", "NodeManager/RamDiskExecuter"),
		}
	}
	return ramdiskExecutorInstance
}

func (rd *ramdiskExecutor) CreateVolumeReplica(replica *localstoragev1alpha1.LocalVolumeReplica) (*localstoragev1alpha1.LocalVolumeReplica, error) {

	devPath := rd.getDevPath(replica.Spec.VolumeName)
	if err := rd.mkdir(devPath); err != nil {
		return nil, err
	}
	if err := rd.mount(ramdiskMountSourceName(replica.Spec.VolumeName), devPath, []string{"-t", "tmpfs", "-o", fmt.Sprintf("size=%d", replica.Spec.RequiredCapacityBytes)}); err != nil {
		return nil, err
	}
	return rd.TestVolumeReplica(replica)
}

func (rd *ramdiskExecutor) DeleteVolumeReplica(replica *localstoragev1alpha1.LocalVolumeReplica) error {
	devPath := rd.getDevPath(replica.Spec.VolumeName)

	if err := rd.umount(devPath); err != nil {
		return err
	}
	if err := rd.rmdir(devPath); err != nil {
		return err
	}
	return nil
}

func (rd *ramdiskExecutor) ExpandVolumeReplica(replica *localstoragev1alpha1.LocalVolumeReplica, newCapacityBytes int64) (*localstoragev1alpha1.LocalVolumeReplica, error) {
	return nil, fmt.Errorf("not supported")
}

func (rd *ramdiskExecutor) TestVolumeReplica(replica *localstoragev1alpha1.LocalVolumeReplica) (*localstoragev1alpha1.LocalVolumeReplica, error) {
	devPath := rd.getDevPath(replica.Spec.VolumeName)
	result := rd.cmdExec.RunCommand(exechelper.ExecParams{
		CmdName: "findmnt",
		CmdArgs: []string{"-n", "--output=source,size", "--target", devPath},
	})
	if result.ExitCode != 0 {
		return nil, result.Error
	}

	items := strings.Fields(result.OutBuf.String())
	if len(items) < 2 {
		return nil, fmt.Errorf("invalid output")
	}
	if !strings.HasPrefix(items[0], ramdiskPrefix) {
		return nil, fmt.Errorf("wrong volume type")
	}

	allocatedCapacityBytes, err := utils.ParseBytes(items[1])
	if err != nil {
		return nil, err
	}

	newReplica := replica.DeepCopy()
	newReplica.Status.State = localstoragev1alpha1.VolumeReplicaStateReady
	newReplica.Status.AllocatedCapacityBytes = allocatedCapacityBytes
	newReplica.Status.DevicePath = devPath
	newReplica.Status.StoragePath = ramdiskPrefix
	newReplica.Status.Synced = true
	return newReplica, nil
}

func (rd *ramdiskExecutor) ExtendPools(availableLocalDisks []*localstoragev1alpha1.LocalDisk) error {
	return nil
}

func (rd *ramdiskExecutor) ConsistencyCheck(crdReplicas map[string]*localstoragev1alpha1.LocalVolumeReplica) {

	rd.logger.Debug("Consistency Checking for RAM disk volume ...")

	replicas, err := rd.GetReplicas()
	if err != nil {
		rd.logger.Error("Failed to collect volume replicas info from OS")
		return
	}

	for volName, crd := range crdReplicas {
		rd.logger.WithField("volume", volName).Debug("Checking VolumeReplica CRD")
		replica, exists := replicas[volName]
		if !exists {
			if crd.Status.State != localstoragev1alpha1.VolumeReplicaStateReady {
				continue
			}
			rd.logger.WithField("volume", volName).WithError(fmt.Errorf("not found on Host")).Warning("Volume replica consistency check failed")
			// for ramdisk, it should be remounted after node reboot
			replica, err = rd.CreateVolumeReplica(crd)
			if err != nil {
				rd.logger.WithField("volume", volName).WithError(err).Error("Failed to recreate RAM disk volume replica")
			} else {
				rd.logger.WithField("volume", volName).Debug("Recreated a RAM disk volume replica")
			}
		}
		if crd.Status.StoragePath != replica.Status.StoragePath {
			rd.logger.WithFields(log.Fields{
				"volume":   volName,
				"crd.path": crd.Status.StoragePath,
				"rep.path": replica.Status.StoragePath,
			}).WithError(fmt.Errorf("mismatched storage path")).Warning("Volume replica consistency check failed")
		}
		if crd.Status.DevicePath != replica.Status.DevicePath {
			rd.logger.WithFields(log.Fields{
				"volume":   volName,
				"crd.path": crd.Status.DevicePath,
				"rep.path": replica.Status.DevicePath,
			}).WithError(fmt.Errorf("mismatched device path")).Warning("Volume replica consistency check failed")
		}
		if crd.Status.AllocatedCapacityBytes != replica.Status.AllocatedCapacityBytes {
			rd.logger.WithFields(log.Fields{
				"volume":       volName,
				"crd.capacity": crd.Status.AllocatedCapacityBytes,
				"rep.capacity": replica.Status.AllocatedCapacityBytes,
			}).WithError(fmt.Errorf("mismatched allocated capacity")).Warning("Volume replica consistency check failed")
		}
	}

	for volName, replica := range replicas {
		rd.logger.WithField("volume", volName).Debug("Checking volume replica on Host")
		crd, exists := crdReplicas[volName]
		if !exists {
			rd.logger.WithField("volume", volName).WithError(fmt.Errorf("not found the CRD")).Warning("Volume replica consistency check failed")
			continue
		}
		if crd.Status.StoragePath != replica.Status.StoragePath {
			rd.logger.WithFields(log.Fields{
				"volume":   volName,
				"crd.path": crd.Status.StoragePath,
				"rep.path": replica.Status.StoragePath,
			}).WithError(fmt.Errorf("mismatched storage path")).Warning("Volume replica consistency check failed")
		}
		if crd.Status.DevicePath != replica.Status.DevicePath {
			rd.logger.WithFields(log.Fields{
				"volume":   volName,
				"crd.path": crd.Status.DevicePath,
				"rep.path": replica.Status.DevicePath,
			}).WithError(fmt.Errorf("mismatched device path")).Warning("Volume replica consistency check failed")
		}
		if crd.Status.AllocatedCapacityBytes != replica.Status.AllocatedCapacityBytes {
			rd.logger.WithFields(log.Fields{
				"volume":       volName,
				"crd.capacity": crd.Status.AllocatedCapacityBytes,
				"rep.capacity": replica.Status.AllocatedCapacityBytes,
			}).WithError(fmt.Errorf("mismatched allocated capacity")).Warning("Volume replica consistency check failed")
		}
	}

	rd.logger.Debug("Consistency check completed")
}

// ================    Helpers   ====================
func ramdiskMountSourceName(volName string) string {
	return fmt.Sprintf("%s-%s", ramdiskPrefix, volName)
}

func (rd *ramdiskExecutor) mount(source string, target string, options []string) error {

	if rd.isMounted(target) {
		return nil
	}

	params := exechelper.ExecParams{
		CmdName: "mount",
		CmdArgs: append(options, source, target),
	}
	res := rd.cmdExec.RunCommand(params)
	if res.ExitCode == 0 {
		return nil
	}
	return res.Error
}

func (rd *ramdiskExecutor) umount(mountpoint string) error {

	if !rd.isMounted(mountpoint) {
		return nil
	}

	params := exechelper.ExecParams{
		CmdName: "umount",
		CmdArgs: []string{mountpoint},
	}
	res := rd.cmdExec.RunCommand(params)
	if res.ExitCode == 0 {
		return nil
	}
	return res.Error
}

func (rd *ramdiskExecutor) isMounted(mountpoint string) bool {
	result := rd.cmdExec.RunCommand(exechelper.ExecParams{
		CmdName: "findmnt",
		CmdArgs: []string{"-n", "--output=source", "--target", mountpoint},
	})
	if result.ExitCode != 0 {
		return false
	}
	return strings.HasPrefix(result.OutBuf.String(), ramdiskPrefix)
}

func (rd *ramdiskExecutor) mkdir(path string) error {
	if rd.dirExists(path) {
		return nil
	}
	params := exechelper.ExecParams{
		CmdName: "mkdir",
		CmdArgs: []string{"-p", path},
	}
	res := rd.cmdExec.RunCommand(params)
	if res.ExitCode == 0 {
		return nil
	}
	return res.Error
}

func (rd *ramdiskExecutor) rmdir(path string) error {
	if !rd.dirExists(path) {
		return nil
	}
	params := exechelper.ExecParams{
		CmdName: "rm",
		CmdArgs: []string{"-rf", path},
	}
	res := rd.cmdExec.RunCommand(params)
	if res.ExitCode == 0 {
		return nil
	}
	return res.Error
}

func (rd *ramdiskExecutor) dirExists(path string) bool {
	params := exechelper.ExecParams{
		CmdName: "test",
		CmdArgs: []string{"-d", path},
	}
	res := rd.cmdExec.RunCommand(params)
	return res.ExitCode == 0
}

func (rd *ramdiskExecutor) getAllDevPaths() map[string]string {

	devPaths := map[string]string{}
	contents, err := ioutil.ReadDir(rd.poolPath)
	if err != nil {
		rd.logger.WithField("dir", rd.poolPath).WithError(err).Warning("Failed to read directory")
		return devPaths
	}

	for _, f := range contents {
		if !f.IsDir() || !strings.HasPrefix(f.Name(), "pvc-") {
			continue
		}
		devPaths[f.Name()] = rd.getDevPath(f.Name())
	}
	rd.logger.WithField("devpaths", devPaths).Debug("Fetched device paths")
	return devPaths
}

func (rd *ramdiskExecutor) getDevPath(volumeName string) string {
	return filepath.Join(rd.poolPath, volumeName)
}
