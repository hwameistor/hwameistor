package hwameistor

import (
	"bytes"
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/types"
	"math"
	"strings"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	hwameistorapi "github.com/hwameistor/hwameistor/pkg/apiserver/api"
	utils "github.com/hwameistor/hwameistor/pkg/apiserver/util"
)

const (
	groupName   = "hwameistor.io"
	versionName = "v1"

	APIVersion                   = "v1alpha1"
	LocalVolumeMigrateKind       = "LocalVolumeMigrate"
	LocalVolumeConvertKind       = "LocalVolumeConvert"
	LocalVolumeMigrateAPIVersion = "hwameistor.io" + "/" + APIVersion
	LocalVolumeConvertAPIVersion = "hwameistor.io" + "/" + APIVersion
	ConvertReplicaNum            = 2
)

// LocalVolumeController
type LocalVolumeController struct {
	client.Client
	record.EventRecorder

	clientset *kubernetes.Clientset
}

// NewLocalVolumeController
func NewLocalVolumeController(client client.Client, clientset *kubernetes.Clientset, recorder record.EventRecorder) *LocalVolumeController {
	return &LocalVolumeController{
		Client:        client,
		EventRecorder: recorder,
		clientset:     clientset,
	}
}

// ListLocalVolume
func (lvController *LocalVolumeController) ListLocalVolume(queryPage hwameistorapi.QueryPage) (*hwameistorapi.VolumeList, error) {
	var volList = &hwameistorapi.VolumeList{}
	vols, err := lvController.listLocalVolume(queryPage)
	fmt.Println("ListLocalVolume vols = %v", vols)
	if err != nil {
		log.WithError(err).Error("Failed to listLocalVolume")
		return nil, err
	}

	volList.Volumes = utils.DataPatination(vols, queryPage.Page, queryPage.PageSize)

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

// listLocalVolume
func (lvController *LocalVolumeController) listLocalVolume(queryPage hwameistorapi.QueryPage) ([]*hwameistorapi.Volume, error) {
	lvList := &apisv1alpha1.LocalVolumeList{}
	if err := lvController.Client.List(context.TODO(), lvList); err != nil {
		log.WithError(err).Error("Failed to list LocalVolumes")
		return nil, err
	}
	fmt.Println("listLocalVolume queryPage = %v, queryPage.VolumeState = %v", queryPage, queryPage.VolumeState)

	var vols []*hwameistorapi.Volume
	for _, lv := range lvList.Items {
		var vol = &hwameistorapi.Volume{}
		vol.LocalVolume = lv

		if (queryPage.VolumeName == "") && (queryPage.VolumeState == apisv1alpha1.VolumeStateEmpty) && (queryPage.NameSpace == "") {
			vols = append(vols, vol)
		} else if (queryPage.VolumeName != "" && strings.Contains(vol.Name, queryPage.VolumeName)) &&
			(queryPage.VolumeState == apisv1alpha1.VolumeStateEmpty) && (queryPage.NameSpace == "") {
			vols = append(vols, vol)
		} else if (queryPage.VolumeState != "" && queryPage.VolumeState == vol.Status.State) &&
			(queryPage.VolumeName == "") && (queryPage.NameSpace == "") {
			vols = append(vols, vol)
		} else if (queryPage.NameSpace != "" && (queryPage.NameSpace == vol.Spec.PersistentVolumeClaimNamespace)) &&
			(queryPage.VolumeName == "") && (queryPage.VolumeState == apisv1alpha1.VolumeStateEmpty) {
			vols = append(vols, vol)
		} else if (queryPage.VolumeName != "" && strings.Contains(vol.Name, queryPage.VolumeName)) &&
			(queryPage.VolumeState != "" && queryPage.VolumeState == vol.Status.State) && (queryPage.NameSpace == "") {
			vols = append(vols, vol)
		} else if (queryPage.VolumeName != "" && strings.Contains(vol.Name, queryPage.VolumeName)) &&
			(queryPage.NameSpace != "" && (queryPage.NameSpace == vol.Spec.PersistentVolumeClaimNamespace)) && (queryPage.VolumeState == apisv1alpha1.VolumeStateEmpty) {
			vols = append(vols, vol)
		} else if (queryPage.VolumeState != "" && queryPage.VolumeState == vol.Status.State) &&
			(queryPage.VolumeName == "") && (queryPage.NameSpace != "" && (queryPage.NameSpace == vol.Spec.PersistentVolumeClaimNamespace)) {
			vols = append(vols, vol)
		} else if (queryPage.VolumeName != "" && strings.Contains(vol.Name, queryPage.VolumeName)) &&
			(queryPage.VolumeState != "" && queryPage.VolumeState == vol.Status.State) &&
			(queryPage.NameSpace != "" && (queryPage.NameSpace == vol.Spec.PersistentVolumeClaimNamespace)) {
			vols = append(vols, vol)
		}
	}

	return vols, nil
}

// GetLocalVolume
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

// getLocalVolumeReplicas
func (lvController *LocalVolumeController) getLocalVolumeReplicas(lvname string) ([]*apisv1alpha1.LocalVolumeReplica, error) {
	lv := &apisv1alpha1.LocalVolume{}
	if err := lvController.Client.Get(context.TODO(), client.ObjectKey{Name: lvname}, lv); err != nil {
		if !errors.IsNotFound(err) {
			log.WithError(err).Error("Failed to query diskume")
		} else {
			log.Info("Not found the diskume")
		}
		return nil, err
	}

	var lvrs []*apisv1alpha1.LocalVolumeReplica
	var replicaNames = lv.Status.Replicas
	for _, replicaname := range replicaNames {
		lvr := &apisv1alpha1.LocalVolumeReplica{}
		if err := lvController.Client.Get(context.TODO(), client.ObjectKey{Name: replicaname}, lvr); err != nil {
			if !errors.IsNotFound(err) {
				log.WithError(err).Error("Failed to query localvolumereplica")
			} else {
				log.Info("Not found the localvolumereplica")
			}
			return nil, err
		}
		lvrs = append(lvrs, lvr)
	}

	return lvrs, nil
}

// GetVolumeReplicas
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
		} else if (queryPage.Synced != "" && convertedSynced == vr.Status.Synced) && (queryPage.VolumeState != "") && (queryPage.VolumeReplicaName == "") {
			vrs = append(vrs, vr)
		} else if (queryPage.VolumeReplicaName != "" && strings.Contains(vr.Name, queryPage.VolumeReplicaName)) && (queryPage.VolumeState != "") && (queryPage.Synced == "") {
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

// GetVolumeOperation
func (lvController *LocalVolumeController) GetVolumeOperation(queryPage hwameistorapi.QueryPage) (*hwameistorapi.VolumeOperationByVolume, error) {

	var volumeOperation = &hwameistorapi.VolumeOperationByVolume{}
	var volumeMigrateOperations = []*hwameistorapi.VolumeMigrateOperation{}
	lvmList := apisv1alpha1.LocalVolumeMigrateList{}
	if err := lvController.Client.List(context.Background(), &lvmList, &client.ListOptions{}); err != nil {
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

	var volumeConvertOperations = []*hwameistorapi.VolumeConvertOperation{}
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

	var volumeExpandOperations = []*hwameistorapi.VolumeExpandOperation{}
	lveList := apisv1alpha1.LocalVolumeExpandList{}
	if err := lvController.Client.List(context.Background(), &lveList, &client.ListOptions{}); err != nil {
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

// GetLocalVolumeMigrateYamlStr
func (lvController *LocalVolumeController) GetLocalVolumeMigrateYamlStr(resourceName string) (*hwameistorapi.YamlData, error) {
	lvm := &apisv1alpha1.LocalVolumeMigrate{}
	if err := lvController.Client.Get(context.TODO(), client.ObjectKey{Name: resourceName}, lvm); err != nil {
		if !errors.IsNotFound(err) {
			log.WithError(err).Error("Failed to query localvolumemigrate")
		} else {
			log.Info("Not found the localvolumemigrate")
		}
		return nil, err
	}

	resourceYamlStr, err := lvController.getLVMResourceYaml(lvm)
	if err != nil {
		log.WithError(err).Error("Failed to getLVMResourceYaml")
		return nil, err
	}
	var yamlData = &hwameistorapi.YamlData{}
	yamlData.Data = resourceYamlStr

	return yamlData, nil
}

// GetLocalVolumeReplicaYamlStr
func (lvController *LocalVolumeController) GetLocalVolumeReplicaYamlStr(resourceName string) (*hwameistorapi.YamlData, error) {
	lvr := &apisv1alpha1.LocalVolumeReplica{}
	if err := lvController.Client.Get(context.TODO(), client.ObjectKey{Name: resourceName}, lvr); err != nil {
		if !errors.IsNotFound(err) {
			log.WithError(err).Error("Failed to query localvolumereplica")
		} else {
			log.Info("Not found the localvolumereplica")
		}
		return nil, err
	}

	resourceYamlStr, err := lvController.getLVRResourceYaml(lvr)
	if err != nil {
		log.WithError(err).Error("Failed to getLVRResourceYaml")
		return nil, err
	}
	var yamlData = &hwameistorapi.YamlData{}
	yamlData.Data = resourceYamlStr

	return yamlData, nil
}

// GetLocalVolumeYamlStr
func (lvController *LocalVolumeController) GetLocalVolumeYamlStr(resourceName string) (*hwameistorapi.YamlData, error) {
	lv := &apisv1alpha1.LocalVolume{}
	if err := lvController.Client.Get(context.TODO(), client.ObjectKey{Name: resourceName}, lv); err != nil {
		if !errors.IsNotFound(err) {
			log.WithError(err).Error("Failed to query localVolume")
		} else {
			log.Info("Not found the localVolume")
		}
		return nil, err
	}

	resourceYamlStr, err := lvController.getLVResourceYaml(lv)
	if err != nil {
		log.WithError(err).Error("Failed to getLVRResourceYaml")
		return nil, err
	}
	var yamlData = &hwameistorapi.YamlData{}
	yamlData.Data = resourceYamlStr

	return yamlData, nil
}

// getLVMResourceYaml
func (lvController *LocalVolumeController) getLVMResourceYaml(lvm *apisv1alpha1.LocalVolumeMigrate) (string, error) {

	buf := new(bytes.Buffer)

	lvm.GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{
		Group:   groupName,
		Version: versionName,
		Kind:    lvm.Kind,
	})
	y := printers.YAMLPrinter{}
	err := y.PrintObj(lvm, buf)
	if err != nil {
		panic(err)
	}

	return buf.String(), nil
}

// getLVRResourceYaml
func (lvController *LocalVolumeController) getLVRResourceYaml(lvr *apisv1alpha1.LocalVolumeReplica) (string, error) {

	buf := new(bytes.Buffer)

	lvr.GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{
		Group:   groupName,
		Version: versionName,
		Kind:    lvr.Kind,
	})
	y := printers.YAMLPrinter{}
	err := y.PrintObj(lvr, buf)
	if err != nil {
		panic(err)
	}

	return buf.String(), nil
}

// getLVResourceYaml
func (lvController *LocalVolumeController) getLVResourceYaml(lv *apisv1alpha1.LocalVolume) (string, error) {

	buf := new(bytes.Buffer)

	lv.GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{
		Group:   groupName,
		Version: versionName,
		Kind:    lv.Kind,
	})
	y := printers.YAMLPrinter{}
	err := y.PrintObj(lv, buf)
	if err != nil {
		panic(err)
	}

	return buf.String(), nil
}

// CreateVolumeMigrate
func (lvController *LocalVolumeController) CreateVolumeMigrate(volName, srcNode, selectedNode string, abort bool) (*hwameistorapi.VolumeMigrateRspBody, error) {

	lvmName := fmt.Sprintf("migrate-%s", volName)

	lvm := &apisv1alpha1.LocalVolumeMigrate{}
	if err := lvController.Client.Get(context.TODO(), client.ObjectKey{Name: lvmName}, lvm); err == nil {
		return nil, errors.NewBadRequest("LocalVolume has already exists lvm task, try it later")
	}

	lv := &apisv1alpha1.LocalVolume{}
	if err := lvController.Client.Get(context.Background(), types.NamespacedName{Name: volName}, lv); err != nil {
		log.WithField("localvolume", lv.Name).WithError(err).Error("Failed to submit localvolume")
		return nil, err
	}

	if lv.Status.PublishedNodeName != "" {
		return nil, errors.NewBadRequest("LocalVolume is still in use by source node, try it later")
	}

	lvm = &apisv1alpha1.LocalVolumeMigrate{
		ObjectMeta: metav1.ObjectMeta{
			Name: lvmName,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       LocalVolumeMigrateKind,
			APIVersion: LocalVolumeMigrateAPIVersion,
		},
		Spec: apisv1alpha1.LocalVolumeMigrateSpec{
			VolumeName: volName,
			SourceNode: srcNode,
			// don't specify the target nodes, so the scheduler will select from the avaliables
			TargetNodesSuggested: []string{selectedNode},
			MigrateAllVols:       true,
			Abort:                abort,
		},
	}
	if abort == false {
		if err := lvController.Client.Create(context.Background(), lvm); err != nil {
			log.WithField("migrate", lvm.Name).WithError(err).Error("Failed to submit a migrate job")
			return nil, err
		}
	} else {
		if err := lvController.Client.Delete(context.TODO(), lvm); err != nil {
			log.WithField("migrate", lvm.Name).WithError(err).Error("Failed to delete a migrate job")
			return nil, err
		}
	}

	var RspBody = &hwameistorapi.VolumeMigrateRspBody{}
	var vmi = &hwameistorapi.VolumeMigrateInfo{}
	vmi.VolumeName = volName
	vmi.SrcNode = srcNode
	vmi.SelectedNode = selectedNode

	RspBody.VolumeMigrateInfo = vmi

	return RspBody, nil
}

// CreateVolumeConvert
func (lvController *LocalVolumeController) CreateVolumeConvert(volName string, abort bool) (*hwameistorapi.VolumeConvertRspBody, error) {
	lvmName := fmt.Sprintf("convert-%s", volName)

	var RspBody = &hwameistorapi.VolumeConvertRspBody{}
	var vci = &hwameistorapi.VolumeConvertInfo{}
	vci.VolumeName = volName
	vci.ReplicaNum = ConvertReplicaNum
	RspBody.VolumeConvertInfo = vci

	lv, err := lvController.GetLocalVolume(volName)
	if err != nil {
		return RspBody, nil
	}
	if lv.Spec.Convertible == false || lv.Spec.ReplicaNumber > 1 {
		return RspBody, errors.NewBadRequest("Cannot create convert crd: check convertible is true and replicanumber == 1")
	}

	lvc := &apisv1alpha1.LocalVolumeConvert{
		ObjectMeta: metav1.ObjectMeta{
			Name: lvmName,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       LocalVolumeConvertKind,
			APIVersion: LocalVolumeConvertAPIVersion,
		},
		Spec: apisv1alpha1.LocalVolumeConvertSpec{
			VolumeName:    volName,
			ReplicaNumber: ConvertReplicaNum,
			Abort:         abort,
		},
	}

	if abort == false {
		if err := lvController.Client.Create(context.Background(), lvc); err != nil {
			log.WithField("convert", lvc.Name).WithError(err).Error("Failed to submit a convert job")
			if errors.IsAlreadyExists(err) {
				return RspBody, nil
			}
			return nil, err
		}
	} else {
		if err := lvController.Client.Delete(context.Background(), lvc); err != nil {
			log.WithField("convert", lvc.Name).WithError(err).Error("Failed to delete a convert job")
			return RspBody, err
		}
	}

	RspBody.VolumeConvertInfo = vci
	return RspBody, nil
}

// GetTargetNodesByManualTargetNodeType
func (lvController *LocalVolumeController) GetTargetNodesByManualTargetNodeType() ([]string, error) {

	lsnList := &apisv1alpha1.LocalStorageNodeList{}
	if err := lvController.Client.List(context.TODO(), lsnList); err != nil {
		log.WithError(err).Error("Failed to list LocalStorageNodes")
		return nil, err
	}

	var nodeNames []string
	for _, lsn := range lsnList.Items {
		if lsn.Status.State == apisv1alpha1.NodeStateReady {
			nodeNames = append(nodeNames, lsn.Name)
		}
	}

	return nodeNames, nil
}

// GetVolumeConvert
func (lvController *LocalVolumeController) GetVolumeConvert(lvname string) (*hwameistorapi.VolumeConvertOperation, error) {

	var vcp = &hwameistorapi.VolumeConvertOperation{}

	lvcList := apisv1alpha1.LocalVolumeConvertList{}
	if err := lvController.Client.List(context.Background(), &lvcList, &client.ListOptions{}); err != nil {
		return vcp, err
	}

	for _, item := range lvcList.Items {
		if item.Spec.VolumeName == lvname {
			vcp.LocalVolumeConvert = item
			break
		}
	}
	return vcp, nil
}

// GetVolumeMigrate
func (lvController *LocalVolumeController) GetVolumeMigrate(lvname string) (hwameistorapi.VolumeMigrateOperation, error) {

	var vcp hwameistorapi.VolumeMigrateOperation

	lvmList := apisv1alpha1.LocalVolumeMigrateList{}
	if err := lvController.Client.List(context.Background(), &lvmList, &client.ListOptions{}); err != nil {
		return vcp, err
	}

	for _, item := range lvmList.Items {
		if item.Spec.VolumeName == lvname {
			vcp.LocalVolumeMigrate = item
			break
		}
	}
	return vcp, nil
}
