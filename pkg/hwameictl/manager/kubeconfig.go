package manager

import (
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/cmdparser/definitions"
	snapshotv1 "github.com/kubernetes-csi/external-snapshotter/client/v6/apis/volumesnapshot/v1"
	storagev1 "k8s.io/api/storage/v1"
)

func BuildKubeClient(kubeConfigPath string) (*kubernetes.Clientset, client.Client, error) {
	loadingRules := clientcmd.NewDefaultPathOptions().LoadingRules

	if kubeConfigPath != definitions.DefaultKubeConfigPath {
		loadingRules.ExplicitPath = kubeConfigPath
		if !Exists(kubeConfigPath) {
			// Check specified kubeconfig file path if exists
			return nil, nil, fmt.Errorf("kubeconfig is not exists at %v", kubeConfigPath)
		}
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{
		Timeout: definitions.Timeout.String(),
	})

	config, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, nil, err
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}

	options := client.Options{
		Scheme: runtime.NewScheme(),
	}

	// Setup Scheme for resources
	if err = scheme.AddToScheme(options.Scheme); err != nil {
		return nil, nil, err
	}

	if err = apisv1alpha1.AddToScheme(options.Scheme); err != nil {
		return nil, nil, err
	}

	if err = storagev1.AddToScheme(options.Scheme); err != nil {
		return nil, nil, err
	}

	if err = snapshotv1.AddToScheme(options.Scheme); err != nil {
		return nil, nil, err
	}

	kClient, err := client.New(config, options)
	return clientSet, kClient, err
}

func Exists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return os.IsExist(err)
	}
	return true
}
