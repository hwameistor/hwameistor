package api

import (
	snapshotv1 "github.com/kubernetes-csi/external-snapshotter/client/v6/apis/volumesnapshot/v1"
)

type SnapshotClass struct {
	snapshotv1.VolumeSnapshotClass
}

type SnapshotClassItemsList struct {
	SnapshotClasses []*SnapshotClass `json:"snapshot_classes"`
}

type SnapshotClassList struct {
	StorageClasses []*SnapshotClass `json:"storage_classes"`
	Page           *Pagination      `json:"pagination,omitempty"`
}
