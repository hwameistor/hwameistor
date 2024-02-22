package main

import (
	"context"
	"flag"
	"fmt"
	localctrl "github.com/hwameistor/hwameistor/pkg/local-storage/member/controller"
	"os"
	"path"
	"runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	apis "github.com/hwameistor/hwameistor/pkg/apis/hwameistor"
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-storage/controller"
	"github.com/hwameistor/hwameistor/pkg/local-storage/member"
	"github.com/hwameistor/hwameistor/pkg/local-storage/utils"
	"github.com/hwameistor/hwameistor/pkg/local-storage/utils/datacopy"
)

const (
	warmupPeriod = 5 * time.Second

	restServerDefaultPort     = 80
	restServerCertsDefaultDir = "/etc/ssl/certs"

	defaultDRBDStartPort      = 43001
	defaultHAVolumeTotalCount = 1000
	defaultDataSyncToolName   = datacopy.SyncToolJuiceSync
)

var (
	nodeName                = flag.String("nodename", "", "Node name")
	namespace               = flag.String("namespace", "", "Namespace of the Pod")
	csiSockAddr             = flag.String("csi-address", "", "CSI endpoint")
	systemMode              = flag.String("system-mode", string(apisv1alpha1.SystemModeDRBD), "dlocal system mode")
	dataSyncToolName        = flag.String("data-sync-tool", defaultDataSyncToolName, "tool to sync the data across the nodes, e.g. juicesync")
	drbdStartPort           = flag.Int("drbd-start-port", defaultDRBDStartPort, "drbd start port, end port=start-port+volume-count-1")
	haVolumeTotalCount      = flag.Int("max-ha-volume-count", defaultHAVolumeTotalCount, "max HA volume count")
	httpPort                = flag.Int("http-port", restServerDefaultPort, "HTTP port for REST server")
	logLevel                = flag.Int("v", 4 /*Log Info*/, "number for the log level verbosity")
	MigrateConcurrentNumber = flag.Int("max-migrate-count", 1, "Limit the number of concurrent migrations")
)

var BUILDVERSION, BUILDTIME, GOVERSION string

func printVersion() {
	//log.Info(fmt.Sprintf("Operator Version: %s", version.Version))
	log.Info(fmt.Sprintf("GitCommit:%q, BuildDate:%q, GoVersion:%q", BUILDVERSION, BUILDTIME, GOVERSION))
}

func setupLogging() {
	// parse log level(default level: info)
	var level log.Level
	if *logLevel >= int(log.TraceLevel) {
		level = log.TraceLevel
	} else if *logLevel <= int(log.PanicLevel) {
		level = log.PanicLevel
	} else {
		level = log.Level(*logLevel)
	}

	log.SetLevel(level)
	log.SetFormatter(&log.JSONFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			s := strings.Split(f.Function, ".")
			funcName := s[len(s)-1]
			fileName := path.Base(f.File)
			return funcName, fmt.Sprintf("%s:%d", fileName, f.Line)
		}})
	log.SetReportCaller(true)
}

func main() {
	//klog.InitFlags(nil)
	// Add flags registered by imported packages (e.g. glog and
	// controller-runtime)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	pflag.Parse()

	printVersion()

	flag.Parse()

	setupLogging()

	if *nodeName == "" {
		log.WithFields(log.Fields{"nodename": *nodeName}).Error("Invalid node name")
		os.Exit(1)
	}

	if *csiSockAddr == "" {
		log.WithFields(log.Fields{"endpoint": *csiSockAddr}).Error("Invalid CSI endpoint")
		os.Exit(1)
	}

	if *namespace == "" {
		log.WithFields(log.Fields{"namespace": *namespace}).Error("Invalid namespace")
		os.Exit(1)
	}

	localctrl.MigrateConcurrentNumber = *MigrateConcurrentNumber

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

	// Setup Cache for field index
	setIndexField(mgr.GetCache())

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
		SyncToolName:     *dataSyncToolName,
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

// setIndexField must be called after scheme has been added
func setIndexField(cache cache.Cache) {
	indexes := []struct {
		field string
		Func  func(client.Object) []string
	}{
		// indexer for LocalVolumeSnapshot
		{
			field: "spec.sourceVolume",
			Func: func(obj client.Object) []string {
				return []string{obj.(*apisv1alpha1.LocalVolumeSnapshot).Spec.SourceVolume}
			},
		},
	}

	for _, index := range indexes {
		if err := cache.IndexField(context.Background(), &apisv1alpha1.LocalVolumeSnapshot{}, index.field, index.Func); err != nil {
			log.Error(err, "failed to setup index field", "field", index.field)
			// indexer is required, exit immediately if it fails, more details see issue: #1209
			os.Exit(1)
		}
		log.Info("setup index field successfully", "field", index.field)
	}
}
