package hwameistor

import (
	"context"
	"fmt"
	hwameistorapi "github.com/hwameistor/hwameistor/pkg/apiserver/api"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	LvmHwameistor  = "lvm.hwameistor.io"
	DiskHwameistor = "disk.hwameistor.io"
)

type StorageClassController struct {
	client.Client
	record.EventRecorder
}

func NewStorageClassController(client client.Client, recorder record.EventRecorder) *StorageClassController {
	return &StorageClassController{
		Client:        client,
		EventRecorder: recorder,
	}
}

func (scController *StorageClassController) ListHwameistorStroageClass(queryPage hwameistorapi.QueryPage) ([]*hwameistorapi.StorageClass, error) {
	scList := &storagev1.StorageClassList{}
	if err := scController.Client.List(context.TODO(), scList); err != nil {
		log.WithError(err).Error("Failed to list StorageClass")
		return nil, err
	}
	log.Infof("listStorageClass queryPage = %v", queryPage)

	var scs []*hwameistorapi.StorageClass
	for _, item := range scList.Items {

		if item.Provisioner == LvmHwameistor || item.Provisioner == DiskHwameistor {
			var sc = &hwameistorapi.StorageClass{}
			sc.StorageClass = item
			scs = append(scs, sc)
		} else {
			continue
		}

	}
	return scs, nil
}

func (scController *StorageClassController) AddHwameistorStroageClass(name, provisioner string, parameters map[string]string) error {
	s := &storagev1.StorageClass{}
	obj := client.ObjectKey{
		Name: name,
	}
	if err := scController.Client.Get(context.TODO(), obj, s); err == nil {
		return fmt.Errorf("This sc-name : %s is already in use ", name)
	} else {
		if !errors.IsNotFound(err) {
			return err
		}
		var reclaimPolicy v1.PersistentVolumeReclaimPolicy = "Delete"
		var volumeBindingMode storagev1.VolumeBindingMode = "WaitForFirstConsumer"
		var allowVolumeExpansion bool = true
		sc := storagev1.StorageClass{
			ObjectMeta: v12.ObjectMeta{
				Name: name,
			},
			Provisioner:          provisioner,
			ReclaimPolicy:        &reclaimPolicy,
			VolumeBindingMode:    &volumeBindingMode,
			AllowVolumeExpansion: &allowVolumeExpansion,
			Parameters:           parameters,
		}

		if err := scController.Client.Create(context.TODO(), &sc); err != nil {
			log.Errorf("create hwameistor storageclass err: %v", err)
			return err
		}
	}

	return nil
}

func (scController *StorageClassController) DeleteHwameistorStroageClass(name string) error {
	s := &storagev1.StorageClass{}
	obj := client.ObjectKey{
		Name: name,
	}
	if err := scController.Client.Get(context.TODO(), obj, s); err != nil {
		return err
	} else {
		err := scController.Client.Delete(context.TODO(), s)
		if err != nil {
			return err
		}
	}
	return nil
}
