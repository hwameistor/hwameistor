package node

import (
	"context"
	"fmt"
	"strings"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-storage/utils/datacopy"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (m *manager) startRcloneVolumeMountTaskWorker(stopCh <-chan struct{}) {

	m.logger.Debug("VolumeBlockMount Assignment Worker is working now")
	go func() {
		for {
			task, shutdown := m.rcloneVolumeMountTaskQueue.Get()
			if shutdown {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Stop the volumeBlockMountTask Assignment worker")
				break
			}
			if err := m.processRcloneVolumeMount(task); err != nil {
				m.logger.WithFields(log.Fields{"task": task, "error": err.Error()}).Error("Failed to process assignment, retry later")
				m.rcloneVolumeMountTaskQueue.AddRateLimited(task)
			} else {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Completed an assignment.")
				m.rcloneVolumeMountTaskQueue.Forget(task)
			}
			m.rcloneVolumeMountTaskQueue.Done(task)
		}
	}()

	<-stopCh
	m.rcloneVolumeMountTaskQueue.Shutdown()
}

func (m *manager) processRcloneVolumeMount(lvName string) error {
	logCtx := m.logger.WithFields(log.Fields{"LocalVolume": lvName, "nodeName": m.name})
	logCtx.Debug("Working on a rclone volume mount Task")

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

	cmName := datacopy.GetConfigMapName(datacopy.RCloneConfigMapName, lvName)
	cm := &corev1.ConfigMap{}
	if err := m.apiClient.Get(context.TODO(), types.NamespacedName{Namespace: m.namespace, Name: cmName}, cm); err != nil {
		logCtx.WithField("configmap", cmName).Error("Not found the rclone configmap")
		return err
	}

	vol := &apisv1alpha1.LocalVolume{}
	if err := m.apiClient.Get(ctx, types.NamespacedName{Name: lvName}, vol); err != nil {
		m.logger.WithField("LocalVolume", lvName).WithError(err).Error("Failed to get LocalVolume")
		return err
	}

	sourceNodeName := cm.Data[datacopy.RCloneConfigSrcNodeNameKey]
	targetNodeName := cm.Data[datacopy.RCloneConfigDstNodeNameKey]
	mountPoint := ""
	if m.name == sourceNodeName {
		mountPoint = datacopy.RCloneSrcMountPoint + lvName
	} else if m.name == targetNodeName {
		mountPoint = datacopy.RCloneDstMountPoint + lvName
	} else {
		return nil
	}

	if cm.Data[datacopy.RCloneConfigSyncDoneKey] == datacopy.RCloneTrue {
		if err := m.mounter.Unmount(mountPoint); err != nil {
			m.logger.WithField("mountpoint", mountPoint).WithError(err).Error("Failed to Unmount volume")
			return err
		}
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
		if err := m.mounter.FormatAndMount(devPath, mountPoint, fsType, []string{}); err != nil {
			m.logger.WithField("mountpoint", mountPoint).WithError(err).Error("Failed to FormatAndMount volume")
			return err
		}
	}

	if m.name == sourceNodeName {
		if cm.Data[datacopy.RCloneConfigSourceNodeReadyKey] == datacopy.RCloneTrue {
			return nil
		}
		cm.Data[datacopy.RCloneConfigSourceNodeReadyKey] = datacopy.RCloneTrue
	} else {
		if cm.Data[datacopy.RCloneConfigRemoteNodeReadyKey] == datacopy.RCloneTrue {
			return nil
		}
		cm.Data[datacopy.RCloneConfigRemoteNodeReadyKey] = datacopy.RCloneTrue
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
