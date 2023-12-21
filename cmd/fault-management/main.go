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
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"path"
	"runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"strings"
	"time"
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

	var clientset kubernetes.Interface
	if clientset, err = kubernetes.NewForConfig(cfg); err != nil {
		logrus.WithError(err).Fatal("Failed to create Kubernetes clientset")
	}

	var hmClientSet *hwclientset.Clientset
	if hmClientSet, err = hwclientset.NewForConfig(cfg); err != nil {
		logrus.WithError(err).Fatal("Failed to create hwameistor clientset")
	}

	var kclient client.Client
	if kclient, err = client.New(cfg, client.Options{}); err != nil {
		logrus.WithError(err).Fatal("Failed to create Kubernetes client")
	}

	hmFactory := hwinformers.NewSharedInformerFactory(hmClientSet, time.Second*10)
	_ = apisv1alpha.AddToScheme(scheme.Scheme)

	// ----------------------------------------
	// informers requested
	factory := informers.NewSharedInformerFactory(clientset, 0)
	pvcInformer := factory.Core().V1().PersistentVolumeClaims()
	pvInformer := factory.Core().V1().PersistentVolumes()
	podInformer := factory.Core().V1().Pods()
	lvInformer := hmFactory.Hwameistor().V1alpha1().LocalVolumes()
	lsnInformer := hmFactory.Hwameistor().V1alpha1().LocalStorageNodes()
	ftInformer := hmFactory.Hwameistor().V1alpha1().FaultTickets()
	scLister := factory.Storage().V1().StorageClasses().Lister()

	ctx := signals.SetupSignalHandler()
	if err = utils.RunWithLease(namespace, podName, fmt.Sprintf("hwameistor-fault-management-%s", nodeName), func(_ context.Context) {
		// run faultmanagement controller
		ftManager := faultmanagement.New(nodeName, namespace, kclient, hmClientSet, podInformer, pvcInformer, pvInformer,
			lvInformer, lsnInformer, ftInformer, scLister)

		// start all the requested informer
		factory.Start(ctx.Done())
		hmFactory.Start(ctx.Done())

		if err = ftManager.Run(ctx.Done()); err != nil {
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
		DisableHTMLEscape: true,
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
