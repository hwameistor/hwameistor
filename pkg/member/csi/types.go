package csi

// consts
const (
	VolumeReplicaDevicePathKey = "volumeReplicaDevicePath"
	VolumeReplicaNameKey       = "volumeReplicaName"
	VolumeReplicaKindKey       = "volumeReplicaKind"
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
