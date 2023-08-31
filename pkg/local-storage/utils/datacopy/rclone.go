package datacopy

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	RCloneConfigMapKey = "rclone.conf"
)

var (
	rcloneImageName = "daocloud.io/daocloud/hwameistor-migrate-rclone:v1.1.2"
)

type RClone struct {
	namespace string
	apiClient k8sclient.Client
}

func (r *RClone) Prepare(targetNodeName, sourceNodeName, lvName string) error {
	ctx := context.TODO()

	cmName := GetConfigMapName(SyncConfigMapName, lvName)

	cm := &corev1.ConfigMap{}
	if err := r.apiClient.Get(context.TODO(), types.NamespacedName{Namespace: r.namespace, Name: cmName}, cm); err == nil {
		logger.WithField("configmap", cmName).Debug("The config of rclone already exists")
		return nil
	}

	remoteNameData := "[remote]" + "\n"
	sourceNameData := "[source]" + "\n"
	typeData := "type = sftp" + "\n"
	remoteHostData := "host = " + targetNodeName + "\n"
	sourceHostData := "host = " + sourceNodeName + "\n"
	keyFileData := "key_file = /config/rclone/" + SyncCertKey + "\n"
	shellTypeData := "shell_type = unix" + "\n"
	md5sumCommandData := "md5sum_command = md5sum" + "\n"
	sha1sumCommandData := "sha1sum_command = sha1sum" + "\n"

	remoteConfig := remoteNameData + typeData + remoteHostData + keyFileData + shellTypeData + md5sumCommandData + sha1sumCommandData
	sourceConfig := sourceNameData + typeData + sourceHostData + keyFileData + shellTypeData + md5sumCommandData + sha1sumCommandData
	data := map[string]string{
		RCloneConfigMapKey:       remoteConfig + sourceConfig,
		SyncConfigVolumeNameKey:  lvName,
		SyncConfigDstNodeNameKey: targetNodeName,
		SyncConfigSrcNodeNameKey: sourceNodeName,
	}
	cm = &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: r.namespace,
			Labels:    map[string]string{},
		},
		Data: data,
	}

	if err := r.apiClient.Create(ctx, cm); err != nil {
		logger.WithError(err).Error("Failed to create MigrateConfigmap")
		return err
	}

	return nil
}

func (r *RClone) StartSync(jobName string, volName string, excludedRunningNodeName string, runningNodeName string) error {
	job := r.getJob(jobName, volName, excludedRunningNodeName, runningNodeName)

	if err := r.apiClient.Create(context.TODO(), job); err != nil {
		if k8serrors.IsAlreadyExists(err) {
			return fmt.Errorf("failed to create sync job, already exists")
		}
		return err
	}

	return nil
}

func (r *RClone) getJob(jobName string, volName string, excludedRunningNodeName string, runningNodeName string) *batchv1.Job {
	if value := os.Getenv("MIGRAGE_RCLONE_IMAGE"); len(value) > 0 {
		rcloneImageName = value
	}

	tmpDstMountPoint := SyncDstMountPoint + volName
	tmpSrcMountPoint := SyncSrcMountPoint + volName
	runCommand := fmt.Sprintf("sync sync %s:%s %s:%s --progress --links;", SyncSrcName, tmpSrcMountPoint, SyncRemoteName, tmpDstMountPoint)

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

	var privileged = true
	baseStruct := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: r.namespace,
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
							Name:  syncMountContainerName,
							Image: rcloneImageName,
							SecurityContext: &corev1.SecurityContext{
								Privileged: &privileged,
							},
							Command: []string{"sh", "-c", runCommand},
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

	syncConfigVolumeMount := corev1.VolumeMount{
		Name:      "rclone-config",
		MountPath: filepath.Join("/config/rclone", RCloneConfigMapKey),
		SubPath:   RCloneConfigMapKey,
	}

	syncKeyConfigVolumeMount := corev1.VolumeMount{
		Name:      "rclone-key-config",
		MountPath: filepath.Join("/config/rclone", SyncCertKey),
		SubPath:   SyncCertKey,
	}

	// Container volume mount declare
	baseStruct.Spec.Template.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{
		syncConfigVolumeMount,
		syncKeyConfigVolumeMount,
	}

	// Template volume declare
	baseStruct.Spec.Template.Spec.Volumes = []corev1.Volume{
		{
			Name: "rclone-config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: GetConfigMapName(SyncConfigMapName, volName)},
					Items: []corev1.KeyToPath{
						{
							Key:  RCloneConfigMapKey,
							Path: RCloneConfigMapKey,
						},
					},
				},
			},
		},
		{
			Name: "rclone-key-config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: SyncKeyConfigMapName},
					Items: []corev1.KeyToPath{
						{
							Key:  SyncCertKey,
							Path: SyncCertKey,
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
