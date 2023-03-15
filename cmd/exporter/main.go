package main

import (
	"flag"
	"fmt"
	"path"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	"github.com/hwameistor/hwameistor/pkg/exporter"
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

	exporter.NewCollectorManager().Run(signals.SetupSignalHandler().Done())

}
