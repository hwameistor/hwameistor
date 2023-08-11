package main

import (
	"context"
	"flag"
	"os"

	"github.com/kubernetes-csi/csi-lib-utils/leaderelection"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/hwameistor/hwameistor/pkg/evictor"
)

const (
	lockName = "hwameistor-volume-evictor"
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
	leClientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.WithError(err).Fatal("Failed to create client set")
	}

	stopCh := make(chan struct{})

	run := func(ctx context.Context) {
		if err := evictor.New(leClientset).Run(stopCh); err != nil {
			log.WithFields(log.Fields{"error": err.Error()}).Error("failed to run evictor")
			os.Exit(1)
		}
	}

	le := leaderelection.NewLeaderElection(leClientset, lockName, run)
	opNamespace, _ := k8sutil.GetOperatorNamespace()
	le.WithNamespace(opNamespace)

	if err := le.Run(); err != nil {
		stopCh <- struct{}{}
		log.Fatalf("failed to initialize leader election: %v", err)
	}

}
