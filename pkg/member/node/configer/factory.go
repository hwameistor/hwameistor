package configer

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	localstoragev1alpha1 "github.com/hwameistor/local-storage/pkg/apis/localstorage/v1alpha1"
)

var ErrNotSupportedConfiger = fmt.Errorf("no configer found for this system mode")

type SyncReplicaStatus func(replicaName string)

func ConfigerFactory(hostname string, config localstoragev1alpha1.SystemConfig, apiClient client.Client, syncFunc SyncReplicaStatus) (Configer, error) {
	switch config.Mode {
	case localstoragev1alpha1.SystemModeDRBD:
		{
			return NewDRBDConfiger(hostname, config, apiClient, syncFunc)
		}
	}
	return nil, ErrNotSupportedConfiger
}
