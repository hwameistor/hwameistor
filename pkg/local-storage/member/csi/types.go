package csi

// consts
const (
	VolumeReplicaDevicePathKey = "volumeReplicaDevicePath"
	VolumeReplicaNameKey       = "volumeReplicaName"
	VolumeReplicaKindKey       = "volumeReplicaKind"
	VolumeEncryptSecretKey     = "volumeEncryptSecret"
	VolumeEncryptTypeKey       = "volumeEncryptType"

	// AnnSelectedNode annotation is added to a PVC that has been triggered by scheduler to
	// be dynamically provisioned. Its value is the name of the selected node.
	AnnSelectedNode = "volume.kubernetes.io/selected-node"
)

// VolumeMetrics struct
type VolumeMetrics struct {
	TotalCapacityBytes int64
	UsedCapacityBytes  int64
	FreeCapacityBytes  int64

	TotalINodeNumber int64
	UsedINodeNumber  int64
	FreeINodeNumber  int64
}
