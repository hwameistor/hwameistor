package manager

import (
	"errors"
	"os"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/apiserver/api"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/cmdparser/definitions"
)

func BuildKubeClient(kubeconfigPath string) (client.Client, error) {
	if !Exists(kubeconfigPath) {
		return nil, errors.New("kubeconfig file is not exists")
	}
	loadingRules := clientcmd.NewDefaultPathOptions().LoadingRules
	loadingRules.ExplicitPath = kubeconfigPath
	overrides := &clientcmd.ConfigOverrides{
		Timeout: definitions.Timeout.String(),
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides)

	config, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	options := client.Options{
		Scheme: runtime.NewScheme(),
	}

	// Setup Scheme for all resources
	if err = api.AddToScheme(options.Scheme); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}
	if err = apisv1alpha1.AddToScheme(options.Scheme); err != nil {
		log.Error(err, "Failed to setup scheme for ldm resources")
		os.Exit(1)
	}

	return client.New(config, options)
}

func Exists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return os.IsExist(err)
	}
	return true
}
