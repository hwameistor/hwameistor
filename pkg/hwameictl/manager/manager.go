package manager

import (
	"context"
	"fmt"
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/apiserver/api"
	"github.com/hwameistor/hwameistor/pkg/apiserver/controller"
	"github.com/hwameistor/hwameistor/pkg/apiserver/manager"
	mgrpkg "sigs.k8s.io/controller-runtime/pkg/manager"
)

func NewServerManager(kubeconfig string) (*manager.ServerManager, error) {
	clientset, cfg, err := BuildKubeClient(kubeconfig)

	// Create a new manager to provide shared dependencies and start components
	mgr, err := mgrpkg.New(cfg, mgrpkg.Options{})
	if err != nil {
		return nil, err
	}

	// Setup Scheme for all resources
	if err = api.AddToScheme(mgr.GetScheme()); err != nil {
		return nil, err
	}

	if err = apisv1alpha1.AddToScheme(mgr.GetScheme()); err != nil {
		return nil, err
	}

	// Setup all Controllers
	if err = controller.AddToManager(mgr); err != nil {
		return nil, err
	}

	go func() {
		err = mgr.Start(context.TODO())
		if err != nil {
			fmt.Println(err)
		}
	}()
	mgr.GetCache().WaitForCacheSync(context.TODO())

	return manager.NewServerManager(mgr, clientset)
}
