package hwameistor

import (
	"context"
	"fmt"
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	hwameistorapi "github.com/hwameistor/hwameistor/pkg/apiserver/api"
	snapshotv1 "github.com/kubernetes-csi/external-snapshotter/client/v6/apis/volumesnapshot/v1"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const defaultNamespace = "default"

type VolumeSnapshotController struct {
	client.Client
	record.EventRecorder
}

func NewVolumeSnapshotController(client client.Client, recorder record.EventRecorder) *VolumeSnapshotController {
	return &VolumeSnapshotController{
		Client:        client,
		EventRecorder: recorder,
	}
}

func (vsController *VolumeSnapshotController) ListVolumeSnapShot() ([]*hwameistorapi.VolumeSnapshot, error) {
	vsList := &snapshotv1.VolumeSnapshotList{}
	if err := vsController.Client.List(context.TODO(), vsList); err != nil {
		log.WithError(err).Error("Failed to list VolumeSnapshot")
		return nil, err
	}
	//log.Infof("listSnapshotClass queryPage = %v", queryPage)

	var vss []*hwameistorapi.VolumeSnapshot
	for _, item := range vsList.Items {

		var vs = &hwameistorapi.VolumeSnapshot{}
		vs.VolumeSnapshot = item
		vss = append(vss, vs)
	}

	return vss, nil
}

func (vsController *VolumeSnapshotController) AddVolumeSnapshot(name, volumesnapshotClassName, pvcName, ns string) error {
	vs := &snapshotv1.VolumeSnapshot{}
	obj := client.ObjectKey{
		Name:      name,
		Namespace: ns,
	}
	if err := vsController.Client.Get(context.TODO(), obj, vs); err == nil {
		return fmt.Errorf("This vs-name : %s is already in use ", name)
	} else {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("get vsc error: %v", err)
		}
		s := &snapshotv1.VolumeSnapshotClass{}
		vsc_obj := client.ObjectKey{
			Name: volumesnapshotClassName,
		}
		if err := vsController.Client.Get(context.TODO(), vsc_obj, s); err != nil {
			return fmt.Errorf("This volumesnapshotClassName-name : %s is not found ", volumesnapshotClassName)
		}

		pvc := &corev1.PersistentVolumeClaim{}
		pvc_obj := client.ObjectKey{
			Name:      pvcName,
			Namespace: defaultNamespace,
		}
		if err := vsController.Client.Get(context.TODO(), pvc_obj, pvc); err != nil {
			return fmt.Errorf("This pvc-name : %s is not found ", pvcName)
		}
		snapshot := &snapshotv1.VolumeSnapshot{
			TypeMeta: metav1.TypeMeta{
				Kind:       "VolumeSnapshot",
				APIVersion: "snapshot.storage.k8s.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: defaultNamespace,
			},
			Spec: snapshotv1.VolumeSnapshotSpec{
				Source: snapshotv1.VolumeSnapshotSource{
					PersistentVolumeClaimName: &pvcName,
				},
				VolumeSnapshotClassName: &volumesnapshotClassName,
			},
			Status: nil,
		}
		if err := vsController.Client.Create(context.TODO(), snapshot); err != nil {
			log.Errorf("create volumesnapshot err: %v", err)
			return err
		}
	}
	return nil
}

func (vsController *VolumeSnapshotController) DeleteVolumeSnapshot(name, ns string) error {
	s := &snapshotv1.VolumeSnapshot{}
	obj := client.ObjectKey{
		Name:      name,
		Namespace: ns,
	}
	if err := vsController.Client.Get(context.TODO(), obj, s); err != nil {
		return err
	} else {
		err := vsController.Client.Delete(context.TODO(), s)
		if err != nil {
			return err
		}
	}
	return nil
}

func (vsController *VolumeSnapshotController) RestoreVolumeSnapshot(name, pvcName, storage, ns string) error {
	if pvcName == "" || storage == "" {
		return fmt.Errorf("sc,pvc,storage Not Null")
	}

	vs := &snapshotv1.VolumeSnapshot{}
	obj := client.ObjectKey{
		Name:      name,
		Namespace: ns,
	}
	if err := vsController.Client.Get(context.TODO(), obj, vs); err != nil {
		return fmt.Errorf("This vs-name : %s is nof found  ", name)
	} else {
		pvc := *vs.Spec.Source.PersistentVolumeClaimName
		p := &corev1.PersistentVolumeClaim{}
		pvc_obj := client.ObjectKey{
			Name:      pvc,
			Namespace: defaultNamespace,
		}
		if err = vsController.Client.Get(context.TODO(), pvc_obj, p); err != nil {
			return err
		}

		sc := p.Spec.StorageClassName

		APIGroup := "snapshot.storage.k8s.io"
		restorePvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pvcName,
				Namespace: defaultNamespace,
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse(storage),
					},
				},
				StorageClassName: sc,
				DataSource: &corev1.TypedLocalObjectReference{
					APIGroup: &APIGroup,
					Kind:     "VolumeSnapshot",
					Name:     vs.Name,
				},
			},
		}
		err := vsController.Client.Create(context.TODO(), restorePvc)
		if err != nil {
			return err
		}
	}
	return nil
}

func (vsController *VolumeSnapshotController) RollbackVolumeSnapshot(vsname, ns string) error {

	vs := &snapshotv1.VolumeSnapshot{}
	obj := client.ObjectKey{
		Name:      vsname,
		Namespace: ns,
	}
	if err := vsController.Client.Get(context.TODO(), obj, vs); err != nil {
		return fmt.Errorf("This vs-name : %s is not found ", vsname)
	} else {
		lvsName := vs.Status.BoundVolumeSnapshotContentName
		restore := &apisv1alpha1.LocalVolumeSnapshotRestore{
			TypeMeta: metav1.TypeMeta{
				Kind:       "LocalVolumeSnapshotRestore",
				APIVersion: "hwameistor.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "restore-" + *lvsName,
				Namespace: defaultNamespace,
			},
			Spec: apisv1alpha1.LocalVolumeSnapshotRestoreSpec{
				SourceVolumeSnapshot: *lvsName,
				RestoreType:          "rollback",
			},
		}
		if err := vsController.Client.Create(context.TODO(), restore); err != nil {
			log.Errorf("create LocalVolumeSnapshotRestore err: %v", err)
			return err
		}

	}
	return nil
}
