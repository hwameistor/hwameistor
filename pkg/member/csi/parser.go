package csi

import (
	"fmt"
	"strconv"
	"strings"

	localstoragev1alpha1 "github.com/hwameistor/local-storage/pkg/apis/localstorage/v1alpha1"
	"github.com/hwameistor/local-storage/pkg/utils"
)

type volumeParameters struct {
	poolClass     string
	poolType      string
	poolName      string
	volumeKind    string
	replicaNumber int64
	striped       bool
	convertible   bool
}

func parseParameters(req RequestParameterHandler) (*volumeParameters, error) {
	params := req.GetParameters()

	poolClass, ok := params[localstoragev1alpha1.VolumeParameterPoolClassKey]
	if !ok {
		return nil, fmt.Errorf("not found pool class")
	}
	poolType, ok := params[localstoragev1alpha1.VolumeParameterPoolTypeKey]
	if !ok {
		return nil, fmt.Errorf("not found pool type")
	}
	volumeKind, ok := params[localstoragev1alpha1.VolumeParameterVolumeKindKey]
	if !ok {
		return nil, fmt.Errorf("not found pool kind")
	}
	poolName, err := utils.BuildStoragePoolName(poolClass, poolType)
	if err != nil {
		return nil, err
	}
	striped := false
	if stripedValue, ok := params[localstoragev1alpha1.VolumeParameterStriped]; ok && strings.ToLower(stripedValue) == "true" {
		striped = true
	}
	replicaNumberStr, ok := params[localstoragev1alpha1.VolumeParameterReplicaNumberKey]
	if !ok {
		return nil, fmt.Errorf("not found volume replica count")
	}
	replicaNumber, err := strconv.Atoi(replicaNumberStr)
	if err != nil {
		return nil, err
	}
	convertible := true
	// for HA volume, already be convertible
	if replicaNumber < 2 {
		convertibleValue, ok := params[localstoragev1alpha1.VolumeParameterConvertible]
		if !ok {
			// for non-HA volume, default to false
			convertible = false
		} else {
			if strings.ToLower(convertibleValue) != "true" {
				convertible = false
			}
		}
	}

	return &volumeParameters{
		poolClass:     poolClass,
		poolType:      poolType,
		poolName:      poolName,
		volumeKind:    volumeKind,
		striped:       striped,
		replicaNumber: int64(replicaNumber),
		convertible:   convertible,
	}, nil
}
