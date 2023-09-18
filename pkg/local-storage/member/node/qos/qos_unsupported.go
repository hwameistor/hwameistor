//go:build !linux
// +build !linux

package qos

import (
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

// IsFilesystemInitialized always returns a false value on unsupported platforms
func (m *VolumeQoSManager) IsFilesystemInitialized(replica *apisv1alpha1.LocalVolumeReplica) bool {
	return false
}
