package main

import (
	"context"
	"flag"
	"fmt"
	datasetManager "github.com/hwameistor/hwameistor/pkg/dataset-manager"
	"github.com/kubernetes-csi/csi-lib-utils/leaderelection"
	"k8s.io/client-go/informers"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"os"
	"time"

	dsclientset "github.com/hwameistor/datastore/pkg/apis/client/clientset/versioned"
	dsinformers "github.com/hwameistor/datastore/pkg/apis/client/informers/externalversions"
	hmclientset "github.com/hwameistor/hwameistor/pkg/apis/client/clientset/versioned"
)

var (
	showVersion          = flag.Bool("version", false, "Show version.")
	enableLeaderElection = flag.Bool("leader-election", false, "Enable leader election.")
	kubeconfig           = flag.String("kubeconfig", "", "Absolute path to the kubeconfig file. Required only when running out of cluster.")
	rsync                = flag.Duration("rsync", 10*time.Minute, "Rsync interval of the controller.")
	version              = "unknown"
)

func main() {
	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")
	flag.Parse()
	klog.Infof("args: %s", strings.Join(os.Args, " "))

	if *showVersion {
		fmt.Println(os.Args[0], version)
		return
	}
	klog.Infof("Version: %s", version)

	// Create the client config. Use kubeconfig if given, otherwise assume in-cluster.
	config, err := buildConfig(*kubeconfig)
	if err != nil {
		klog.Error(err.Error())
		os.Exit(1)
	}

	// Create the kubeClientset for in-cluster resources
	kubeClientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Error(err.Error())
		os.Exit(1)
	}

	// Create the kubeClientset for datastore resources
	dsClient, err := dsclientset.NewForConfig(config)
	if err != nil {
		klog.Error(err.Error())
		os.Exit(1)
	}

	// Create the kubeClientset for hwameistor resources
	hmClient, err := hmclientset.NewForConfig(config)
	if err != nil {
		klog.Error(err.Error())
		os.Exit(1)
	}

	// Create the kubeClientset for datastore resources
	coreFactory := informers.NewSharedInformerFactory(kubeClientset, *rsync)
	pvInformer := coreFactory.Core().V1().PersistentVolumes()
	dsFactory := dsinformers.NewSharedInformerFactory(dsClient, *rsync)
	datasetInformer := dsFactory.Datastore().V1alpha1().DataSets()

	ctr := datasetManager.New(kubeClientset, dsClient, hmClient, datasetInformer, pvInformer)
	run := func(ctx context.Context) {
		stopCh := ctx.Done()
		dsFactory.Start(stopCh)
		coreFactory.Start(stopCh)
		ctr.Run(stopCh)
	}

	leClientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatalf("Failed to create leaderelection client: %v", err)
	}

	if *enableLeaderElection {
		lockName := "hwameistor-dataset-manager-master"
		le := leaderelection.NewLeaderElection(leClientset, lockName, run)
		if err = le.Run(); err != nil {
			klog.Fatalf("Failed to initialize leader election: %v", err)
		}
	} else {
		run(context.TODO())
	}
}

func buildConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}
