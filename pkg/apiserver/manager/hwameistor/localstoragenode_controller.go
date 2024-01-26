package hwameistor

import (
	"context"
	"errors"
	"fmt"
	"k8s.io/apimachinery/pkg/types"
	"math"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	k8sv1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	hwameistorapi "github.com/hwameistor/hwameistor/pkg/apiserver/api"
	utils "github.com/hwameistor/hwameistor/pkg/apiserver/util"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/handler/localdisk"
)

type LocalStorageNodeController struct {
	client.Client
	record.EventRecorder

	clientset *kubernetes.Clientset
	ldHandler *localdisk.Handler
}

func NewLocalStorageNodeController(client client.Client, clientset *kubernetes.Clientset, recorder record.EventRecorder) *LocalStorageNodeController {
	diskHandler := localdisk.NewLocalDiskHandler(client, recorder)
	return &LocalStorageNodeController{
		Client:        client,
		EventRecorder: recorder,
		clientset:     clientset,
		ldHandler:     diskHandler,
	}
}
func (lsnController *LocalStorageNodeController) SetLdHandler(handler *localdisk.Handler) {
	lsnController.ldHandler = handler
}
func (lsnController *LocalStorageNodeController) GetLocalStorageNode(key client.ObjectKey) (*apisv1alpha1.LocalStorageNode, error) {
	lsn := &apisv1alpha1.LocalStorageNode{}
	if err := lsnController.Client.Get(context.TODO(), key, lsn); err != nil {
		if !k8serrors.IsNotFound(err) {
			log.WithError(err).Error("Failed to query lsn")
		} else {
			log.Printf("GetLocalStorageNode: not found lsn")
			log.WithError(err)
		}
		return nil, err
	}
	return lsn, nil
}

func (lsnController *LocalStorageNodeController) StorageNodeList(queryPage hwameistorapi.QueryPage) (*hwameistorapi.StorageNodeList, error) {
	var storageNodeList = &hwameistorapi.StorageNodeList{}
	sns, err := lsnController.ListLocalStorageNode(queryPage)
	if err != nil {
		log.WithError(err).Error("Failed to list ListLocalStorageNode")
		return nil, err
	}

	var storageNodes []*hwameistorapi.StorageNode

	storageNodeList.StorageNodes = utils.DataPatination(sns, queryPage.Page, queryPage.PageSize)
	if len(sns) == 0 {
		storageNodeList.StorageNodes = storageNodes
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

func (lsnController *LocalStorageNodeController) ListLocalStorageNode(queryPage hwameistorapi.QueryPage) ([]*hwameistorapi.StorageNode, error) {
	lsnList := &apisv1alpha1.LocalStorageNodeList{}
	if err := lsnController.Client.List(context.TODO(), lsnList); err != nil {
		log.WithError(err).Error("Failed to list LocalStorageNodes")
		return nil, err
	}

	var sns []*hwameistorapi.StorageNode
	for i := range lsnList.Items {
		var sn = &hwameistorapi.StorageNode{}

		sn.LocalStorageNode = lsnList.Items[i]
		localDiskNode, err := lsnController.GetLocalDiskNode(lsnList.Items[i].Name)
		if err != nil {
			log.WithError(err).Error("Failed to get localDiskNode")
			return nil, err
		}

		if queryPage.PoolName != "" {
			for _, pool := range lsnList.Items[i].Status.Pools {
				if pool.Name == queryPage.PoolName {
					sn.TotalDisk = len(pool.Disks)
				}
			}
		}

		sn.LocalDiskNode = *localDiskNode
		k8sNode, K8sNodeState := lsnController.getK8SNode(lsnList.Items[i].Name)
		sn.K8sNode = k8sNode
		sn.K8sNodeState = K8sNodeState

		log.Infof("ListLocalStorageNode queryPage.Name = %v, queryPage.DriverState = %v, queryPage.NodeState = %v", queryPage.Name, queryPage.DriverState, queryPage.NodeState)

		// filter out node with mismatched names
		if !(queryPage.Name == "" || strings.Contains(sn.LocalStorageNode.Name, queryPage.Name)) {
			continue
		}
		// filter out node with mismatched driver state
		if !(queryPage.DriverState == "" || queryPage.DriverState == sn.LocalStorageNode.Status.State) {
			continue
		}
		// filter out node with mismatched node state
		switch queryPage.NodeState {
		case "":
			sns = append(sns, sn)
		case hwameistorapi.NodeStateReady, hwameistorapi.NodeStateNotReady:
			if queryPage.NodeState == sn.K8sNodeState {
				sns = append(sns, sn)
			}
		case hwameistorapi.NodeStateUnknown:
			if sn.K8sNodeState != hwameistorapi.NodeStateReady && sn.K8sNodeState != hwameistorapi.NodeStateNotReady {
				sns = append(sns, sn)
			}
		default:
			if queryPage.NodeState == sn.K8sNodeState {
				sns = append(sns, sn)
			}
		}
	}

	return sns, nil
}

func (lsnController *LocalStorageNodeController) getK8SNode(nodeName string) (*k8sv1.Node, hwameistorapi.State) {
	k8snode, err := lsnController.clientset.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
	if err != nil {
		return k8snode, hwameistorapi.NodeStateNotReady
	}

	if k8snode.Name == nodeName {
		return k8snode, hwameistorapi.State(k8snode.Status.Conditions[len(k8snode.Status.Conditions)-1].Type)
	}

	return k8snode, hwameistorapi.NodeStateNotReady
}

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

func (lsnController *LocalStorageNodeController) convertStorageNode(lsn apisv1alpha1.LocalStorageNode) *hwameistorapi.StorageNode {
	sn := &hwameistorapi.StorageNode{}

	sn.LocalStorageNode = lsn

	return sn
}

func (lsnController *LocalStorageNodeController) GetStorageNode(nodeName string) (*hwameistorapi.StorageNode, error) {

	var sn = &hwameistorapi.StorageNode{}
	lsn := &apisv1alpha1.LocalStorageNode{}
	objectKey := client.ObjectKey{Name: nodeName}
	lsn, err := lsnController.GetLocalStorageNode(objectKey)
	if err != nil {
		log.WithError(err).Error("failed to get localstoragenode,nodeName: %v", nodeName)
		return nil, err
	}
	sn.LocalStorageNode = *lsn
	k8sNode, K8sNodeState := lsnController.getK8SNode(lsn.Name)
	sn.K8sNode = k8sNode
	sn.K8sNodeState = K8sNodeState

	return sn, nil
}

func (lsnController *LocalStorageNodeController) GetStorageNodeByPool(nodeName, poolName string) (*hwameistorapi.StorageNode, error) {
	var queryPage hwameistorapi.QueryPage
	queryPage.NodeName = nodeName
	queryPage.PoolName = poolName
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

	log.Infof("GetStorageNodeMigrate queryPage = %v", queryPage)
	volumeMigrateOperations, err := lsnController.getStorageNodeMigrateOperations(queryPage)
	if err != nil {
		log.WithError(err).Error("Failed to getStorageNodeMigrateOperations")
		return nil, err
	}

	var vmos []*hwameistorapi.VolumeMigrateOperation
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

func (lsnController *LocalStorageNodeController) getStorageNodeMigrateOperations(queryPage hwameistorapi.QueryPage) ([]*hwameistorapi.VolumeMigrateOperation, error) {
	lvmList := apisv1alpha1.LocalVolumeMigrateList{}
	if err := lsnController.Client.List(context.Background(), &lvmList, &client.ListOptions{}); err != nil {
		return nil, err
	}

	log.Infof("getStorageNodeMigrateOperations lvmList = %v", lvmList)
	var vmos []*hwameistorapi.VolumeMigrateOperation
	for i := range lvmList.Items {
		lvm := lvmList.Items[i]
		if lvm.Spec.SourceNode == queryPage.NodeName || lvm.Status.TargetNode == queryPage.NodeName {
			var vmo = &hwameistorapi.VolumeMigrateOperation{}
			vmo.LocalVolumeMigrate = lvm
			if queryPage.OperationName == "" && queryPage.Name == "" && queryPage.NodeState == hwameistorapi.NodeStateEmpty {
				vmos = append(vmos, vmo)
			} else if (queryPage.OperationName != "" && queryPage.OperationName == lvm.Name) && queryPage.VolumeName == "" && (queryPage.OperationState == apisv1alpha1.VolumeStateEmpty || queryPage.OperationState == lvm.Status.State) {
				vmos = append(vmos, vmo)
			} else if (queryPage.OperationName != "" && queryPage.OperationName == lvm.Name) && (queryPage.VolumeName != "" && queryPage.VolumeName == lvm.Spec.VolumeName) && (queryPage.OperationState == apisv1alpha1.VolumeStateEmpty || queryPage.OperationState == lvm.Status.State) {
				vmos = append(vmos, vmo)
			} else if (queryPage.OperationName == "") && (queryPage.VolumeName != "" && queryPage.VolumeName == lvm.Spec.VolumeName) && (queryPage.OperationState == apisv1alpha1.VolumeStateEmpty || queryPage.OperationState == lvm.Status.State) {
				vmos = append(vmos, vmo)
			} else if (queryPage.OperationName == "") && (queryPage.VolumeName == "") && (queryPage.OperationState == apisv1alpha1.VolumeStateEmpty || queryPage.OperationState == lvm.Status.State) {
				vmos = append(vmos, vmo)
			}
		}
	}

	return vmos, nil
}

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

func (lsnController *LocalStorageNodeController) getAvailableDiskCapacity(nodeName, devPath, diskClass string) int64 {
	var availableDiskCapacity int64

	nodeKey := client.ObjectKey{
		Name: nodeName,
	}

	if lsn, err := lsnController.GetLocalStorageNode(nodeKey); err != nil {
		log.Errorf("GetLocalStorageNode err = %v", err)
	} else {
		for _, pool := range lsn.Status.Pools {
			if pool.Class == diskClass {
				for _, disk := range pool.Disks {
					if disk.DevPath == devPath {
						availableDiskCapacity = disk.CapacityBytes
						log.WithField("availableDiskCapacity", availableDiskCapacity).WithField("devPath", devPath).
							WithField("diskClass", diskClass).WithField("nodeName", nodeName).
							Info("availableDiskCapacity is found")
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
	var localDisks []*hwameistorapi.LocalDiskInfo

	disks, err := lsnController.ListStorageNodeDisks(queryPage)
	if err != nil {
		log.WithError(err).Error("Failed to ListStorageNodeDisks")
		return nil, err
	}
	log.Infof("LocalDiskListByNode disks = %v", disks)

	var pagination = &hwameistorapi.Pagination{}
	pagination.Page = queryPage.Page
	pagination.PageSize = queryPage.PageSize
	pagination.Total = uint32(len(disks))

	localDiskList.NodeName = queryPage.Name

	if len(disks) == 0 {
		localDiskList.LocalDisks = localDisks
		return localDiskList, nil
	} else {
		pagination.Pages = int32(math.Ceil(float64(len(disks)) / float64(queryPage.PageSize)))
	}

	localDiskList.Page = pagination
	localDiskList.LocalDisks = utils.DataPatination(disks, queryPage.Page, queryPage.PageSize)

	return localDiskList, nil
}

func (lsnController *LocalStorageNodeController) GetLocalDiskNode(nodeName string) (*apisv1alpha1.LocalDiskNode, error) {
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

func (lsnController *LocalStorageNodeController) ListStorageNodeDisks(queryPage hwameistorapi.QueryPage) ([]*hwameistorapi.LocalDiskInfo, error) {
	diskList, err := lsnController.ldHandler.ListNodeLocalDisk(queryPage.NodeName)
	if err != nil {
		log.WithError(err).Error("Failed to list LocalDisks")
		return nil, err
	}

	var disks []*hwameistorapi.LocalDiskInfo
	for i := range diskList.Items {
		var disk = &hwameistorapi.LocalDiskInfo{}
		disk.LocalDisk = diskList.Items[i]

		if diskList.Items[i].Spec.DiskAttributes.Type == hwameistorapi.DiskClassNameHDD {
			disk.LocalStoragePooLName = hwameistorapi.PoolNameForHDD
		} else if diskList.Items[i].Spec.DiskAttributes.Type == hwameistorapi.DiskClassNameSSD {
			disk.LocalStoragePooLName = hwameistorapi.PoolNameForSSD
		} else if diskList.Items[i].Spec.DiskAttributes.Type == hwameistorapi.DiskClassNameNVMe {
			disk.LocalStoragePooLName = hwameistorapi.PoolNameForNVMe
		}

		disk.TotalCapacityBytes = diskList.Items[i].Spec.Capacity
		availableCapacityBytes := lsnController.getAvailableDiskCapacity(queryPage.NodeName, diskList.Items[i].Spec.DevicePath, diskList.Items[i].Spec.DiskAttributes.Type)
		disk.AvailableCapacityBytes = availableCapacityBytes
		diskShortName := strings.Split(diskList.Items[i].Spec.DevicePath, hwameistorapi.DEV)[1]
		disk.DiskPathShort = diskShortName

		log.Infof("ListStorageNodeDisks queryPage.DiskState = %v", queryPage.DiskState)
		if queryPage.DiskState == "" || (queryPage.DiskState != "" && queryPage.DiskState == disk.Status.State) {
			disks = append(disks, disk)
		}

	}

	return disks, nil
}

func (lsnController *LocalStorageNodeController) ReserveStorageNodeDisk(queryPage hwameistorapi.QueryPage) (*hwameistorapi.DiskReservedRspBody, error) {
	var RspBody = &hwameistorapi.DiskReservedRspBody{}
	var diskReservedRsp hwameistorapi.DiskReservedRsp
	deviceShortPath := queryPage.DeviceShortPath
	RspBody.DiskReservedRsp = diskReservedRsp

	diskName := utils.ConvertNodeName(queryPage.NodeName) + "-" + deviceShortPath

	// query localdisk
	localDisks, err := lsnController.ldHandler.ListLocalDiskByNodeDevicePath(queryPage.NodeName, hwameistorapi.DEV+queryPage.DeviceShortPath)
	if err != nil {
		log.Errorf("failed to get localDisk %s", err.Error())
		return RspBody, err
	}
	ld := &localDisks[0]
	log.Infof("ReserveStorageNodeDisk ld = %v", ld)
	lsnController.ldHandler = lsnController.ldHandler.For(ld)
	lsnController.ldHandler.ReserveDisk()

	err = lsnController.ldHandler.Update()
	if err != nil {
		return RspBody, err
	}

	diskReservedRsp.ReservedRsp = hwameistorapi.LocalDiskReserved
	diskReservedRsp.DiskName = diskName

	RspBody.DiskReservedRsp = diskReservedRsp

	return RspBody, nil
}

func (lsnController *LocalStorageNodeController) RemoveReserveStorageNodeDisk(queryPage hwameistorapi.QueryPage) (*hwameistorapi.DiskRemoveReservedRspBody, error) {
	var RspBody = &hwameistorapi.DiskRemoveReservedRspBody{}
	var diskRemoveReservedRsp hwameistorapi.DiskRemoveReservedRsp

	deviceShortPath := queryPage.DeviceShortPath
	RspBody.DiskRemoveReservedRsp = diskRemoveReservedRsp

	diskName := utils.ConvertNodeName(queryPage.NodeName) + "-" + deviceShortPath

	//ld, err := diskHandler.GetLocalDisk(client.ObjectKey{Name: diskName})
	localDisks, err := lsnController.ldHandler.ListLocalDiskByNodeDevicePath(queryPage.NodeName, hwameistorapi.DEV+queryPage.DeviceShortPath)
	if err != nil {
		log.Errorf("failed to get localDisk %s", err.Error())
		return RspBody, err
	}
	ld := &localDisks[0]
	ld.Spec.Reserved = false
	lsnController.ldHandler = lsnController.ldHandler.For(ld)

	err = lsnController.ldHandler.Update()
	if err != nil {
		return RspBody, err
	}

	diskRemoveReservedRsp.RemoveReservedRsp = hwameistorapi.LocalDiskReleaseReserved
	diskRemoveReservedRsp.DiskName = diskName

	RspBody.DiskRemoveReservedRsp = diskRemoveReservedRsp
	return RspBody, nil
}

func (lsnController *LocalStorageNodeController) SetStorageNodeDiskOwner(queryPage hwameistorapi.QueryPage) (*hwameistorapi.DiskOwnerRspBody, error) {
	var RspBody = &hwameistorapi.DiskOwnerRspBody{}
	var diskOwnerRsp hwameistorapi.DiskOwnerRsp
	deviceShortPath := queryPage.DeviceShortPath
	RspBody.DiskOwnerRsp = diskOwnerRsp

	diskName := utils.ConvertNodeName(queryPage.NodeName) + "-" + deviceShortPath

	localDisks, err := lsnController.ldHandler.ListLocalDiskByNodeDevicePath(queryPage.NodeName, hwameistorapi.DEV+queryPage.DeviceShortPath)
	if err != nil {
		log.Errorf("failed to get localDisk %s", err.Error())
		return RspBody, err
	}
	if len(localDisks) == 0 {
		log.WithField("nodeName", queryPage.NodeName).WithField("devPath", hwameistorapi.DEV+queryPage.DeviceShortPath).Errorf("no localdisks found")
		return RspBody, fmt.Errorf("no localdisks found,nodeName:%v,devPath:%v", queryPage.NodeName, hwameistorapi.DEV+queryPage.DeviceShortPath)
	}
	ld := &localDisks[0]
	//Unable to operate the operating system disk
	if ld.Spec.Owner != "" {
		log.Errorf("Only unclaimed disks can modify the disk owner")
		return RspBody, err
	}

	log.Infof("SetStorageNodeDiskOwner ld = %v", ld)
	lsnController.ldHandler = lsnController.ldHandler.For(ld)
	lsnController.ldHandler.SetOwner(queryPage.Owner)

	err = lsnController.ldHandler.Update()
	if err != nil {
		return RspBody, err
	}

	diskOwnerRsp.Owner = queryPage.Owner
	diskOwnerRsp.DiskName = diskName

	RspBody.DiskOwnerRsp = diskOwnerRsp

	return RspBody, nil
}

func (lsnController *LocalStorageNodeController) GetStorageNodeDisk(page hwameistorapi.QueryPage) (*hwameistorapi.LocalDiskInfo, error) {
	var ldi = &hwameistorapi.LocalDiskInfo{}

	devicePath := hwameistorapi.DEV + page.DiskName
	localDisks, err := lsnController.ldHandler.ListLocalDiskByNodeDevicePath(page.NodeName, devicePath)
	if err != nil {
		log.Errorf("failed to get localDisk by path %s", err.Error())
		return ldi, err
	}
	if len(localDisks) == 0 {
		log.WithField("nodeName", page.NodeName).WithField("devPath", devicePath).Errorf("no localdisks found")
		return nil, fmt.Errorf("no localdisks found,nodeName:%v,devPath:%v", page.NodeName, devicePath)
	}
	ldi.LocalDisk = localDisks[0]
	ldi.TotalCapacityBytes = localDisks[0].Spec.Capacity
	diskShortName := strings.Split(localDisks[0].Spec.DevicePath, hwameistorapi.DEV)[1]
	ldi.DiskPathShort = diskShortName
	if localDisks[0].Spec.DiskAttributes.Type == hwameistorapi.DiskClassNameHDD {
		ldi.LocalStoragePooLName = hwameistorapi.PoolNameForHDD
	} else if localDisks[0].Spec.DiskAttributes.Type == hwameistorapi.DiskClassNameSSD {
		ldi.LocalStoragePooLName = hwameistorapi.PoolNameForSSD
	} else if localDisks[0].Spec.DiskAttributes.Type == hwameistorapi.DiskClassNameNVMe {
		ldi.LocalStoragePooLName = hwameistorapi.PoolNameForNVMe
	}
	ldi.AvailableCapacityBytes = lsnController.getAvailableDiskCapacity(page.NodeName, devicePath, ldi.Spec.DiskAttributes.Type)

	return ldi, nil
}

func (lsnController *LocalStorageNodeController) StorageNodePoolsList(queryPage hwameistorapi.QueryPage) (*hwameistorapi.StoragePoolList, error) {
	spl, err := lsnController.getStorageNodePoolList(queryPage.NodeName)
	if err != nil {
		return nil, err
	}

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

	log.Infof("StorageNodePoolsList spl = %v", spl)
	return spl, nil
}

func (lsnController *LocalStorageNodeController) StorageNodePoolGet(queryPage hwameistorapi.QueryPage) (*hwameistorapi.StoragePool, error) {
	var sp = &hwameistorapi.StoragePool{}

	nodeKey := client.ObjectKey{
		Name: queryPage.NodeName,
	}

	if lsn, err := lsnController.GetLocalStorageNode(nodeKey); err == nil {
		for _, pool := range lsn.Status.Pools {
			if pool.Name == queryPage.PoolName {
				var snp hwameistorapi.StorageNodePool
				snp.LocalPool = pool
				snp.NodeName = queryPage.NodeName
				sp.PoolName = pool.Name
				sp.StorageNodePools = append(sp.StorageNodePools, snp)
				sp.NodeNames = append(sp.NodeNames, queryPage.NodeName)
				sp.AllocatedCapacityBytes = pool.UsedCapacityBytes
				sp.TotalCapacityBytes = pool.TotalCapacityBytes
				sp.CreateTime = lsn.CreationTimestamp.Time
				break
			}
		}
	}

	return sp, nil
}

func (lsnController *LocalStorageNodeController) StorageNodePoolDisksList(page hwameistorapi.QueryPage) (*hwameistorapi.LocalDisksItemsList, error) {
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
					// get localdisk which is specified node and devpath
					localDisks, err := lsnController.ldHandler.ListLocalDiskByNodeDevicePath(page.NodeName, disk.DevPath)
					if err != nil {
						log.Errorf("failed to get localDisk %s", err.Error())
						return ldilist, err
					}
					if len(localDisks) == 0 {
						log.Errorf("failed to get localDisk,nodeName: %v,devPath: %v", page.NodeName, disk.DevPath)
						return ldilist, err
					}
					localDisk := &localDisks[0]
					log.Infof("StorageNodePoolDisksList ldname = %v", localDisk.Name)

					ldi.LocalDisk = *localDisk
					ldi.TotalCapacityBytes = localDisk.Spec.Capacity
					ldilist.LocalDisks = append(ldilist.LocalDisks, ldi)
				}
			}
		}
	}

	return ldilist, nil
}

func (lsnController *LocalStorageNodeController) StorageNodePoolDiskGet(page hwameistorapi.QueryPage) (*hwameistorapi.LocalDiskInfo, error) {
	var ldi = &hwameistorapi.LocalDiskInfo{}

	nodeKey := client.ObjectKey{
		Name: page.NodeName,
	}
	// get localdisk information
	devicePath := hwameistorapi.DEV + page.DiskName
	localDisks, e := lsnController.ldHandler.ListLocalDiskByNodeDevicePath(page.NodeName, devicePath)
	if e != nil {
		log.WithField("nodeName", page.NodeName).WithField("devPath", devicePath).WithError(e).Errorf("Failed to get localdisk")
		return nil, e
	}
	if len(localDisks) == 0 {
		log.WithField("nodeName", page.NodeName).WithField("devPath", devicePath).WithError(e).Errorf("no localdisks found")
		return nil, fmt.Errorf("no localdisks found,nodeName:%v,devPath:%v", page.NodeName, devicePath)
	}
	localDisk := &localDisks[0]
	// get the localstorage disk
	if lsn, err := lsnController.GetLocalStorageNode(nodeKey); err == nil {
		for _, pool := range lsn.Status.Pools {
			if pool.Name == page.PoolName {
				for _, disk := range pool.Disks {
					if localDisk.Spec.DevicePath == disk.DevPath {
						ldi.LocalStoragePooLName = pool.Name
						ldi.AvailableCapacityBytes = disk.CapacityBytes
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

func (lsnController *LocalStorageNodeController) getStorageNodePoolList(nodeName string) (*hwameistorapi.StoragePoolList, error) {
	var spl = &hwameistorapi.StoragePoolList{}

	nodeKey := client.ObjectKey{
		Name: nodeName,
	}
	log.Infof("StorageNodePoolsList nodeKey = %v", nodeKey)
	if lsn, err := lsnController.GetLocalStorageNode(nodeKey); err == nil {
		for _, pool := range lsn.Status.Pools {
			var sp = &hwameistorapi.StoragePool{}
			var snp hwameistorapi.StorageNodePool
			snp.LocalPool = pool
			snp.NodeName = nodeName
			sp.StorageNodePools = append(sp.StorageNodePools, snp)
			sp.PoolName = pool.Name
			sp.NodeNames = append(sp.NodeNames, nodeName)
			sp.AllocatedCapacityBytes = pool.UsedCapacityBytes
			sp.TotalCapacityBytes = pool.TotalCapacityBytes
			sp.CreateTime = lsn.CreationTimestamp.Time
			spl.StoragePools = append(spl.StoragePools, sp)
		}
	} else {
		log.Infof("StorageNodePoolsList err = %v", err)
		return nil, err
	}

	return spl, nil
}

func (lsnController *LocalStorageNodeController) UpdateLocalStorageNode(nodeName string, enable bool) error {
	// check localDiskNode exists
	lsn, err := lsnController.GetLocalStorageNode(types.NamespacedName{Name: nodeName})
	if err != nil {
		return err
	}

	// get the node resource
	node := &corev1.Node{}
	if err = lsnController.Client.Get(context.TODO(), types.NamespacedName{Name: nodeName}, node); err != nil {
		return err
	}

	return lsnController.updateStorageNodeEnable(lsn, node, enable)
}

func (lsnController *LocalStorageNodeController) updateStorageNodeEnable(lsn *apisv1alpha1.LocalStorageNode, node *corev1.Node, enable bool) error {
	if node.Labels == nil {
		node.Labels = map[string]string{}
	}

	_, exists := node.Labels["lvm.hwameistor.io/enable"]
	// set default value
	if !exists {
		node.Labels["lvm.hwameistor.io/enable"] = "true"
	}

	if node.Labels["lvm.hwameistor.io/enable"] == strconv.FormatBool(enable) {
		// if its equal, return ok
		return nil
	}

	if enable {
		// enable Node
		node.Labels["lvm.hwameistor.io/enable"] = "true"
	} else {
		// disable Node, check the pools are not used
		for _, pool := range lsn.Status.Pools {
			if pool.UsedVolumeCount != 0 {
				return errors.New("LocalVolumes must be empty")
			}
		}
		node.Labels["lvm.hwameistor.io/enable"] = "false"
	}

	return lsnController.Update(context.TODO(), node)
}
