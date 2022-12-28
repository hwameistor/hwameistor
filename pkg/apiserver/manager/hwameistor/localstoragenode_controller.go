package hwameistor

import (
	"bytes"
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"math"
	"strings"

	hwameistorapi "github.com/hwameistor/hwameistor/pkg/apiserver/api"
	utils "github.com/hwameistor/hwameistor/pkg/apiserver/util"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/handler/localdisk"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/client-go/kubernetes"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// LocalStorageNodeController
type LocalStorageNodeController struct {
	client.Client
	record.EventRecorder

	clientset *kubernetes.Clientset
}

// NewLocalStorageNodeController
func NewLocalStorageNodeController(client client.Client, clientset *kubernetes.Clientset, recorder record.EventRecorder) *LocalStorageNodeController {
	return &LocalStorageNodeController{
		Client:        client,
		EventRecorder: recorder,
		clientset:     clientset,
	}
}

// GetLocalStorageNode
func (lsnController *LocalStorageNodeController) GetLocalStorageNode(key client.ObjectKey) (*apisv1alpha1.LocalStorageNode, error) {
	lsn := &apisv1alpha1.LocalStorageNode{}
	if err := lsnController.Client.Get(context.TODO(), key, lsn); err != nil {
		if !errors.IsNotFound(err) {
			log.WithError(err).Error("Failed to query lsn")
		} else {
			log.Printf("GetLocalStorageNode: not found lsn")
			log.WithError(err)
		}
		return nil, err
	}
	return lsn, nil
}

// StorageNodeList
func (lsnController *LocalStorageNodeController) StorageNodeList(queryPage hwameistorapi.QueryPage) (*hwameistorapi.StorageNodeList, error) {

	var storageNodeList = &hwameistorapi.StorageNodeList{}
	sns, err := lsnController.ListLocalStorageNode(queryPage)
	if err != nil {
		log.WithError(err).Error("Failed to list ListLocalStorageNode")
		return nil, err
	}

	var storagenodes = []*hwameistorapi.StorageNode{}

	storageNodeList.StorageNodes = utils.DataPatination(sns, queryPage.Page, queryPage.PageSize)
	if len(sns) == 0 {
		storageNodeList.StorageNodes = storagenodes
	}

	var pagination = &hwameistorapi.Pagination{}
	pagination.Page = queryPage.Page
	pagination.PageSize = queryPage.PageSize
	pagination.Total = uint32(len(sns))
	if len(sns) == 0 {
		pagination.Pages = 0
	} else {
		pagination.Pages = int32(math.Ceil(float64(len(sns)) / float64(queryPage.PageSize)))
	}

	storageNodeList.Page = pagination

	return storageNodeList, nil
}

// ListLocalStorageNode
func (lsnController *LocalStorageNodeController) ListLocalStorageNode(queryPage hwameistorapi.QueryPage) ([]*hwameistorapi.StorageNode, error) {

	lsnList := &apisv1alpha1.LocalStorageNodeList{}
	if err := lsnController.Client.List(context.TODO(), lsnList); err != nil {
		log.WithError(err).Error("Failed to list LocalStorageNodes")
		return nil, err
	}

	var sns []*hwameistorapi.StorageNode
	for i := range lsnList.Items {
		//claimedLocaldisks, err := lsnController.listClaimedLocalDiskByNode(lsnList.Items[i].Name)
		//if err != nil {
		//	log.WithError(err).Error("Failed to list listClaimedLocalDiskByNode")
		//	return nil, err
		//}
		//
		//localdisks, err := lsnController.ListStorageNodeDisks(queryPage)
		//if err != nil {
		//	log.WithError(err).Error("Failed to ListStorageNodeDisks")
		//	return nil, err
		//}
		sn := lsnController.convertStorageNode(lsnList.Items[i])
		sn.K8sNodeState = lsnController.getK8SNodeStatus(lsnList.Items[i].Name)

		fmt.Println("ListLocalStorageNode queryPage.Name = %v, queryPage.DriverState = %v, queryPage.NodeState = %v", queryPage.Name, queryPage.DriverState, queryPage.NodeState)
		if (queryPage.Name == "") && (queryPage.NodeState == hwameistorapi.NodeStateEmpty) && (queryPage.DriverState == "") {
			sns = append(sns, sn)
		} else if (queryPage.Name != "" && strings.Contains(sn.LocalStorageNode.Name, queryPage.Name)) && (queryPage.NodeState == hwameistorapi.NodeStateEmpty) && (queryPage.DriverState == "") {
			sns = append(sns, sn)
		} else if (queryPage.Name == "") && (queryPage.NodeState == hwameistorapi.NodeStateReadyAndNotReady && (sn.K8sNodeState == hwameistorapi.NodeStateReady || sn.K8sNodeState == hwameistorapi.NodeStateNotReady)) && (queryPage.DriverState == "") {
			sns = append(sns, sn)
		} else if (queryPage.Name == "") && (queryPage.NodeState != hwameistorapi.NodeStateUnknown && queryPage.DriverState == sn.LocalStorageNode.Status.State) && (queryPage.DriverState == "") {
			sns = append(sns, sn)
		} else if (queryPage.Name == "") && (queryPage.NodeState == hwameistorapi.NodeStateEmpty) && (queryPage.DriverState != "" && queryPage.DriverState == sn.LocalStorageNode.Status.State) {
			sns = append(sns, sn)
		} else if (queryPage.Name == "") && (queryPage.NodeState == hwameistorapi.NodeStateReadyAndNotReady && (sn.K8sNodeState == hwameistorapi.NodeStateReady || sn.K8sNodeState == hwameistorapi.NodeStateNotReady)) && (queryPage.DriverState != "" && queryPage.DriverState == sn.LocalStorageNode.Status.State) {
			sns = append(sns, sn)
		} else if (queryPage.Name == "") && (queryPage.NodeState != hwameistorapi.NodeStateUnknown && queryPage.DriverState == sn.LocalStorageNode.Status.State) && (queryPage.DriverState != "" && queryPage.DriverState == sn.LocalStorageNode.Status.State) {
			sns = append(sns, sn)
		} else if (queryPage.Name != "" && strings.Contains(sn.LocalStorageNode.Name, queryPage.Name)) && (queryPage.NodeState == hwameistorapi.NodeStateEmpty) && (queryPage.DriverState != "" && queryPage.DriverState == sn.LocalStorageNode.Status.State) {
			sns = append(sns, sn)
		} else if (queryPage.Name != "" && strings.Contains(sn.LocalStorageNode.Name, queryPage.Name)) && (queryPage.NodeState == hwameistorapi.NodeStateReadyAndNotReady && (sn.K8sNodeState == hwameistorapi.NodeStateReady || sn.K8sNodeState == hwameistorapi.NodeStateNotReady)) && (queryPage.DriverState == "") {
			sns = append(sns, sn)
		} else if (queryPage.Name != "" && strings.Contains(sn.LocalStorageNode.Name, queryPage.Name)) && (queryPage.NodeState != hwameistorapi.NodeStateUnknown && queryPage.NodeState == sn.K8sNodeState) && (queryPage.DriverState == "") {
			sns = append(sns, sn)
		} else if (queryPage.Name != "" && strings.Contains(sn.LocalStorageNode.Name, queryPage.Name)) && (queryPage.NodeState == hwameistorapi.NodeStateReadyAndNotReady && (sn.K8sNodeState == hwameistorapi.NodeStateReady || sn.K8sNodeState == hwameistorapi.NodeStateNotReady)) && (queryPage.DriverState != "" && queryPage.DriverState == sn.LocalStorageNode.Status.State) {
			sns = append(sns, sn)
		} else if (queryPage.Name != "" && strings.Contains(sn.LocalStorageNode.Name, queryPage.Name)) && (queryPage.NodeState != hwameistorapi.NodeStateUnknown && queryPage.NodeState == sn.K8sNodeState) && (queryPage.DriverState != "" && queryPage.DriverState == sn.LocalStorageNode.Status.State) {
			sns = append(sns, sn)
		}
	}

	return sns, nil
}

// getK8SNodeStatus
func (lsnController *LocalStorageNodeController) getK8SNodeStatus(nodeName string) hwameistorapi.State {
	// list K8S nodes
	nodes, err := lsnController.clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.WithError(err).Error("Failed to list k8s nodes")
		return hwameistorapi.NodeStateNotReady
	}
	for _, node := range nodes.Items {
		if node.Name == nodeName {
			return hwameistorapi.State(node.Status.Conditions[len(node.Status.Conditions)-1].Type)
		}
	}
	return ""
}

// convertStorageNode
func (lsnController *LocalStorageNodeController) convertStorageNode(lsn apisv1alpha1.LocalStorageNode) *hwameistorapi.StorageNode {
	sn := &hwameistorapi.StorageNode{}

	sn.LocalStorageNode = lsn

	//sn.NodeName = lsn.Name
	//sn.IP = lsn.Spec.StorageIP

	//if lsn.Status.State == apisv1alpha1.NodeStateReady {
	//	for _, pool := range lsn.Status.Pools {
	//		if pool.Class == hwameistorapi.DiskClassNameHDD {
	//			sn.TotalHDDCapacityBytes = pool.TotalCapacityBytes
	//			sn.AllocatedHDDCapacityBytes = pool.UsedCapacityBytes
	//			//sn.FreeCapacityBytes += pool.FreeCapacityBytes
	//		} else if pool.Class == hwameistorapi.DiskClassNameSSD {
	//			sn.TotalSSDCapacityBytes = pool.TotalCapacityBytes
	//			sn.AllocatedSSDCapacityBytes = pool.UsedCapacityBytes
	//			//sn.FreeCapacityBytes += pool.FreeCapacityBytes
	//		}
	//	}
	//}

	return sn
}

// GetStorageNode
func (lsnController *LocalStorageNodeController) GetStorageNode(nodeName string) (*hwameistorapi.StorageNode, error) {
	var queryPage hwameistorapi.QueryPage
	queryPage.NodeName = nodeName
	sns, err := lsnController.ListLocalStorageNode(queryPage)
	if err != nil {
		log.WithError(err).Error("Failed to ListLocalStorageNode")
		return nil, err
	}

	for _, sn := range sns {
		if sn.LocalStorageNode.Name == nodeName {
			return sn, nil
		}
	}

	return nil, nil
}

func (lsnController *LocalStorageNodeController) GetStorageNodeMigrate(queryPage hwameistorapi.QueryPage) (*hwameistorapi.VolumeOperationListByNode, error) {
	var volumeOperationListByNode = &hwameistorapi.VolumeOperationListByNode{}

	volumeMigrateOperations, err := lsnController.getStorageNodeMigrateOperations(queryPage.NodeName)
	if err != nil {
		log.WithError(err).Error("Failed to getStorageNodeMigrateOperations")
		return nil, err
	}

	var vmos = []*hwameistorapi.VolumeMigrateOperation{}
	volumeOperationListByNode.VolumeMigrateOperations = utils.DataPatination(volumeMigrateOperations, queryPage.Page, queryPage.PageSize)
	volumeOperationListByNode.NodeName = queryPage.NodeName

	var pagination = &hwameistorapi.Pagination{}
	pagination.Page = queryPage.Page
	pagination.PageSize = queryPage.PageSize
	if len(volumeMigrateOperations) == 0 {
		pagination.Pages = 0
		volumeOperationListByNode.VolumeMigrateOperations = vmos
	} else {
		pagination.Pages = int32(math.Ceil(float64(len(volumeMigrateOperations)) / float64(queryPage.PageSize)))
	}
	pagination.Total = uint32(len(volumeMigrateOperations))
	volumeOperationListByNode.Page = pagination

	return volumeOperationListByNode, nil
}

// GetStorageNodeMigrate
func (lsnController *LocalStorageNodeController) getStorageNodeMigrateOperations(nodeName string) ([]*hwameistorapi.VolumeMigrateOperation, error) {
	lvmList := apisv1alpha1.LocalVolumeMigrateList{}
	if err := lsnController.Client.List(context.Background(), &lvmList, &client.ListOptions{}); err != nil {
		return nil, err
	}

	fmt.Println("getStorageNodeMigrateOperations lvmList = %v", lvmList)
	var vmos []*hwameistorapi.VolumeMigrateOperation
	for _, lvm := range lvmList.Items {
		//if len(lvm.Spec.TargetNodesSuggested) != 0 {
		//if lvm.Spec.TargetNodesSuggested[0] == nodeName || lvm.Spec.SourceNode == nodeName {
		var vmo = &hwameistorapi.VolumeMigrateOperation{}
		vmo.LocalVolumeMigrate = lvm
		//vmo.Name = lvm.Name
		//vmo.SourceNode = lvm.Spec.SourceNode
		////vmo.TargetNode = lvm.Spec.TargetNodesSuggested[0]
		//vmo.VolumeName = lvm.Spec.VolumeName
		//vmo.StartTime = lvm.CreationTimestamp.Time
		//vmo.State = hwameistorapi.StateConvert(lvm.Status.State)
		vmos = append(vmos, vmo)
		//}
		//}
	}

	return vmos, nil
}

// listClaimedLocalDiskByNode
func (lsnController *LocalStorageNodeController) listClaimedLocalDiskByNode(nodeName string) ([]apisv1alpha1.LocalDisk, error) {
	diskList := &apisv1alpha1.LocalDiskList{}
	if err := lsnController.Client.List(context.TODO(), diskList); err != nil {
		log.WithError(err).Error("Failed to list LocalDisks")
		return nil, err
	}

	var claimedLocalDisks []apisv1alpha1.LocalDisk
	for i := range diskList.Items {
		if diskList.Items[i].Spec.NodeName == nodeName {
			if diskList.Items[i].Status.State == apisv1alpha1.LocalDiskBound {
				claimedLocalDisks = append(claimedLocalDisks, diskList.Items[i])
			}
		}
	}

	return claimedLocalDisks, nil
}

// getAvailableDiskCapacity
func (lsnController *LocalStorageNodeController) getAvailableDiskCapacity(nodeName, devPath, diskClass string) int64 {
	var availableDiskCapacity int64

	nodeKey := client.ObjectKey{
		Name: nodeName,
	}
	if lsn, err := lsnController.GetLocalStorageNode(nodeKey); err != nil {
		for _, pool := range lsn.Status.Pools {
			if pool.Class == diskClass {
				for _, disk := range pool.Disks {
					if disk.DevPath == devPath {
						availableDiskCapacity = disk.CapacityBytes
						break
					}
				}
			}
		}
	}

	return availableDiskCapacity
}

func (lsnController *LocalStorageNodeController) LocalDiskListByNode(queryPage hwameistorapi.QueryPage) (*hwameistorapi.LocalDiskListByNode, error) {

	var localDiskList = &hwameistorapi.LocalDiskListByNode{}
	var localdisks = []*hwameistorapi.LocalDiskInfo{}

	disks, err := lsnController.ListStorageNodeDisks(queryPage)
	if err != nil {
		log.WithError(err).Error("Failed to ListStorageNodeDisks")
		return nil, err
	}
	fmt.Println("LocalDiskListByNode disks = %v", disks)

	var pagination = &hwameistorapi.Pagination{}
	pagination.Page = queryPage.Page
	pagination.PageSize = queryPage.PageSize
	pagination.Total = uint32(len(disks))

	localDiskList.NodeName = queryPage.Name

	if len(disks) == 0 {
		localDiskList.LocalDisks = localdisks
		return localDiskList, nil
	} else {
		pagination.Pages = int32(math.Ceil(float64(len(disks)) / float64(queryPage.PageSize)))
	}

	localDiskList.Page = pagination
	localDiskList.LocalDisks = utils.DataPatination(disks, queryPage.Page, queryPage.PageSize)

	return localDiskList, nil
}

// ListStorageNodeDisks
func (lsnController *LocalStorageNodeController) GetcdLocalDiskNodes(nodeName string) (*apisv1alpha1.LocalDiskNode, error) {

	ldnList := &apisv1alpha1.LocalDiskNodeList{}
	if err := lsnController.Client.List(context.TODO(), ldnList); err != nil {
		log.WithError(err).Error("Failed to list LocalStorageNodes")
		return nil, err
	}

	var ldnVal = apisv1alpha1.LocalDiskNode{}
	for _, ldn := range ldnList.Items {
		if ldn.Name == nodeName {
			ldnVal = ldn
		}
	}

	return &ldnVal, nil
}

// ListStorageNodeDisks
func (lsnController *LocalStorageNodeController) ListStorageNodeDisks(queryPage hwameistorapi.QueryPage) ([]*hwameistorapi.LocalDiskInfo, error) {

	diskList := &apisv1alpha1.LocalDiskList{}
	if err := lsnController.Client.List(context.TODO(), diskList); err != nil {
		log.WithError(err).Error("Failed to list LocalDisks")
		return nil, err
	}

	var disks []*hwameistorapi.LocalDiskInfo
	for i := range diskList.Items {
		if diskList.Items[i].Spec.NodeName == queryPage.NodeName {
			var disk = &hwameistorapi.LocalDiskInfo{}
			disk.LocalDisk = diskList.Items[i]

			if diskList.Items[i].Spec.DiskAttributes.Type == hwameistorapi.DiskClassNameHDD {
				disk.LocalStoragePooLName = hwameistorapi.PoolNameForHDD
			} else if diskList.Items[i].Spec.DiskAttributes.Type == hwameistorapi.DiskClassNameSSD {
				disk.LocalStoragePooLName = hwameistorapi.PoolNameForSSD
			}

			disk.TotalCapacityBytes = diskList.Items[i].Spec.Capacity
			availableCapacityBytes := lsnController.getAvailableDiskCapacity(queryPage.NodeName, diskList.Items[i].Spec.DevicePath, diskList.Items[i].Spec.DiskAttributes.Type)
			disk.AvailableCapacityBytes = availableCapacityBytes

			fmt.Println("ListStorageNodeDisks queryPage.DiskState = %v", queryPage.DiskState)
			if queryPage.DiskState == "" || (queryPage.DiskState != "" && queryPage.DiskState == disk.Status.State) {
				disks = append(disks, disk)
			}
		}
	}

	return disks, nil
}

// convertLocalDiskState
func (lsnController *LocalStorageNodeController) convertLocalDiskState(state apisv1alpha1.LocalDiskState) hwameistorapi.State {
	switch state {
	case apisv1alpha1.LocalDiskBound:
		return hwameistorapi.LocalDiskBound

	case apisv1alpha1.LocalDiskPending:
		return hwameistorapi.LocalDiskPending

	case apisv1alpha1.LocalDiskAvailable:
		return hwameistorapi.LocalDiskAvailable

	case apisv1alpha1.LocalDiskEmpty:
		return hwameistorapi.LocalDiskEmpty

	}

	return hwameistorapi.LocalDiskUnknown
}

// convertDriverStatus
func (lsnController *LocalStorageNodeController) convertDriverStatus(state apisv1alpha1.State) hwameistorapi.State {

	switch state {
	case apisv1alpha1.NodeStateReady:
		return hwameistorapi.DriverStateReady

	case apisv1alpha1.NodeStateMaintain:
		return hwameistorapi.DriverStateMaintain

	case apisv1alpha1.NodeStateOffline:
		return hwameistorapi.DriverStateOffline
	}

	return ""
}

// GetLocalVolumeMigrateYamlStr
func (lsnController *LocalStorageNodeController) GetStorageNodeVolumeMigrateYamlStr(resourceName string) (*hwameistorapi.YamlData, error) {

	lvmList := apisv1alpha1.LocalVolumeMigrateList{}
	if err := lsnController.Client.List(context.Background(), &lvmList, &client.ListOptions{}); err != nil {
		return nil, err
	}
	fmt.Println("GetStorageNodeVolumeMigrateYamlStr lvmList = %v", lvmList)
	var resourceYamlStr string
	var err error
	for _, item := range lvmList.Items {
		if item.Name == resourceName {
			resourceYamlStr, err = lsnController.getResourceYaml(&item)
			fmt.Println("GetLocalVolumeMigrateYamlStr resourceYamlStr = %v", resourceYamlStr)

			if err != nil {
				log.WithError(err).Error("Failed to getResourceYaml")
				return nil, err
			}
		}
	}

	var yamlData = &hwameistorapi.YamlData{}
	yamlData.Data = resourceYamlStr

	return yamlData, nil
}

// getResourceYaml
func (lsnController *LocalStorageNodeController) getResourceYaml(res *apisv1alpha1.LocalVolumeMigrate) (string, error) {

	buf := new(bytes.Buffer)
	fmt.Println("getResourceYaml res.(type) = %v", res)

	res.GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{
		Group:   groupName,
		Version: versionName,
		Kind:    res.Kind,
	})
	y := printers.YAMLPrinter{}
	err := y.PrintObj(res, buf)
	if err != nil {
		panic(err)
	}

	return buf.String(), nil
}

// ReserveStorageNodeDisk
func (lsnController *LocalStorageNodeController) ReserveStorageNodeDisk(queryPage hwameistorapi.QueryPage, diskHandler *localdisk.Handler) (*hwameistorapi.DiskReservedRspBody, error) {

	var RspBody = &hwameistorapi.DiskReservedRspBody{}
	var diskReservedRsp hwameistorapi.DiskReservedRsp
	nodeName := queryPage.NodeName
	diskName := queryPage.DiskName
	RspBody.DiskReservedRsp = diskReservedRsp

	//diskShortName := strings.Split(diskName, "/dev/")[1]
	localDiskName := utils.ConvertNodeName(nodeName) + "-" + diskName

	ld, err := diskHandler.GetLocalDisk(client.ObjectKey{Name: localDiskName})
	if err != nil {
		log.Errorf("failed to get localDisk %s", err.Error())
		return RspBody, err
	}
	fmt.Println("ReserveStorageNodeDisk ld = %v", ld)
	diskHandler = diskHandler.For(ld)
	diskHandler.ReserveDisk()

	err = diskHandler.Update()
	if err != nil {
		return RspBody, err
	}

	diskReservedRsp.ReservedRsp = hwameistorapi.LocalDiskReserved
	diskReservedRsp.DiskName = diskName

	RspBody.DiskReservedRsp = diskReservedRsp

	return RspBody, nil
}

// ReleaseReserveStorageNodeDisk
func (lsnController *LocalStorageNodeController) ReleaseReserveStorageNodeDisk(queryPage hwameistorapi.QueryPage, diskHandler *localdisk.Handler) (*hwameistorapi.DiskRemoveReservedRspBody, error) {

	var RspBody = &hwameistorapi.DiskRemoveReservedRspBody{}
	var diskRemoveReservedRsp hwameistorapi.DiskRemoveReservedRsp
	nodeName := queryPage.NodeName
	diskName := queryPage.DiskName
	RspBody.DiskRemoveReservedRsp = diskRemoveReservedRsp

	//diskShortName := strings.Split(diskName, "/dev/")[1]
	localDiskName := utils.ConvertNodeName(nodeName) + "-" + diskName

	ld, err := diskHandler.GetLocalDisk(client.ObjectKey{Name: localDiskName})
	if err != nil {
		log.Errorf("failed to get localDisk %s", err.Error())
		return RspBody, err
	}
	ld.Spec.Reserved = false
	diskHandler = diskHandler.For(ld)

	err = diskHandler.Update()
	if err != nil {
		return RspBody, err
	}

	diskRemoveReservedRsp.RemoveReservedRsp = hwameistorapi.LocalDiskReleaseReserved
	diskRemoveReservedRsp.DiskName = diskName

	RspBody.DiskRemoveReservedRsp = diskRemoveReservedRsp
	return RspBody, nil
}

// GetStorageNodeDisk
func (lsnController *LocalStorageNodeController) GetStorageNodeDisk(page hwameistorapi.QueryPage, diskHandler *localdisk.Handler) (*hwameistorapi.LocalDiskInfo, error) {

	var ldi = &hwameistorapi.LocalDiskInfo{}
	nodeKey := client.ObjectKey{
		Name: page.NodeName,
	}
	if lsn, err := lsnController.GetLocalStorageNode(nodeKey); err == nil {
		fmt.Println("GetStorageNodeDisk lsn = %v", lsn)
		for _, pool := range lsn.Status.Pools {
			for _, disk := range pool.Disks {
				if strings.Contains(disk.DevPath, page.DiskName) {
					ldi.LocalStoragePooLName = pool.Name
					ldi.AvailableCapacityBytes = disk.CapacityBytes

					ldname := utils.ConvertNodeName(lsn.Spec.HostName) + "-" + page.DiskName

					localDisk, err := diskHandler.GetLocalDisk(client.ObjectKey{Name: ldname})
					if err != nil {
						log.Errorf("failed to get localDisk %s", err.Error())
						return ldi, err
					}
					ldi.LocalDisk = *localDisk
					ldi.TotalCapacityBytes = localDisk.Spec.Capacity
				}
			}
		}
	} else {
		return ldi, err
	}
	return ldi, nil
}

// StorageNodePoolsList
func (lsnController *LocalStorageNodeController) StorageNodePoolsList(queryPage hwameistorapi.QueryPage, handler *localdisk.Handler) (*hwameistorapi.StoragePoolList, error) {
	var spl = &hwameistorapi.StoragePoolList{}

	nodeKey := client.ObjectKey{
		Name: queryPage.NodeName,
	}
	fmt.Println("StorageNodePoolsList nodeKey = %v", nodeKey)
	if lsn, err := lsnController.GetLocalStorageNode(nodeKey); err == nil {
		fmt.Println("StorageNodePoolsList lsn.Status.Pools = %v", lsn.Status.Pools)
		for _, pool := range lsn.Status.Pools {
			var sp = &hwameistorapi.StoragePool{}
			sp.LocalPool = pool
			sp.NodeNames = append(sp.NodeNames, queryPage.NodeName)
			sp.AllocatedCapacityBytes = pool.UsedCapacityBytes
			sp.CreateTime = lsn.CreationTimestamp.Time
			spl.StoragePools = append(spl.StoragePools, sp)
		}
	} else {
		fmt.Println("StorageNodePoolsList err = %v", err)
	}

	fmt.Println("StorageNodePoolsList spl.StoragePools[0] = %v", len(spl.StoragePools))

	var pagination = &hwameistorapi.Pagination{}
	pagination.Page = queryPage.Page
	pagination.PageSize = queryPage.PageSize
	pagination.Total = uint32(len(spl.StoragePools))

	var storagePools []*hwameistorapi.StoragePool
	if len(spl.StoragePools) == 0 {
		spl.StoragePools = storagePools
		return spl, nil
	} else {
		pagination.Pages = int32(math.Ceil(float64(len(spl.StoragePools)) / float64(queryPage.PageSize)))
	}

	spl.Page = pagination
	spl.StoragePools = utils.DataPatination(spl.StoragePools, queryPage.Page, queryPage.PageSize)

	fmt.Println("StorageNodePoolsList spl = %v", spl)
	return spl, nil
}

// StorageNodePoolGet
func (lsnController *LocalStorageNodeController) StorageNodePoolGet(queryPage hwameistorapi.QueryPage, handler *localdisk.Handler) (*hwameistorapi.StoragePool, error) {
	var sp = &hwameistorapi.StoragePool{}

	nodeKey := client.ObjectKey{
		Name: queryPage.NodeName,
	}

	if lsn, err := lsnController.GetLocalStorageNode(nodeKey); err == nil {
		for _, pool := range lsn.Status.Pools {
			if pool.Name == queryPage.PoolName {
				sp.LocalPool = pool
				sp.NodeNames = append(sp.NodeNames, queryPage.NodeName)
				sp.AllocatedCapacityBytes = pool.UsedCapacityBytes
				sp.CreateTime = lsn.CreationTimestamp.Time
				break
			}
		}
	}

	return sp, nil
}

// StorageNodePoolDisksList
func (lsnController *LocalStorageNodeController) StorageNodePoolDisksList(page hwameistorapi.QueryPage, diskHandler *localdisk.Handler) (*hwameistorapi.LocalDisksItemsList, error) {
	var ldilist = &hwameistorapi.LocalDisksItemsList{}

	nodeKey := client.ObjectKey{
		Name: page.NodeName,
	}
	if lsn, err := lsnController.GetLocalStorageNode(nodeKey); err == nil {
		for _, pool := range lsn.Status.Pools {
			if pool.Name == page.PoolName {
				for _, disk := range pool.Disks {
					var ldi = &hwameistorapi.LocalDiskInfo{}
					ldi.LocalStoragePooLName = pool.Name
					ldi.AvailableCapacityBytes = disk.CapacityBytes

					diskName := strings.Split(disk.DevPath, "/dev/")[1]
					ldname := utils.ConvertNodeName(lsn.Spec.HostName) + "-" + diskName

					fmt.Println("StorageNodePoolDisksList ldname = %v", ldname)
					localDisk, err := diskHandler.GetLocalDisk(client.ObjectKey{Name: ldname})
					if err != nil {
						log.Errorf("failed to get localDisk %s", err.Error())
						return ldilist, err
					}
					ldi.LocalDisk = *localDisk
					ldi.TotalCapacityBytes = localDisk.Spec.Capacity
					ldilist.LocalDisks = append(ldilist.LocalDisks, ldi)
				}
			}
		}
	}

	return ldilist, nil
}

// StorageNodePoolDiskGet
func (lsnController *LocalStorageNodeController) StorageNodePoolDiskGet(page hwameistorapi.QueryPage, diskHandler *localdisk.Handler) (*hwameistorapi.LocalDiskInfo, error) {
	var ldi = &hwameistorapi.LocalDiskInfo{}

	nodeKey := client.ObjectKey{
		Name: page.NodeName,
	}

	if lsn, err := lsnController.GetLocalStorageNode(nodeKey); err == nil {
		for _, pool := range lsn.Status.Pools {
			if pool.Name == page.PoolName {
				for _, disk := range pool.Disks {
					localDisk, err := diskHandler.GetLocalDisk(client.ObjectKey{Name: page.DiskName})
					if err != nil {
						log.Errorf("failed to get localDisk %s", err.Error())
						return ldi, err
					}
					if localDisk.Spec.DevicePath == disk.DevPath {
						ldi.LocalStoragePooLName = pool.Name
						ldi.AvailableCapacityBytes = disk.CapacityBytes

						localDisk, err := diskHandler.GetLocalDisk(client.ObjectKey{Name: page.DiskName})
						if err != nil {
							log.Errorf("failed to get localDisk %s", err.Error())
							return ldi, err
						}
						ldi.LocalDisk = *localDisk
						ldi.TotalCapacityBytes = localDisk.Spec.Capacity
						break
					}
				}
			}
		}
	}

	return ldi, nil
}
