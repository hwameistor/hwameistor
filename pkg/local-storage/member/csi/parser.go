package csi

import (
	"fmt"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"strconv"
	"strings"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-storage/utils"
)

const (
	pvcNameKey      = "csi.storage.k8s.io/pvc/name"
	pvcNamespaceKey = "csi.storage.k8s.io/pvc/namespace"
)

type volumeParameters struct {
	poolClass     string
	poolType      string
	poolName      string
	replicaNumber int64
	convertible   bool
	pvcName       string
	pvcNamespace  string
	throughput    string
	iops          string
	snapshot      string
}

func parseParameters(req *csi.CreateVolumeRequest) (*volumeParameters, error) {
	params := req.GetParameters()

	poolClass, ok := params[apisv1alpha1.VolumeParameterPoolClassKey]
	if !ok {
		return nil, fmt.Errorf("not found pool class")
	}
	poolName, err := utils.BuildStoragePoolName(poolClass)
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

	pvcNamespace, ok := params[pvcNamespaceKey]
	if !ok {
		return nil, fmt.Errorf("not found pvc namespace")
	}
	pvcName, ok := params[pvcNameKey]
	if !ok {
		return nil, fmt.Errorf("not found pvc name")
	}

	snapshot := ""
	if req.VolumeContentSource != nil && req.VolumeContentSource.GetSnapshot() != nil {
		snapshot = req.VolumeContentSource.GetSnapshot().SnapshotId
	}

	return &volumeParameters{
		poolClass:     poolClass,
		// poolType:      poolType,
		poolName:      poolName,
		replicaNumber: int64(replicaNumber),
		convertible:   convertible,
		pvcNamespace:  pvcNamespace,
		pvcName:       pvcName,
		throughput:    params[apisv1alpha1.VolumeParameterThroughput],
		iops:          params[apisv1alpha1.VolumeParameterIOPS],
		snapshot:      snapshot,
	}, nil
}
