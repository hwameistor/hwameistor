//go:build linux
// +build linux

package qos

import (
	"k8s.io/mount-utils"
	utilexec "k8s.io/utils/exec"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

// IsFilesystemInitialized checks if the filesystem is initialized. It must be called before configuring QoS for a volume.
// If the filesystem is not initialized, we should skip volume QoS configuration in order to void the mkfs process hangs,
// See #958 for details.
func (m *VolumeQoSManager) IsFilesystemInitialized(replica *apisv1alpha1.LocalVolumeReplica) bool {
	mounter := mount.SafeFormatAndMount{
		Interface: mount.New("/bin/mount"),
		Exec:      utilexec.New(),
	}
	source := getVolumeDevicePath(replica)
	existingFormat, err := mounter.GetDiskFormat(source)
	if err != nil {
		return false
	}
	return existingFormat != ""
}
