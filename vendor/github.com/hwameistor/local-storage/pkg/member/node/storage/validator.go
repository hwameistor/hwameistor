package storage

import (
	"fmt"

	apisv1alpha1 "github.com/hwameistor/local-storage/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/local-storage/pkg/utils"
)

type validator struct{}

func newValidator() *validator {
	return &validator{}
}

func (cr *validator) checkReplicaExists(vr *apisv1alpha1.LocalVolumeReplica, reg LocalRegistry) bool {
	return reg.HasVolumeReplica(vr)
}

func (cr *validator) checkVolumeReplicaStruct(vr *apisv1alpha1.LocalVolumeReplica) error {
	if vr == nil {
		return fmt.Errorf("failed to exec create replica. Invalid VolumeReplica given")
	}

	if len(vr.Spec.VolumeName) == 0 || len(vr.Spec.PoolName) == 0 || vr.Spec.RequiredCapacityBytes < 0 {
		return fmt.Errorf("failed to exec create replica. Invalid arguments given. name:%s, pool:%s, size:%d", vr.Spec.VolumeName, vr.Spec.PoolName, vr.Spec.RequiredCapacityBytes)
	}

	return nil
}

func (cr *validator) canCreateVolumeReplica(vr *apisv1alpha1.LocalVolumeReplica, reg LocalRegistry) error {
	if err := cr.checkVolumeReplicaStruct(vr); err != nil {
		return err
	}
	if cr.checkReplicaExists(vr, reg) {
		return ErrorReplicaExists
	}
	if err := cr.checkPoolVolumeCount(vr, reg); err != nil {
		return err
	}
	if err := cr.checkPoolCapacity(vr, reg); err != nil {
		return err
	}
	if err := cr.checkPerVolumeCapacityLimit(vr, reg); err != nil {
		return err
	}
	return nil
}

func (cr *validator) canDeleteVolumeReplica(vr *apisv1alpha1.LocalVolumeReplica, reg LocalRegistry) error {
	if err := cr.checkVolumeReplicaStruct(vr); err != nil {
		return err
	}
	if !cr.checkReplicaExists(vr, reg) {
		return ErrorReplicaNotFound
	}
	return nil
}

func (cr *validator) canExpandVolumeReplica(vr *apisv1alpha1.LocalVolumeReplica, newCapacityBytes int64, reg LocalRegistry) error {
	if err := cr.checkVolumeReplicaStruct(vr); err != nil {
		return err
	}
	if !cr.checkReplicaExists(vr, reg) {
		return ErrorReplicaNotFound
	}

	// validate for to-extend capacity
	newVr := vr.DeepCopy()
	newVr.Spec.RequiredCapacityBytes = newCapacityBytes - vr.Status.AllocatedCapacityBytes
	if err := cr.checkPoolCapacity(newVr, reg); err != nil {
		return err
	}
	newVr.Spec.RequiredCapacityBytes = newCapacityBytes
	if err := cr.checkPerVolumeCapacityLimit(newVr, reg); err != nil {
		return err
	}
	return nil
}

func (cr *validator) checkPoolVolumeCount(vr *apisv1alpha1.LocalVolumeReplica, reg LocalRegistry) error {
	pools := reg.Pools()
	if pool, has := pools[vr.Spec.PoolName]; has {
		if pool.FreeVolumeCount <= 0 {
			return ErrorInsufficientRequestResources
		}
	} else {
		return ErrorPoolNotFound
	}
	return nil
}

func (cr *validator) checkPoolCapacity(vr *apisv1alpha1.LocalVolumeReplica, reg LocalRegistry) error {
	pools := reg.Pools()
	if pool, has := pools[vr.Spec.PoolName]; has {
		if pool.FreeCapacityBytes < utils.NumericToLVMBytes(vr.Spec.RequiredCapacityBytes) {
			return ErrorInsufficientRequestResources
		}
	} else {
		return ErrorPoolNotFound
	}
	return nil
}

func (cr *validator) checkPerVolumeCapacityLimit(vr *apisv1alpha1.LocalVolumeReplica, reg LocalRegistry) error {
	pools := reg.Pools()
	if pool, has := pools[vr.Spec.PoolName]; has {
		if pool.VolumeCapacityBytesLimit < utils.NumericToLVMBytes(vr.Spec.RequiredCapacityBytes) {
			return ErrorOverLimitedRequestResource
		}
	} else {
		return ErrorPoolNotFound
	}
	return nil
}
