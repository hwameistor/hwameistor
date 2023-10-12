package storage

import (
	"fmt"
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/exechelper"
	"github.com/hwameistor/hwameistor/pkg/exechelper/nsexecutor"
	log "github.com/sirupsen/logrus"
	"path"
	"sync"
)

type ddExecutor struct {
	cmdExec exechelper.Executor
	logger  *log.Entry
}

var (
	ddExecutorInstance *ddExecutor
	once               sync.Once
)

const (
	DiskDumpCMD       = "dd"
	DiskDumpTimeout   = 60 * 10 // seconds
	DiskDumpBlockSize = "bs=10M"
)

func newDDExecutor() *ddExecutor {
	once.Do(func() {
		ddExecutorInstance = &ddExecutor{
			cmdExec: nsexecutor.New(),
			logger:  log.WithField("Module", "NodeManager/ddExecutor"),
		}
	})

	return ddExecutorInstance
}

func (dd *ddExecutor) RollbackVolumeReplicaSnapshot(snapshotRestore *apisv1alpha1.LocalVolumeReplicaSnapshotRestore) error {
	panic("not implemented")
	return nil
}

func (dd *ddExecutor) RestoreVolumeReplicaSnapshot(snapshotRestore *apisv1alpha1.LocalVolumeReplicaSnapshotRestore) error {
	poolName := snapshotRestore.Spec.TargetPoolName
	outPutDevicePath := composePoolVolumePath(poolName, snapshotRestore.Spec.TargetVolume)
	inputDevicePath := composePoolVolumePath(poolName, snapshotRestore.Spec.SourceVolumeSnapshot)

	// example：dd if=/dev/LocalStorage_PoolHDD/snapshot of=/dev/LocalStorage_PoolHDD/volume-new bs=10M
	dataCopyCommand := exechelper.ExecParams{
		CmdName: DiskDumpCMD,
		CmdArgs: []string{
			fmt.Sprintf("if=%s", inputDevicePath),
			fmt.Sprintf("of=%s", outPutDevicePath),
			DiskDumpBlockSize,
		},
		Timeout: DiskDumpTimeout,
	}

	dd.logger.WithField("restoreVolume", outPutDevicePath).Info("Start restoring snapshot")
	if result := dd.cmdExec.RunCommand(dataCopyCommand); result.ExitCode != 0 {
		dd.logger.WithError(result.Error).WithField("restoreVolume", outPutDevicePath).Info("Failed to restore snapshot")
		return result.Error
	}

	dd.logger.WithField("restoreVolume", outPutDevicePath).Info("Successfully restored snapshot")
	return nil
}

const SysDevicePath = "/dev"

func composePoolVolumePath(poolName, volume string) string {
	return path.Join(SysDevicePath, poolName, volume)
}
