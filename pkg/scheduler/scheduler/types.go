package scheduler

import (
	v1 "k8s.io/api/core/v1"
)

// MaxHAVolumeCount
const MaxHAVolumeCount = 1000

//go:generate mockgen -source=types.go -destination=../genscheduler/volume_scheduler.go  -package=genscheduler
type VolumeScheduler interface {
	Filter(existingLocalVolume []string, unboundPVCs []*v1.PersistentVolumeClaim, node *v1.Node) (bool, error)
	Reserve(unboundPVCs []*v1.PersistentVolumeClaim, node string) error
	Unreserve(unboundPVCs []*v1.PersistentVolumeClaim, node string) error
	Score(unboundPVCs []*v1.PersistentVolumeClaim, node string) (int64, error)
	CSIDriverName() string
}
