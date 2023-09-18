package main

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	autoresizer "github.com/hwameistor/hwameistor/pkg/pvc-autoresizer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	v1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	log "github.com/sirupsen/logrus"
	apimachineryruntime "k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

var (
	scheme   = apimachineryruntime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
}

func main() {
	setupLogging()

	kubeconfig, err := config.GetConfig()
	if err != nil {
		log.Errorf("Get kubeconfig err: %v", err)
		os.Exit(1)
	}

	cli, err := client.New(kubeconfig, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		log.Errorf("New Client err: %v", err)
		os.Exit(1)
	}

	options := manager.Options{
		Namespace: "", // watch all namespaces
	}

	mgr, err := manager.New(kubeconfig, options)
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}


	stopChan := signals.SetupSignalHandler()

	if err := v1alpha1.AddToScheme(mgr.GetScheme()); err != nil {
		log.WithError(err).Error("Failed to setup scheme for all HwameiStor resources")
		os.Exit(1)
	}

	go func() {
		log.Info("Starting the manager of all local storage resources.")
		if err := mgr.Start(stopChan); err != nil {
			log.WithError(err).Error("Failed to run resources manager")
			os.Exit(1)
		}
	}()

	pvcWorkQueue := workqueue.NewNamedRateLimitingQueue(
		workqueue.NewItemExponentialFailureRateLimiter(time.Second, 16*time.Second), 
		"pvc-attacher",
	)
	pvcAttacher := autoresizer.NewPVCAttacher(cli, pvcWorkQueue)
	go pvcAttacher.StartPVCInformer(cli, stopChan)
	go pvcAttacher.Start(stopChan.Done())
	// go autoresizer.StartResizePolicyEventHandler(cli, mgr.GetCache())
	go autoresizer.StartResizePolicyEventHandlerV2(cli, pvcWorkQueue, mgr.GetCache())
	go autoresizer.NewAutoResizer(cli, stopChan).Start()

	select {
	case <-stopChan.Done():
		log.Info("Receive exit signal.")
		time.Sleep(3 * time.Second)
		os.Exit(1)
	}
}

func setupLogging() {
	log.SetFormatter(&log.JSONFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			s := strings.Split(f.Function, ".")
			funcName := s[len(s)-1]
			fileName := path.Base(f.File)
			return funcName, fmt.Sprintf("%s:%d", fileName, f.Line)
		}})
	log.SetReportCaller(true)
}