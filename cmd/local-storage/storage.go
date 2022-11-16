package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	apis "github.com/hwameistor/hwameistor/pkg/apis/hwameistor"
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-storage/controller"
	"github.com/hwameistor/hwameistor/pkg/local-storage/member"
	"github.com/hwameistor/hwameistor/pkg/local-storage/utils"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

const (
	warmupPeriod = 5 * time.Second

	restServerDefaultPort     = 80
	restServerCertsDefaultDir = "/etc/ssl/certs"

	defaultDRBDStartPort      = 43001
	defaultHAVolumeTotalCount = 1000
)

var (
	debug              = flag.Bool("debug", false, "debug mode, false by default")
	nodeName           = flag.String("nodename", "", "Node name")
	namespace          = flag.String("namespace", "", "Namespace of the Pod")
	csiSockAddr        = flag.String("csi-address", "", "CSI endpoint")
	systemMode         = flag.String("system-mode", string(apisv1alpha1.SystemModeDRBD), "dlocal system mode")
	drbdStartPort      = flag.Int("drbd-start-port", defaultDRBDStartPort, "drbd start port, end port=start-port+volume-count-1")
	haVolumeTotalCount = flag.Int("max-ha-volume-count", defaultHAVolumeTotalCount, "max HA volume count")
	httpPort           = flag.Int("http-port", restServerDefaultPort, "HTTP port for REST server")
)

var BUILDVERSION, BUILDTIME, GOVERSION string

func printVersion() {
	//log.Info(fmt.Sprintf("Operator Version: %s", version.Version))
	log.Info(fmt.Sprintf("GitCommit:%q, BuildDate:%q, GoVersion:%q", BUILDVERSION, BUILDTIME, GOVERSION))
}

func setupLogging(enableDebug bool) {
	if enableDebug {
		log.SetLevel(log.DebugLevel)
	}

	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
		// log with funcname, file fileds. eg: func=processNode file="node_task_worker.go:43"
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			s := strings.Split(f.Function, ".")
			funcname := s[len(s)-1]
			filename := path.Base(f.File)
			return funcname, fmt.Sprintf("%s:%d", filename, f.Line)
		},
	})
	log.SetReportCaller(true)
}

func main() {
	klog.InitFlags(nil)
	// Add flags registered by imported packages (e.g. glog and
	// controller-runtime)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	pflag.Parse()

	printVersion()

	flag.Parse()

	setupLogging(*debug)

	if *nodeName == "" {
		log.WithFields(log.Fields{"nodename": *nodeName}).Error("Invalid node name")
		os.Exit(1)
	}

	if *csiSockAddr == "" {
		log.WithFields(log.Fields{"endpoint": *csiSockAddr}).Error("Invalid CSI endpint")
		os.Exit(1)
	}

	if *namespace == "" {
		log.WithFields(log.Fields{"namespace": *namespace}).Error("Invalid namespace")
		os.Exit(1)
	}

	systemConfig, err := getSystemConfig()
	if err != nil {
		log.Fatalf("invalid system config: %s", err)
	}

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatalf("Failed to get API server config: %s", err.Error())
	}

	// Set default manager options
	options := manager.Options{
		Namespace: "", // watch all namespaces
	}

	// Create a new manager to provide shared dependencies and start components
	mgr, err := manager.New(cfg, options)
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	log.Info("Registering Components...")

	// Setup Scheme for all resources of Local Storage Member
	if err := apisv1alpha1.AddToScheme(mgr.GetScheme()); err != nil {
		log.WithError(err).Error("Failed to setup scheme for all HwameiStor resources")
		os.Exit(1)
	}

	// if err := corev1.AddToScheme(mgr.GetScheme()); err != nil {
	// 	log.WithError(err).Error("Failed to setup scheme for all core resources")
	// 	os.Exit(1)
	// }

	// Setup all Controllers for Local Storage Member
	if err := controller.AddToManager(mgr); err != nil {
		log.WithError(err).Error("Failed to setup controllers for all local storage resources")
		os.Exit(1)
	}

	//initialize the local storage node/member as:
	log.Info("Configuring the Local Storage Member")
	storageMember := member.Member().ConfigureBase(*nodeName, *namespace, systemConfig, mgr.GetClient(), mgr.GetCache(),
		mgr.GetEventRecorderFor(fmt.Sprintf("%s/%s", "localstoragemanager", *nodeName))).
		ConfigureNode(mgr.GetScheme()).
		ConfigureController(mgr.GetScheme()).
		ConfigureCSIDriver(apisv1alpha1.CSIDriverName, *csiSockAddr).
		ConfigureRESTServer(*httpPort)

	runFunc := func(ctx context.Context) {
		stopCtx := signals.SetupSignalHandler()
		stopCh := make(chan struct{}, 1)

		// NOTE: Adapter for controller-runtime from 0.6.9 to 0.9.0
		// Convert context.Cancel signal to chan event
		go func() {
			<-stopCtx.Done()
			stopCh <- struct{}{}
		}()

		// Start the resource controllers manager
		go func() {
			log.Info("Starting the manager of all local storage resources.")
			if err := mgr.Start(stopCtx); err != nil {
				log.WithError(err).Error("Failed to run resources manager")
				os.Exit(1)
			}
		}()

		time.Sleep(warmupPeriod)

		log.Info("Starting to run local storage member")
		storageMember.Run(stopCh)

		// This will run forever until channel receives stop signal
		<-stopCh
		log.Debug("Stopped Local Storage Member")
	}

	log.Debug("Starting the Local Storage Member with heartbeat")
	if err := utils.RunWithLease(*namespace, *nodeName, fmt.Sprintf("%s-%s", apis.NodeLeaseNamePrefix, *nodeName), runFunc); err != nil {
		log.WithError(err).Error("failed to initialize node heartbeat lease")
		os.Exit(1)
	}

	log.Debug("Completely stopped")
}

func validateSystemConfig() error {
	var errMsgs []string
	switch apisv1alpha1.SystemMode(*systemMode) {
	case apisv1alpha1.SystemModeDRBD:
	default:
		errMsgs = append(errMsgs, fmt.Sprintf("system mode %s not supported", *systemMode))
	}

	if len(errMsgs) != 0 {
		return fmt.Errorf(strings.Join(errMsgs, "; "))
	}
	return nil
}

func getSystemConfig() (apisv1alpha1.SystemConfig, error) {
	if err := validateSystemConfig(); err != nil {
		return apisv1alpha1.SystemConfig{}, err
	}

	config := apisv1alpha1.SystemConfig{
		Mode:             apisv1alpha1.SystemMode(*systemMode),
		MaxHAVolumeCount: *haVolumeTotalCount,
	}

	switch config.Mode {
	case apisv1alpha1.SystemModeDRBD:
		{
			config.DRBD = &apisv1alpha1.DRBDSystemConfig{
				StartPort: *drbdStartPort,
			}
		}
	}
	return config, nil
}
