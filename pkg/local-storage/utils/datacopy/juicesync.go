package datacopy

import (
	"context"
	"fmt"
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"os"
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	juiceSyncImageName = "ghcr.io/hwameistor/hwameistor-juicesync:v1.0.4-01"
)

type JuiceSync struct {
	namespace string
	apiClient k8sclient.Client
}

func (js *JuiceSync) Prepare(targetNodeName, sourceNodeName, volName string) (err error) {
	ctx := context.TODO()

	cmName := GetConfigMapName(SyncConfigMapName, volName)

	cm := &corev1.ConfigMap{}
	if err = js.apiClient.Get(ctx, types.NamespacedName{Namespace: js.namespace, Name: cmName}, cm); err == nil {
		logger.WithField("configmap", cmName).Debug("The config of data sync already exists")
		return nil
	}

	data := map[string]string{
		SyncConfigVolumeNameKey:     volName,
		SyncConfigSourceNodeNameKey: sourceNodeName,
		SyncConfigTargetNodeNameKey: targetNodeName,
	}
	cm = &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: js.namespace,
			Labels:    map[string]string{},
		},
		Data: data,
	}

	if err = js.apiClient.Create(ctx, cm); err != nil {
		logger.WithError(err).Error("Failed to create MigrateConfigmap")
		return err
	}

	return nil
}

// getStorageNodeIP returns the StorageIP configured in the corresponding LocalStorageNode
func (js *JuiceSync) getStorageNodeIP(nodeName string) (string, error) {
	storageNode := apisv1alpha1.LocalStorageNode{}
	if err := js.apiClient.Get(context.TODO(), k8sclient.ObjectKey{Name: nodeName}, &storageNode); err != nil {
		return "", err
	}
	return storageNode.Spec.StorageIP, nil
}

func (js *JuiceSync) StartSync(jobName, volName, excludedRunningNodeName, runningNodeName string) error {
	job := js.buildJob(jobName, volName, excludedRunningNodeName, runningNodeName)

	if err := js.apiClient.Create(context.TODO(), job); err != nil {
		if k8serrors.IsAlreadyExists(err) {
			return fmt.Errorf("failed to create sync job, already exists")
		}
		return err
	}

	return nil
}

func (js *JuiceSync) buildJob(jobName string, volName string, excludedRunningNodeName string, runningNodeName string) *batchv1.Job {
	if value := os.Getenv("MIGRAGE_JUICESYNC_IMAGE"); len(value) > 0 {
		juiceSyncImageName = value
	}

	runCommand := "sync_hwameistor_volumes.sh"

	nodeSelectExpression := []corev1.NodeSelectorRequirement{}
	if len(strings.TrimSpace(runningNodeName)) > 0 {
		nodeSelectExpression = append(nodeSelectExpression, corev1.NodeSelectorRequirement{
			Key:      SyncJobAffinityKey,
			Operator: corev1.NodeSelectorOpIn,
			Values: []string{
				runningNodeName,
			},
		})
	} else if len(strings.TrimSpace(excludedRunningNodeName)) > 0 {
		nodeSelectExpression = append(nodeSelectExpression, corev1.NodeSelectorRequirement{
			Key:      SyncJobAffinityKey,
			Operator: corev1.NodeSelectorOpNotIn,
			Values: []string{
				excludedRunningNodeName,
			},
		})

	}

	// use StorageNodeIP instead of nodeName, more details see #1195
	var sourceNodeIP, targetNodeIP string
	cm, err := js.getConfigMap(volName)
	if err != nil {
		logger.WithError(err).Error("failed to get config map for job")
		return nil
	} else {
		if sourceNodeIP, err = js.getStorageNodeIP(cm.Data[SyncConfigSourceNodeNameKey]); err != nil {
			logger.WithError(err).Error("failed to get source node ip from config map for job")
			return nil
		}
		if targetNodeIP, err = js.getStorageNodeIP(cm.Data[SyncConfigTargetNodeNameKey]); err != nil {
			logger.WithError(err).Error("failed to get target node ip from config map for job")
			return nil
		}
	}

	var privileged = true
	baseStruct := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: js.namespace,
			//Annotations: annotations,
			Labels: map[string]string{
				"app": SyncJobLabelApp,
			},
			Finalizers: []string{SyncJobFinalizer},
		},
		Spec: batchv1.JobSpec{
			// Require feature gate
			//TTLSecondsAfterFinished: &ttl,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": SyncJobLabelApp,
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy: "Never",
					Containers: []corev1.Container{
						{
							Name:            syncMountContainerName,
							Image:           juiceSyncImageName,
							ImagePullPolicy: corev1.PullIfNotPresent,
							SecurityContext: &corev1.SecurityContext{
								Privileged: &privileged,
							},
							Command: []string{"sh", "-c", runCommand},
							EnvFrom: []corev1.EnvFromSource{
								{
									ConfigMapRef: &corev1.ConfigMapEnvSource{
										LocalObjectReference: corev1.LocalObjectReference{Name: GetConfigMapName(SyncConfigMapName, volName)},
									},
								},
							},
							// overwrite sourceNode and targetNode with src/dst node ip
							Env: []corev1.EnvVar{
								{
									Name:  SyncConfigSourceNodeNameKey,
									Value: sourceNodeIP,
								},
								{
									Name:  SyncConfigTargetNodeNameKey,
									Value: targetNodeIP,
								},
							},
						},
					},
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{
									{
										MatchExpressions: nodeSelectExpression,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	syncKeyConfigVolumeMount := corev1.VolumeMount{
		Name:      "key-config",
		MountPath: "/root/.ssh/id_rsa",
		SubPath:   SyncPrivateKeyFileName,
	}

	// Container volume mount declare
	baseStruct.Spec.Template.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{
		syncKeyConfigVolumeMount,
	}

	// Template volume declare
	baseStruct.Spec.Template.Spec.Volumes = []corev1.Volume{
		{
			Name: "key-config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: SyncKeyConfigMapName},
					Items: []corev1.KeyToPath{
						{
							Key:  SyncPrivateKeyFileName,
							Path: SyncPrivateKeyFileName,
						},
					},
				},
			},
		},
	}

	hostVolumeDevMount := corev1.VolumeMount{
		Name: "host-dev", MountPath: "/dev",
	}
	hostVolume := corev1.Volume{
		Name: "host-dev",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "/dev",
			},
		},
	}

	etchostsVolume := corev1.Volume{
		Name: "etc-hosts",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "/etc/hosts",
			},
		},
	}
	etchostsVolumeMount := corev1.VolumeMount{
		Name: "etc-hosts", MountPath: "/etc/hosts",
	}

	hostCopyVolumeMountMnt := corev1.VolumeMount{
		Name: "host-mnt", MountPath: "/mnt/",
	}
	hostVolumeMnt := corev1.Volume{
		Name: "host-mnt",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "/mnt",
			},
		},
	}
	baseStruct.Spec.Template.Spec.Containers[0].VolumeMounts = append(baseStruct.Spec.Template.Spec.Containers[0].VolumeMounts, hostVolumeDevMount)
	baseStruct.Spec.Template.Spec.Volumes = append(baseStruct.Spec.Template.Spec.Volumes, hostVolume)

	baseStruct.Spec.Template.Spec.Containers[0].VolumeMounts = append(baseStruct.Spec.Template.Spec.Containers[0].VolumeMounts, hostCopyVolumeMountMnt)
	baseStruct.Spec.Template.Spec.Volumes = append(baseStruct.Spec.Template.Spec.Volumes, hostVolumeMnt)

	baseStruct.Spec.Template.Spec.Containers[0].VolumeMounts = append(baseStruct.Spec.Template.Spec.Containers[0].VolumeMounts, etchostsVolumeMount)
	baseStruct.Spec.Template.Spec.Volumes = append(baseStruct.Spec.Template.Spec.Volumes, etchostsVolume)

	return baseStruct
}

func (js *JuiceSync) getConfigMap(volName string) (*corev1.ConfigMap, error) {
	cmName := GetConfigMapName(SyncConfigMapName, volName)
	cm := &corev1.ConfigMap{}
	return cm, js.apiClient.Get(context.TODO(), types.NamespacedName{Namespace: js.namespace, Name: cmName}, cm)
}
