package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	clientset "github.com/hwameistor/hwameistor/pkg/apis/client/clientset/versioned"
	informers "github.com/hwameistor/hwameistor/pkg/apis/client/informers/externalversions"
	localdiskactioncontroller "github.com/hwameistor/hwameistor/pkg/local-disk-action-controller"
	"github.com/hwameistor/hwameistor/pkg/utils"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
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

	client, err := clientset.NewForConfig(cfg)
	if err != nil {
		log.WithError(err).Fatal("Failed to create k8s client set")
	}
	ctx := utils.SetupSignalHandler()

	factory := informers.NewSharedInformerFactory(client, time.Second*30)

	controller := localdiskactioncontroller.NewLocalDiskActionController(client,
		factory.Hwameistor().V1alpha1().LocalDisks(),
		factory.Hwameistor().V1alpha1().LocalDiskActions())

	factory.Start(ctx.Done())
	if err := controller.Run(ctx); err != nil {
		log.WithField("detail", err.Error()).Error("failed to run localdisk controller")
		os.Exit(1)
	}

	log.Info("controller exits successfully")
}
