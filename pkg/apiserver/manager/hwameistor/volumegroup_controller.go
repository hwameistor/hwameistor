package hwameistor

import (
	"context"
	"fmt"
	"strings"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	hwameistorapi "github.com/hwameistor/hwameistor/pkg/apiserver/api"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// VolumeGroupController
type VolumeGroupController struct {
	client.Client
	record.EventRecorder

	clientset *kubernetes.Clientset
}

// NewVolumeGroupController
func NewVolumeGroupController(client client.Client, clientset *kubernetes.Clientset, recorder record.EventRecorder) *VolumeGroupController {
	return &VolumeGroupController{
		Client:        client,
		EventRecorder: recorder,
		clientset:     clientset,
	}
}

// ListVolumesByVolumeGroup
func (vgController *VolumeGroupController) GetVolumeGroupByVolumeGroupName(vgName string) (hwameistorapi.VolumeGroup, error) {
	var vis = []hwameistorapi.VolumeInfo{}

	var vg = hwameistorapi.VolumeGroup{}
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

	fmt.Println("ListVolumesByVolumeGroup lvg.Name = %v, lvg.Spec.Volumes = %v", lvg.Name, lvg.Spec.Volumes)
	vg.Name = lvg.Name

	for _, volumeinfo := range lvg.Spec.Volumes {
		var vi = hwameistorapi.VolumeInfo{}
		var vol = &hwameistorapi.Volume{}

		vol.Name = volumeinfo.LocalVolumeName
		if vol.Name == "" {
			vol.Name = volumeinfo.LocalVolumeName
		}
		fmt.Println("ListVolumesByVolumeGroup vgvi.VolumeName = %v", vol.Name)
		lv := &apisv1alpha1.LocalVolume{}
		if err := vgController.Client.Get(context.TODO(), client.ObjectKey{Name: vol.Name}, lv); err != nil {
			if !errors.IsNotFound(err) {
				log.WithError(err).Error("Failed to query localvolume")
			} else {
				log.Info("Not found the localvolume")
			}
			return vg, err
		}
		vol.LocalVolume = *lv

		for _, replicas := range lv.Spec.Config.Replicas {
			vi.NodeNames = append(vi.NodeNames, replicas.Hostname)
		}
		vi.Volume = vol

		vis = append(vis, vi)
	}

	vg.Name = vgName

	return vg, nil
}

func (vgController *VolumeGroupController) ListVolumeGroup() (*hwameistorapi.VolumeGroupList, error) {

	var vglist = &hwameistorapi.VolumeGroupList{}
	var vgs = []hwameistorapi.VolumeGroup{}
	lvList := &apisv1alpha1.LocalVolumeList{}
	if err := vgController.Client.List(context.TODO(), lvList); err != nil {
		log.WithError(err).Error("Failed to list LocalVolumes")
		return nil, err
	}

	var vgnames []string
	for _, lv := range lvList.Items {
		var vgsnamestr string = strings.Join(vgnames, " ")
		if lv.Spec.VolumeGroup != "" && !strings.Contains(vgsnamestr, lv.Spec.VolumeGroup) {
			vgnames = append(vgnames, lv.Spec.VolumeGroup)

			vg, err := vgController.GetVolumeGroupByVolumeGroupName(lv.Spec.VolumeGroup)
			if err != nil {
				log.WithError(err).Error("Failed to GetVolumeGroupByVolumeGroupName")
				continue
			}

			vgs = append(vgs, vg)
		}
	}
	vglist.VolumeGroupNames = vgnames
	vglist.VolumeGroups = vgs

	return vglist, nil
}
