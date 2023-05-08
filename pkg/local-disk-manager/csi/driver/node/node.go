package node

import (
	"context"
	"fmt"
	volume "github.com/hwameistor/hwameistor/pkg/local-disk-manager/member/controller/volume"

	"github.com/container-storage-interface/spec/lib/go/csi"

	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils"
)

type Server struct {
	// vm manager volume create,delete,query
	vm volume.Manager

	// supportNodeCapability
	supportNodeCapability []*csi.NodeServiceCapability

	// Config
	Config `json:"config"`
}

type Config struct {
	NodeName string `json:"nodeName"`
}

func NewServer() *Server {
	server := &Server{}
	server.initConfig()
	server.initNodeCapabilities()
	server.vm = volume.New()
	return server
}

func (s *Server) initConfig() {
	s.Config.NodeName = utils.GetNodeName()
}

func (s *Server) initNodeCapabilities() {
	caps := []csi.NodeServiceCapability_RPC_Type{
		//NodeServiceCapability_RPC_GET_VOLUME_STATS,
		//NodeServiceCapability_RPC_VOLUME_CONDITION,
	}
	for _, c := range caps {
		s.supportNodeCapability = append(s.supportNodeCapability, newNodeServiceCapability(c))
	}
}

func newNodeServiceCapability(cap csi.NodeServiceCapability_RPC_Type) *csi.NodeServiceCapability {
	return &csi.NodeServiceCapability{
		Type: &csi.NodeServiceCapability_Rpc{
			Rpc: &csi.NodeServiceCapability_RPC{
				Type: cap,
			},
		},
	}
}

func (s *Server) NodeStageVolume(context.Context, *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *Server) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *Server) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	if err := s.validateNodePublishRequest(req); err != nil {
		return nil, err
	}

	return &csi.NodePublishVolumeResponse{}, s.vm.NodePublishVolume(ctx, req)
}

func (s *Server) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	if err := s.validateNodeUnPublishRequest(req); err != nil {
		return nil, err
	}

	return &csi.NodeUnpublishVolumeResponse{}, s.vm.NodeUnpublishVolume(ctx, req.GetVolumeId(), req.GetTargetPath())
}

func (s *Server) NodeGetVolumeStats(ctx context.Context, req *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *Server) NodeExpandVolume(context.Context, *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *Server) NodeGetCapabilities(context.Context, *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	return &csi.NodeGetCapabilitiesResponse{Capabilities: s.supportNodeCapability}, nil
}

func (s *Server) NodeGetInfo(context.Context, *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	return &csi.NodeGetInfoResponse{
		NodeId: s.NodeName,
		AccessibleTopology: &csi.Topology{
			Segments: map[string]string{
				volume.TopologyNodeKey: s.NodeName,
			}},
	}, nil
}

func (s *Server) validateNodePublishRequest(req *csi.NodePublishVolumeRequest) error {
	if req.GetVolumeId() == "" {
		return fmt.Errorf("VolumeId is empty")
	}

	if req.GetTargetPath() == "" {
		return fmt.Errorf("TargetPath is empty")
	}

	if req.GetVolumeCapability() == nil {
		return fmt.Errorf("VolumeCapbility is empty")
	}

	return nil
}

func (s *Server) validateNodeUnPublishRequest(req *csi.NodeUnpublishVolumeRequest) error {
	if req.GetVolumeId() == "" {
		return fmt.Errorf("VolumeId is empty")
	}

	if req.GetTargetPath() == "" {
		return fmt.Errorf("TargetPath is empty")
	}

	return nil
}
