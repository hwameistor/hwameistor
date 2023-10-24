package csi

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apis "github.com/hwameistor/hwameistor/pkg/apis/hwameistor"
	"github.com/hwameistor/hwameistor/pkg/exechelper"
	"github.com/hwameistor/hwameistor/pkg/exechelper/nsexecutor"
	"github.com/hwameistor/hwameistor/pkg/local-storage/member/node/qos"
)

const (
	driverVersion = "1.0"
)

// Driver interface
//
//go:generate mockgen -source=plugin.go -destination=../../member/csi/plugin_mock.go  -package=csi
type Driver interface {
	Run(stopCh <-chan struct{})
}

// plugin - local storage system CSI plugin struct including controller, node, identity
type plugin struct {
	name    string
	version string

	nodeName   string
	namespace  string
	sockAddr   string
	grpcServer Server

	storageMember apis.LocalStorageMember
	mounter       Mounter

	//lock      sync.Mutex
	apiClient client.Client

	volumeQoSManager *qos.VolumeQoSManager

	cmdExecutor exechelper.Executor
	logger      *log.Entry

	pCaps  []*csi.PluginCapability
	csCaps []*csi.ControllerServiceCapability
	nsCaps []*csi.NodeServiceCapability
	vCaps  []*csi.VolumeCapability
}

// New - create a new plugin instance
func New(nodeName string, namespace string, driverName string, sockAddr string, storageMember apis.LocalStorageMember, cli client.Client) Driver {

	logger := log.WithField("Module", "CSIPlugin")

	volumeQoSManager, err := qos.NewVolumeQoSManager(nodeName, cli)
	if err != nil {
		panic(err)
	}

	return &plugin{
		name:             driverName,
		version:          driverVersion,
		nodeName:         nodeName,
		namespace:        namespace,
		sockAddr:         sockAddr,
		grpcServer:       NewGRPCServer(logger),
		storageMember:    storageMember,
		mounter:          NewLinuxMounter(logger),
		cmdExecutor:      nsexecutor.New(),
		apiClient:        cli,
		volumeQoSManager: volumeQoSManager,
		logger:           logger,
	}
}

// Run - run the plugin
func (p *plugin) Run(stopCh <-chan struct{}) {

	p.logger.Debug("Initialize CSI plugin ...")
	defer p.logger.Debug("End of Initialize CSI plugin")

	p.initCapabilities()

	//initialize the grpc server for listening on the unix socket
	p.grpcServer.Init(p.sockAddr)

	p.logger.Debug("Starting to run CSI driver")

	go p.startServer(stopCh)
}

func (p *plugin) startServer(stopCh <-chan struct{}) {
	// start the grpc service to communicate with kubernetes
	p.grpcServer.Run(p, p, p)

	<-stopCh
	p.logger.Info("Got a stop signal to terminate driver")
	p.grpcServer.GracefulStop()
}

func (p *plugin) initCapabilities() {
	p.initPluginCapabilities()
	p.initControllerServiceCapabilities()
	p.initNodeServiceCapabilities()
	p.initVolumeCapability()
}

func (p *plugin) initPluginCapabilities() {
	caps := []csi.PluginCapability_Service_Type{
		csi.PluginCapability_Service_CONTROLLER_SERVICE,
		csi.PluginCapability_Service_VOLUME_ACCESSIBILITY_CONSTRAINTS,
	}
	for _, c := range caps {
		p.logger.WithField("capability", c.String()).Debug("Enabling plugin capability.")
		p.pCaps = append(p.pCaps, newPluginCapability(c))
	}
}

func (p *plugin) initControllerServiceCapabilities() {
	caps := []csi.ControllerServiceCapability_RPC_Type{
		// for volume
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
		csi.ControllerServiceCapability_RPC_LIST_VOLUMES,
		csi.ControllerServiceCapability_RPC_EXPAND_VOLUME,
		csi.ControllerServiceCapability_RPC_GET_VOLUME,
		csi.ControllerServiceCapability_RPC_VOLUME_CONDITION,
		// for storage capacity
		//csi.ControllerServiceCapability_RPC_GET_CAPACITY, // don't support it as the scheduler will take care of it
		// for snapshot
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT,
		csi.ControllerServiceCapability_RPC_LIST_SNAPSHOTS,
		// for clone
		csi.ControllerServiceCapability_RPC_CLONE_VOLUME,
	}
	for _, c := range caps {
		p.logger.WithField("capability", c.String()).Debug("Enabling controller service capability.")
		p.csCaps = append(p.csCaps, newControllerServiceCapability(c))
	}
}

func (p *plugin) initNodeServiceCapabilities() {
	caps := []csi.NodeServiceCapability_RPC_Type{
		csi.NodeServiceCapability_RPC_EXPAND_VOLUME,
		csi.NodeServiceCapability_RPC_GET_VOLUME_STATS,
		csi.NodeServiceCapability_RPC_VOLUME_CONDITION,
	}
	for _, c := range caps {
		p.logger.WithField("capability", c.String()).Debug("Enabling node service capability.")
		p.nsCaps = append(p.nsCaps, newNodeServiceCapability(c))
	}
}

func (p *plugin) initVolumeCapability() {
	p.vCaps = []*csi.VolumeCapability{
		{ // Tell CO we can provisiion readWriteOnce filesystem volumes.
			AccessType: &csi.VolumeCapability_Mount{
				Mount: &csi.VolumeCapability_MountVolume{},
			},
			AccessMode: &csi.VolumeCapability_AccessMode{
				Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER},
		},
		{ // Tell CO we can provisiion readWriteOnce raw block volumes.
			AccessType: &csi.VolumeCapability_Block{
				Block: &csi.VolumeCapability_BlockVolume{},
			},
			AccessMode: &csi.VolumeCapability_AccessMode{
				Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
			},
		},
	}
	for _, c := range p.vCaps {
		p.logger.WithField("capability", c).Debug("Enabling volume capability")
	}
}
