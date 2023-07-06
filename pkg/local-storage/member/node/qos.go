package node

import apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"

func (m *manager) configureQoS(replica *apisv1alpha1.LocalVolumeReplica) error {
	// If the volume is not formatted, we will skip volume QoS configuration in order to void
	// the mkfs process hangs.
	if !m.volumeQoSManager.IsFilesystemInitialized(replica) {
		// It will be retried on the csi.NodePublishVolume as a workaround.
		// TODO: remove this workaround if we found a better solution.
		return nil
	}
	return m.volumeQoSManager.ConfigureQoSForLocalVolumeReplica(replica)
}
