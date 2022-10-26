package datacopy

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/hwameistor/hwameistor/pkg/local-storage/exechelper"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	RcloneSrcName            = "source"
	RcloneRemoteName         = "remote"
	RCloneKeyDir             = "/root/.ssh"
	RCloneKeyComment         = "RClonePubKey"
	sshkeygenCmd             = "ssh-keygen"
	RCloneSrcMountPoint      = "/mnt/hwameistor/src/"
	RCloneDstMountPoint      = "/mnt/hwameistor/dst/"
	RClonePubKeyFileName     = "rclone.pub"
	RClonePrivateKeyFileName = "rclone"
	RCloneKeyConfigMapName   = "rclone-key-config"
	RCloneConfigMapName      = "rclone-config"
	RCloneConfigMapKey       = "rclone.conf"
	RCloneCertKey            = "rclone-ssh-keys"
	RcloneJobLabelApp        = "hwameistor-datasync-rclone"
	rcloneJobAffinityKey     = "kubernetes.io/hostname"

	RCloneConfigSrcNodeNameKey     = "sourceNode"
	RCloneConfigDstNodeNameKey     = "targetNode"
	RCloneConfigVolumeNameKey      = "localVolume"
	RCloneConfigSourceNodeReadyKey = "sourceReady"
	RCloneConfigRemoteNodeReadyKey = "targetReady"
	RCloneConfigSyncDoneKey        = "completed"

	RCloneTrue  string = "yes"
	RCloneFalse string = "no"

	RCloneJobFinalizer = "hwameistor.io/rclone-job-protect"
)

type Rclone struct {
	rcloneImage              string
	rcloneMountContainerName string
	rcloneConfigMapNamespace string
	dcm                      *DataCopyManager
	cmdExec                  exechelper.Executor
}

func (rcl *Rclone) generateSSHPubAndPrivateKeyCM() (string, string, error) {
	logger.Debug("GenerateSSHPubAndPrivateKey start ")

	paramsRemove := exechelper.ExecParams{
		CmdName: "rm",
		CmdArgs: []string{"-rf", filepath.Join(RCloneKeyDir, RClonePrivateKeyFileName)},
		Timeout: 0,
	}
	resultRemove := rcl.cmdExec.RunCommand(paramsRemove)
	if resultRemove.ExitCode != 0 {
		return "", "", fmt.Errorf("rm -rf %s err: %d, %s", filepath.Join(RCloneKeyDir, RClonePrivateKeyFileName), resultRemove.ExitCode, resultRemove.ErrBuf.String())
	}

	paramsMkdir := exechelper.ExecParams{
		CmdName: "mkdir",
		CmdArgs: []string{"-p", RCloneKeyDir},
		Timeout: 0,
	}
	resultMkdir := rcl.cmdExec.RunCommand(paramsMkdir)
	if resultMkdir.ExitCode != 0 {
		return "", "", fmt.Errorf("mkdir -p %s err: %d, %s", RCloneKeyDir, resultMkdir.ExitCode, resultMkdir.ErrBuf.String())
	}

	params := exechelper.ExecParams{
		CmdName: sshkeygenCmd,
		CmdArgs: []string{"-q", "-b 4096", "-C" + RCloneKeyComment, "-f", filepath.Join(RCloneKeyDir, RClonePrivateKeyFileName)},
		Timeout: 0,
	}
	result := rcl.cmdExec.RunCommand(params)
	if result.ExitCode != 0 {
		return "", "", fmt.Errorf("ssh-keygen %s err: %d, %s", RCloneKeyComment, result.ExitCode, result.ErrBuf.String())
	}

	paramsCatRclone := exechelper.ExecParams{
		CmdName: "cat",
		CmdArgs: []string{filepath.Join(RCloneKeyDir, RClonePrivateKeyFileName)},
		Timeout: 0,
	}
	resultCatRclone := rcl.cmdExec.RunCommand(paramsCatRclone)
	if resultCatRclone.ExitCode != 0 {
		return "", "", fmt.Errorf("cat %s err: %d, %s", filepath.Join(RCloneKeyDir, RClonePrivateKeyFileName), resultCatRclone.ExitCode, resultCatRclone.ErrBuf.String())
	}

	paramsCatRclonePub := exechelper.ExecParams{
		CmdName: "cat",
		CmdArgs: []string{filepath.Join(RCloneKeyDir, RClonePubKeyFileName)},
		Timeout: 0,
	}
	resultCatRclonePub := rcl.cmdExec.RunCommand(paramsCatRclonePub)
	if resultCatRclonePub.ExitCode != 0 {
		return "", "", fmt.Errorf("cat %s err: %d, %s", filepath.Join(RCloneKeyDir, RClonePubKeyFileName), resultCatRclonePub.ExitCode, resultCatRclonePub.ErrBuf.String())
	}
	rclonePubKeyData := resultCatRclonePub.OutBuf.String()
	rclonePrivateKeyData := resultCatRclone.OutBuf.String()

	return rclonePubKeyData, rclonePrivateKeyData, nil
}

func (rcl *Rclone) GenerateRcloneKeyConfigMap() *corev1.ConfigMap {

	var rcloneCM = &corev1.ConfigMap{}
	rclonePubKeyData, rclonePrivateKeyData, err := rcl.generateSSHPubAndPrivateKeyCM()

	if err != nil {
		logger.WithError(err).Errorf("generateRcloneKeyConfigMap generateSSHPubAndPrivateKeyCM")
		return rcloneCM
	}

	configData := map[string]string{
		RClonePubKeyFileName:     rclonePubKeyData,
		RClonePrivateKeyFileName: rclonePrivateKeyData,
		RCloneCertKey:            rclonePrivateKeyData + "\n" + rclonePubKeyData,
	}

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      RCloneKeyConfigMapName,
			Namespace: rcl.rcloneConfigMapNamespace,
		},
		Data: configData,
	}

	return configMap
}

func (rcl *Rclone) StartRCloneJob(jobName, lvName, srcNodeName string, waitUntilSuccess bool, timeout time.Duration) error {
	job := rcl.getBaseJobStruct(jobName, lvName)

	affinity := &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      rcloneJobAffinityKey,
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

	tmpDstMountPoint := RCloneDstMountPoint + lvName
	tmpSrcMountPoint := RCloneSrcMountPoint + lvName

	rcloneCommand := fmt.Sprintf("rclone sync %s:%s %s:%s --progress;", RcloneSrcName, tmpSrcMountPoint, RcloneRemoteName, tmpDstMountPoint)
	job.Spec.Template.Spec.Containers[0].Command = []string{"sh", "-c", rcloneCommand}

	if err := rcl.dcm.k8sControllerClient.Create(rcl.dcm.ctx, job); err != nil {
		if k8serrors.IsAlreadyExists(err) {
			return fmt.Errorf("failed to create rclone job, already exists")
		}
		return err
	}

	return nil
}

// func (rcl *Rclone) WaitMigrateJobTaskDone(jobName, lvName string, waitUntilSuccess bool, timeout time.Duration) error {

// 	resCh := make(chan *DataCopyStatus)
// 	rcl.dcm.RegisterRelatedJob(jobName, resCh)
// 	defer rcl.dcm.DeregisterRelatedJob(jobName)

// 	if waitUntilSuccess {
// 		if timeout == 0 {
// 			timeout = DefaultCopyTimeout
// 		}

// 		select {
// 		case res := <-resCh:
// 			if res.Phase == DataCopyStatusFailed {
// 				return fmt.Errorf("failed to run job %s, message is [TODO] %s", res.JobName, res.Message)
// 			} else {
// 				return nil
// 			}
// 		case <-time.After(timeout):
// 			return fmt.Errorf("failed to copy data from srcMountPointName: %s to dstMountPointName: %s, timeout after %s", RCloneSrcMountPoint+lvName, RCloneDstMountPoint+lvName, timeout)
// 		}
// 	}

// 	return nil
// }

// User should fill up with Spec.Template.Spec.Containers[0].Command
func (rcl *Rclone) getBaseJobStruct(jobName, volName string) *batchv1.Job {
	var privileged = true
	baseStruct := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: rcl.rcloneConfigMapNamespace,
			//Annotations: annotations,
			Labels: map[string]string{
				"app": RcloneJobLabelApp,
			},
			Finalizers: []string{RCloneJobFinalizer},
		},
		Spec: batchv1.JobSpec{
			// Require feature gate
			//TTLSecondsAfterFinished: &ttl,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": RcloneJobLabelApp,
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy: "Never",
					Containers: []corev1.Container{
						{
							Name:  rcl.rcloneMountContainerName,
							Image: rcl.rcloneImage,
							SecurityContext: &corev1.SecurityContext{
								Privileged: &privileged,
							},
						},
					},
				},
			},
		},
	}

	rcloneConfigVolumeMount := corev1.VolumeMount{
		Name:      "rclone-config",
		MountPath: filepath.Join("/config/rclone", RCloneConfigMapKey),
		SubPath:   RCloneConfigMapKey,
	}

	rcloneKeyConfigVolumeMount := corev1.VolumeMount{
		Name:      "rclone-key-config",
		MountPath: filepath.Join("/config/rclone", RCloneCertKey),
		SubPath:   RCloneCertKey,
	}

	// Container volume mount declare
	baseStruct.Spec.Template.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{
		rcloneConfigVolumeMount,
		rcloneKeyConfigVolumeMount,
	}

	// Template volume declare
	baseStruct.Spec.Template.Spec.Volumes = []corev1.Volume{
		{
			Name: "rclone-config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: GetConfigMapName(RCloneConfigMapName, volName)},
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
					LocalObjectReference: corev1.LocalObjectReference{Name: RCloneKeyConfigMapName},
					Items: []corev1.KeyToPath{
						{
							Key:  RCloneCertKey,
							Path: RCloneCertKey,
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

// TODO
func (rcl *Rclone) progressWatchingFunc() *Progress {
	return nil
}

func GetConfigMapName(str1, str2 string) string {
	return fmt.Sprintf("%s-%s", str1, str2)
}
