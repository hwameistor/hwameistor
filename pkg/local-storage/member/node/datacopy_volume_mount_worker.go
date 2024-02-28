package node

import (
	"context"
	"fmt"
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/exechelper"
	"github.com/hwameistor/hwameistor/pkg/exechelper/nsexecutor"
	"github.com/hwameistor/hwameistor/pkg/local-storage/utils/datacopy"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"os"
	"strings"
	"time"
)

func (m *manager) startSyncVolumeMountTaskWorker(stopCh <-chan struct{}) {

	m.logger.Debug("VolumeBlockMount Assignment Worker is working now")
	go func() {
		for {
			task, shutdown := m.syncVolumeMountTaskQueue.Get()
			if shutdown {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Stop the volumeBlockMountTask Assignment worker")
				break
			}
			if err := m.processSyncVolumeMount(task); err != nil {
				m.logger.WithFields(log.Fields{"task": task, "error": err.Error()}).Error("Failed to process assignment, retry later")
				m.syncVolumeMountTaskQueue.AddRateLimited(task)
			} else {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Completed an assignment.")
				m.syncVolumeMountTaskQueue.Forget(task)
			}
			m.syncVolumeMountTaskQueue.Done(task)
		}
	}()

	<-stopCh
	m.syncVolumeMountTaskQueue.Shutdown()
}

func (m *manager) processSyncVolumeMount(lvName string) error {
	logCtx := m.logger.WithFields(log.Fields{"LocalVolume": lvName, "nodeName": m.name})
	logCtx.Debug("Working on a data sync volume mount Task")

	m.lock.Lock()
	defer m.lock.Unlock()

	ctx := context.TODO()

	replicas, err := m.getReplicasForVolume(lvName)
	if err != nil {
		logCtx.Error("Failed to list VolumeReplica")
		return err
	}
	if len(replicas) != 2 {
		return fmt.Errorf("incorrect number of the volume replicas")
	}
	for _, replica := range replicas {
		if replica.Status.State != apisv1alpha1.VolumeReplicaStateReady {
			return fmt.Errorf("volume replica is not ready")
		}
	}

	cmName := datacopy.GetConfigMapName(datacopy.SyncConfigMapName, lvName)
	cm := &corev1.ConfigMap{}
	if err := m.apiClient.Get(ctx, types.NamespacedName{Namespace: m.namespace, Name: cmName}, cm); err != nil {
		logCtx.WithField("configmap", cmName).Error("Not found the data sync configmap")
		return err
	}

	vol := &apisv1alpha1.LocalVolume{}
	if err := m.apiClient.Get(ctx, types.NamespacedName{Name: lvName}, vol); err != nil {
		m.logger.WithField("LocalVolume", lvName).WithError(err).Error("Failed to get LocalVolume")
		return err
	}

	if verify, ok := cm.Data[datacopy.SyncConfigSyncVerifyKey]; ok && verify == datacopy.SyncVerifyFail {
		if err := m.apiClient.Delete(ctx, cm); err != nil {
			logCtx.WithField("configmap", cmName).WithError(err).Error("Failed to delete sync configmap")
			return err
		}
		return nil
	}

	sourceNodeName := cm.Data[datacopy.SyncConfigSourceNodeNameKey]
	targetNodeName := cm.Data[datacopy.SyncConfigTargetNodeNameKey]
	mountPoint := ""
	if m.name == sourceNodeName {
		mountPoint = datacopy.SyncSourceMountPoint + lvName
		_, exist := cm.Data[datacopy.SyncConfigSourceNodeDataHash]
		if verify, ok := cm.Data[datacopy.SyncConfigSyncVerifyKey]; ok && verify == datacopy.SyncVerifyStart && !exist {
			start := time.Now()
			datahash, err := m.ComputeDataHash(mountPoint)
			if err != nil {
				logCtx.WithField("mountPoint", mountPoint).Error("Failed to calculate data hash")
				return err
			}
			elapsed := time.Since(start)
			logCtx.Debugf("compute filehash time ：%s", elapsed)
			cm.Data[datacopy.SyncConfigSourceNodeDataHash] = datahash
			if err := m.apiClient.Update(ctx, cm); err != nil {
				m.logger.WithField("configmap", cm.Name).WithError(err).Error("Failed to update rclone's config")
				return err
			}
		}
	} else if m.name == targetNodeName {
		mountPoint = datacopy.SyncTargetMountPoint + lvName
		_, exist := cm.Data[datacopy.SyncConfigTargetNodeDataHash]
		if verify, ok := cm.Data[datacopy.SyncConfigSyncVerifyKey]; ok && verify == datacopy.SyncVerifyStart && !exist {
			datahash, err := m.ComputeDataHash(mountPoint)
			if err != nil {
				logCtx.WithField("mountPoint", mountPoint).Error("Failed to calculate data hash")
				return err
			}
			cm.Data[datacopy.SyncConfigTargetNodeDataHash] = datahash
			if err := m.apiClient.Update(ctx, cm); err != nil {
				m.logger.WithField("configmap", cm.Name).WithError(err).Error("Failed to update rclone's config")
				return err
			}
		} else if cm.Data[datacopy.SyncConfigSyncVerifyKey] == datacopy.SyncVerifyFail {

			err := m.CleanData(mountPoint)
			if err != nil {
				logCtx.WithField("mountPoint", mountPoint).WithError(err).Error("Failed to delete no-complete data")
				return err
			}
			if err := m.apiClient.Delete(ctx, cm); err != nil {
				logCtx.WithField("configmap", cmName).WithError(err).Error("Failed to delete sync configmap")
				return err
			}
		}
	} else {
		return nil
	}

	if cm.Data[datacopy.SyncConfigSyncCompleteKey] == datacopy.SyncTrue {
		m.logger.WithField("mountpoint", mountPoint).Debug("Trying to umount volume")

		if err := m.mounter.Unmount(mountPoint); err != nil {
			if !os.IsNotExist(err) {
				m.logger.WithField("mountpoint", mountPoint).WithError(err).Error("Failed to Unmount volume")
				return err
			} else {
				logCtx.Debugf("mountPoint delete success:%s", mountPoint)
			}
		}

		if m.name == sourceNodeName {
			cm.Data[datacopy.SyncConfigSourceNodeCompleteKey] = datacopy.SyncTrue
		} else {
			cm.Data[datacopy.SyncConfigTargetNodeCompleteKey] = datacopy.SyncTrue
		}
		if err = m.apiClient.Update(ctx, cm); err != nil {
			m.logger.WithField("configmap", cm.Name).WithError(err).Error("Failed to update config for target node complete")
			return err
		}
		m.logger.WithField("configmap", cm.Name).Debug("Successes to update config for nodes complete")
		return nil
	}

	newLvName := strings.Replace(lvName, "-", "--", -1)
	devPath := fmt.Sprintf("/dev/mapper/%s-%s", vol.Spec.PoolName, newLvName)

	fsType := vol.Status.PublishedFSType
	if len(fsType) == 0 {
		// in case of the upgrade
		fsType = "xfs"
	}
	// return directly if device has already mounted at TargetPath
	if !isStringInArray(mountPoint, m.mounter.GetDeviceMountPoints(devPath)) {
		m.logger.WithField("mountpoint", mountPoint).Debug("Trying to format and mount volume")
		if err := m.mounter.FormatAndMount(devPath, mountPoint, fsType, []string{}); err != nil {
			m.logger.WithField("mountpoint", mountPoint).WithError(err).Error("Failed to FormatAndMount volume")
			return err
		}
	}

	if m.name == sourceNodeName {
		if cm.Data[datacopy.SyncConfigSourceNodeReadyKey] == datacopy.SyncTrue {
			return nil
		}
		cm.Data[datacopy.SyncConfigSourceNodeReadyKey] = datacopy.SyncTrue
		cm.Data[datacopy.SyncConfigSourceMountPointKey] = mountPoint
	} else {
		if cm.Data[datacopy.SyncConfigTargetNodeReadyKey] == datacopy.SyncTrue {
			return nil
		}
		cm.Data[datacopy.SyncConfigTargetNodeReadyKey] = datacopy.SyncTrue
		cm.Data[datacopy.SyncConfigTargetMountPointKey] = mountPoint
	}
	if err := m.apiClient.Update(ctx, cm); err != nil {
		m.logger.WithField("configmap", cm.Name).WithError(err).Error("Failed to update rclone's config")
		return err
	}

	return nil
}

func (m *manager) getReplicasForVolume(volName string) ([]*apisv1alpha1.LocalVolumeReplica, error) {
	// todo
	replicaList := &apisv1alpha1.LocalVolumeReplicaList{}
	if err := m.apiClient.List(context.TODO(), replicaList); err != nil {
		return nil, err
	}

	var replicas []*apisv1alpha1.LocalVolumeReplica
	for i := range replicaList.Items {
		if replicaList.Items[i].Spec.VolumeName == volName {
			replicas = append(replicas, &replicaList.Items[i])
		}
	}
	return replicas, nil
}

func (m *manager) ComputeDataHash(mountPoint string) (string, error) {
	params := exechelper.ExecParams{
		//find . -type f -exec md5sum {} + | md5sum
		CmdName: "sh",
		CmdArgs: []string{"-c", "cd " + mountPoint + " &&  find . -type f -exec md5sum {} + | sort | md5sum"},
		Timeout: 1200,
	}
	result := nsexecutor.New().RunCommand(params)
	if result.ExitCode == 0 {
		return result.OutBuf.String(), nil
	}
	return "", result.Error
}

func (m *manager) CleanData(mountPoint string) error {
	params := exechelper.ExecParams{
		//find . -type f -exec md5sum {} + | md5sum
		CmdName: "sh",
		CmdArgs: []string{"-c", "cd " + mountPoint + " &&  rm -rf * "},
		Timeout: 600,
	}
	result := nsexecutor.New().RunCommand(params)
	if result.ExitCode == 0 {
		return nil
	}
	return result.Error
}

//func (m manager) ComputerLvmHash(vol *apisv1alpha1.LocalVolume) (string, error) {
//	//todo Problem: long block hash times
//	newLvName := strings.Replace(vol.Name, "-", "--", -1)
//	devPath := fmt.Sprintf("/dev/mapper/%s-%s", vol.Spec.PoolName, newLvName)
//	params := exechelper.ExecParams{
//		//md5sum LocalStorage_PoolHDD-pvc--3913863f--53a0--48d6--b3b4--e296a81b875d | awk '{print $1}'
//		CmdName: "sh",
//		CmdArgs: []string{"-c", "md5sum " + devPath + " | awk '{print $1}'"},
//		Timeout: 1200,
//	}
//	result := nsexecutor.New().RunCommand(params)
//	if result.ExitCode == 0 {
//		return result.OutBuf.String(), nil
//	}
//	return "", result.Error
//}
