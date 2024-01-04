package hwameistor

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	hwameistorapi "github.com/hwameistor/hwameistor/pkg/apiserver/api"
	utils "github.com/hwameistor/hwameistor/pkg/apiserver/util"
)

const (
	ConvertReplicaNum = 2
	DESC              = "DESC"
	ASC               = "ASC"
)

type LocalVolumeController struct {
	client.Client
	record.EventRecorder
}

func NewLocalVolumeController(client client.Client, recorder record.EventRecorder) *LocalVolumeController {
	return &LocalVolumeController{
		Client:        client,
		EventRecorder: recorder,
	}
}

func (lvController *LocalVolumeController) ListLocalVolume(queryPage hwameistorapi.QueryPage) (*hwameistorapi.VolumeList, error) {
	var volList = &hwameistorapi.VolumeList{}
	vols, err := lvController.listLocalVolume(queryPage)
	log.Infof("ListLocalVolume vols = %v", vols)
	if err != nil {
		log.WithError(err).Error("Failed to listLocalVolume")
		return nil, err
	}

	if queryPage.SortDir == DESC || queryPage.SortDir == "" {
		switch queryPage.Sort {
		case "name":
			a := utils.ByVolumeNameDesc(vols)
			sort.Sort(a)
		case "", "time":
			a := utils.ByVolumeTimeDesc(vols)
			sort.Sort(a)
		case "namespace":
			a := utils.ByVolumeNsDesc(vols)
			sort.Sort(a)
		}
	} else if queryPage.SortDir == ASC {
		switch queryPage.Sort {
		case "name":
			a := utils.ByVolumeNameAsc(vols)
			sort.Sort(a)
		case "", "time":
			a := utils.ByVolumeTimeAsc(vols)
			sort.Sort(a)
		case "namespace":
			a := utils.ByVolumeNsAsc(vols)
			sort.Sort(a)
		}
	}

	//Pagination
	volList.Volumes = utils.DataPatination(vols, queryPage.Page, queryPage.PageSize)
	if len(vols) == 0 {
		volList.Volumes = []*hwameistorapi.Volume{}
	}

	var pagination = &hwameistorapi.Pagination{}
	pagination.Page = queryPage.Page
	pagination.PageSize = queryPage.PageSize
	pagination.Total = uint32(len(vols))
	if len(vols) == 0 {
		pagination.Pages = 0
	} else {
		pagination.Pages = int32(math.Ceil(float64(len(vols)) / float64(queryPage.PageSize)))
	}
	volList.Page = pagination

	return volList, nil
}

func (lvController *LocalVolumeController) listLocalVolume(queryPage hwameistorapi.QueryPage) ([]*hwameistorapi.Volume, error) {
	lvList := &apisv1alpha1.LocalVolumeList{}
	if err := lvController.Client.List(context.TODO(), lvList); err != nil {
		log.WithError(err).Error("Failed to list LocalVolumes")
		return nil, err
	}
	log.Infof("listLocalVolume queryPage = %v, queryPage.VolumeState = %v", queryPage, queryPage.VolumeState)

	var vols []*hwameistorapi.Volume
	for _, lv := range lvList.Items {
		var vol = &hwameistorapi.Volume{}
		vol.LocalVolume = lv
		if (queryPage.VolumeName == "") && (queryPage.VolumeState == "") && (queryPage.NameSpace == "") && (queryPage.VolumeGroup == "") {
			vols = append(vols, vol)
		} else if (queryPage.VolumeName != "" && strings.Contains(vol.Name, queryPage.VolumeName)) && (queryPage.VolumeState == "") &&
			(queryPage.NameSpace == "") && (queryPage.VolumeGroup == "") {
			vols = append(vols, vol)
		} else if (queryPage.VolumeName != "" && strings.Contains(vol.Name, queryPage.VolumeName)) && (queryPage.VolumeState != "") &&
			(queryPage.VolumeState == vol.Status.State) && (queryPage.NameSpace == "") && (queryPage.VolumeGroup == "") {
			vols = append(vols, vol)
		} else if (queryPage.VolumeName != "" && strings.Contains(vol.Name, queryPage.VolumeName)) && (queryPage.VolumeState == "") &&
			(queryPage.NameSpace != "") && (queryPage.NameSpace == vol.Spec.PersistentVolumeClaimNamespace) && (queryPage.VolumeGroup == "") {
			vols = append(vols, vol)
		} else if (queryPage.VolumeName != "" && strings.Contains(vol.Name, queryPage.VolumeName)) && (queryPage.VolumeState == "") &&
			(queryPage.NameSpace == "") && (queryPage.VolumeGroup != "") && strings.Contains(vol.Spec.VolumeGroup, queryPage.VolumeGroup) {
			vols = append(vols, vol)
		} else if (queryPage.VolumeName == "") && (queryPage.VolumeState != "") && (queryPage.VolumeState == vol.Status.State) &&
			(queryPage.NameSpace == "") && (queryPage.VolumeGroup == "") {
			vols = append(vols, vol)
		} else if (queryPage.VolumeName == "") && (queryPage.VolumeState != "") && (queryPage.VolumeState == vol.Status.State) &&
			(queryPage.NameSpace != "") && (queryPage.NameSpace == vol.Spec.PersistentVolumeClaimNamespace) && (queryPage.VolumeGroup == "") {
			vols = append(vols, vol)
		} else if (queryPage.VolumeName == "") && (queryPage.VolumeState != "") && (queryPage.VolumeState == vol.Status.State) &&
			(queryPage.NameSpace == "") && (queryPage.VolumeGroup != "") && strings.Contains(vol.Spec.VolumeGroup, queryPage.VolumeGroup) {
			vols = append(vols, vol)
		} else if (queryPage.VolumeName == "") && (queryPage.VolumeState == "") &&
			(queryPage.NameSpace != "") && (queryPage.NameSpace == vol.Spec.PersistentVolumeClaimNamespace) && (queryPage.VolumeGroup == "") {
			vols = append(vols, vol)
		} else if (queryPage.VolumeName == "") && (queryPage.VolumeState == "") && (queryPage.NameSpace != "") &&
			(queryPage.NameSpace == vol.Spec.PersistentVolumeClaimNamespace) && (queryPage.VolumeGroup != "") &&
			strings.Contains(vol.Spec.VolumeGroup, queryPage.VolumeGroup) {
			vols = append(vols, vol)
		} else if (queryPage.VolumeName == "") && (queryPage.VolumeState == "") &&
			(queryPage.NameSpace == "") && (queryPage.VolumeGroup != "") && strings.Contains(vol.Spec.VolumeGroup, queryPage.VolumeGroup) {
			vols = append(vols, vol)
		} else if (queryPage.VolumeName != "") && strings.Contains(vol.Name, queryPage.VolumeName) && (queryPage.VolumeState != "") &&
			(queryPage.VolumeState == vol.Status.State) && (queryPage.NameSpace != "") &&
			(queryPage.NameSpace == vol.Spec.PersistentVolumeClaimNamespace) && (queryPage.VolumeGroup == "") {
			vols = append(vols, vol)
		} else if (queryPage.VolumeName != "") && strings.Contains(vol.Name, queryPage.VolumeName) && (queryPage.VolumeState != "") &&
			(queryPage.VolumeState == vol.Status.State) && (queryPage.NameSpace == "") &&
			(queryPage.VolumeGroup != "") && strings.Contains(vol.Spec.VolumeGroup, queryPage.VolumeGroup) {
			vols = append(vols, vol)
		} else if (queryPage.VolumeName != "") && strings.Contains(vol.Name, queryPage.VolumeName) && (queryPage.VolumeState == "") &&
			(queryPage.NameSpace != "") && (queryPage.NameSpace == vol.Spec.PersistentVolumeClaimNamespace) &&
			(queryPage.VolumeGroup != "") && strings.Contains(vol.Spec.VolumeGroup, queryPage.VolumeGroup) {
			vols = append(vols, vol)
		} else if (queryPage.VolumeName == "") && (queryPage.VolumeState != "") && (queryPage.VolumeState == vol.Status.State) &&
			(queryPage.NameSpace != "") && (queryPage.NameSpace == vol.Spec.PersistentVolumeClaimNamespace) &&
			(queryPage.VolumeGroup != "") && strings.Contains(vol.Spec.VolumeGroup, queryPage.VolumeGroup) {
			vols = append(vols, vol)
		} else if (queryPage.VolumeName != "") && strings.Contains(vol.Name, queryPage.VolumeName) &&
			(queryPage.VolumeState != "") && (queryPage.VolumeState == vol.Status.State) &&
			(queryPage.NameSpace != "") && (queryPage.NameSpace == vol.Spec.PersistentVolumeClaimNamespace) &&
			(queryPage.VolumeGroup != "") && strings.Contains(vol.Spec.VolumeGroup, queryPage.VolumeGroup) {
			vols = append(vols, vol)
		}
	}

	return vols, nil
}

func (lvController *LocalVolumeController) GetLocalVolume(lvname string) (*hwameistorapi.Volume, error) {
	var queryPage hwameistorapi.QueryPage
	queryPage.VolumeName = lvname

	lvs, err := lvController.listLocalVolume(queryPage)
	if err != nil {
		log.WithError(err).Error("Failed to listLocalVolume")
		return nil, err
	}

	for _, lv := range lvs {
		if lv.Name == lvname {
			return lv, nil
		}
	}
	return nil, nil
}

func (lvController *LocalVolumeController) getLocalVolumeReplicas(lvname string) ([]*apisv1alpha1.LocalVolumeReplica, error) {
	lv := &apisv1alpha1.LocalVolume{}
	if err := lvController.Client.Get(context.TODO(), client.ObjectKey{Name: lvname}, lv); err != nil {
		if !errors.IsNotFound(err) {
			log.WithError(err).Error("Failed to query LocalVolume")
		} else {
			log.Info("Not found the LocalVolume")
		}
		return nil, err
	}

	var lvrs []*apisv1alpha1.LocalVolumeReplica
	var replicaNames = lv.Status.Replicas
	for _, replicaName := range replicaNames {
		lvr := &apisv1alpha1.LocalVolumeReplica{}
		if err := lvController.Client.Get(context.TODO(), client.ObjectKey{Name: replicaName}, lvr); err != nil {
			if !errors.IsNotFound(err) {
				log.WithError(err).Error("Failed to query LocalVolumeReplica")
			} else {
				log.Info("Not found the LocalVolumeReplica")
			}
			return nil, err
		}
		lvrs = append(lvrs, lvr)
	}

	return lvrs, nil
}

func (lvController *LocalVolumeController) GetVolumeReplicas(queryPage hwameistorapi.QueryPage) (*hwameistorapi.VolumeReplicaList, error) {
	lvrs, err := lvController.getLocalVolumeReplicas(queryPage.VolumeName)
	if err != nil {
		log.WithError(err).Error("Failed to getLocalVolumeReplicas")
		return nil, err
	}

	var vrList = &hwameistorapi.VolumeReplicaList{}
	var vrs []*hwameistorapi.VolumeReplica
	for _, lvr := range lvrs {
		var vr = &hwameistorapi.VolumeReplica{}
		vr.LocalVolumeReplica = *lvr
		//vr.Name = lvr.Name
		//vr.NodeName = lvr.Spec.NodeName
		//vr.DevicePath = lvr.Status.DevicePath
		//vr.RequiredCapacityBytes = lvr.Spec.RequiredCapacityBytes
		//vr.StoragePath = lvr.Status.StoragePath
		//vr.Synced = lvr.Status.Synced
		//vr.State = hwameistorapi.StateConvert(lvr.Status.State)

		var convertedSynced bool
		if strings.Contains("true", queryPage.Synced) || strings.Contains("True", queryPage.Synced) {
			convertedSynced = true
		} else if strings.Contains("false", queryPage.Synced) || strings.Contains("False", queryPage.Synced) {
			convertedSynced = false
		}

		if queryPage.VolumeReplicaName == "" && queryPage.VolumeState == apisv1alpha1.VolumeStateEmpty && queryPage.Synced == "" {
			vrs = append(vrs, vr)
		} else if (queryPage.VolumeReplicaName != "" && strings.Contains(vr.Name, queryPage.VolumeReplicaName)) &&
			queryPage.VolumeState == apisv1alpha1.VolumeStateEmpty && queryPage.Synced == "" {
			vrs = append(vrs, vr)
		} else if (queryPage.VolumeState != "" && queryPage.VolumeState == vr.Status.State) && (queryPage.VolumeReplicaName == "") && (queryPage.Synced == "") {
			vrs = append(vrs, vr)
		} else if (queryPage.Synced != "" && convertedSynced == vr.Status.Synced) && (queryPage.VolumeReplicaName == "") && (queryPage.VolumeState == apisv1alpha1.VolumeStateEmpty) {
			vrs = append(vrs, vr)
		} else if (queryPage.Synced != "" && convertedSynced == vr.Status.Synced) && (queryPage.VolumeReplicaName != "" && strings.Contains(vr.Name, queryPage.VolumeReplicaName)) && (queryPage.VolumeState == apisv1alpha1.VolumeStateEmpty) {
			vrs = append(vrs, vr)
		} else if (queryPage.Synced != "" && convertedSynced == vr.Status.Synced) && (queryPage.VolumeState != "" && queryPage.VolumeState == vr.Status.State) && queryPage.VolumeReplicaName == "" {
			vrs = append(vrs, vr)
		} else if (queryPage.VolumeReplicaName != "" && strings.Contains(vr.Name, queryPage.VolumeReplicaName)) && (queryPage.VolumeState != "" && queryPage.VolumeState == vr.Status.State) && (queryPage.Synced == "") {
			vrs = append(vrs, vr)
		} else if (queryPage.VolumeReplicaName != "" && strings.Contains(vr.Name, queryPage.VolumeReplicaName)) &&
			(queryPage.VolumeState != "" && vr.Status.State == queryPage.VolumeState) &&
			(queryPage.Synced != "" && convertedSynced == vr.Status.Synced) {
			vrs = append(vrs, vr)
		}

	}
	vrList.VolumeReplicas = vrs
	vrList.VolumeName = queryPage.VolumeName

	return vrList, nil
}

func (lvController *LocalVolumeController) GetVolumeOperation(queryPage hwameistorapi.QueryPage) (*hwameistorapi.VolumeOperationByVolume, error) {
	var volumeOperation = &hwameistorapi.VolumeOperationByVolume{}
	var volumeMigrateOperations []*hwameistorapi.VolumeMigrateOperation
	lvmList := apisv1alpha1.LocalVolumeMigrateList{}
	if err := lvController.Client.List(context.Background(), &lvmList); err != nil {
		return nil, err
	}

	for _, item := range lvmList.Items {
		if item.Spec.VolumeName == queryPage.VolumeName {
			var volumeMigrateOperation = &hwameistorapi.VolumeMigrateOperation{}
			volumeMigrateOperation.LocalVolumeMigrate = item

			if queryPage.VolumeEventName == "" && queryPage.VolumeState == apisv1alpha1.VolumeStateEmpty {
				volumeMigrateOperations = append(volumeMigrateOperations, volumeMigrateOperation)
			} else if (queryPage.VolumeEventName != "" && strings.Contains(volumeMigrateOperation.Name, queryPage.VolumeEventName)) &&
				queryPage.VolumeState == apisv1alpha1.VolumeStateEmpty {
				volumeMigrateOperations = append(volumeMigrateOperations, volumeMigrateOperation)
			} else if (queryPage.VolumeState != "" && queryPage.VolumeState == volumeMigrateOperation.Status.State) && (queryPage.VolumeEventName == "") {
				volumeMigrateOperations = append(volumeMigrateOperations, volumeMigrateOperation)
			} else if (queryPage.VolumeEventName != "" && strings.Contains(volumeMigrateOperation.Name, queryPage.VolumeEventName)) && (queryPage.VolumeState == apisv1alpha1.VolumeStateEmpty) {
				volumeMigrateOperations = append(volumeMigrateOperations, volumeMigrateOperation)
			} else if (queryPage.VolumeEventName != "" && strings.Contains(volumeMigrateOperation.Name, queryPage.VolumeEventName)) && (queryPage.VolumeState != "" && queryPage.VolumeState == volumeMigrateOperation.Status.State) {
				volumeMigrateOperations = append(volumeMigrateOperations, volumeMigrateOperation)
			}
		}
	}

	var volumeConvertOperations []*hwameistorapi.VolumeConvertOperation
	lvcList := apisv1alpha1.LocalVolumeConvertList{}
	if err := lvController.Client.List(context.Background(), &lvcList, &client.ListOptions{}); err != nil {
		return nil, err
	}

	for _, item := range lvcList.Items {
		if item.Spec.VolumeName == queryPage.VolumeName {
			var volumeConvertOperation = &hwameistorapi.VolumeConvertOperation{}
			volumeConvertOperation.LocalVolumeConvert = item

			if queryPage.VolumeEventName == "" && queryPage.VolumeState == apisv1alpha1.VolumeStateEmpty {
				volumeConvertOperations = append(volumeConvertOperations, volumeConvertOperation)
			} else if (queryPage.VolumeEventName != "" && strings.Contains(volumeConvertOperation.Name, queryPage.VolumeEventName)) &&
				queryPage.VolumeState == apisv1alpha1.VolumeStateEmpty {
				volumeConvertOperations = append(volumeConvertOperations, volumeConvertOperation)
			} else if (queryPage.VolumeState != "" && queryPage.VolumeState == volumeConvertOperation.Status.State) && (queryPage.VolumeEventName == "") {
				volumeConvertOperations = append(volumeConvertOperations, volumeConvertOperation)
			} else if (queryPage.VolumeEventName != "" && strings.Contains(volumeConvertOperation.Name, queryPage.VolumeEventName)) && (queryPage.VolumeState == apisv1alpha1.VolumeStateEmpty) {
				volumeConvertOperations = append(volumeConvertOperations, volumeConvertOperation)
			} else if (queryPage.VolumeEventName != "" && strings.Contains(volumeConvertOperation.Name, queryPage.VolumeEventName)) && (queryPage.VolumeState != "" && queryPage.VolumeState == volumeConvertOperation.Status.State) {
				volumeConvertOperations = append(volumeConvertOperations, volumeConvertOperation)
			}
		}
	}

	var volumeExpandOperations []*hwameistorapi.VolumeExpandOperation
	lveList := apisv1alpha1.LocalVolumeExpandList{}
	if err := lvController.Client.List(context.Background(), &lveList); err != nil {
		return nil, err
	}

	for _, item := range lveList.Items {
		if item.Spec.VolumeName == queryPage.VolumeName {
			var volumeExpandOperation = &hwameistorapi.VolumeExpandOperation{}
			volumeExpandOperation.LocalVolumeExpand = item

			if queryPage.VolumeEventName == "" && queryPage.VolumeState == apisv1alpha1.VolumeStateEmpty {
				volumeExpandOperations = append(volumeExpandOperations, volumeExpandOperation)
			} else if (queryPage.VolumeEventName != "" && strings.Contains(volumeExpandOperation.Name, queryPage.VolumeEventName)) &&
				queryPage.VolumeState == apisv1alpha1.VolumeStateEmpty {
				volumeExpandOperations = append(volumeExpandOperations, volumeExpandOperation)
			} else if (queryPage.VolumeState != "" && queryPage.VolumeState == volumeExpandOperation.Status.State) && (queryPage.VolumeEventName == "") {
				volumeExpandOperations = append(volumeExpandOperations, volumeExpandOperation)
			} else if (queryPage.VolumeEventName != "" && strings.Contains(volumeExpandOperation.Name, queryPage.VolumeEventName)) && (queryPage.VolumeState == apisv1alpha1.VolumeStateEmpty) {
				volumeExpandOperations = append(volumeExpandOperations, volumeExpandOperation)
			} else if (queryPage.VolumeEventName != "" && strings.Contains(volumeExpandOperation.Name, queryPage.VolumeEventName)) && (queryPage.VolumeState != "" && queryPage.VolumeState == volumeExpandOperation.Status.State) {
				volumeExpandOperations = append(volumeExpandOperations, volumeExpandOperation)
			}
		}
	}

	volumeOperation.VolumeMigrateOperations = volumeMigrateOperations
	volumeOperation.VolumeConvertOperations = volumeConvertOperations
	volumeOperation.VolumeExpandOperations = volumeExpandOperations
	volumeOperation.VolumeName = queryPage.VolumeName
	return volumeOperation, nil
}

// CreateVolumeMigrate it creates a migrate crd or set a migrate task abort, if abort, we need volName only
func (lvController *LocalVolumeController) CreateVolumeMigrate(volName, srcNode, selectedNode string, abort bool) (*hwameistorapi.VolumeMigrateRspBody, error) {
	rsp := &hwameistorapi.VolumeMigrateRspBody{
		VolumeMigrateInfo: &hwameistorapi.VolumeMigrateInfo{
			VolumeName:   volName,
			SrcNode:      srcNode,
			SelectedNode: selectedNode,
		},
	}

	if abort {
		// Abort the migrate operation
		lvm, err := lvController.GetVolumeMigrate(volName)
		if err != nil {
			return nil, err
		}
		if lvm == nil {
			return nil, fmt.Errorf("LocalVolumeMigrate is not exists")
		}
		lvm.Spec.Abort = true
		return rsp, lvController.Client.Update(context.TODO(), lvm)
	}

	// Create the migrate operation crd
	lv := &apisv1alpha1.LocalVolume{}
	if err := lvController.Client.Get(context.Background(), types.NamespacedName{Name: volName}, lv); err != nil {
		return nil, err
	}
	if lv.Status.PublishedNodeName == srcNode {
		return nil, fmt.Errorf("LocalVolume is still in use by source node, try it later")
	}

	lvm := &apisv1alpha1.LocalVolumeMigrate{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("migrate-%s", volName),
		},
		Spec: apisv1alpha1.LocalVolumeMigrateSpec{
			VolumeName:           volName,
			SourceNode:           srcNode,
			MigrateAllVols:       true,
			TargetNodesSuggested: []string{},
		},
	}
	// Don't specify the target nodes, so the scheduler will select from the available nodes
	if selectedNode != "" {
		lvm.Spec.TargetNodesSuggested = append(lvm.Spec.TargetNodesSuggested, selectedNode)
	}
	return rsp, lvController.Client.Create(context.Background(), lvm)
}

func (lvController *LocalVolumeController) CreateVolumeConvert(volName string, abort bool) (*hwameistorapi.VolumeConvertRspBody, error) {
	rsp := &hwameistorapi.VolumeConvertRspBody{
		VolumeConvertInfo: &hwameistorapi.VolumeConvertInfo{
			VolumeName: volName,
			ReplicaNum: ConvertReplicaNum,
		},
	}

	if abort {
		// Abort the convert operation
		lvc, err := lvController.GetVolumeConvert(volName)
		if err != nil {
			return nil, err
		}
		if lvc == nil {
			return nil, fmt.Errorf("LocalVolumeConvert is not exists")
		}
		lvc.Spec.Abort = true
		return rsp, lvController.Client.Update(context.TODO(), lvc)
	}

	lv, err := lvController.GetLocalVolume(volName)
	if err != nil {
		return nil, err
	}
	if lv == nil {
		return nil, fmt.Errorf("volume %v is not exists", volName)
	}

	// Create the convert operation crd
	if lv.Spec.Convertible == false || lv.Spec.ReplicaNumber > 1 {
		return nil, fmt.Errorf("convertible is false or RplicaNumber is not 1")
	}

	lvc := &apisv1alpha1.LocalVolumeConvert{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("convert-%s", volName),
		},
		Spec: apisv1alpha1.LocalVolumeConvertSpec{
			VolumeName:    volName,
			ReplicaNumber: ConvertReplicaNum,
		},
	}
	return rsp, lvController.Client.Create(context.Background(), lvc)
}

func (lvController *LocalVolumeController) CreateVolumeExpand(volName string, targetCapacity string, abort bool) (*hwameistorapi.VolumeExpandRspBody, error) {
	rsp := &hwameistorapi.VolumeExpandRspBody{
		VolumeExpandInfo: &hwameistorapi.VolumeExpandInfo{
			VolumeName: volName,
		},
	}

	if abort {
		// Abort the expand operation
		lve, err := lvController.GetVolumeExpand(volName)
		if err != nil {
			return nil, err
		}
		if lve == nil {
			return nil, fmt.Errorf("LocalVolumeExpand is not exists")
		}
		lve.Spec.Abort = true
		return rsp, lvController.Client.Update(context.TODO(), lve)
	}

	// Get the LocalVolume
	lv, err := lvController.GetLocalVolume(volName)
	if err != nil {
		return nil, err
	}
	if lv == nil {
		return nil, fmt.Errorf("volume %v is not exists", volName)
	}
	currentCapacityBytes := lv.Spec.RequiredCapacityBytes

	//Determine whether there is enough space
	nodes := lv.Spec.Accessibility.Nodes
	var freeCount int64
	for _, name := range nodes {
		node := &apisv1alpha1.LocalStorageNode{}
		if err := lvController.Client.Get(context.Background(), types.NamespacedName{Name: name}, node); err != nil {
			return nil, err
		}
		if freeCount == 0 || node.Status.Pools[lv.Spec.PoolName].FreeCapacityBytes < freeCount {
			freeCount = node.Status.Pools[lv.Spec.PoolName].FreeCapacityBytes
		}
	}
	// Parse the targetCapacity
	quantity, err := resource.ParseQuantity(targetCapacity)
	if err != nil {
		return nil, err
	}
	targetCapacityBytes := quantity.Value()
	if targetCapacityBytes-currentCapacityBytes > freeCount {
		log.Errorf("Insufficient available space, freeCount is %d", freeCount)
		return nil, fmt.Errorf("Insufficient available space ")
	}

	// Get the pvc
	pvc := &corev1.PersistentVolumeClaim{}
	pvcKey := types.NamespacedName{
		Namespace: lv.Spec.PersistentVolumeClaimNamespace,
		Name:      lv.Spec.PersistentVolumeClaimName,
	}
	err = lvController.Client.Get(context.TODO(), pvcKey, pvc)
	if err != nil {
		return nil, err
	}
	rsp.VolumeExpandInfo.TargetCapacityBytes = quantity.Value()

	// Update the pvc's request capacity
	pvc.Spec.Resources.Requests["storage"] = quantity

	return rsp, lvController.Client.Update(context.TODO(), pvc)
}

func (lvController *LocalVolumeController) GetVolumeConvert(lvName string) (*apisv1alpha1.LocalVolumeConvert, error) {
	lvcList, err := lvController.ListVolumeConvert()
	if err != nil {
		return nil, err
	}

	for _, item := range lvcList.Items {
		if item.Spec.VolumeName == lvName {
			return &item, nil
		}
	}
	return nil, nil
}

func (lvController *LocalVolumeController) ListVolumeConvert() (*apisv1alpha1.LocalVolumeConvertList, error) {
	lvcList := &apisv1alpha1.LocalVolumeConvertList{}
	if err := lvController.Client.List(context.Background(), lvcList); err != nil {
		return nil, err
	}
	return lvcList, nil
}

func (lvController *LocalVolumeController) GetVolumeMigrate(lvName string) (*apisv1alpha1.LocalVolumeMigrate, error) {
	lvmList, err := lvController.ListVolumeMigrate()
	if err != nil {
		return nil, err
	}
	for _, item := range lvmList.Items {
		if item.Spec.VolumeName == lvName {
			return &item, nil
		}
	}
	return nil, nil
}

func (lvController *LocalVolumeController) ListVolumeMigrate() (*apisv1alpha1.LocalVolumeMigrateList, error) {
	lvmList := &apisv1alpha1.LocalVolumeMigrateList{}
	if err := lvController.Client.List(context.Background(), lvmList); err != nil {
		return nil, err
	}
	return lvmList, nil
}

func (lvController *LocalVolumeController) GetVolumeExpand(lvName string) (*apisv1alpha1.LocalVolumeExpand, error) {
	lveList, err := lvController.ListVolumeExpand()
	if err != nil {
		return nil, err
	}

	for _, item := range lveList.Items {
		if item.Spec.VolumeName == lvName {
			return &item, nil
		}
	}
	return nil, nil
}

func (lvController *LocalVolumeController) ListVolumeExpand() (*apisv1alpha1.LocalVolumeExpandList, error) {
	lveList := &apisv1alpha1.LocalVolumeExpandList{}
	if err := lvController.Client.List(context.Background(), lveList); err != nil {
		return nil, err
	}
	return lveList, nil
}
