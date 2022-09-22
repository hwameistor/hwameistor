package node

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"strings"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/types"
)

func (m *manager) startVolumeBlockMountTaskWorker(stopCh <-chan struct{}) {

	m.logger.Debug("VolumeBlockMount Assignment Worker is working now")
	go func() {
		for {
			task, shutdown := m.volumeBlockMountTaskQueue.Get()
			if shutdown {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Stop the volumeBlockMountTask Assignment worker")
				break
			}
			if err := m.processVolumeBlockMountTaskAssignment(task); err != nil {
				m.logger.WithFields(log.Fields{"task": task, "error": err.Error()}).Error("Failed to process assignment, retry later")
				m.volumeBlockMountTaskQueue.AddRateLimited(task)
			} else {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Completed an assignment.")
				m.volumeBlockMountTaskQueue.Forget(task)
			}
			m.volumeBlockMountTaskQueue.Done(task)
		}
	}()

	<-stopCh
	m.volumeBlockMountTaskQueue.Shutdown()
}

func (m *manager) processVolumeBlockMountTaskAssignment(lvNamespacedName string) error {
	logCtx := m.logger.WithFields(log.Fields{"VolumeBlockMount": lvNamespacedName, "nodeName": m.name})
	logCtx.Debug("Working on a VolumeBlockMount Task")

	m.lock.Lock()
	defer m.lock.Unlock()

	splitRes := strings.Split(lvNamespacedName, "/")
	var ns, lvname string
	if len(splitRes) >= 2 {
		ns = splitRes[0]
		lvname = splitRes[1]
	}

	cmList := &corev1.ConfigMapList{}
	if err := m.apiClient.List(context.TODO(), cmList); err != nil {
		m.logger.WithError(err).Error("Failed to get cmList")
		return err
	}

	var sourceNodeName, targetNodeName, syncDone string
	for _, cm := range cmList.Items {
		if cm.Name == rcloneConfigMapName {
			sourceNodeName = cm.Data["sourceNodeName"]
			targetNodeName = cm.Data["targetNodeName"]
			syncDone = cm.Data["syncDone"]
		}
	}

	tmpVol := &apisv1alpha1.LocalVolume{}
	if err := m.apiClient.Get(context.TODO(), types.NamespacedName{Name: lvname}, tmpVol); err != nil {
		m.logger.WithError(err).Error("Failed to get LocalVolume")
		return err
	}

	newLvName := strings.Replace(lvname, "-", "--", -1)
	devPath := fmt.Sprintf("/dev/mapper/%s-%s", tmpVol.Spec.PoolName, newLvName)

	tmpDstMountPoint := dstMountPoint + lvname
	tmpSrcMountPoint := srcMountPoint + lvname

	var flags []string
	fstype := "xfs"

	if m.name == sourceNodeName {

		// return directly if device has already mounted at TargetPath
		if !isStringInArray(tmpSrcMountPoint, m.mounter.GetDeviceMountPoints(devPath)) {
			if err := m.mounter.FormatAndMount(devPath, tmpSrcMountPoint, fstype, flags); err != nil {
				m.logger.WithError(err).Error("Failed to FormatAndMount tmpSrcMountPoint")
				return err
			}
		} else {
			if syncDone == "True" {
				m.logger.Debug("processVolumeBlockMountTaskAssignment Unmount tmpSrcMountPoint = %v, sourceNodeName = %v", tmpSrcMountPoint, sourceNodeName)
				if err := m.mounter.Unmount(tmpSrcMountPoint); err != nil {
					m.logger.WithError(err).Error("Failed to Unmount tmpSrcMountPoint")
					return err
				}
			}
		}

		// return directly if device has already mounted at TargetPath
		if isStringInArray(tmpDstMountPoint, m.mounter.GetDeviceMountPoints(devPath)) {
			if syncDone == "True" {
				m.logger.Debug("processVolumeBlockMountTaskAssignment Unmount tmpDstMountPoint = %v, sourceNodeName = %v", tmpDstMountPoint, sourceNodeName)
				if err := m.mounter.Unmount(tmpDstMountPoint); err != nil {
					m.logger.WithError(err).Error("Failed to Unmount tmpSrcMountPoint")
					return err
				}
			}
		}
	}

	if m.name == targetNodeName {
		// return directly if device has already mounted at TargetPath
		if !isStringInArray(tmpDstMountPoint, m.mounter.GetDeviceMountPoints(devPath)) {
			if err := m.mounter.FormatAndMount(devPath, tmpDstMountPoint, fstype, flags); err != nil {
				m.logger.WithError(err).Error("Failed to FormatAndMount tmpDstMountPoint")
				return err
			}
		} else {
			if syncDone == "True" {
				m.logger.Debug("processVolumeBlockMountTaskAssignment Unmount tmpDstMountPoint = %v, targetNodeName = %v", tmpDstMountPoint, targetNodeName)
				if err := m.mounter.Unmount(tmpDstMountPoint); err != nil {
					m.logger.WithError(err).Error("Failed to Unmount tmpDstMountPoint")
					return err
				}
			}
		}

		// return directly if device has already mounted at TargetPath
		if isStringInArray(tmpSrcMountPoint, m.mounter.GetDeviceMountPoints(devPath)) {
			if syncDone == "True" {
				m.logger.Debug("processVolumeBlockMountTaskAssignment Unmount tmpSrcMountPoint = %v, targetNodeName = %v", tmpSrcMountPoint, targetNodeName)
				if err := m.mounter.Unmount(tmpSrcMountPoint); err != nil {
					m.logger.WithError(err).Error("Failed to Unmount tmpSrcMountPoint")
					return err
				}
			}
		}
	}

	if syncDone == "True" {
		tmpConfigMap := &corev1.ConfigMap{}
		err := m.apiClient.Get(context.TODO(), types.NamespacedName{Namespace: ns, Name: rcloneConfigMapName}, tmpConfigMap)
		if err == nil {
			if delErr := m.apiClient.Delete(context.TODO(), tmpConfigMap); delErr != nil {
				m.logger.WithError(err).Error("Failed to delete Configmap")
				return delErr
			}
		}
	}
	return nil
}
