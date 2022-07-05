package controller

import (
	"context"
	"fmt"

	. "github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/csi/volumemanager"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	// supportControllerCapability
	supportControllerCapability []*ControllerServiceCapability

	// vm manager volume create,delete,query
	vm volumemanager.VolumeManager
}

func NewServer() *Server {
	server := &Server{}
	server.vm = volumemanager.NewLocalDiskVolumeManager()
	server.initControllerCapability()
	return server
}

func (s *Server) CreateVolume(ctx context.Context, req *CreateVolumeRequest) (*CreateVolumeResponse, error) {
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

	return &CreateVolumeResponse{
		Volume: &Volume{
			CapacityBytes: volume.Capacity,
			VolumeId:      volume.Name,
			VolumeContext: volume.VolumeContext,
		},
	}, nil
}

func (s *Server) DeleteVolume(ctx context.Context, req *DeleteVolumeRequest) (*DeleteVolumeResponse, error) {
	log.Infof("Calling %s ...", utils.FuncName())

	if req.GetVolumeId() == "" {
		return nil, fmt.Errorf("volumeid is empty")
	}

	return &DeleteVolumeResponse{}, s.vm.DeleteVolume(ctx, req.GetVolumeId())
}

func (s *Server) ControllerPublishVolume(context.Context, *ControllerPublishVolumeRequest) (*ControllerPublishVolumeResponse, error) {
	log.Infof("Calling %s ...", utils.FuncName())

	// just return success here

	return &ControllerPublishVolumeResponse{}, nil
}

func (s *Server) ControllerUnpublishVolume(context.Context, *ControllerUnpublishVolumeRequest) (*ControllerUnpublishVolumeResponse, error) {
	log.Infof("Calling %s ...", utils.FuncName())

	// just return success here

	return &ControllerUnpublishVolumeResponse{}, nil
}
func (s *Server) ValidateVolumeCapabilities(context.Context, *ValidateVolumeCapabilitiesRequest) (*ValidateVolumeCapabilitiesResponse, error) {
	log.Infof("Calling %s ...", utils.FuncName())

	return nil, fmt.Errorf("not implemented")
}

func (s *Server) ListVolumes(context.Context, *ListVolumesRequest) (*ListVolumesResponse, error) {
	log.Infof("Calling %s ...", utils.FuncName())

	return nil, fmt.Errorf("not implemented")
}

func (s *Server) GetCapacity(context.Context, *GetCapacityRequest) (*GetCapacityResponse, error) {
	log.Infof("Calling %s ...", utils.FuncName())

	return nil, fmt.Errorf("not implemented")
}

func (s *Server) ControllerGetCapabilities(context.Context, *ControllerGetCapabilitiesRequest) (*ControllerGetCapabilitiesResponse, error) {
	log.Infof("Calling %s ...", utils.FuncName())
	return &ControllerGetCapabilitiesResponse{Capabilities: s.supportControllerCapability}, nil
}

func (s *Server) CreateSnapshot(context.Context, *CreateSnapshotRequest) (*CreateSnapshotResponse, error) {
	log.Infof("Calling %s ...", utils.FuncName())

	return nil, fmt.Errorf("not implemented")
}

func (s *Server) DeleteSnapshot(context.Context, *DeleteSnapshotRequest) (*DeleteSnapshotResponse, error) {
	log.Infof("Calling %s ...", utils.FuncName())

	return nil, fmt.Errorf("not implemented")
}

func (s *Server) ListSnapshots(context.Context, *ListSnapshotsRequest) (*ListSnapshotsResponse, error) {
	log.Infof("Calling %s ...", utils.FuncName())

	return nil, fmt.Errorf("not implemented")
}

func (s *Server) ControllerExpandVolume(context.Context, *ControllerExpandVolumeRequest) (*ControllerExpandVolumeResponse, error) {
	log.Infof("Calling %s ...", utils.FuncName())

	return nil, fmt.Errorf("not implemented")
}

func (s *Server) ControllerGetVolume(context.Context, *ControllerGetVolumeRequest) (*ControllerGetVolumeResponse, error) {
	log.Infof("Calling %s ...", utils.FuncName())

	return nil, fmt.Errorf("not implemented")
}

func (s *Server) getControllerCapability() []*ControllerServiceCapability {
	return s.supportControllerCapability
}

func (s *Server) verifyControllerCapability(needCap ControllerServiceCapability_RPC_Type) error {
	for _, existCap := range s.getControllerCapability() {
		if existCap.GetRpc().Type == needCap {
			return nil
		}
	}
	return status.Errorf(codes.InvalidArgument, "unsupported controller capability %s", needCap)
}

func (s *Server) validateCreateVolumeRequest(req *CreateVolumeRequest) error {
	// verify Capability
	if err := s.verifyControllerCapability(ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		return err
	}

	return nil
}

func (s *Server) initControllerCapability() {
	caps := []ControllerServiceCapability_RPC_Type{
		// for volume
		ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		//ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
		ControllerServiceCapability_RPC_LIST_VOLUMES,
		ControllerServiceCapability_RPC_GET_VOLUME,
	}
	for _, c := range caps {
		s.supportControllerCapability = append(s.supportControllerCapability, newControllerServiceCapability(c))
	}
}

func newControllerServiceCapability(cap ControllerServiceCapability_RPC_Type) *ControllerServiceCapability {
	return &ControllerServiceCapability{
		Type: &ControllerServiceCapability_Rpc{
			Rpc: &ControllerServiceCapability_RPC{
				Type: cap,
			},
		},
	}
}
