package main

import (
	"fmt"
	"net/http"
	"path"
	"runtime"
	"strings"

	"github.com/hwameistor/hwameistor/pkg/metrics"
	log "github.com/sirupsen/logrus"
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

	stopCh := make(chan struct{})

	handler := metrics.NewHandler()
	handler.Run(stopCh)

	http.Handle("/metrics", handler)
	err := http.ListenAndServe(":8080", handler)
	if err != nil {
		panic(err)
	}
}
