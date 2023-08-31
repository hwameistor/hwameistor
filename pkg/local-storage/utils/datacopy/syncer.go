package datacopy

import (
	"fmt"

	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	SyncSrcName        = "source"
	SyncRemoteName     = "remote"
	SyncConfigMapName  = "sync-config"
	SyncJobLabelApp    = "hwameistor-datasync"
	SyncJobAffinityKey = "kubernetes.io/hostname"

	SyncSrcMountPoint = "/mnt/hwameistor/src/"
	SyncDstMountPoint = "/mnt/hwameistor/dst/"

	SyncConfigSrcNodeNameKey     = "sourceNode"
	SyncConfigDstNodeNameKey     = "targetNode"
	SyncConfigVolumeNameKey      = "localVolume"
	SyncConfigSourceNodeReadyKey = "sourceReady"
	SyncConfigRemoteNodeReadyKey = "targetReady"
	SyncConfigSyncDoneKey        = "completed"

	SyncTrue  string = "yes"
	SyncFalse string = "no"

	SyncJobFinalizer = "hwameistor.io/sync-job-protect"

	SyncToolRClone    = "rclone"
	SyncToolJuiceSync = "juicesync"
)

type DataSyncer interface {
	Prepare(targetNodeName, sourceNodeName, lvName string) error
	StartSync(jobName, lvName, excludedRunningNodeName, runningNodeName string) error
}

func NewSyncer(syncerName string, namespace string, client k8sclient.Client) DataSyncer {
	if syncerName == SyncToolJuiceSync {
		return &JuiceSync{
			namespace: namespace,
			apiClient: client,
		}
	}

	return &RClone{
		namespace: namespace,
		apiClient: client,
	}
}

func GetConfigMapName(str1, str2 string) string {
	return fmt.Sprintf("%s-%s", str1, str2)
}
