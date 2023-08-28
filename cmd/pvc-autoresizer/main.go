package main

import (
	"os"
	"time"

	autoresizer "github.com/hwameistor/hwameistor/pkg/pvc-autoresizer"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	log "github.com/sirupsen/logrus"
	apimachineryruntime "k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
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

	stopChan := signals.SetupSignalHandler()
	go autoresizer.NewAutoResizer(cli, stopChan).Start()
	go autoresizer.NewHooker(cli, stopChan).Start()

	select {
	case <-stopChan.Done():
		log.Info("Receive exit signal.")
		time.Sleep(3 * time.Second)
		os.Exit(1)
	}
}