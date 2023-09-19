package hwameistor

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

type LocalDiskController struct {
	client.Client
	record.EventRecorder
}

func NewLocalDiskController(client client.Client, recorder record.EventRecorder) *LocalDiskController {
	return &LocalDiskController{
		Client:        client,
		EventRecorder: recorder,
	}
}

func (ldController *LocalDiskController) ListLocalDisk() (*apisv1alpha1.LocalDiskList, error) {
	diskList := &apisv1alpha1.LocalDiskList{}
	if err := ldController.Client.List(context.TODO(), diskList); err != nil {
		log.WithError(err).Error("Failed to list LocalDisks")
	}
	return diskList, nil
}

func (ldController *LocalDiskController) GetLocalDisk(localDiskName string) (*apisv1alpha1.LocalDisk, error) {
	disk := &apisv1alpha1.LocalDisk{}
	if err := ldController.Client.Get(context.TODO(), types.NamespacedName{Name: localDiskName}, disk); err != nil {
		if !errors.IsNotFound(err) {
			log.WithError(err).Error("Failed to query LocalDisk")
		} else {
			log.Info("Not found the LocalDisk")
		}
		return nil, err
	}
	return disk, nil
}

func (ldController *LocalDiskController) GetLocalDiskByPath(nodeName, shortPath string) (*apisv1alpha1.LocalDisk, error) {
	diskList, err := ldController.ListLocalDisk()
	if err != nil {
		return nil, err
	}
	for _, disk := range diskList.Items {
		if disk.Spec.NodeName == nodeName && disk.Spec.DevicePath == fmt.Sprintf("/dev/%s", shortPath) {
			return &disk, nil
		}
	}
	return nil, fmt.Errorf("not found the LocalDisk")
}

func (ldController *LocalDiskController) SetLocalDiskOwner(localDiskName string, owner string) error {
	disk, err := ldController.GetLocalDisk(localDiskName)
	if err != nil {
		return err
	}
	disk.Spec.Owner = owner
	return ldController.Client.Update(context.TODO(), disk)
}
