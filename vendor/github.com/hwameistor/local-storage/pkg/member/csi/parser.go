package csi

import (
	"fmt"
	"strconv"
	"strings"

	apisv1alpha1 "github.com/hwameistor/local-storage/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/local-storage/pkg/utils"
)

type volumeParameters struct {
	poolClass     string
	poolType      string
	poolName      string
	replicaNumber int64
	convertible   bool
}

func parseParameters(req RequestParameterHandler) (*volumeParameters, error) {
	params := req.GetParameters()

	poolClass, ok := params[apisv1alpha1.VolumeParameterPoolClassKey]
	if !ok {
		return nil, fmt.Errorf("not found pool class")
	}
	poolType, ok := params[apisv1alpha1.VolumeParameterPoolTypeKey]
	if !ok {
		return nil, fmt.Errorf("not found pool type")
	}
	poolName, err := utils.BuildStoragePoolName(poolClass, poolType)
	if err != nil {
		return nil, err
	}
	replicaNumberStr, ok := params[apisv1alpha1.VolumeParameterReplicaNumberKey]
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
		convertibleValue, ok := params[apisv1alpha1.VolumeParameterConvertible]
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
		replicaNumber: int64(replicaNumber),
		convertible:   convertible,
	}, nil
}
