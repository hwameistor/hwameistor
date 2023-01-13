package rest

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/soheilhy/cmux"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apis "github.com/hwameistor/hwameistor/pkg/apis/hwameistor"
)

// Server interface
type Server interface {
	Run(stopCh <-chan struct{})
}

type restServer struct {
	name      string
	namespace string

	httpPort int

	member apis.LocalStorageMember

	apiClient client.Client

	logger *log.Entry
}

// New creates a rest server
func New(name string, namespace string, httpPort int, member apis.LocalStorageMember, cli client.Client) Server {
	return &restServer{
		name:      name,
		namespace: namespace,
		httpPort:  httpPort,
		member:    member,
		apiClient: cli,
		logger:    log.WithField("Module", "RESTServer"),
	}
}

// Run the rest server
func (rs *restServer) Run(stopCh <-chan struct{}) {

	go rs.startServer(stopCh)
}

func (rs *restServer) startServer(stopCh <-chan struct{}) {

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", rs.httpPort))
	if err != nil {
		rs.logger.Fatalf("Failed to listen: %v", err)
	}
	log.WithFields(log.Fields{
		"endpoint": listener.Addr().String(),
		"protocol": listener.Addr().Network(),
	}).Info("Listening on the same port for both REST and gRPC connections.")

	tcpm := cmux.New(listener)

	// Declare the match for different services required.
	//go rs.serveGPRC(tcpm.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc")))

	go rs.serveHTTP(tcpm.Match(cmux.Any()))

	if err := tcpm.Serve(); !strings.Contains(err.Error(), "use of closed network connection") {
		log.Fatal(err)
	}

	// Waiting for the stop signal
	<-stopCh
	rs.logger.Info("Got a stop signal to terminate REST server")
}

/*
func (rs *restServer) serveGPRC(listener net.Listener) {
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(utils.LogGRPC),
	}
	grpcServer := grpc.NewServer(opts...)

	localstoragev1.RegisterStorageDriverServer(grpcServer, rs)
	log.Debug("Registered MetricsCollector server.")
	reflection.Register(grpcServer)

	log.Debug("Starting gRPC server")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("While serving gRpc request: %v", err)
	}
}
*/

func (rs *restServer) serveHTTP(listener net.Listener) {

	router := mux.NewRouter().StrictSlash(true)
	for _, route := range rs.buildRoutes() {
		router.
			Name(route.Name).
			Methods(route.Method).
			Path(route.Pattern).
			Handler(route.HandlerFunc)
		//Handler(utils.LogREST(route.HandlerFunc, route.Name))
	}
	// start server on HTTP port
	rs.logger.WithFields(log.Fields{"https.port": rs.httpPort}).Debug("starting HTTP server")
	if err := http.Serve(listener, router); err != nil {
		rs.logger.WithError(err).Fatal("REST server run into problem")
	}
}
