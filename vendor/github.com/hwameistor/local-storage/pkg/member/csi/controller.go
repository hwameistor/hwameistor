package csi

import (
	"fmt"
	"math"
	"time"

	localapis "github.com/hwameistor/local-storage/pkg/apis"
	apisv1alpha1 "github.com/hwameistor/local-storage/pkg/apis/hwameistor/v1alpha1"

	csi "github.com/container-storage-interface/spec/lib/go/csi"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

var _ csi.ControllerServer = (*plugin)(nil)

//ControllerGetCapabilities implementation
func (p *plugin) ControllerGetCapabilities(ctx context.Context, req *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	p.logger.Debug("ControllerGetCapabilities")

	return &csi.ControllerGetCapabilitiesResponse{Capabilities: p.csCaps}, nil
}

//CreateVolume implementation, idempotent
func (p *plugin) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	if req.VolumeContentSource != nil {
		if req.VolumeContentSource.GetSnapshot() != nil {
			return p.restoreVolumeFromSnapshot(ctx, req)
		}
		if req.VolumeContentSource.GetVolume() != nil {
			return p.cloneVolume(ctx, req)
		}
		p.logger.WithFields(log.Fields{"type": req.VolumeContentSource.Type}).Error("Invalid VolumeContentSource")
	}

	logCtx := p.logger.WithFields(log.Fields{
		"volume":           req.Name,
		"requiredCapacity": req.CapacityRange.RequiredBytes,
		"limitedCapacity":  req.CapacityRange.LimitBytes,
		"parameters":       req.Parameters,
		"topology":         req.AccessibilityRequirements,
		"capabilities":     req.VolumeCapabilities,
	})
	logCtx.Debug("CreateVolume")

	resp := &csi.CreateVolumeResponse{}
	params, err := parseParameters(req)
	if err != nil {
		p.logger.WithError(err).Error("Failed to parse parameters")
		return resp, err
	}
	for i := 0; i < 3; i++ {
		vol := &apisv1alpha1.LocalVolume{}
		if err := p.apiClient.Get(ctx, types.NamespacedName{Name: req.Name}, vol); err != nil {
			if !errors.IsNotFound(err) {
				p.logger.WithFields(log.Fields{"volName": req.Name, "error": err.Error()}).Error("Failed to query volume")
				return resp, err
			}
			vol.Name = req.Name
			vol.Spec.PoolName = params.poolName
			vol.Spec.ReplicaNumber = int64(params.replicaNumber)
			vol.Spec.RequiredCapacityBytes = req.CapacityRange.RequiredBytes
			vol.Spec.Convertible = params.convertible
			if req.AccessibilityRequirements != nil && len(req.AccessibilityRequirements.Requisite) == 1 {
				if nodeName, ok := req.AccessibilityRequirements.Requisite[0].Segments[localapis.TopologyNodeKey]; ok {
					vol.Spec.Accessibility.Node = nodeName
				}
			}

			p.logger.WithFields(log.Fields{"volume": vol}).Debug("Creating a volume")
			if err := p.apiClient.Create(ctx, vol); err != nil {
				p.logger.WithFields(log.Fields{"volume": vol, "error": err.Error()}).Error("Failed to create a volume")
				return resp, err
			}
		} else if vol.Status.State == apisv1alpha1.VolumeStateReady {
			resp.Volume = &csi.Volume{
				VolumeId:      vol.Name,
				CapacityBytes: vol.Status.AllocatedCapacityBytes,
				VolumeContext: req.Parameters,
			}
			return resp, nil
		}
		time.Sleep(5 * time.Second)
	}
	return resp, fmt.Errorf("volume is still in creating")
}

func (p *plugin) restoreVolumeFromSnapshot(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	logCtx := p.logger.WithFields(log.Fields{
		"volume":           req.Name,
		"snapshot":         req.VolumeContentSource.GetSnapshot().SnapshotId,
		"requiredCapacity": req.CapacityRange.RequiredBytes,
		"limitedCapacity":  req.CapacityRange.LimitBytes,
		"parameters":       req.Parameters,
		"topology":         req.AccessibilityRequirements,
		"capabilities":     req.VolumeCapabilities,
	})
	logCtx.Debug("restoreVolumeFromSnapshot")

	return &csi.CreateVolumeResponse{}, fmt.Errorf("not implemented")
}

func (p *plugin) cloneVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	logCtx := p.logger.WithFields(log.Fields{
		"volume":           req.Name,
		"snapshot":         req.VolumeContentSource.GetVolume().VolumeId,
		"requiredCapacity": req.CapacityRange.RequiredBytes,
		"limitedCapacity":  req.CapacityRange.LimitBytes,
		"parameters":       req.Parameters,
		"topology":         req.AccessibilityRequirements,
		"capabilities":     req.VolumeCapabilities,
	})
	logCtx.Debug("cloneVolume")

	return &csi.CreateVolumeResponse{}, fmt.Errorf("not implemented")
}

//ControllerGetVolume implementation, idempotent
func (p *plugin) ControllerGetVolume(ctx context.Context, req *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	p.logger.WithFields(log.Fields{"volume": req.VolumeId}).Debug("ControllerGetVolume")

	// get volume from apiClient
	resp := &csi.ControllerGetVolumeResponse{}
	vol := &apisv1alpha1.LocalVolume{}
	if err := p.apiClient.Get(ctx, types.NamespacedName{Name: req.VolumeId}, vol); err != nil {
		if !errors.IsNotFound(err) {
			p.logger.WithFields(log.Fields{"volName": req.VolumeId, "error": err.Error()}).Error("Failed to query volume")
			return resp, err
		}
		// not found volume, should be deleted already
		return resp, nil
	}

	resp.Volume = &csi.Volume{
		CapacityBytes:      vol.Status.AllocatedCapacityBytes,
		VolumeId:           req.VolumeId,
		AccessibleTopology: []*csi.Topology{},
	}
	resp.Status = &csi.ControllerGetVolumeResponse_VolumeStatus{
		PublishedNodeIds: []string{},
		VolumeCondition:  &csi.VolumeCondition{},
	}
	if vol.Status.PublishedNodeName != "" {
		resp.Status.PublishedNodeIds = append(resp.Status.PublishedNodeIds, vol.Status.PublishedNodeName)
	}
	volReplica := &apisv1alpha1.LocalVolumeReplica{}
	for _, replicaID := range vol.Status.Replicas {
		if err := p.apiClient.Get(ctx, types.NamespacedName{Name: replicaID}, volReplica); err != nil {
			p.logger.WithFields(log.Fields{"replica": replicaID, "error": err.Error()}).Error("Failed to query volume replica")
			return resp, err
		}
		if volReplica.Spec.NodeName == "" {
			continue
		}
		resp.Volume.AccessibleTopology = append(resp.Volume.AccessibleTopology,
			&csi.Topology{
				Segments: map[string]string{
					localapis.TopologyNodeKey: volReplica.Spec.NodeName,
				},
			},
		)
	}

	if vol.Status.State != apisv1alpha1.VolumeStateReady {
		resp.Status.VolumeCondition.Abnormal = true
		resp.Status.VolumeCondition.Message = "The volume is not ready"
	} else {
		resp.Status.VolumeCondition.Abnormal = false
		resp.Status.VolumeCondition.Message = "The volume is ready"
	}

	return resp, nil
}

//DeleteVolume implementation, idempotent
func (p *plugin) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	p.logger.WithFields(log.Fields{
		"volume":  req.VolumeId,
		"secrets": req.Secrets,
	}).Debug("DeleteVolume")

	resp := &csi.DeleteVolumeResponse{}
	vol := &apisv1alpha1.LocalVolume{}
	if err := p.apiClient.Get(ctx, types.NamespacedName{Name: req.VolumeId}, vol); err != nil {
		if !errors.IsNotFound(err) {
			p.logger.WithFields(log.Fields{"volName": req.VolumeId, "error": err.Error()}).Error("Failed to query volume")
			return resp, err
		}
		// not found volume, should be deleted already
		return resp, nil
	}

	if vol.Status.State == apisv1alpha1.VolumeStateDeleted {
		return resp, nil
	}
	if vol.Status.State == apisv1alpha1.VolumeStateToBeDeleted {
		return resp, fmt.Errorf("volume in deleting")
	}
	vol.Spec.Delete = true
	if err := p.apiClient.Update(ctx, vol); err != nil {
		p.logger.WithFields(log.Fields{"volName": vol.Name, "state": vol.Status.State, "error": err.Error()}).Error("Failed to set volume status")
		return resp, err
	}
	return resp, fmt.Errorf("volume in deleting")
}

//ControllerPublishVolume implementation, idempotent
func (p *plugin) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	p.logger.WithFields(log.Fields{
		"volume":        req.VolumeId,
		"node":          req.NodeId,
		"volumeContext": req.VolumeContext,
		"readonly":      req.Readonly,
		"secrets":       req.Secrets,
		"AccessMode":    req.VolumeCapability.AccessMode.Mode,
		"AccessType":    req.VolumeCapability.AccessType,
	}).Debug("ControllerPublishVolume")

	resp := &csi.ControllerPublishVolumeResponse{PublishContext: map[string]string{}}

	vol := &apisv1alpha1.LocalVolume{}
	if err := p.apiClient.Get(ctx, types.NamespacedName{Name: req.VolumeId}, vol); err != nil {
		p.logger.WithFields(log.Fields{"volName": req.VolumeId, "error": err.Error()}).Error("Failed to query volume")
		return resp, err
	}
	volReplica := &apisv1alpha1.LocalVolumeReplica{}
	for _, replicaName := range vol.Status.Replicas {
		if err := p.apiClient.Get(ctx, types.NamespacedName{Name: replicaName}, volReplica); err != nil {
			p.logger.WithFields(log.Fields{"replica": replicaName, "error": err.Error()}).Error("Failed to query volume replica")
			return resp, err
		}
		if volReplica.Spec.NodeName != req.NodeId {
			continue
		}
		if volReplica.Status.State != apisv1alpha1.VolumeReplicaStateReady {
			p.logger.WithFields(log.Fields{"replica": replicaName, "state": volReplica.Status.State}).Error("volume replica is not ready")
			return resp, fmt.Errorf("replica not ready")
		}
		if !volReplica.Status.Synced {
			p.logger.WithFields(log.Fields{"replica": replicaName, "synced": volReplica.Status.Synced}).Error("volume replica is out of date")
			return resp, fmt.Errorf("replica out of date")
		}
		if volReplica.Status.DevicePath == "" {
			p.logger.WithFields(log.Fields{"replica": replicaName, "devicePath": volReplica.Status.DevicePath}).Error("invalid volume replica device path")
			return resp, fmt.Errorf("invalid device path of volume replica")

		}
		vol.Status.PublishedNodeName = req.NodeId
		if err := p.apiClient.Status().Update(ctx, vol); err != nil {
			p.logger.WithFields(log.Fields{"volume": vol.Name, "node": req.NodeId}).Error("Failed to update volume with published node info")
			return resp, err
		}
		p.logger.WithFields(log.Fields{"volume": req.VolumeId, "node": req.NodeId, "devicePath": volReplica.Status.DevicePath}).Debug("Found valid volume replica")
		resp.PublishContext[VolumeReplicaDevicePathKey] = volReplica.Status.DevicePath
		resp.PublishContext[VolumeReplicaNameKey] = volReplica.Name
		return resp, nil
	}

	p.logger.WithFields(log.Fields{"volume": req.VolumeId, "node": req.NodeId}).Error("not found valid volume replica")
	return resp, fmt.Errorf("not found valid volume replica")
}

//ControllerUnpublishVolume implementation, idempotent
func (p *plugin) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	p.logger.WithFields(log.Fields{
		"volume":  req.VolumeId,
		"node":    req.NodeId,
		"secrets": req.Secrets,
	}).Debug("ControllerUnpublishVolume")

	resp := &csi.ControllerUnpublishVolumeResponse{}

	vol := &apisv1alpha1.LocalVolume{}
	if err := p.apiClient.Get(ctx, types.NamespacedName{Name: req.VolumeId}, vol); err != nil {
		p.logger.WithFields(log.Fields{"volume": req.VolumeId, "error": err.Error()}).Error("Failed to query volume")
		return resp, err
	}

	if vol.Status.PublishedNodeName != req.NodeId {
		p.logger.WithFields(log.Fields{"volume": req.VolumeId, "node": req.NodeId}).Error("Wrong published node in request")
		return resp, fmt.Errorf("wrong published node")
	}
	vol.Status.PublishedNodeName = ""
	if err := p.apiClient.Status().Update(ctx, vol); err != nil {
		p.logger.WithFields(log.Fields{"volume": vol.Name, "node": req.NodeId}).Error("Failed to update volume to unpublish node")
		return resp, err
	}

	return resp, nil
}

//ValidateVolumeCapabilities implementation
func (p *plugin) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	p.logger.WithFields(log.Fields{
		"volume":             req.VolumeId,
		"parameters":         req.Parameters,
		"secrets":            req.Secrets,
		"volumeCapabilities": req.VolumeCapabilities,
	}).Debug("ValidateVolumeCapabilities")

	resp := &csi.ValidateVolumeCapabilitiesResponse{}

	vol := &apisv1alpha1.LocalVolume{}
	if err := p.apiClient.Get(ctx, types.NamespacedName{Name: req.VolumeId}, vol); err != nil {
		if errors.IsNotFound(err) {
			p.logger.WithFields(log.Fields{"volName": req.VolumeId, "error": err.Error()}).Error("not found volume")

		} else {
			p.logger.WithFields(log.Fields{"volName": req.VolumeId, "error": err.Error()}).Error("Failed to query volume")
		}
		return resp, err
	}

	for _, reqCap := range req.VolumeCapabilities {
		if reqCap.AccessMode != nil && reqCap.AccessMode.Mode != csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER {
			return resp, nil
		}
	}
	resp.Confirmed = &csi.ValidateVolumeCapabilitiesResponse_Confirmed{VolumeCapabilities: p.vCaps}
	return resp, nil

}

//ListVolumes implementation, consider the case of multiple pages
func (p *plugin) ListVolumes(ctx context.Context, req *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	p.logger.WithFields(log.Fields{
		"maxEntries": req.MaxEntries,
		"token":      req.StartingToken,
	}).Debug("ListVolumes")

	// Limit: not support token/multi-pages implementation currently

	resp := &csi.ListVolumesResponse{
		Entries: []*csi.ListVolumesResponse_Entry{},
	}
	volumeList := &apisv1alpha1.LocalVolumeList{}
	if err := p.apiClient.List(ctx, volumeList); err != nil {
		p.logger.WithFields(log.Fields{"error": err}).Error("Failed to list volumes")
		return resp, err
	}

	replicaList := &apisv1alpha1.LocalVolumeReplicaList{}
	if err := p.apiClient.List(ctx, replicaList); err != nil {
		p.logger.WithFields(log.Fields{"error": err}).Error("Failed to list volume replicas")
		return resp, err
	}
	replicas := map[string]*apisv1alpha1.LocalVolumeReplica{}
	for i, replica := range replicaList.Items {
		replicas[replica.Name] = &replicaList.Items[i]
	}

	for i, vol := range volumeList.Items {
		// return all volumes when MaxEntries == 0, otherwise limited by MaxEntries
		if req.MaxEntries > 0 && int32(i) >= req.MaxEntries {
			break
		}
		entry := &csi.ListVolumesResponse_Entry{
			Volume: &csi.Volume{
				CapacityBytes:      vol.Status.AllocatedCapacityBytes,
				VolumeId:           vol.Name,
				AccessibleTopology: []*csi.Topology{},
			},
			Status: &csi.ListVolumesResponse_VolumeStatus{
				PublishedNodeIds: []string{},
				VolumeCondition:  &csi.VolumeCondition{},
			},
		}
		if vol.Status.PublishedNodeName != "" {
			entry.Status.PublishedNodeIds = append(entry.Status.PublishedNodeIds, vol.Status.PublishedNodeName)
		}
		if vol.Status.State != apisv1alpha1.VolumeStateReady {
			entry.Status.VolumeCondition.Abnormal = true
			entry.Status.VolumeCondition.Message = "The volume is not ready"
		} else {
			entry.Status.VolumeCondition.Abnormal = false
			entry.Status.VolumeCondition.Message = "The volume is ready"
		}
		for _, replicaName := range vol.Status.Replicas {
			if replica, ok := replicas[replicaName]; ok {
				if replica.Spec.NodeName != "" {
					entry.Volume.AccessibleTopology = append(entry.Volume.AccessibleTopology,
						&csi.Topology{
							Segments: map[string]string{
								localapis.TopologyNodeKey: replica.Spec.NodeName,
							},
						},
					)
				}
			}
		}
		resp.Entries = append(resp.Entries, entry)
	}

	return resp, nil
}

//GetCapacity implementation
func (p *plugin) GetCapacity(ctx context.Context, req *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	p.logger.WithFields(log.Fields{
		"volumeCapabilities": req.VolumeCapabilities,
		"parameters":         req.Parameters,
		"AccessibleTopology": req.AccessibleTopology.Segments,
	}).Debug("GetCapacity")

	resp := &csi.GetCapacityResponse{}

	// DO NOT SUPPORT: The scheduler should take it into account

	return resp, fmt.Errorf("not supported")
}

//CreateSnapshot implementation, idempotent
func (p *plugin) CreateSnapshot(ctx context.Context, req *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	p.logger.WithFields(log.Fields{
		"name":         req.Name,
		"sourceVolume": req.SourceVolumeId,
		"parameters":   req.Parameters,
		"secrets":      req.Secrets,
	}).Debug("CreateSnapshot")

	return nil, fmt.Errorf("not Implemented")
}

//DeleteSnapshot implementation, idempotent
func (p *plugin) DeleteSnapshot(ctx context.Context, req *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	p.logger.WithFields(log.Fields{
		"id":      req.SnapshotId,
		"secrets": req.Secrets,
	}).Debug("DeleteSnapshot")

	return nil, fmt.Errorf("not Implemented")
}

//ListSnapshots implementation, idempotent
func (p *plugin) ListSnapshots(ctx context.Context, req *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	p.logger.WithFields(log.Fields{
		"sourceVolume": req.SourceVolumeId,
		"snapshotID":   req.SnapshotId,
		"secrets":      req.Secrets,
		"maxEntries":   req.MaxEntries,
		"token":        req.StartingToken,
	}).Debug("ValidateVolumeCapabilities")

	return nil, fmt.Errorf("not Implemented")
}

//ControllerExpandVolume - it will expand volume size in storage pool, idempotent
func (p *plugin) ControllerExpandVolume(ctx context.Context, req *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	logCtx := p.logger.WithFields(log.Fields{"volume": req.VolumeId})
	logCtx.WithFields(log.Fields{
		"RequiredCapacity": req.CapacityRange.RequiredBytes,
		"AccessMode":       req.VolumeCapability.AccessMode.Mode,
		"AccessType":       req.VolumeCapability.AccessType,
	}).Debug("ControllerExpandVolume")

	resp := &csi.ControllerExpandVolumeResponse{}

	vol := &apisv1alpha1.LocalVolume{}
	if err := p.apiClient.Get(ctx, types.NamespacedName{Name: req.VolumeId}, vol); err != nil {
		logCtx.WithError(err).Error("Failed to query volume")
		return resp, err
	}

	if req.CapacityRange == nil {
		logCtx.Error("Invalid volume capacity request")
		return resp, fmt.Errorf("invalid capacity request")
	}

	// new capacity is less than the current, disallow it
	if (req.CapacityRange.RequiredBytes + apisv1alpha1.VolumeExpansionCapacityBytesMin) < vol.Status.AllocatedCapacityBytes {
		logCtx.WithFields(log.Fields{"currentCapacity": vol.Status.AllocatedCapacityBytes, "newCapacity": req.CapacityRange.RequiredBytes}).Error("Can't reduce volume capacity")
		return resp, fmt.Errorf("can't reduce capacity")
	}

	// new capacity is close to the current (diff < 10MB), treat it as a successful expansion
	if math.Abs(float64(req.CapacityRange.RequiredBytes-vol.Status.AllocatedCapacityBytes)) <= float64(apisv1alpha1.VolumeExpansionCapacityBytesMin) {
		logCtx.WithFields(log.Fields{"currentCapacity": vol.Status.AllocatedCapacityBytes, "newCapacity": req.CapacityRange.RequiredBytes}).Info("Volume capacity expand completed")
		resp.CapacityBytes = vol.Status.AllocatedCapacityBytes
		resp.NodeExpansionRequired = true
		return resp, nil
	}

	// new capacity is bigger than the current (diff > 10MB), expand it
	expand := &apisv1alpha1.LocalVolumeExpand{}
	if err := p.apiClient.Get(ctx, types.NamespacedName{Name: req.VolumeId}, expand); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("Failed to query volume expansion action")
			return resp, err
		}
		// no expansion in progress, create a new one
		expand.Name = req.VolumeId
		expand.Spec.VolumeName = req.VolumeId
		expand.Spec.RequiredCapacityBytes = req.CapacityRange.RequiredBytes
		if err := p.apiClient.Create(ctx, expand); err != nil {
			logCtx.WithError(err).Error("Failed to submit volume expansion request")
			return resp, err
		}
		logCtx.WithField("newCapacity", expand.Spec.RequiredCapacityBytes).Info("Submitted volume expansion request")
		// still return error, will check volume size at next visit
		return resp, fmt.Errorf("volume expansion not completed yet")
	}

	logCtx.WithFields(log.Fields{"spec": expand.Spec, "status": expand.Status}).Debug("Volume expansion is still in progress")
	return resp, fmt.Errorf("volume expansion in progress")
}
