package main

import (
	"fmt"
	"github.com/hwameistor/hwameistor/cmd/admission/app"
	"github.com/hwameistor/hwameistor/pkg/webhook"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"net/http"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	options := webhook.NewServerOption()
	pflag.CommandLine.AddFlagSet(options.AddFlags(&pflag.FlagSet{}))
	pflag.Parse()

	certPath := filepath.Join(options.CertDir, options.TLSCert)
	keyPath := filepath.Join(options.CertDir, options.TLSKey)

	if certPath == "" || keyPath == "" {
		log.Fatal("--cert-dir, --tls-cert-file, --tls-private-key-file is required")
	}

	mux := http.NewServeMux()
	mux.Handle("/mutate", app.RegisterHwameiStorMutateWebhooks(*options))
	server := &http.Server{
		Addr:    ":18443",
		Handler: mux,
	}

	log.Infof("admission server at %v, using tls-cert-file %s, tls-private-key-file %s", server.Addr, certPath, keyPath)
	log.Fatal(server.ListenAndServeTLS(certPath, keyPath))
}

func setupLogging() {
	log.SetLevel(log.TraceLevel)
	log.SetFormatter(&log.JSONFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			s := strings.Split(f.Function, ".")
			funcName := s[len(s)-1]
			fileName := path.Base(f.File)
			return funcName, fmt.Sprintf("%s:%d", fileName, f.Line)
		}})
	log.SetReportCaller(true)
}

func init() {
	setupLogging()
}
