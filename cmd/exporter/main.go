package main

import (
	"flag"

	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	"github.com/hwameistor/hwameistor/pkg/exporter"
)

func setupLogging() {
	log.SetLevel(log.DebugLevel)
}

func main() {
	flag.Parse()
	setupLogging()

	exporter.NewCollectorManager().Run(signals.SetupSignalHandler().Done())
}
