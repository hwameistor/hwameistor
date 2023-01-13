package helm

import (
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/hwameistor/hwameistor/pkg/exechelper"
)

const (
	HelmToolJobAffinityKey = "kubernetes.io/hostname"
	HelmToolJobLabelApp    = "migrate-HelmTool"
)

type HelmTool struct {
	HelmToolImage              string
	HelmToolMountContainerName string
	HelmToolConfigMapNamespace string
	cmdExec                    exechelper.Executor
}

func (ht *HelmTool) StartHelmToolJob(jobName, lvName, srcNodeName string, waitUntilSuccess bool, timeout time.Duration) error {
	job := ht.getBaseJobStruct(jobName, lvName)

	affinity := &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      HelmToolJobAffinityKey,
								Operator: corev1.NodeSelectorOpNotIn,
								Values: []string{
									srcNodeName,
								},
							},
						},
					},
				},
			},
		},
	}
	job.Spec.Template.Spec.Affinity = affinity

	return nil
}

// User should fill up with Spec.Template.Spec.Containers[0].Command
func (ht *HelmTool) getBaseJobStruct(jobName, volName string) *batchv1.Job {
	var privileged = true
	baseStruct := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: ht.HelmToolConfigMapNamespace,
			//Annotations: annotations,
			Labels: map[string]string{
				"app": HelmToolJobLabelApp,
			},
		},
		Spec: batchv1.JobSpec{
			// Require feature gate
			//TTLSecondsAfterFinished: &ttl,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": HelmToolJobLabelApp,
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy: "Never",
					Containers: []corev1.Container{
						{
							Name:  ht.HelmToolMountContainerName,
							Image: ht.HelmToolImage,
							SecurityContext: &corev1.SecurityContext{
								Privileged: &privileged,
							},
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
