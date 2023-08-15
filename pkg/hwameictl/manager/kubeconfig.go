package manager

import (
	"errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

func BuildKubeClient(kubeconfigPath string) (*kubernetes.Clientset, *rest.Config, error) {
	if !Exists(kubeconfigPath) {
		return nil, nil, errors.New("kubeconfig file is not exists")
	}

	loadingRules := clientcmd.NewDefaultPathOptions().LoadingRules
	loadingRules.ExplicitPath = kubeconfigPath
	overrides := &clientcmd.ConfigOverrides{
		//Context:         api.Context{},
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides)

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, nil, err
	}

	clientset, err := kubernetes.NewForConfig(restConfig)

	if err != nil {
		return nil, nil, err
	}

	return clientset, restConfig, nil
}

func Exists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return os.IsExist(err)
	}
	return true
}
