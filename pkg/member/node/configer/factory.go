package configer

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	udsv1alpha1 "github.com/HwameiStor/local-storage/pkg/apis/uds/v1alpha1"
)

var ErrNotSupportedConfiger = fmt.Errorf("no configer found for this system mode")

type SyncReplicaStatus func(replicaName string)

func ConfigerFactory(hostname string, config udsv1alpha1.SystemConfig, apiClient client.Client, syncFunc SyncReplicaStatus) (Configer, error) {
	switch config.Mode {
	case udsv1alpha1.SystemModeDRBD:
		{
			return NewDRBDConfiger(hostname, config, apiClient, syncFunc)
		}
	}
	return nil, ErrNotSupportedConfiger
}
