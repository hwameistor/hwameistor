package scheduler

import (
	framework "k8s.io/kubernetes/pkg/scheduler/framework"

	diskscheduler "github.com/hwameistor/hwameistor/pkg/local-disk-manager/csi/scheduler"
)

func NewDiskVolumeScheduler(f framework.Handle) VolumeScheduler {
	return diskscheduler.NewDiskVolumeSchedulerPlugin(f.SharedInformerFactory().Storage().V1().StorageClasses().Lister())
}
