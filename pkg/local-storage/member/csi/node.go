package csi

import (
	"fmt"
	"path"
	"strings"

	"github.com/hwameistor/hwameistor/pkg/local-storage/member/node/encrypt"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/container-storage-interface/spec/lib/go/csi"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	apis "github.com/hwameistor/hwameistor/pkg/apis/hwameistor"
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

var _ csi.NodeServer = (*plugin)(nil)

// NodeGetCapabilities - it will query node's capabilities
func (p *plugin) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	p.logger.Debug("NodeGetCapabilities")

	return &csi.NodeGetCapabilitiesResponse{Capabilities: p.nsCaps}, nil
}

// NodeGetInfo - it will query node's info
func (p *plugin) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	p.logger.Debug("NodeGetInfo")

	return &csi.NodeGetInfoResponse{
		NodeId: p.nodeName,
		// MaxVolumesPerNode: 1048576, // TODO: should set it? how?
		AccessibleTopology: &csi.Topology{
			Segments: map[string]string{
				apis.TopologyNodeKey: p.nodeName,
			}},
	}, nil
}

// NodeStageVolume - it will mount the volume to a global mountpoint which can be shared by multi-pods
func (p *plugin) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	p.logger.WithFields(log.Fields{"volume": req.VolumeId}).Debug("NodeStageVolume")

	// DO NOT SUPPORT: local volume can't be shared

	return nil, fmt.Errorf("not supported")
}

// NodeUnstageVolume - it will umount the volume from a global mountpoint
func (p *plugin) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	p.logger.WithFields(log.Fields{"volume": req.VolumeId}).Debug("NodeUnstageVolume")

	// DO NOT SUPPORT: local volume can't be shared

	return nil, fmt.Errorf("not supported")
}

// NodePublishVolume - it will mount the volume from a global mountpoint to the pod's mountpoint (bind mount)
func (p *plugin) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	p.logger.WithFields(log.Fields{
		"volume":         req.VolumeId,
		"targetPath":     req.TargetPath,
		"publishContext": req.PublishContext,
		"volmeContext":   req.VolumeContext,
		"secrets":        req.Secrets,
		"AccessMode":     req.VolumeCapability.AccessMode.Mode,
		"AccessType":     req.VolumeCapability.AccessType,
	}).Debug("NodePublishVolume")

	// req.StagingTargetPath should be empty as no NodeStageVolume
	resp := &csi.NodePublishVolumeResponse{}

	if req.GetVolumeId() == "" {
		p.logger.Error("Invalid volume id")
		return resp, fmt.Errorf("invalid volume id")
	}

	if req.GetTargetPath() == "" {
		p.logger.Error("Invalid target path")
		return resp, fmt.Errorf("invalid target path")
	}

	if req.GetVolumeCapability() == nil {
		p.logger.Error("invalid volume capability")
		return resp, fmt.Errorf("invalid volume capability")
	}

	// format the volume, and mount to the target path
	devicePath, ok := req.PublishContext[VolumeReplicaDevicePathKey]
	if !ok {
		p.logger.Error("not found volume replica device path")
		return resp, fmt.Errorf("not found device path")
	}

	encryptType := req.PublishContext[VolumeEncryptTypeKey]
	encryptSecret := req.PublishContext[VolumeEncryptSecretKey]

	// return directly if volume has already mounted at TargetPath
	yes, _ := p.mounter.IsMountPoint(req.TargetPath)
	if yes {
		p.logger.WithFields(log.Fields{
			"volume":     req.VolumeId,
			"targetPath": req.TargetPath,
			"devicePath": devicePath,
		}).Debug("device has already mounted at target path")
		return resp, nil
	}

	// mount volume
	volumeCap := req.GetVolumeCapability()
	switch volumeCap.GetAccessType().(type) {
	case *csi.VolumeCapability_Block:
		// raw block
		err := p.mounter.MountRawBlock(devicePath, req.TargetPath)
		if err != nil {
			return resp, err
		}
	case *csi.VolumeCapability_Mount:
		// filesystem block
		mnt := req.GetVolumeCapability().GetMount()
		// mount xfs with nouuid, just in case that uuid is already mounted, especially for volume restored from snapshot
		if mnt.FsType == "xfs" {
			mnt.MountFlags = append(mnt.MountFlags, "nouuid")
		}

		// decrypt the device and use the returned volume path
		var err error
		if encryptType != "" {
			if devicePath, err = openEncryptedDevice(p.apiClient, encryptType, devicePath, encryptSecret); err != nil {
				return resp, err
			}
		}

		if err = p.mounter.FormatAndMount(devicePath, req.TargetPath, mnt.FsType, mnt.MountFlags); err != nil {
			return resp, err
		}
	default:
		return resp, fmt.Errorf("invalid access type")
	}

	err := p.volumeQoSManager.RefreshQoSForLocalVolumeName(req.VolumeId)
	if err != nil {
		p.logger.WithFields(log.Fields{"volume": req.VolumeId, "error": err.Error()}).Error("Failed to refresh QoS for LocalVolume")
		return resp, err
	}

	// As we are supporting the restore of a volume to a bigger size and
	// creating bigger size clone from a volume, we need to check filesystem
	// resize is required, if required resize filesystem.
	if req.GetVolumeCapability().GetBlock() == nil {
		ok, err := p.mounter.NeedResize(devicePath, req.TargetPath)
		if err != nil {
			return resp, fmt.Errorf(
				"need resize check failed on devicePath %s and targetPath %s, error: %v",
				devicePath,
				req.TargetPath,
				err)
		}
		if ok {
			p.logger.WithFields(log.Fields{
				"volume":     req.VolumeId,
				"devicePath": devicePath,
				"targetPath": req.TargetPath,
			}).Info("Need to resize filesystem")
			err = p.expandFileSystemByMountPoint(req.TargetPath)
			if err != nil {
				return resp, fmt.Errorf("failed to expand file system by mount point %s", req.TargetPath)
			}
		}
	}

	return resp, nil
}

// NodeUnpublishVolume -  it will umount the volume from the pod's mountpoint
func (p *plugin) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	p.logger.WithFields(log.Fields{
		"volume":     req.VolumeId,
		"targetPath": req.TargetPath,
	}).Debug("NodeUnpublishVolume")

	resp := &csi.NodeUnpublishVolumeResponse{}

	if req.GetVolumeId() == "" {
		p.logger.Error("Invalid volume id")
		return resp, fmt.Errorf("invalid volume id")
	}

	if req.GetTargetPath() == "" {
		p.logger.Error("Invalid target path")
		return resp, fmt.Errorf("invalid target path")
	}

	// umount and delete the target path, more info see: #1322
	err := p.mounter.Unmount(req.TargetPath)
	if err != nil {
		p.logger.WithError(err).Error("Failed to unmount targetPath")
		return resp, err
	}

	return resp, p.closeEncryptedVolumeIfNeeded(req.VolumeId)
}

func (p *plugin) closeEncryptedVolumeIfNeeded(volumeName string) error {
	vol := &apisv1alpha1.LocalVolume{}
	if err := p.apiClient.Get(context.Background(), types.NamespacedName{Name: volumeName}, vol); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	if !vol.Spec.VolumeEncrypt.Enable {
		return nil
	}

	return encrypt.NewLUKS().CloseVolume(path.Join("/dev/mapper", volumeName+"-encrypt"))
}

// NodeGetVolumeStats - it will query volume status
func (p *plugin) NodeGetVolumeStats(ctx context.Context, req *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	logCtx := p.logger.WithFields(log.Fields{
		"volume":     req.VolumeId,
		"volumePath": req.VolumePath,
	})
	logCtx.Debug("NodeGetVolumeStats")

	resp := &csi.NodeGetVolumeStatsResponse{}
	vol := &apisv1alpha1.LocalVolume{}
	if err := p.apiClient.Get(ctx, types.NamespacedName{Name: req.VolumeId}, vol); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("Failed to query volume")
		} else {
			// not found volume, should be deleted already
			logCtx.WithError(err).Error("Not found volume")
		}

		return resp, err
	}

	if vol.Status.State != apisv1alpha1.VolumeStateReady {
		resp.VolumeCondition = &csi.VolumeCondition{
			Abnormal: true,
			Message:  "The volume is not ready",
		}
	} else {
		resp.VolumeCondition = &csi.VolumeCondition{
			Abnormal: false,
			Message:  "The volume is ready",
		}
	}

	// it's impossible to get usage of the raw block device
	metrics, err := getVolumeMetrics(req.VolumePath)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get volume metrics")
		return resp, err
	}

	logCtx.WithFields(log.Fields{"capacityUsed": metrics.UsedCapacityBytes, "iNodes": metrics.TotalINodeNumber, "iNodesUsed": metrics.UsedINodeNumber}).Debug("Got volume metrics")

	vol.Status.UsedCapacityBytes = metrics.UsedCapacityBytes
	vol.Status.TotalInodes = metrics.TotalINodeNumber
	vol.Status.UsedInodes = metrics.UsedINodeNumber
	if err = p.apiClient.Status().Update(ctx, vol); err != nil {
		logCtx.WithFields(log.Fields{"volume": vol.Name, "status": vol.Status}).WithError(err).Error("Failed to update LocalVolume capacity status")
		return resp, err
	}

	resp.Usage = []*csi.VolumeUsage{
		{
			Unit:      csi.VolumeUsage_BYTES,
			Total:     metrics.TotalCapacityBytes,
			Available: metrics.FreeCapacityBytes,
			Used:      metrics.UsedCapacityBytes,
		},
		{
			Unit:      csi.VolumeUsage_INODES,
			Total:     metrics.TotalINodeNumber,
			Available: metrics.FreeINodeNumber,
			Used:      metrics.UsedINodeNumber,
		},
	}

	return resp, nil
}

// NodeExpandVolume - it will expand a volume by rescanning block and resizing fs
func (p *plugin) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	p.logger.WithFields(log.Fields{
		"volume":                 req.VolumeId,
		"volumePath":             req.VolumePath,
		"stagingTargetPath":      req.StagingTargetPath,
		"Capacity.RequiredBytes": req.CapacityRange.RequiredBytes,
	}).Debug("NodeExpandVolume")

	resp := &csi.NodeExpandVolumeResponse{
		CapacityBytes: req.CapacityRange.RequiredBytes,
	}

	// expand fs only when volumeMode is not block
	if req.GetVolumeCapability().GetBlock() == nil {
		return resp, p.expandFileSystemByMountPoint(req.VolumePath)
	}

	return resp, nil
}

func getEncryptSecretKey(cli client.Client, namespacedName string) (string, error) {
	secret := &v1.Secret{}
	ss := strings.Split(namespacedName, "/")
	if len(ss) != 2 {
		return "", fmt.Errorf("invalid secret namespaced name %s", namespacedName)
	}

	key := client.ObjectKey{Name: ss[1], Namespace: ss[0]}
	err := cli.Get(context.Background(), key, secret)
	if err != nil {
		return "", err
	}

	if secret.Data == nil || secret.Data["key"] == nil {
		return "", fmt.Errorf("key is not exist in %s", namespacedName)
	}

	return string(secret.Data["key"]), nil
}

func openEncryptedDevice(cli client.Client, encryptType, devicePath, encryptSecret string) (string, error) {
	// decrypt the device and use the returned volume path
	switch encryptType {
	case "LUKS":
		if secret, err := getEncryptSecretKey(cli, encryptSecret); err != nil {
			return "", err
		} else {
			// overwrite devicePath with encrypted devicePath
			if devicePath, err = encrypt.NewLUKS().OpenVolume(devicePath, secret); err != nil {
				return "", err
			}
		}
	default:
		// do nothing
	}

	return devicePath, nil
}
