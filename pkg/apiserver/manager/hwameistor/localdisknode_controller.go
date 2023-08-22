package hwameistor

import (
	"context"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

type LocalDiskNodeController struct {
	client.Client
	record.EventRecorder
}

func NewLocalDiskNodeController(client client.Client, recorder record.EventRecorder) *LocalDiskNodeController {
	return &LocalDiskNodeController{
		Client:        client,
		EventRecorder: recorder,
	}
}

func (ldController *LocalDiskNodeController) ListLocalDiskNode() (*apisv1alpha1.LocalDiskNodeList, error) {
	localDiskList := &apisv1alpha1.LocalDiskNodeList{}
	if err := ldController.Client.List(context.TODO(), localDiskList); err != nil {
		log.WithError(err).Error("Failed to list LocalDiskNodes")
	}
	return localDiskList, nil
}

func (ldController *LocalDiskNodeController) GetLocalDiskNode(key client.ObjectKey) (*apisv1alpha1.LocalDiskNode, error) {
	node := &apisv1alpha1.LocalDiskNode{}
	if err := ldController.Client.Get(context.TODO(), key, node); err != nil {
		if !errors.IsNotFound(err) {
			log.WithError(err).Error("Failed to query localDiskNode")
		} else {
			log.Info("Not found the localDiskNode")
		}
		return nil, err
	}
	return node, nil
}
