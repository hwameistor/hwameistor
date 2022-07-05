package identity

import (
	"context"

	csi "github.com/container-storage-interface/spec/lib/go/csi"
)

// Server
type Server struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

const (
	VendorVersion = "v1alpha1"
	DriverName    = "disk.hwameistor.io"
)

// NewServer
func NewServer() *Server {
	return &Server{
		Name:    DriverName,
		Version: VendorVersion,
	}
}

// GetPluginInfo
func (ids *Server) GetPluginInfo(context.Context, *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	return &csi.GetPluginInfoResponse{
		Name:          ids.Name,
		VendorVersion: ids.Version,
	}, nil
}

// GetPluginCapabilities
func (ids *Server) GetPluginCapabilities(context.Context, *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	return &csi.GetPluginCapabilitiesResponse{
		Capabilities: []*csi.PluginCapability{
			{
				Type: &csi.PluginCapability_Service_{
					Service: &csi.PluginCapability_Service{
						Type: csi.PluginCapability_Service_CONTROLLER_SERVICE,
					},
				},
			}, {
				Type: &csi.PluginCapability_Service_{
					Service: &csi.PluginCapability_Service{
						Type: csi.PluginCapability_Service_VOLUME_ACCESSIBILITY_CONSTRAINTS,
					},
				},
			},
		},
	}, nil
}

// Probe
func (ids *Server) Probe(context.Context, *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	return &csi.ProbeResponse{}, nil
}
