package hwameistor

import (
	"context"
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	log "github.com/sirupsen/logrus"
)

// LocalDiskController
type LocalDiskController struct {
	client.Client
	record.EventRecorder
}

// NewLocalDiskController
func NewLocalDiskController(client client.Client, recorder record.EventRecorder) *LocalDiskController {
	return &LocalDiskController{
		Client:        client,
		EventRecorder: recorder,
	}
}

// ListLocalDisk
func (ldController *LocalDiskController) ListLocalDisk() (*apisv1alpha1.LocalDiskList, error) {
	diskList := &apisv1alpha1.LocalDiskList{}
	if err := ldController.Client.List(context.TODO(), diskList); err != nil {
		log.WithError(err).Error("Failed to list LocalDisks")
	}
	return diskList, nil
}

// GetLocalDisk
func (ldController *LocalDiskController) GetLocalDisk(key client.ObjectKey) (*apisv1alpha1.LocalDisk, error) {
	disk := &apisv1alpha1.LocalDisk{}
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
