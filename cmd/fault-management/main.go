package main

import (
	"context"
	"flag"
	"fmt"
	hwclientset "github.com/hwameistor/hwameistor/pkg/apis/client/clientset/versioned"
	"github.com/hwameistor/hwameistor/pkg/apis/client/clientset/versioned/scheme"
	hwinformers "github.com/hwameistor/hwameistor/pkg/apis/client/informers/externalversions"
	apisv1alpha "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	faultmanagement "github.com/hwameistor/hwameistor/pkg/fault-management"
	"github.com/hwameistor/hwameistor/pkg/local-storage/utils"
	"github.com/sirupsen/logrus"
	"path"
	"runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"strings"
)

var (
	BUILDVERSION, BUILDTIME, GOVERSION string
	logLevel                           = flag.Int("v", 5 /*Log Debug*/, "number for the log level verbosity")
	namespace                          = utils.GetNamespace()
	nodeName                           = utils.GetNodeName()
	podName                            = utils.GetPodName()
)

func printVersion() {
	logrus.Info(fmt.Sprintf("GitCommit:%q, BuildDate:%q, GoVersion:%q", BUILDVERSION, BUILDTIME, GOVERSION))
}

func main() {
	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to get kubernetes cluster config")
	}

	var hwClientSet *hwclientset.Clientset
	if hwClientSet, err = hwclientset.NewForConfig(cfg); err != nil {
		logrus.WithError(err).Fatal("Failed to create hwameistor clientset")
	}

	var kclient client.Client
	if kclient, err = client.New(cfg, client.Options{}); err != nil {
		logrus.WithError(err).Fatal("Failed to create Kubernetes client")
	}

	hwShareInformer := hwinformers.NewSharedInformerFactory(hwClientSet, 0)
	_ = apisv1alpha.AddToScheme(scheme.Scheme)

	ctx := signals.SetupSignalHandler()
	if err = utils.RunWithLease(namespace, podName, "hwameistor-fault-management", func(_ context.Context) {
		faultmanager := faultmanagement.New(nodeName, namespace, kclient,
			hwShareInformer.Hwameistor().V1alpha1().FaultTickets())
		hwShareInformer.Start(ctx.Done())

		// run faultmanagement controller
		if err = faultmanager.Run(ctx.Done()); err != nil {
			logrus.WithError(err).Fatal("Failed to run faultmanagement controller")
			return
		}
	}); err != nil {
		logrus.WithError(err).Fatal("failed to start hwameistor fault management controller")
	}
	return
}

func setupLogging() {
	// parse log level(default level: info)
	var level logrus.Level
	if *logLevel >= int(logrus.TraceLevel) {
		level = logrus.TraceLevel
	} else if *logLevel <= int(logrus.PanicLevel) {
		level = logrus.PanicLevel
	} else {
		level = logrus.Level(*logLevel)
	}

	logrus.SetLevel(level)
	logrus.SetFormatter(&logrus.JSONFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			s := strings.Split(f.Function, ".")
			funcName := s[len(s)-1]
			fileName := path.Base(f.File)
			return funcName, fmt.Sprintf("%s:%d", fileName, f.Line)
		}})
	logrus.SetReportCaller(true)
}

func init() {
	flag.Parse()
	setupLogging()
	printVersion()
}
