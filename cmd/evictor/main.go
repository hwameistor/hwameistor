package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"

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

var (
	logLevel = flag.Int("v", 4 /*Log Info*/, "number for the log level verbosity")
)

func setupLogging() {
	// parse log level(default level: info)
	var level log.Level
	if *logLevel >= int(log.TraceLevel) {
		level = log.TraceLevel
	} else if *logLevel <= int(log.PanicLevel) {
		level = log.PanicLevel
	} else {
		level = log.Level(*logLevel)
	}

	log.SetLevel(level)
	log.SetFormatter(&log.JSONFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			s := strings.Split(f.Function, ".")
			funcName := s[len(s)-1]
			fileName := path.Base(f.File)
			return funcName, fmt.Sprintf("%s:%d", fileName, f.Line)
		}})
	log.SetReportCaller(true)
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
