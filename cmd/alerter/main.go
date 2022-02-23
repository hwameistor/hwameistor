package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/kubernetes-csi/csi-lib-utils/leaderelection"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	"github.com/hwameistor/local-storage/pkg/alerter"
)

const (
	lockName = "localstorage-local-storage-alerter"
)

var (
	debug         = flag.Bool("debug", false, "debug mode")
	isVirtualNode = flag.Bool("is-virtual-machine", false, "is virtual machine environment, false by default")
	opNamespace   = flag.String("namespace", "default", "k8s namespace")
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

	flag.Parse()

	setupLogging(*debug)
	printVersion()

	run := func(ctx context.Context) {
		if err := alerter.NewManager(*isVirtualNode).Run(signals.SetupSignalHandler()); err != nil {
			log.WithFields(log.Fields{"error": err.Error()}).Error("failed to run local storage alerter")
			os.Exit(1)
		}
	}

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal(err)
	}

	leClientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatal(err)
	}

	le := leaderelection.NewLeaderElection(leClientset, lockName, run)
	le.WithNamespace(*opNamespace)

	if err := le.Run(); err != nil {
		log.Fatalf("failed to initialize leader election: %v", err)
	}
}
