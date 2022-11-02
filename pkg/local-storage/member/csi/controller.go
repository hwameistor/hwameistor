package csi

import (
	"fmt"
	"math"
	"strconv"
	"time"

	apis "github.com/hwameistor/hwameistor/pkg/apis/hwameistor"
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	csi "github.com/container-storage-interface/spec/lib/go/csi"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ csi.ControllerServer = (*plugin)(nil)

// ControllerGetCapabilities implementation
func (p *plugin) ControllerGetCapabilities(ctx context.Context, req *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	p.logger.Debug("ControllerGetCapabilities")

	return &csi.ControllerGetCapabilitiesResponse{Capabilities: p.csCaps}, nil
}

// CreateVolume implementation, idempotent
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

	lvg, err := p.getLocalVolumeGroupOrCreate(req, params)
	if err != nil {
		p.logger.WithError(err).Error("Failed to get or create LocalVolumeGroup")
		return resp, err
	}

	for i := 0; i < 2; i++ {
		vol := &apisv1alpha1.LocalVolume{}
		if err := p.apiClient.Get(ctx, types.NamespacedName{Name: req.Name}, vol); err != nil {
			if !errors.IsNotFound(err) {
				p.logger.WithFields(log.Fields{"volName": req.Name, "error": err.Error()}).Error("Failed to query volume")
				return resp, err
			}
			vol.Name = req.Name
			vol.Spec.PoolName = params.poolName
			vol.Spec.ReplicaNumber = params.replicaNumber
			vol.Spec.RequiredCapacityBytes = req.CapacityRange.RequiredBytes
			vol.Spec.Convertible = params.convertible
			vol.Spec.PersistentVolumeClaimName = params.pvcName
			vol.Spec.PersistentVolumeClaimNamespace = params.pvcNamespace
			vol.Spec.VolumeGroup = lvg.Name
			vol.Spec.Accessibility.Nodes = lvg.Spec.Accessibility.Nodes

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

func (p *plugin) getLocalVolumeGroupOrCreate(req *csi.CreateVolumeRequest, params *volumeParameters) (*apisv1alpha1.LocalVolumeGroup, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if req.AccessibilityRequirements == nil || len(req.AccessibilityRequirements.Requisite) != 1 {
		p.logger.WithFields(log.Fields{"volume": req.Name, "accessibility": req.AccessibilityRequirements}).Error("Not found accessibility requirements")
		return nil, fmt.Errorf("not found accessibility requirements")
	}
	requiredNodeName := req.AccessibilityRequirements.Requisite[0].Segments[apis.TopologyNodeKey]

	// fetch the local volume group by PVC
	lvg, err := p.getLocalVolumeGroupByPVC(params.pvcNamespace, params.pvcName)
	if err != nil {
		return nil, err
	}
	if lvg != nil && len(lvg.Name) > 0 {
		return lvg, nil
	}
	// not found the local volume group, create it
	p.logger.WithFields(log.Fields{"pvc": params.pvcName, "namespace": params.pvcNamespace}).Debug("Not found the LocalVolumeGroup")
	// get the pod with the volume firstly, and then get all the hwameistor volumes associated with the pod
	lvs, err := p.getAssociatedVolumes(params.pvcNamespace, params.pvcName)
	if err != nil {
		p.logger.WithFields(log.Fields{"pvc": params.pvcName, "namespace": params.pvcNamespace}).WithError(err).Error("Not found associated volumes")
		return nil, fmt.Errorf("not found associated volumes")
	}
	candidateNodes := p.storageMember.Controller().VolumeScheduler().GetNodeCandidates(lvs)
	selectedNodes := []string{}
	foundThisNode := false
	for _, nn := range candidateNodes {
		if len(selectedNodes) == int(params.replicaNumber) {
			break
		}
		if nn.Name == requiredNodeName {
			foundThisNode = true
			selectedNodes = append(selectedNodes, nn.Name)
		} else {
			if len(selectedNodes) == int(params.replicaNumber-1) {
				if foundThisNode {
					selectedNodes = append(selectedNodes, nn.Name)
				}
			} else {
				selectedNodes = append(selectedNodes, nn.Name)
			}
		}
	}
	if !foundThisNode {
		p.logger.WithField("requireNode", requiredNodeName).Errorf("requireNode is not exist in candidateNodes")
		return nil, fmt.Errorf("requireNode %s is not ready", requiredNodeName)
	}
	if len(selectedNodes) < int(params.replicaNumber) {
		p.logger.WithFields(log.Fields{"nodes": selectedNodes, "replica": params.replicaNumber}).Error("No enough nodes")
		return nil, fmt.Errorf("no enough nodes")
	}
	lvg = &apisv1alpha1.LocalVolumeGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name: genUUID(),
		},
		Spec: apisv1alpha1.LocalVolumeGroupSpec{
			Namespace: params.pvcNamespace,
			Volumes:   []apisv1alpha1.VolumeInfo{},
			Accessibility: apisv1alpha1.AccessibilityTopology{
				Nodes: selectedNodes,
			},
		},
	}
	for _, lv := range lvs {
		lvg.Spec.Volumes = append(lvg.Spec.Volumes, apisv1alpha1.VolumeInfo{PersistentVolumeClaimName: lv.Spec.PersistentVolumeClaimName})
	}
	log.WithFields(log.Fields{"lvg": lvg.Name}).Debug("Creating a new LVG ...")
	if err := p.apiClient.Create(context.Background(), lvg, &client.CreateOptions{}); err != nil {
		log.WithField("lvg", lvg.Name).WithError(err).Error("Failed to create LVG")
		return nil, err
	}

	return lvg, nil
}

func (p *plugin) getLocalVolumeGroupByPVC(pvcNamespace string, pvcName string) (*apisv1alpha1.LocalVolumeGroup, error) {
	lvgList := apisv1alpha1.LocalVolumeGroupList{}
	if err := p.apiClient.List(context.Background(), &lvgList, &client.ListOptions{}); err != nil {
		return nil, err
	}
	for i, lvg := range lvgList.Items {
		if lvg.Spec.Namespace != pvcNamespace {
			continue
		}
		for _, vol := range lvg.Spec.Volumes {
			if vol.PersistentVolumeClaimName == pvcName {
				return &lvgList.Items[i], nil
			}
		}
	}
	return &apisv1alpha1.LocalVolumeGroup{}, nil
}

func (p *plugin) getAssociatedVolumes(namespace string, pvcName string) ([]*apisv1alpha1.LocalVolume, error) {
	podList := corev1.PodList{}
	if err := p.apiClient.List(context.Background(), &podList, &client.ListOptions{Namespace: namespace}); err != nil {
		p.logger.WithError(err).Error("Failed to list Pods")
		return []*apisv1alpha1.LocalVolume{}, err
	}
	for i, pod := range podList.Items {
		for _, vol := range pod.Spec.Volumes {
			if vol.PersistentVolumeClaim == nil {
				continue
			}
			if vol.PersistentVolumeClaim.ClaimName == pvcName {
				return p.getHwameiStorPVCs(&podList.Items[i])
			}
		}
	}
	log.Debug("getAssociatedVolumes return empty")
	return []*apisv1alpha1.LocalVolume{}, nil

}

func (p *plugin) getHwameiStorPVCs(pod *corev1.Pod) ([]*apisv1alpha1.LocalVolume, error) {
	lvs := []*apisv1alpha1.LocalVolume{}
	p.logger.WithField("pog", pod.Name).Debug("Query hwameistor PVCs")

	ctx := context.Background()
	for _, vol := range pod.Spec.Volumes {
		if vol.PersistentVolumeClaim == nil {
			continue
		}
		pvc := &corev1.PersistentVolumeClaim{}
		if err := p.apiClient.Get(ctx, types.NamespacedName{Namespace: pod.Namespace, Name: vol.PersistentVolumeClaim.ClaimName}, pvc); err != nil {
			// if pvc can't be found in the cluster, the pod should not be able to be scheduled
			p.logger.WithField("pvc", vol.PersistentVolumeClaim.ClaimName).WithError(err).Error("Failed to get PVC")
			return lvs, err
		}
		if pvc.Spec.StorageClassName == nil {
			// should not be the CSI pvc, ignore
			continue
		}
		sc := &storagev1.StorageClass{}
		if err := p.apiClient.Get(ctx, types.NamespacedName{Name: *pvc.Spec.StorageClassName}, sc); err != nil {
			// can't found storageclass in the cluster, the pod should not be able to be scheduled
			p.logger.WithField("storageclass", *pvc.Spec.StorageClassName).WithError(err).Error("Failed to get StorageClass")
			return lvs, err
		}
		if sc.Provisioner == apisv1alpha1.CSIDriverName {
			if lv, err := constructLocalVolumeForPVC(pvc, sc); err == nil {
				lvs = append(lvs, lv)
			}
		}
	}
	return lvs, nil
}

func constructLocalVolumeForPVC(pvc *corev1.PersistentVolumeClaim, sc *storagev1.StorageClass) (*apisv1alpha1.LocalVolume, error) {

	lv := apisv1alpha1.LocalVolume{}
	poolName, err := buildStoragePoolName(
		sc.Parameters[apisv1alpha1.VolumeParameterPoolClassKey],
		sc.Parameters[apisv1alpha1.VolumeParameterPoolTypeKey])
	if err != nil {
		return nil, err
	}

	lv.Spec.PersistentVolumeClaimNamespace = pvc.Namespace
	lv.Spec.PersistentVolumeClaimName = pvc.Name
	lv.Spec.PoolName = poolName
	storage := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
	lv.Spec.RequiredCapacityBytes = storage.Value()
	replica, _ := strconv.Atoi(sc.Parameters[apisv1alpha1.VolumeParameterReplicaNumberKey])
	lv.Spec.ReplicaNumber = int64(replica)
	return &lv, nil
}

func buildStoragePoolName(poolClass string, poolType string) (string, error) {

	if poolClass == apisv1alpha1.DiskClassNameHDD && poolType == apisv1alpha1.PoolTypeRegular {
		return apisv1alpha1.PoolNameForHDD, nil
	}
	if poolClass == apisv1alpha1.DiskClassNameSSD && poolType == apisv1alpha1.PoolTypeRegular {
		return apisv1alpha1.PoolNameForSSD, nil
	}
	if poolClass == apisv1alpha1.DiskClassNameNVMe && poolType == apisv1alpha1.PoolTypeRegular {
		return apisv1alpha1.PoolNameForNVMe, nil
	}

	return "", fmt.Errorf("invalid pool info")
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

// ControllerGetVolume implementation, idempotent
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
					apis.TopologyNodeKey: volReplica.Spec.NodeName,
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

// DeleteVolume implementation, idempotent
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

// ControllerPublishVolume implementation, idempotent
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
		if req.VolumeCapability.GetBlock() != nil {
			vol.Status.PublishedRawBlock = true
		} else if req.VolumeCapability.GetMount() != nil {
			if len(req.VolumeCapability.GetMount().FsType) == 0 {
				vol.Status.PublishedFSType = "ext4"
			} else {
				vol.Status.PublishedFSType = req.VolumeCapability.GetMount().FsType
			}
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

// ControllerUnpublishVolume implementation, idempotent
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

	// when publish node is empty, return directly
	// https://github.com/hwameistor/hwameistor/issues/296
	if vol.Status.PublishedNodeName == "" {
		p.logger.WithFields(log.Fields{"volume": req.VolumeId, "node": req.NodeId}).Debug("Volume has already unpublished")
		return resp, nil
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

// ValidateVolumeCapabilities implementation
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

// ListVolumes implementation, consider the case of multiple pages
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
								apis.TopologyNodeKey: replica.Spec.NodeName,
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

// GetCapacity implementation
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

// CreateSnapshot implementation, idempotent
func (p *plugin) CreateSnapshot(ctx context.Context, req *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	p.logger.WithFields(log.Fields{
		"name":         req.Name,
		"sourceVolume": req.SourceVolumeId,
		"parameters":   req.Parameters,
		"secrets":      req.Secrets,
	}).Debug("CreateSnapshot")

	return nil, fmt.Errorf("not Implemented")
}

// DeleteSnapshot implementation, idempotent
func (p *plugin) DeleteSnapshot(ctx context.Context, req *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	p.logger.WithFields(log.Fields{
		"id":      req.SnapshotId,
		"secrets": req.Secrets,
	}).Debug("DeleteSnapshot")

	return nil, fmt.Errorf("not Implemented")
}

// ListSnapshots implementation, idempotent
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

// ControllerExpandVolume - it will expand volume size in storage pool, idempotent
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
