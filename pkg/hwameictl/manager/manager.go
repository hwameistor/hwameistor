package manager

import (
	"github.com/hwameistor/hwameistor/pkg/apiserver/manager/hwameistor"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/hwameistor/hwameistor/pkg/hwameictl/cmdparser/definitions"
)

func buildControllerParameters() (client.Client, record.EventRecorder, error) {
	var recorder record.EventRecorder
	c, err := BuildKubeClient(definitions.Kubeconfig)
	if err != nil {
		return nil, nil, err
	}
	return c, recorder, nil
}

func NewLocalVolumeController() (*hwameistor.LocalVolumeController, error) {
	c, recorder, err := buildControllerParameters()
	if err != nil {
		return nil, err
	}
	return hwameistor.NewLocalVolumeController(c, recorder), nil
}
