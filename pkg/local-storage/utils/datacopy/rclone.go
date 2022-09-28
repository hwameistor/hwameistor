package datacopy

import (
	"fmt"
	"github.com/hwameistor/hwameistor/pkg/local-storage/exechelper"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	RcloneSrcName          = "source"
	RcloneRemoteName       = "remote"
	rcloneKeyDir           = "~/.ssh"
	rcloneKeyFilePath      = "~/.ssh/rclone"
	rcloneAllKeyFilePath   = "~/.ssh/rclone-merged"
	rcloneKeyName          = "rclonekey"
	sshkeygenCmd           = "ssh-keygen"
	srcMountPoint          = "/mnt/src/"
	dstMountPoint          = "/mnt/dst/"
	rclonePubKeyFilePath   = "~/.ssh/rclone.pub"
	rcloneKeyConfigMapName = "rclone-key-config"
	rcloneJobAffinityKey   = "kubernetes.io/hostname"
)

type Rclone struct {
	rcloneImage              string
	rcloneMountContainerName string
	rcloneConfigMapName      string
	rcloneKeyConfigMapName   string
	rcloneConfigMapNamespace string
	rcloneConfigMapKey       string
	rcloneCertKey            string
	skipRcloneConfiguration  bool
	dcm                      *DataCopyManager
	cmdExec                  exechelper.Executor
}

func (rcl *Rclone) generateSSHPubAndPrivateKeyCM() (string, string, error) {
	logger.Debug("GenerateSSHPubAndPrivateKey start ")

	paramsRemove := exechelper.ExecParams{
		CmdName: "rm",
		CmdArgs: []string{"-rf", rcloneKeyFilePath},
		Timeout: 0,
	}
	resultRemove := rcl.cmdExec.RunCommand(paramsRemove)
	if resultRemove.ExitCode != 0 {
		return "", "", fmt.Errorf("rm -rf %s err: %d, %s", rcloneKeyFilePath, resultRemove.ExitCode, resultRemove.ErrBuf.String())
	}

	paramsRemoveAll := exechelper.ExecParams{
		CmdName: "rm",
		CmdArgs: []string{"-rf", rcloneAllKeyFilePath},
		Timeout: 0,
	}
	resultRemoveAll := rcl.cmdExec.RunCommand(paramsRemoveAll)
	if resultRemoveAll.ExitCode != 0 {
		return "", "", fmt.Errorf("rm -rf %s err: %d, %s", rcloneAllKeyFilePath, resultRemoveAll.ExitCode, resultRemoveAll.ErrBuf.String())
	}

	paramsMkdir := exechelper.ExecParams{
		CmdName: "mkdir",
		CmdArgs: []string{"-p", rcloneKeyDir},
		Timeout: 0,
	}
	resultMkdir := rcl.cmdExec.RunCommand(paramsMkdir)
	if resultMkdir.ExitCode != 0 {
		return "", "", fmt.Errorf("mkdir -p %s err: %d, %s", rcloneKeyDir, resultMkdir.ExitCode, resultMkdir.ErrBuf.String())
	}

	params := exechelper.ExecParams{
		CmdName: sshkeygenCmd,
		CmdArgs: []string{"-q", "-b 4096", "-CrcloneKey", "-f", rcloneKeyFilePath},
		Timeout: 0,
	}
	result := rcl.cmdExec.RunCommand(params)
	if result.ExitCode != 0 {
		return "", "", fmt.Errorf("ssh-keygen %s err: %d, %s", rcloneKeyName, result.ExitCode, result.ErrBuf.String())
	}

	paramsCatRclone := exechelper.ExecParams{
		CmdName: "cat",
		CmdArgs: []string{rcloneKeyFilePath},
		Timeout: 0,
	}
	resultCatRclone := rcl.cmdExec.RunCommand(paramsCatRclone)
	if resultCatRclone.ExitCode != 0 {
		return "", "", fmt.Errorf("cat %s err: %d, %s", rcloneKeyFilePath, resultCatRclone.ExitCode, resultCatRclone.ErrBuf.String())
	}

	paramsCatRclonePub := exechelper.ExecParams{
		CmdName: "cat",
		CmdArgs: []string{rclonePubKeyFilePath},
		Timeout: 0,
	}
	resultCatRclonePub := rcl.cmdExec.RunCommand(paramsCatRclonePub)
	if resultCatRclonePub.ExitCode != 0 {
		return "", "", fmt.Errorf("cat %s err: %d, %s", rclonePubKeyFilePath, resultCatRclonePub.ExitCode, resultCatRclonePub.ErrBuf.String())
	}
	rclonePubKeyData := resultCatRclonePub.OutBuf.String()
	rclonePrivateKeyData := resultCatRclone.OutBuf.String()

	return rclonePubKeyData, rclonePrivateKeyData, nil
}

func (rcl *Rclone) GenerateRcloneKeyConfigMap(ns string) *corev1.ConfigMap {

	var rcloneCM = &corev1.ConfigMap{}
	rclonePubKeyData, rclonePrivateKeyData, err := rcl.generateSSHPubAndPrivateKeyCM()

	if err != nil {
		logger.WithError(err).Errorf("generateRcloneKeyConfigMap generateSSHPubAndPrivateKeyCM")
		return rcloneCM
	}

	configData := map[string]string{
		"rclone.pub":   rclonePubKeyData,
		"rclone":       rclonePrivateKeyData,
		"rclonemerged": rclonePrivateKeyData + "\n" + rclonePubKeyData,
	}

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rcloneKeyConfigMapName,
			Namespace: ns,
		},
		Data: configData,
	}

	return configMap
}

func (rcl *Rclone) SrcMountPointToRemoteMountPoint(jobName, namespace, lvPoolName, lvName, srcNodeName, dstNodeName string, waitUntilSuccess bool, timeout time.Duration) error {
	jobStruct := rcl.getBaseJobStruct(jobName, namespace)

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
	jobStruct.Spec.Template.Spec.Affinity = affinity

	tmpDstMountPoint := dstMountPoint + lvName
	tmpSrcMountPoint := srcMountPoint + lvName

	logger.Debugf("DstMountPointBind  rclone sync  %s:%s %s:%s", RcloneSrcName, tmpSrcMountPoint, RcloneRemoteName, tmpDstMountPoint)
	//rcloneCommand := fmt.Sprintf("rclone mkdir %s; rclone mkdir %s; rclone mount --allow-non-empty --allow-other --daemon %s:%s %s; "+
	//	"rclone mount --allow-non-empty --allow-other --daemon %s:%s %s; rclone sync %s %s --progress; sleep 1800s;",
	rcloneCommand := fmt.Sprintf("sleep 18s; rclone sync %s:%s %s:%s --progress;", RcloneSrcName, tmpSrcMountPoint, RcloneRemoteName, tmpDstMountPoint)
	jobStruct.Spec.Template.Spec.Containers[0].Command = []string{"sh", "-c", rcloneCommand}

	if err := rcl.dcm.k8sControllerClient.Create(rcl.dcm.ctx, jobStruct); err != nil {
		if k8serrors.IsAlreadyExists(err) {
			fmt.Errorf("Failed to create MigrateJob, Job already exists")
		} else {
			return err
		}
	}

	return nil
}

func (rcl *Rclone) WaitMigrateJobTaskDone(jobName, lvName string, waitUntilSuccess bool, timeout time.Duration) error {

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
			return fmt.Errorf("Failed to copy data from srcMountPointName: %s to dstMountPointName: %s, timeout after %s", srcMountPoint+lvName, dstMountPoint+lvName, timeout)
		}
	}

	return nil
}

// User should fill up with Spec.Template.Spec.Containers[0].Command
func (rcl *Rclone) getBaseJobStruct(jobName, namespace string) *batchv1.Job {
	var privileged = true
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

	// Return without config
	if rcl.skipRcloneConfiguration {
		return baseStruct
	}

	rcloneConfigVolumeMount := corev1.VolumeMount{
		Name:      "rclone-config",
		MountPath: "/config/rclone/rclone.conf",
		SubPath:   "rclone.conf",
	}

	rcloneKeyConfigVolumeMount := corev1.VolumeMount{
		Name:      "rclone-key-config",
		MountPath: "/config/rclone/rclonemerged",
		SubPath:   "rclonemerged",
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
		{
			Name: "rclone-key-config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: rcl.rcloneKeyConfigMapName},
					Items: []corev1.KeyToPath{
						{
							Key:  rcl.rcloneCertKey,
							Path: "rclonemerged",
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
