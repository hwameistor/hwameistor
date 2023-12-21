package api

import (
	snapshotv1 "github.com/kubernetes-csi/external-snapshotter/client/v6/apis/volumesnapshot/v1"
)

type VolumeSnapshot struct {
	snapshotv1.VolumeSnapshot
}

type VolumeSnapshotItemsList struct {
	VolumeSnapshots []*VolumeSnapshot `json:"volumeSnapshots"`
}

type VolumeSnapshotSnapshotList struct {
	VolumeSnapshots []*VolumeSnapshot `json:"volumeSnapshots"`
	Page            *Pagination       `json:"pagination,omitempty"`
}
