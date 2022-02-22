package storage

import (
	udsv1alpha1 "github.com/HwameiStor/local-storage/pkg/apis/uds/v1alpha1"
)

func newLocalPoolExecutor(lm *LocalManager, volumeKind string) LocalPoolExecutor {
	if volumeKind == udsv1alpha1.VolumeKindDisk {
		return newDiskExecutor(lm)
	}
	if volumeKind == udsv1alpha1.VolumeKindLVM {
		return newLVMExecutor(lm)
	}
	if volumeKind == udsv1alpha1.VolumeKindRAM {
		return newRAMDiskExecutor(lm)
	}
	return nil
}

func newLocalVolumeExecutor(lm *LocalManager, volumeKind string) LocalVolumeExecutor {
	if volumeKind == udsv1alpha1.VolumeKindDisk {
		return newDiskExecutor(lm)
	}
	if volumeKind == udsv1alpha1.VolumeKindLVM {
		return newLVMExecutor(lm)
	}
	if volumeKind == udsv1alpha1.VolumeKindRAM {
		return newRAMDiskExecutor(lm)
	}
	return nil
}
