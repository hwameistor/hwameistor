package main

import (
	"context"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	localstoragev1alpha1 "github.com/hwameistor/local-storage/pkg/apis/localstorage/v1alpha1"
	"github.com/hwameistor/local-storage/pkg/member/node/healths"
	"github.com/hwameistor/local-storage/pkg/utils"

	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

const (
	warmupPeriod = 5 * time.Second
)

var BUILDVERSION, BUILDTIME, GOVERSION string

func printVersion() {
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
	setupLogging(true)
	printVersion()

	nodeName := os.Getenv("MY_NAME")
	if len(nodeName) == 0 {
		log.Error("No MY_NAME specified")
		os.Exit(1)
	}
	namespace := os.Getenv("MY_NAMESPACE")
	if len(namespace) == 0 {
		log.Error("No MY_NAMESPACE specified")
		os.Exit(1)
	}

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
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

	schemeBuilder := &scheme.Builder{GroupVersion: localstoragev1alpha1.SchemeGroupVersion}
	schemeBuilder.Register(&localstoragev1alpha1.PhysicalDisk{}, &localstoragev1alpha1.PhysicalDiskList{})
	if err := schemeBuilder.AddToScheme(mgr.GetScheme()); err != nil {
		// Setup Scheme for physical disk
		log.WithError(err).Error("Failed to setup scheme for all resources")
		os.Exit(1)
	}

	runFunc := func(ctx context.Context) {
		stopCh := signals.SetupSignalHandler()
		// Start the resource controllers manager
		go func() {
			log.Info("Starting the manager of all local storage resources.")
			if err := mgr.Start(stopCh); err != nil {
				log.WithError(err).Error("Failed to run resources manager")
				os.Exit(1)
			}
		}()

		time.Sleep(warmupPeriod)

		log.Info("Starting to check disk health")
		healths.NewDiskHealthManager(nodeName, mgr.GetClient()).Run(stopCh)

		// This will run forever until channel receives stop signal
		<-stopCh
		log.Debug("Stopped checking")
	}

	if err := utils.RunWithLease(namespace, nodeName, fmt.Sprintf("localstorage-disk-checker-%s", nodeName), runFunc); err != nil {
		log.WithError(err).Error("failed to initialize node heartbeat lease")
		os.Exit(1)
	}

	log.Debug("Completely stopped")
}
