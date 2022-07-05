package node

import (
	"context"
	"fmt"

	. "github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/csi/volumemanager"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils"
)

type Server struct {
	// vm manager volume create,delete,query
	vm volumemanager.VolumeManager

	// supportNodeCapability
	supportNodeCapability []*NodeServiceCapability

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
	server.vm = volumemanager.NewLocalDiskVolumeManager()
	return server
}

func (s *Server) initConfig() {
	s.Config.NodeName = utils.GetNodeName()
}

func (s *Server) initNodeCapabilities() {
	caps := []NodeServiceCapability_RPC_Type{
		//NodeServiceCapability_RPC_GET_VOLUME_STATS,
		//NodeServiceCapability_RPC_VOLUME_CONDITION,
	}
	for _, c := range caps {
		s.supportNodeCapability = append(s.supportNodeCapability, newNodeServiceCapability(c))
	}
}

func newNodeServiceCapability(cap NodeServiceCapability_RPC_Type) *NodeServiceCapability {
	return &NodeServiceCapability{
		Type: &NodeServiceCapability_Rpc{
			Rpc: &NodeServiceCapability_RPC{
				Type: cap,
			},
		},
	}
}

func (s *Server) NodeStageVolume(context.Context, *NodeStageVolumeRequest) (*NodeStageVolumeResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *Server) NodeUnstageVolume(ctx context.Context, req *NodeUnstageVolumeRequest) (*NodeUnstageVolumeResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *Server) NodePublishVolume(ctx context.Context, req *NodePublishVolumeRequest) (*NodePublishVolumeResponse, error) {
	if err := s.validateNodePublishRequest(req); err != nil {
		return nil, err
	}

	return &NodePublishVolumeResponse{}, s.vm.NodePublishVolume(ctx, req)
}

func (s *Server) NodeUnpublishVolume(ctx context.Context, req *NodeUnpublishVolumeRequest) (*NodeUnpublishVolumeResponse, error) {
	if err := s.validateNodeUnPublishRequest(req); err != nil {
		return nil, err
	}

	return &NodeUnpublishVolumeResponse{}, s.vm.NodeUnpublishVolume(ctx, req.GetVolumeId(), req.GetTargetPath())
}

func (s *Server) NodeGetVolumeStats(ctx context.Context, req *NodeGetVolumeStatsRequest) (*NodeGetVolumeStatsResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *Server) NodeExpandVolume(context.Context, *NodeExpandVolumeRequest) (*NodeExpandVolumeResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *Server) NodeGetCapabilities(context.Context, *NodeGetCapabilitiesRequest) (*NodeGetCapabilitiesResponse, error) {
	return &NodeGetCapabilitiesResponse{Capabilities: s.supportNodeCapability}, nil
}

func (s *Server) NodeGetInfo(context.Context, *NodeGetInfoRequest) (*NodeGetInfoResponse, error) {
	return &NodeGetInfoResponse{
		NodeId: s.NodeName,
		AccessibleTopology: &Topology{
			Segments: map[string]string{
				volumemanager.TopologyNodeKey: s.NodeName,
			}},
	}, nil
}

func (s *Server) validateNodePublishRequest(req *NodePublishVolumeRequest) error {
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

func (s *Server) validateNodeUnPublishRequest(req *NodeUnpublishVolumeRequest) error {
	if req.GetVolumeId() == "" {
		return fmt.Errorf("VolumeId is empty")
	}

	if req.GetTargetPath() == "" {
		return fmt.Errorf("TargetPath is empty")
	}

	return nil
}
