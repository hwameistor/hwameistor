package driver

import (
	"errors"
	"os"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/csi/driver/controller"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/csi/driver/identity"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/csi/driver/node"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/csi/driver/server"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils"
)

const (
	VendorVersion = identity.VendorVersion
)

// driver define the parameters of the disk volume csi driver
type driver struct {
	Config

	// identityServer
	identityServer *identity.Server

	// nodeServer
	nodeServer *node.Server

	// controllerServer
	controllerServer *controller.Server

	// gRPC calls involving any of the fields below must be serialized
	// by locking this mutex before starting. Internal helper
	// functions assume that the mutex has been locked.
	mutex sync.Mutex
}

type Config struct {
	Enable        bool   `json:"enable"`
	DriverName    string `json:"driverName"`
	Endpoint      string `json:"endpoint"`
	NodeID        string `json:"nodeId"`
	VendorVersion string `json:"vendorVersion"`
}

func (cfg *Config) ToMap() map[string]interface{} {
	return utils.StructToMap(cfg, "json")
}

// NewDiskDriver
func NewDiskDriver(cfg Config) *driver {
	return &driver{
		Config:           cfg,
		identityServer:   identity.NewServer(),
		controllerServer: controller.NewServer(),
		nodeServer:       node.NewServer(),
	}
}

func (driver *driver) Run() {
	if driver.DriverName == "" {
		log.WithError(errors.New("no driver name is provided")).Error("Failed to start csi driver")
		os.Exit(1)
	}

	if driver.Endpoint == "" {
		log.WithError(errors.New("no endpoint is provided")).Error("Failed to start csi driver")
		os.Exit(1)
	}

	if driver.NodeID == "" {
		log.WithError(errors.New("no nodeid is provided")).Error("Failed to start csi driver")
		os.Exit(1)
	}

	logBase := log.WithFields((driver).ToMap())
	logBase.Info("Driver info:")

	grpcServer := server.NewNonBlockingGRPCServer()
	grpcServer.Start(driver.Endpoint, driver.identityServer, driver.controllerServer, driver.nodeServer)

	grpcServer.Wait()
	return
}
