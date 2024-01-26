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

func NewVolumeSnapShotController() (*hwameistor.VolumeSnapshotController, error) {
	_, kClient, recorder, err := buildControllerParameters()
	return hwameistor.NewVolumeSnapshotController(kClient, recorder), err
}

func NewLocalVolumeSnapshotController() (*hwameistor.LocalSnapshotController, error) {
	_, kClient, recorder, err := buildControllerParameters()
	return hwameistor.NewLocalSnapshotController(kClient, recorder), err
}

func NewStorageClassController() (*hwameistor.StorageClassController, error) {
	_, kClient, recorder, err := buildControllerParameters()
	return hwameistor.NewStorageClassController(kClient, recorder), err
}

func NewSnapShotClassController() (*hwameistor.SnapShotClassController, error) {
	_, kClient, recorder, err := buildControllerParameters()
	return hwameistor.NewSnapShotClassController(kClient, recorder), err
}

func NewMetricsController() (*hwameistor.MetricController, error) {
	clientSet, kClient, recorder, err := buildControllerParameters()
	return hwameistor.NewMetricController(kClient, clientSet, recorder), err
}
