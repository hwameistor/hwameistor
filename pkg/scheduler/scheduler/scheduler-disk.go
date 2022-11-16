package scheduler

import (
	diskscheduler "github.com/hwameistor/hwameistor/pkg/local-disk-manager/csi/scheduler"
	framework "k8s.io/kubernetes/pkg/scheduler/framework"
)

func NewDiskVolumeScheduler(f framework.Handle) VolumeScheduler {
	return diskscheduler.NewDiskVolumeSchedulerPlugin(f.SharedInformerFactory().Storage().V1().StorageClasses().Lister())
}
