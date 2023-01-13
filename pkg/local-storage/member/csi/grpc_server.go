package csi

import (
	"net"
	"os"
	"path"

	csi "github.com/container-storage-interface/spec/lib/go/csi"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/hwameistor/hwameistor/pkg/local-storage/utils"
)

// Server - interface of grpc server which is for k8s communication
//
//go:generate mockgen -source=grpc_server.go -destination=../../member/csi/grpc_server_mock.go  -package=csi
type Server interface {
	Init(endpoint string)
	Run(ids csi.IdentityServer, cs csi.ControllerServer, ns csi.NodeServer)
	GracefulStop()
	Stop()
}

type server struct {
	grpcServer *grpc.Server
	listener   net.Listener

	logger *log.Entry
}

var _ Server = (*server)(nil)

// NewGRPCServer - create a grpc server instance
func NewGRPCServer(logger *log.Entry) Server {
	return &server{
		logger: logger,
	}
}

func (s *server) Init(endpoint string) {

	proto, addr, err := parseEndpoint(endpoint)
	if err != nil {
		s.logger.Fatal(err.Error())
	}

	s.logger.WithFields(log.Fields{
		"proto": proto,
		"addr":  addr,
	}).Debug("GRPC endpoint")

	if proto == "unix" {
		addr = "/" + addr
		if err := os.Remove(addr); err != nil && !os.IsNotExist(err) {
			s.logger.Fatalf("Failed to remove %s, error: %s", addr, err.Error())
		} else {
			// Need to make directory at the first time the csi service runs.
			dir := path.Dir(addr)
			if exist, _ := pathExists(dir); !exist {
				s.logger.WithFields(log.Fields{
					"dir": dir,
				}).Info("Mkdir")
				os.MkdirAll(dir, 0755)
			}
		}
	}

	s.logger.WithFields(log.Fields{
		"proto": proto,
		"addr":  addr,
	}).Debug("Trying to listen to the endpoint")
	listener, err := net.Listen(proto, addr)
	if err != nil {
		s.logger.Fatalf("Failed to listen: %v", err)
	}

	s.logger.WithFields(log.Fields{
		"name": listener.Addr().String(),
		"net":  listener.Addr().Network(),
	}).Info("Listening for GRPC connections.")

	if listener.Addr().Network() == "unix" {
		if err := os.Chmod(listener.Addr().String(), 0777); err != nil {
			s.logger.Fatal(err)
		}
	}
	s.listener = listener
}

func (s *server) Run(ids csi.IdentityServer, cs csi.ControllerServer, ns csi.NodeServer) {
	s.logger.Debug("Start gRPC server ...")
	defer s.logger.Debug("End of Start gRPC server")

	if s.listener == nil {
		s.logger.Fatalf("Listener is not initialized yet")
	}

	go s.serve(ids, cs, ns)
}

func (s *server) GracefulStop() {
	s.logger.Debug("Stop gRPC server gracefully ...")
	defer s.logger.Debug("End of Stop gRPC server gracefully")

	s.grpcServer.GracefulStop()
	s.listener.Close()
}

func (s *server) Stop() {
	s.logger.Info("Stop gRPC server ...")
	defer s.logger.Info("End of Stop gRPC server")

	s.grpcServer.Stop()
	s.listener.Close()
}

func (s *server) serve(ids csi.IdentityServer, cs csi.ControllerServer, ns csi.NodeServer) {

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(utils.LogGRPC),
	}
	grpcServer := grpc.NewServer(opts...)
	s.grpcServer = grpcServer

	if ids != nil {
		csi.RegisterIdentityServer(grpcServer, ids)
		s.logger.Debug("Registered CSI identity server.")
	}
	if cs != nil {
		csi.RegisterControllerServer(grpcServer, cs)
		s.logger.Debug("Registered CSI controller server.")
	}
	if ns != nil {
		csi.RegisterNodeServer(grpcServer, ns)
		s.logger.Debug("Registered CSI node server.")
	}
	reflection.Register(grpcServer)

	grpcServer.Serve(s.listener)
}
