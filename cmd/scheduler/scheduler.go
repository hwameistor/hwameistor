// k8s scheduler with local-storage replica scheduling
// scheduling for pod which mount local-storage volume
package main

import (
	"math/rand"
	"time"

	log "github.com/sirupsen/logrus"
	"k8s.io/kubernetes/cmd/kube-scheduler/app"

	"github.com/hwameistor/hwameistor/pkg/scheduler/scheduler"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	command := app.NewSchedulerCommand(
		app.WithPlugin(scheduler.Name, scheduler.New),
	)

	log.Info("Starting the HwameiStor scheduler ...")
	if err := command.Execute(); err != nil {
		log.WithError(err).Fatal("Failed to start the HwameiStor scheduler")
	}

}
