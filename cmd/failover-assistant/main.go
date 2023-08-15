package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	failoverassistant "github.com/hwameistor/hwameistor/pkg/failover-assistant"

	"github.com/kubernetes-csi/csi-lib-utils/leaderelection"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

const (
	lockName                       = "hwameistor-failover-assistant"
	failoverCoolDownDefaultMinutes = 15
)

var BUILDVERSION, BUILDTIME, GOVERSION string

func printVersion() {
	log.Info(fmt.Sprintf("GitCommit:%q, BuildDate:%q, GoVersion:%q", BUILDVERSION, BUILDTIME, GOVERSION))
}

func setupLogging(debug bool) {
	if debug {
		log.SetLevel(log.DebugLevel)
	}
}

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", true, "debug mode")
	flag.Parse()

	setupLogging(debug)
	printVersion()

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
		if err := failoverassistant.New(leClientset).Run(stopCh); err != nil {
			log.WithFields(log.Fields{"error": err.Error()}).Error("failed to run failover assistant")
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
