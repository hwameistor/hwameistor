package hwameistor

import (
	"context"
	"strings"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	hwameistorapi "github.com/hwameistor/hwameistor/pkg/apiserver/api"
)

type VolumeGroupController struct {
	client.Client
	record.EventRecorder

	clientset *kubernetes.Clientset
}

func NewVolumeGroupController(client client.Client, clientset *kubernetes.Clientset, recorder record.EventRecorder) *VolumeGroupController {
	return &VolumeGroupController{
		Client:        client,
		EventRecorder: recorder,
		clientset:     clientset,
	}
}

func (vgController *VolumeGroupController) GetVolumeGroupByVolumeGroupName(vgName string) (hwameistorapi.VolumeGroup, error) {
	var vg = hwameistorapi.VolumeGroup{}
	var vols []apisv1alpha1.LocalVolume

	lvg := &apisv1alpha1.LocalVolumeGroup{}
	if err := vgController.Client.Get(context.TODO(), client.ObjectKey{Name: vgName}, lvg); err != nil {
		if !errors.IsNotFound(err) {
			log.WithError(err).Error("Failed to query LocalVolumeGroup")
		} else {
			log.Info("Not found the LocalVolumeGroup")
		}
		return vg, err
	}
	vg.LocalVolumeGroup = *lvg

	log.Infof("ListVolumesByVolumeGroup lvg.Name = %v, lvg.Spec.Volumes = %v", lvg.Name, lvg.Spec.Volumes)
	vg.Name = lvg.Name

	for _, volumeInfo := range lvg.Spec.Volumes {
		volName := volumeInfo.LocalVolumeName
		log.Infof("ListVolumesByVolumeGroup volName = %v", volName)
		lv := &apisv1alpha1.LocalVolume{}
		if err := vgController.Client.Get(context.TODO(), client.ObjectKey{Name: volName}, lv); err != nil {
			if !errors.IsNotFound(err) {
				log.WithError(err).Error("Failed to query localvolume")
			} else {
				log.Info("Not found the localvolume")
			}
			return vg, err
		}
		vols = append(vols, *lv)
	}

	vg.Volumes = vols
	return vg, nil
}

func (vgController *VolumeGroupController) ListVolumeGroup() (*hwameistorapi.VolumeGroupList, error) {
	var vgList = &hwameistorapi.VolumeGroupList{}
	var vgs []hwameistorapi.VolumeGroup
	lvList := &apisv1alpha1.LocalVolumeList{}
	if err := vgController.Client.List(context.TODO(), lvList); err != nil {
		log.WithError(err).Error("Failed to list LocalVolumes")
		return nil, err
	}

	var vgNames []string
	for _, lv := range lvList.Items {
		var vgsNameStr = strings.Join(vgNames, " ")
		if lv.Spec.VolumeGroup != "" && !strings.Contains(vgsNameStr, lv.Spec.VolumeGroup) {
			vgNames = append(vgNames, lv.Spec.VolumeGroup)
			vg, err := vgController.GetVolumeGroupByVolumeGroupName(lv.Spec.VolumeGroup)
			if err != nil {
				log.WithError(err).Error("Failed to GetVolumeGroupByVolumeGroupName")
				continue
			}
			vgs = append(vgs, vg)
		}
	}
	vgList.VolumeGroups = vgs

	return vgList, nil
}
