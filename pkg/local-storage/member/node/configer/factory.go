package configer

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

var ErrNotSupportedConfiger = fmt.Errorf("no configer found for this system mode")

type SyncReplicaStatus func(replicaName string)

func ConfigerFactory(hostname string, config apisv1alpha1.SystemConfig, apiClient client.Client, syncFunc SyncReplicaStatus) (Configer, error) {
	switch config.Mode {
	case apisv1alpha1.SystemModeDRBD:
		{
			return NewDRBDConfiger(hostname, config, apiClient, syncFunc)
		}
	}
	return nil, ErrNotSupportedConfiger
}
