package hwameistor

import (
	"context"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

// LocalDiskNodeController
type LocalDiskNodeController struct {
	client.Client
	record.EventRecorder
}

// NewLocalDiskNodeController
func NewLocalDiskNodeController(client client.Client, recorder record.EventRecorder) *LocalDiskNodeController {
	return &LocalDiskNodeController{
		Client:        client,
		EventRecorder: recorder,
	}
}

// ListLocalDiskNode
func (ldController *LocalDiskNodeController) ListLocalDiskNode() (*apisv1alpha1.LocalDiskNodeList, error) {
	localdiskList := &apisv1alpha1.LocalDiskNodeList{}
	if err := ldController.Client.List(context.TODO(), localdiskList); err != nil {
		log.WithError(err).Error("Failed to list LocalDiskNodes")
	}
	return localdiskList, nil
}

// GetLocalDiskNode
func (ldController *LocalDiskNodeController) GetLocalDiskNode(key client.ObjectKey) (*apisv1alpha1.LocalDiskNode, error) {
	disk := &apisv1alpha1.LocalDiskNode{}
	if err := ldController.Client.Get(context.TODO(), key, disk); err != nil {
		if !errors.IsNotFound(err) {
			log.WithError(err).Error("Failed to query diskume")
		} else {
			log.Info("Not found the diskume")
		}
		return nil, err
	}
	return disk, nil
}
