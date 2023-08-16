package qos

import (
	"context"
	"fmt"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/mount-utils"
	utilexec "k8s.io/utils/exec"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type VolumeQoSManager struct {
	nodeName string
	cgroups  VolumeCgroupsManager

	client client.Client
}

func NewVolumeQoSManager(nodeName string, client client.Client) (*VolumeQoSManager, error) {
	cgroups, err := NewVolumeCgroupsManager()
	if err != nil {
		return nil, err
	}
	m := &VolumeQoSManager{
		nodeName: nodeName,
		cgroups:  cgroups,
		client:   client,
	}
	return m, nil
}

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

// RefreshQoSForLocalVolume re-configures the QoS for a volume.
func (m *VolumeQoSManager) RefreshQoSForLocalVolumeName(volumeName string) error {
	replicas := &apisv1alpha1.LocalVolumeReplicaList{}
	err := m.client.List(context.TODO(), replicas, &client.ListOptions{})
	if err != nil {
		return err
	}

	var (
		targetReplica *apisv1alpha1.LocalVolumeReplica
		found         bool
	)

	for _, replica := range replicas.Items {
		if replica.Spec.VolumeName == volumeName && replica.Spec.NodeName == m.nodeName {
			targetReplica = &replica
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("failed to find the replica for volume %s on node %s", volumeName, m.nodeName)
	}
	return m.ConfigureQoSForLocalVolumeReplica(targetReplica)
}

// ConfigureQoSForLocalVolumeReplica configures the QoS for a volume.
func (m *VolumeQoSManager) ConfigureQoSForLocalVolumeReplica(replica *apisv1alpha1.LocalVolumeReplica) error {
	devPath := getVolumeDevicePath(replica)
	iops, throughput, err := parseVolumeQoSValues(replica.Spec.VolumeQoS)
	if err != nil {
		return err
	}
	return m.cgroups.ConfigureQoSForDevice(devPath, iops, throughput)
}

// parseVolumeQoSValues parses the volume QoS values.
func parseVolumeQoSValues(qos apisv1alpha1.VolumeQoS) (int64, int64, error) {
	var (
		iops       = resource.MustParse("0")
		throughput = resource.MustParse("0")
		err        error
	)

	if qos.IOPS != "" {
		iops, err = resource.ParseQuantity(qos.IOPS)
		if err != nil {
			return 0, 0, err
		}
	}
	if qos.Throughput != "" {
		throughput, err = resource.ParseQuantity(qos.Throughput)
		if err != nil {
			return 0, 0, err
		}
	}
	return iops.Value(), throughput.Value(), nil
}

// getVolumeDevicePath returns the device path of a volume.
func getVolumeDevicePath(replica *apisv1alpha1.LocalVolumeReplica) string {
	storagePath := replica.Status.StoragePath
	if len(storagePath) != 0 {
		return storagePath
	}
	return replica.Status.DevicePath
}
