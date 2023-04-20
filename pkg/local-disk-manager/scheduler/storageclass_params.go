package scheduler

import (
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/member/controller/volume"
)

type StorageClassParams struct {
	DiskType string `json:"diskType"`
}

func parseParams(params map[string]string) *StorageClassParams {
	return &StorageClassParams{
		DiskType: params[volume.VolumeParameterDiskTypeKey],
	}
}
