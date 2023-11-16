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

	SyncSourceMountPoint = "/mnt/hwameistor/src/"
	SyncTargetMountPoint = "/mnt/hwameistor/dst/"

	SyncConfigSourceMountPointKey   = "sourceMountPoint"
	SyncConfigTargetMountPointKey   = "targetMountPoint"
	SyncConfigSourceNodeNameKey     = "sourceNode"
	SyncConfigTargetNodeNameKey     = "targetNode"
	SyncConfigVolumeNameKey         = "localVolume"
	SyncConfigSourceNodeReadyKey    = "sourceReady"
	SyncConfigTargetNodeReadyKey    = "targetReady"
	SyncConfigSourceNodeCompleteKey = "sourceCompleted"
	SyncConfigTargetNodeCompleteKey = "targetCompleted"
	SyncConfigSyncCompleteKey       = "syncCompleted"

	SyncTrue  string = "yes"
	SyncFalse string = "no"

	SyncJobFinalizer = "hwameistor.io/sync-job-protect"

	SyncToolJuiceSync = "juicesync"
)

type DataSyncer interface {
	Prepare(targetNodeName, sourceNodeName, lvName string) error
	StartSync(jobName, lvName, excludedRunningNodeName, runningNodeName string) error
}

func NewSyncer(syncerName string, namespace string, client k8sclient.Client) DataSyncer {
	return &JuiceSync{
		namespace: namespace,
		apiClient: client,
	}
}

func GetConfigMapName(str1, str2 string) string {
	return fmt.Sprintf("%s-%s", str1, str2)
}
