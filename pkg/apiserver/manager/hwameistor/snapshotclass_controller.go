package hwameistor

import (
	"context"
	"fmt"
	hwameistorapi "github.com/hwameistor/hwameistor/pkg/apiserver/api"
	snapshotv1 "github.com/kubernetes-csi/external-snapshotter/client/v6/apis/volumesnapshot/v1"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
)

const Hwameistor = "lvm.hwameistor.io"

type SnapShotClassController struct {
	client.Client
	record.EventRecorder
}

func NewSnapShotClassController(client client.Client, recorder record.EventRecorder) *SnapShotClassController {
	return &SnapShotClassController{
		Client:        client,
		EventRecorder: recorder,
	}
}

func (sscController *SnapShotClassController) ListSnapShotClass() ([]*hwameistorapi.SnapshotClass, error) {
	sscList := &snapshotv1.VolumeSnapshotClassList{}
	if err := sscController.Client.List(context.TODO(), sscList); err != nil {
		log.WithError(err).Error("Failed to list SnapshotClass")
		return nil, err
	}
	//log.Infof("listSnapshotClass queryPage = %v", queryPage)

	var sscs []*hwameistorapi.SnapshotClass
	for _, item := range sscList.Items {

		var ssc = &hwameistorapi.SnapshotClass{}
		ssc.VolumeSnapshotClass = item
		sscs = append(sscs, ssc)
	}

	return sscs, nil
}

func (scController *SnapShotClassController) AddHwameistorSnapshotClass(name, snapsize string) error {
	s := &snapshotv1.VolumeSnapshotClass{}
	obj := client.ObjectKey{
		Name: name,
	}
	if err := scController.Client.Get(context.TODO(), obj, s); err == nil {
		return fmt.Errorf("This ssc-name : %s is already in use ", name)
	} else {
		if !errors.IsNotFound(err) {
			return err
		}
		snapshotclass := &snapshotv1.VolumeSnapshotClass{
			TypeMeta: v12.TypeMeta{
				Kind:       "VolumeSnapshotClass",
				APIVersion: "snapshot.storage.k8s.io/v1",
			},
			ObjectMeta: v12.ObjectMeta{
				Name:      name,
				Namespace: "default",
				Annotations: map[string]string{
					"snapshot.storage.kubernetes.io/is-default-class": "true",
				},
			},
			Driver:         "lvm.hwameistor.io",
			DeletionPolicy: "Delete",
		}

		if snapsize != "" {
			var size int64
			num, err := strconv.ParseInt(snapsize[:len(snapsize)-1], 10, 64)
			if err != nil {
				log.WithError(err).Error("Failed to convert snapsize to numeric value ")
			}

			switch snapsize[len(snapsize)-1:] {
			case "G":
				size = 1024 * 1024 * 1024 * num
			case "M":
				size = 1024 * 1024 * num
			case "K":
				size = 1024 * num
			default:
				log.WithError(err).Error("The unit is invalid. Please use G, M, K.")
				return err
			}

			snapsize = strconv.FormatInt(size, 10)
			snapshotclass.Parameters = map[string]string{
				"snapsize": snapsize,
			}
		}

		if err := scController.Client.Create(context.TODO(), snapshotclass); err != nil {
			log.Errorf("create hwameistor snapshotclass err: %v", err)
			return err
		}
	}

	return nil
}

func (scController *SnapShotClassController) DeleteHwameistorSnapshotClass(name string) error {
	s := &snapshotv1.VolumeSnapshotClass{}
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
