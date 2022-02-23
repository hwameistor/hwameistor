// k8s scheduler with local-storage replica scheduling
// scheduling for pod which mount local-storage volume
package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"k8s.io/component-base/logs"
	"k8s.io/klog"
	"k8s.io/kubernetes/cmd/kube-scheduler/app"

	"github.com/hwameiStor/local-storage/pkg/scheduler"
)

var BUILDVERSION, BUILDTIME, GOVERSION string

func printVersion() {
	klog.V(1).Info(fmt.Sprintf("GitCommit:%q, BuildDate:%q, GoVersion:%q", BUILDVERSION, BUILDTIME, GOVERSION))
}

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	printVersion()
	rand.Seed(time.Now().UTC().UnixNano())

	command := app.NewSchedulerCommand(
		app.WithPlugin(scheduler.Name, scheduler.New),
	)

	if err := command.Execute(); err != nil {
		klog.V(1).Info(err)
		os.Exit(1)
	}

}
