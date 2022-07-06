package csi

import (
	csi "github.com/container-storage-interface/spec/lib/go/csi"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

var _ csi.IdentityServer = (*plugin)(nil)

// Probe implementation
func (p *plugin) Probe(ctx context.Context, req *csi.ProbeRequest) (*csi.ProbeResponse, error) {

	return &csi.ProbeResponse{}, nil
}

// GetPluginInfo implementation
func (p *plugin) GetPluginInfo(ctx context.Context, req *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	p.logger.WithFields(log.Fields{"request": req}).Debug("GetPluginInfo")

	return &csi.GetPluginInfoResponse{
		Name:          p.name,
		VendorVersion: p.version,
	}, nil
}

// GetPluginCapabilities implementation
func (p *plugin) GetPluginCapabilities(ctx context.Context, req *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	p.logger.WithFields(log.Fields{"request": req}).Debug("GetPluginCapabilities")

	return &csi.GetPluginCapabilitiesResponse{Capabilities: p.pCaps}, nil
}
