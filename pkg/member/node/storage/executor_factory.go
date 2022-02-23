package storage

import (
	localstoragev1alpha1 "github.com/hwameistor/local-storage/pkg/apis/localstorage/v1alpha1"
)

func newLocalPoolExecutor(lm *LocalManager, volumeKind string) LocalPoolExecutor {
	if volumeKind == localstoragev1alpha1.VolumeKindDisk {
		return newDiskExecutor(lm)
	}
	if volumeKind == localstoragev1alpha1.VolumeKindLVM {
		return newLVMExecutor(lm)
	}
	if volumeKind == localstoragev1alpha1.VolumeKindRAM {
		return newRAMDiskExecutor(lm)
	}
	return nil
}

func newLocalVolumeExecutor(lm *LocalManager, volumeKind string) LocalVolumeExecutor {
	if volumeKind == localstoragev1alpha1.VolumeKindDisk {
		return newDiskExecutor(lm)
	}
	if volumeKind == localstoragev1alpha1.VolumeKindLVM {
		return newLVMExecutor(lm)
	}
	if volumeKind == localstoragev1alpha1.VolumeKindRAM {
		return newRAMDiskExecutor(lm)
	}
	return nil
}
