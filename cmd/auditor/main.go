package main

import (
	"context"
	"flag"
	"os"

	clusterapiv1alpha1 "github.com/hwameistor/hwameistor-operator/api/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/auditor"
	"github.com/kubernetes-csi/csi-lib-utils/leaderelection"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

const (
	lockName = "hwameistor-auditor"
)

func setupLogging() {
	log.SetLevel(log.DebugLevel)
}

func main() {
	flag.Parse()
	setupLogging()

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		log.WithError(err).Fatal("Failed to get kubernetes cluster config")
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
	if err := clusterapiv1alpha1.AddToScheme(mgr.GetScheme()); err != nil {
		log.WithError(err).Error("Failed to setup scheme for all HwameiStor resources")
		os.Exit(1)
	}

	go func() {
		log.Info("Starting the manager of all local storage resources.")
		if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
			log.WithError(err).Error("Failed to run resources manager")
			os.Exit(1)
		}
	}()

	leClientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.WithError(err).Fatal("Failed to create client set")
	}

	stopCh := make(chan struct{})
	runFunc := func(ctx context.Context) {
		if err := auditor.New(leClientset).Run(mgr.GetCache(), stopCh); err != nil {
			log.WithFields(log.Fields{"error": err.Error()}).Error("failed to run auditor")
			os.Exit(1)
		}
	}

	le := leaderelection.NewLeaderElection(leClientset, lockName, runFunc)
	opNamespace, _ := k8sutil.GetOperatorNamespace()
	le.WithNamespace(opNamespace)

	if err := le.Run(); err != nil {
		stopCh <- struct{}{}
		log.Fatalf("failed to initialize leader election: %v", err)
	}

}
