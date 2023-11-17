package api

import apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"

type Snapshot struct {
	apisv1alpha1.LocalVolumeSnapshot
}
type SnapshotList struct {
	Snapshots []*Snapshot `json:"items"`
	// page 信息
	Page *Pagination `json:"pagination,omitempty"`
}
