package main

import (
	"context"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/hwameistor/hwameistor/pkg/evictor"
	"github.com/kubernetes-csi/csi-lib-utils/leaderelection"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

const (
	lockName = "hwameistor-volume-evictor"
)

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

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		log.WithError(err).Fatal("Failed to get kubernetes cluster config")
	}
	leClientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.WithError(err).Fatal("Failed to create client set")
	}

	run := func(ctx context.Context) {
		if err := evictor.New(leClientset).Run(signals.SetupSignalHandler()); err != nil {
			log.WithFields(log.Fields{"error": err.Error()}).Error("failed to run evictor")
			os.Exit(1)
		}
	}

	le := leaderelection.NewLeaderElection(leClientset, lockName, run)
	opNamespace, _ := k8sutil.GetOperatorNamespace()
	le.WithNamespace(opNamespace)

	if err := le.Run(); err != nil {
		log.Fatalf("failed to initialize leader election: %v", err)
	}

}
