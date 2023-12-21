package hwameistor

import (
	"context"
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	hwameistorapi "github.com/hwameistor/hwameistor/pkg/apiserver/api"
	utils "github.com/hwameistor/hwameistor/pkg/apiserver/util"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"math"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

type LocalSnapshotController struct {
	client.Client
	record.EventRecorder
}

func NewLocalSnapshotController(client client.Client, recorder record.EventRecorder) *LocalSnapshotController {
	return &LocalSnapshotController{
		Client:        client,
		EventRecorder: recorder,
	}
}

func (lvsController *LocalSnapshotController) ListLocalSnapshot(queryPage hwameistorapi.QueryPage) (*hwameistorapi.SnapshotList, error) {
	lvsList := &hwameistorapi.SnapshotList{}
	shots, err := lvsController.listLocalSnapshot(queryPage)
	log.Infof("ListSnapshot shots = %v", shots)
	if err != nil {
		log.WithError(err).Error("Failed to ListSnapshot")
		return nil, err
	}

	lvsList.Snapshots = utils.DataPatination(shots, queryPage.Page, queryPage.PageSize)
	if len(shots) == 0 {
		lvsList.Snapshots = []*hwameistorapi.Snapshot{}
	}

	var pagination = &hwameistorapi.Pagination{}
	pagination.Page = queryPage.Page
	pagination.PageSize = queryPage.PageSize
	pagination.Total = uint32(len(shots))
	if len(shots) == 0 {
		pagination.Pages = 0
	} else {
		pagination.Pages = int32(math.Ceil(float64(len(shots)) / float64(queryPage.PageSize)))
	}
	lvsList.Page = pagination

	return lvsList, nil
}

func (lvsController *LocalSnapshotController) listLocalSnapshot(queryPage hwameistorapi.QueryPage) ([]*hwameistorapi.Snapshot, error) {
	lvsList := &apisv1alpha1.LocalVolumeSnapshotList{}
	if err := lvsController.Client.List(context.TODO(), lvsList); err != nil {
		log.WithError(err).Error("Failed to list LocalSnapshot")
		return nil, err
	}
	log.Infof("listLocalSnapshot queryPage = %v, queryPage.SnapshotState = %v", queryPage, queryPage.SnapshotState)

	var snaps []*hwameistorapi.Snapshot
	for _, lvs := range lvsList.Items {
		var vol = &hwameistorapi.Snapshot{}
		vol.LocalVolumeSnapshot = lvs

		if queryPage.SnapshotName == "" && queryPage.SnapshotState == "" && queryPage.VolumeName == "" {
			snaps = append(snaps, vol)
		} else if queryPage.SnapshotName != "" && strings.Contains(vol.Name, queryPage.SnapshotName) && queryPage.SnapshotState == "" && queryPage.VolumeName == "" {
			snaps = append(snaps, vol)
		} else if queryPage.SnapshotName == "" && queryPage.SnapshotState != "" && queryPage.SnapshotState == vol.Status.State && queryPage.VolumeName == "" {
			snaps = append(snaps, vol)
		} else if queryPage.SnapshotName == "" && queryPage.SnapshotState == "" && queryPage.VolumeName != "" && queryPage.VolumeName == vol.Spec.SourceVolume {
			snaps = append(snaps, vol)
		} else if queryPage.SnapshotName != "" && strings.Contains(vol.Name, queryPage.SnapshotName) && queryPage.SnapshotState != "" && queryPage.SnapshotState == vol.Status.State && queryPage.VolumeName == "" {
			snaps = append(snaps, vol)
		} else if queryPage.SnapshotName != "" && strings.Contains(vol.Name, queryPage.SnapshotName) && queryPage.SnapshotState == apisv1alpha1.VolumeStateEmpty && queryPage.VolumeName != "" && queryPage.VolumeName == vol.Spec.SourceVolume {
			snaps = append(snaps, vol)
		} else if queryPage.SnapshotName == "" && queryPage.SnapshotState != "" && queryPage.SnapshotState == vol.Status.State && queryPage.VolumeName != "" && queryPage.VolumeName == vol.Spec.SourceVolume {
			snaps = append(snaps, vol)
		} else if queryPage.SnapshotName != "" && strings.Contains(vol.Name, queryPage.SnapshotName) && queryPage.SnapshotState != "" && queryPage.SnapshotState == vol.Status.State && queryPage.VolumeName != "" && queryPage.VolumeName == vol.Spec.SourceVolume {
			snaps = append(snaps, vol)
		}
	}

	return snaps, nil
}

func (lvsController *LocalSnapshotController) GetLocalSnapshot(lvsname string) (*hwameistorapi.Snapshot, error) {
	var queryPage hwameistorapi.QueryPage
	queryPage.SnapshotName = lvsname

	shots, err := lvsController.listLocalSnapshot(queryPage)
	if err != nil {
		log.WithError(err).Error("Failed to listLocalVolume")
		return nil, err
	}

	for _, lvs := range shots {
		if lvs.Name == lvsname {
			return lvs, nil
		}
	}
	return nil, nil
}

func (lvsController *LocalSnapshotController) RollbackVolumeSnapshot(lvsname string) error {

	s := &apisv1alpha1.LocalVolumeSnapshot{}
	obj := client.ObjectKey{
		Name:      lvsname,
		Namespace: defaultNamespace,
	}

	err := lvsController.Client.Get(context.TODO(), obj, s)
	if err != nil {
		return err
	}
	restore := &apisv1alpha1.LocalVolumeSnapshotRestore{
		TypeMeta: metav1.TypeMeta{
			Kind:       "LocalVolumeSnapshotRestore",
			APIVersion: "hwameistor.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "restore-" + lvsname,
			Namespace: defaultNamespace,
		},
		Spec: apisv1alpha1.LocalVolumeSnapshotRestoreSpec{
			SourceVolumeSnapshot: lvsname,
			RestoreType:          "rollback",
		},
	}
	if err := lvsController.Client.Create(context.TODO(), restore); err != nil {
		log.Errorf("create LocalVolumeSnapshotRestore err: %v", err)
		return err
	}
	return nil
}
