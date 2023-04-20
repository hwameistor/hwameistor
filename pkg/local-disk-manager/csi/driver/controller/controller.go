package controller

import (
	"context"
	"fmt"
	volume "github.com/hwameistor/hwameistor/pkg/local-disk-manager/member/controller/volume"

	"github.com/container-storage-interface/spec/lib/go/csi"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils"
)

type Server struct {
	// supportControllerCapability
	supportControllerCapability []*csi.ControllerServiceCapability

	// vm manager volume create,delete,query
	vm volume.Manager
}

func NewServer() *Server {
	server := &Server{}
	server.vm = volume.New()
	server.initControllerCapability()
	return server
}

func (s *Server) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	log.Infof("Calling %s ...", utils.FuncName())

	// validate request
	if err := s.validateCreateVolumeRequest(req); err != nil {
		log.WithError(err).Error("CreateVolumeRequest is invalid")
		return nil, err
	}

	volume, err := s.vm.GetVolumeInfo(req.GetName())
	if err != nil {
		log.WithError(err).Error("Failed to get volume info")
		return nil, err
	}

	// The difference between Update and Create is the logic of select disk
	// In Create: we assume that the data volume is not bound to a disk,
	// so we will comprehensively consider the nodes filtered by the scheduler and
	// the internal parameters of the data volume request, and finally filter a qualified devPath
	//
	// In Update: we will not modify the topology info and the devPath of the volume
	if volume.Exist {
		if volume, err = s.vm.UpdateVolume(req.GetName(), req); err != nil {
			log.WithError(err).Error("Failed to UpdateVolume")
			return nil, err
		}
		log.Infof("Volume %s update success", req.GetName())
	} else {
		if volume, err = s.vm.CreateVolume(req.GetName(), req); err != nil {
			log.WithError(err).Error("Failed to CreateVolume")
			return nil, err
		}
		log.Infof("Volume %s created success", req.GetName())
	}

	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			CapacityBytes: volume.Capacity,
			VolumeId:      volume.Name,
			VolumeContext: volume.VolumeContext,
		},
	}, nil
}

func (s *Server) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	log.Infof("Calling %s ...", utils.FuncName())

	if req.GetVolumeId() == "" {
		return nil, fmt.Errorf("volumeid is empty")
	}

	return &csi.DeleteVolumeResponse{}, s.vm.DeleteVolume(ctx, req.GetVolumeId())
}

func (s *Server) ControllerPublishVolume(context.Context, *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	log.Infof("Calling %s ...", utils.FuncName())

	// just return success here

	return &csi.ControllerPublishVolumeResponse{}, nil
}

func (s *Server) ControllerUnpublishVolume(context.Context, *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	log.Infof("Calling %s ...", utils.FuncName())

	// just return success here

	return &csi.ControllerUnpublishVolumeResponse{}, nil
}
func (s *Server) ValidateVolumeCapabilities(context.Context, *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	log.Infof("Calling %s ...", utils.FuncName())

	return nil, fmt.Errorf("not implemented")
}

func (s *Server) ListVolumes(context.Context, *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	log.Infof("Calling %s ...", utils.FuncName())

	return nil, fmt.Errorf("not implemented")
}

func (s *Server) GetCapacity(context.Context, *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	log.Infof("Calling %s ...", utils.FuncName())

	return nil, fmt.Errorf("not implemented")
}

func (s *Server) ControllerGetCapabilities(context.Context, *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	log.Infof("Calling %s ...", utils.FuncName())
	return &csi.ControllerGetCapabilitiesResponse{Capabilities: s.supportControllerCapability}, nil
}

func (s *Server) CreateSnapshot(context.Context, *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	log.Infof("Calling %s ...", utils.FuncName())

	return nil, fmt.Errorf("not implemented")
}

func (s *Server) DeleteSnapshot(context.Context, *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	log.Infof("Calling %s ...", utils.FuncName())

	return nil, fmt.Errorf("not implemented")
}

func (s *Server) ListSnapshots(context.Context, *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	log.Infof("Calling %s ...", utils.FuncName())

	return nil, fmt.Errorf("not implemented")
}

func (s *Server) ControllerExpandVolume(context.Context, *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	log.Infof("Calling %s ...", utils.FuncName())

	return nil, fmt.Errorf("not implemented")
}

func (s *Server) ControllerGetVolume(context.Context, *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	log.Infof("Calling %s ...", utils.FuncName())

	return nil, fmt.Errorf("not implemented")
}

func (s *Server) getControllerCapability() []*csi.ControllerServiceCapability {
	return s.supportControllerCapability
}

func (s *Server) verifyControllerCapability(needCap csi.ControllerServiceCapability_RPC_Type) error {
	for _, existCap := range s.getControllerCapability() {
		if existCap.GetRpc().Type == needCap {
			return nil
		}
	}
	return status.Errorf(codes.InvalidArgument, "unsupported controller capability %s", needCap)
}

func (s *Server) validateCreateVolumeRequest(req *csi.CreateVolumeRequest) error {
	// verify Capability
	if err := s.verifyControllerCapability(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		return err
	}

	return nil
}

func (s *Server) initControllerCapability() {
	caps := []csi.ControllerServiceCapability_RPC_Type{
		// for volume
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		//ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
		csi.ControllerServiceCapability_RPC_LIST_VOLUMES,
		csi.ControllerServiceCapability_RPC_GET_VOLUME,
	}
	for _, c := range caps {
		s.supportControllerCapability = append(s.supportControllerCapability, newControllerServiceCapability(c))
	}
}

func newControllerServiceCapability(cap csi.ControllerServiceCapability_RPC_Type) *csi.ControllerServiceCapability {
	return &csi.ControllerServiceCapability{
		Type: &csi.ControllerServiceCapability_Rpc{
			Rpc: &csi.ControllerServiceCapability_RPC{
				Type: cap,
			},
		},
	}
}
