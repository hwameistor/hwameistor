package datacopy

import (
	"fmt"
	"path/filepath"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	RcloneRemoteName             = "remote"
	migrateVolumeMountPathPrefix = "/var/data/"
	migrateVolumePrefix          = "migrate-data-"
	migratePvcSuffix             = "-migrate"
)

type Rclone struct {
	rcloneImage              string
	rcloneContrinerName      string
	rcloneConfigMapName      string
	rcloneConfigMapNamespace string
	rcloneConfigMapKey       string
	skipRcloneConfiguration  bool
	dcm                      *DataCopyManager
}

// StartCopyJob will return a K8s job struct based on rclone to copy data
func (rcl *Rclone) PVCToRemotePVC(jobName, srcPVCName, srcSubPath, dstPVCName, dstSubPath, userData, namespace string, waitUntilSuccess bool, timeout time.Duration) error {
	jobStruct := rcl.getBaseJobStruct(jobName, userData, namespace)

	jobStruct.Spec.Template.Spec.Volumes = append(jobStruct.Spec.Template.Spec.Volumes, []corev1.Volume{
		{
			Name: "src",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: srcPVCName,
					ReadOnly:  true,
				},
			},
		},
	}...)

	jobStruct.Spec.Template.Spec.Containers[0].VolumeMounts = append(jobStruct.Spec.Template.Spec.Containers[0].VolumeMounts, []corev1.VolumeMount{
		{
			Name:      "src",
			MountPath: migrateVolumeMountPathPrefix + srcPVCName,
		},
	}...)

	logger.Debugf("PVCToRemotePVC rclone sync --create-empty-src-dirs /var/data/%s remote:/var/data/%s", srcPVCName, srcPVCName)
	jobStruct.Spec.Template.Spec.Containers[0].Command = []string{"rclone", "sync", "--create-empty-src-dirs",
		filepath.Join(migrateVolumeMountPathPrefix, srcPVCName),
		RcloneRemoteName + ":" + filepath.Join(migrateVolumeMountPathPrefix, srcPVCName),
	}

	if err := rcl.dcm.k8sControllerClient.Create(rcl.dcm.ctx, jobStruct); err != nil {
		if k8serrors.IsAlreadyExists(err) {
			fmt.Errorf("Failed to create MigrateJob, Job already exists")
		} else {
			return err
		}
	}

	return nil
}

func (rcl *Rclone) WaitMigrateJobTaskDone(jobName, srcPVCName, dstPVCName string, waitUntilSuccess bool, timeout time.Duration) error {

	resCh := make(chan *DataCopyStatus)
	rcl.dcm.RegisterRelatedJob(jobName, resCh)
	defer rcl.dcm.DeregisterRelatedJob(jobName)

	if waitUntilSuccess {
		if timeout == 0 {
			timeout = DefaultCopyTimeout
		}

		select {
		case res := <-resCh:
			if res.Phase == DataCopyStatusFailed {
				return fmt.Errorf("Failed to run job %s, message is [TODO] %s", res.JobName, res.Message)
			} else {
				return nil
			}
		case <-time.After(timeout):
			return fmt.Errorf("Failed to copy data from PVC: %s to PVC: %s, timeout after %s", srcPVCName, dstPVCName, timeout)
		}
	}

	return nil
}

// User should fill up with Spec.Template.Spec.Containers[0].Command
func (rcl *Rclone) getBaseJobStruct(jobName, userData, namespace string) *batchv1.Job {
	// Base job struct just with annotation and necessary container
	//dataCopyStatus := &DataCopyStatus{
	//	UserData: userData,
	//	JobName:  jobName,
	//	Phase:    DataCopyStatusRunning,
	//}
	//b, _ := json.Marshal(dataCopyStatus)
	//annotations := map[string]string{
	//	rcl.dcm.dataCopyJobStatusAnnotationName: string(b),
	//}
	baseStruct := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: namespace,
			//Annotations: annotations,
			Labels: map[string]string{
				"app": "migrate-rclone",
			},
		},
		Spec: batchv1.JobSpec{
			// Require feature gate
			//TTLSecondsAfterFinished: &ttl,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "migrate-rclone",
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy: "Never",
					Containers: []corev1.Container{
						{
							Name:  rcl.rcloneContrinerName,
							Image: rcl.rcloneImage,
						},
					},
				},
			},
		},
	}

	// Return without config
	if rcl.skipRcloneConfiguration {
		return baseStruct
	}

	// Container volume mount declare
	baseStruct.Spec.Template.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{
		{
			Name:      "rclone-config",
			MountPath: "/config/rclone/",
		},
	}

	// Template volume declare
	baseStruct.Spec.Template.Spec.Volumes = []corev1.Volume{
		{
			Name: "rclone-config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: rcl.rcloneConfigMapName},
					Items: []corev1.KeyToPath{
						{
							Key:  rcl.rcloneConfigMapKey,
							Path: "rclone.conf",
						},
					},
				},
			},
		},
	}
	return baseStruct
}

func (rcl *Rclone) EnsureRcloneConfigMapToTargetNamespace(targetNamespace string) error {
	// Do not copy when origin cm can be used
	if rcl.rcloneConfigMapNamespace != targetNamespace {
		keyToFindStaleCM := k8sclient.ObjectKey{
			Name:      rcl.rcloneConfigMapName,
			Namespace: targetNamespace,
		}

		staleCM := &corev1.ConfigMap{}
		if err := rcl.dcm.k8sControllerClient.Get(rcl.dcm.ctx, keyToFindStaleCM, staleCM); err != nil {
			if !k8serrors.IsNotFound(err) {
				logger.WithError(err).Errorf("Failed to check stale configmap name: %s namespace: %s", rcl.rcloneConfigMapName, targetNamespace)
				return err
			}
		} else {
			if err := rcl.dcm.k8sControllerClient.Delete(rcl.dcm.ctx, staleCM); err != nil {
				logger.WithError(err).Errorf("Failed to delete stale configmap name: %s namespace: %s", rcl.rcloneConfigMapName, targetNamespace)
				return err
			}
		}

		keyToFindOriginCM := k8sclient.ObjectKey{
			Name:      rcl.rcloneConfigMapName,
			Namespace: rcl.rcloneConfigMapNamespace,
		}

		originCM := &corev1.ConfigMap{}
		if err := rcl.dcm.k8sControllerClient.Get(rcl.dcm.ctx, keyToFindOriginCM, originCM); err != nil {
			logger.WithError(err).Errorf("Failed to get origin configmap name: %s namespace: %s", rcl.rcloneConfigMapName, rcl.rcloneConfigMapNamespace)
			return err
		}

		originCMCopyed := originCM.DeepCopy()
		cmForTargetNamespace := &corev1.ConfigMap{}
		cmForTargetNamespace.Name = originCMCopyed.Name
		cmForTargetNamespace.Immutable = originCMCopyed.Immutable
		cmForTargetNamespace.Data = originCMCopyed.Data
		cmForTargetNamespace.BinaryData = originCMCopyed.BinaryData
		cmForTargetNamespace.Namespace = targetNamespace

		if err := rcl.dcm.k8sControllerClient.Create(rcl.dcm.ctx, cmForTargetNamespace); err != nil {
			logger.WithError(err).Errorf("Failed to copy configmap name: %s namespace: %s to target namespace: %s", rcl.rcloneConfigMapName, rcl.rcloneConfigMapNamespace, targetNamespace)
			return err
		}
	}
	rcl.skipRcloneConfiguration = false

	return nil
}

// TODO
func (rcl *Rclone) progressWatchingFunc() *Progress {
	return nil
}
