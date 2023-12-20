package manager

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/hwameistor/hwameistor/pkg/apiserver/manager/hwameistor"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/cmdparser/definitions"
)

func buildControllerParameters() (*kubernetes.Clientset, client.Client, record.EventRecorder, error) {
	var recorder record.EventRecorder
	clientSet, kClient, err := BuildKubeClient(definitions.KubeConfigPath)
	return clientSet, kClient, recorder, err
}

func NewLocalVolumeController() (*hwameistor.LocalVolumeController, error) {
	_, kClient, recorder, err := buildControllerParameters()
	return hwameistor.NewLocalVolumeController(kClient, recorder), err
}

func NewLocalStorageNodeController() (*hwameistor.LocalStorageNodeController, error) {
	clientSet, kClient, recorder, err := buildControllerParameters()
	return hwameistor.NewLocalStorageNodeController(kClient, clientSet, recorder), err
}

func NewLocalStoragePoolController() (*hwameistor.LocalStoragePoolController, error) {
	clientSet, kClient, recorder, err := buildControllerParameters()
	return hwameistor.NewLocalStoragePoolController(kClient, clientSet, recorder), err
}

func NewLocalDiskController() (*hwameistor.LocalDiskController, error) {
	_, kClient, recorder, err := buildControllerParameters()
	return hwameistor.NewLocalDiskController(kClient, recorder), err
}

func NewSnapShotClassController() (*hwameistor.SnapShotClassController, error) {
	_, kClient, recorder, err := buildControllerParameters()
	return hwameistor.NewSnapShotClassController(kClient, recorder), err
}
