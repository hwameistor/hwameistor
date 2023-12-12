package main

import (
	"context"
	"flag"
	"fmt"
	prediction "github.com/hwameistor/hwameistor/pkg/disk-health-prediction"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
	"runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"strings"
	"time"
)

var logLevel = flag.Int("v", 4 /*Log Info*/, "number for the log level verbosity")

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

	setupLogging()

	stopCh := signals.SetupSignalHandler()

	go startDiskHealthPrediction(stopCh)

	select {
	case <-stopCh.Done():
		log.Info("Receive exit signal.")
		time.Sleep(3 * time.Second)
		os.Exit(1)
	}
}

func startDiskHealthPrediction(c context.Context) {
	log.Info("starting disk prediction")
	go prediction.NewPredictor().WithSyncPeriod(time.Hour * 12).StartTimerPredict(c)
}
