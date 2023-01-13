package hwameistor

import (
	"context"
	"fmt"
	"math"
	"strings"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	hwameistorapi "github.com/hwameistor/hwameistor/pkg/apiserver/api"
	utils "github.com/hwameistor/hwameistor/pkg/apiserver/util"
)

// LocalStoragePoolController
type LocalStoragePoolController struct {
	client.Client
	record.EventRecorder

	clientset *kubernetes.Clientset
}

// NewLocalStoragePoolController
func NewLocalStoragePoolController(client client.Client, clientset *kubernetes.Clientset, recorder record.EventRecorder) *LocalStoragePoolController {
	return &LocalStoragePoolController{
		Client:        client,
		EventRecorder: recorder,
		clientset:     clientset,
	}
}

// StoragePoolList
func (lspController *LocalStoragePoolController) StoragePoolList(queryPage hwameistorapi.QueryPage) (*hwameistorapi.StoragePoolList, error) {

	var storagePoolList = &hwameistorapi.StoragePoolList{}
	sps, err := lspController.listLocalStoragePools(queryPage)
	if err != nil {
		log.WithError(err).Error("Failed to listLocalStoragePool")
		return nil, err
	}

	var storagePools = []*hwameistorapi.StoragePool{}
	storagePoolList.StoragePools = utils.DataPatination(sps, queryPage.Page, queryPage.PageSize)
	if len(sps) == 0 {
		storagePoolList.StoragePools = storagePools
	}

	var pagination = &hwameistorapi.Pagination{}
	pagination.Page = queryPage.Page
	pagination.PageSize = queryPage.PageSize
	pagination.Total = uint32(len(sps))
	pagination.Pages = int32(math.Ceil(float64(len(sps)) / float64(queryPage.PageSize)))

	storagePoolList.Page = pagination

	return storagePoolList, nil
}

// listLocalStoragePools
func (lspController *LocalStoragePoolController) listLocalStoragePools(queryPage hwameistorapi.QueryPage) ([]*hwameistorapi.StoragePool, error) {

	storagePoolNodesCollectionMap, err := lspController.makeStoragePoolNodesCollectionMap()
	if err != nil {
		log.WithError(err).Error("Failed to makeStoragePoolNodesCollectionMap")
		return nil, err
	}
	var sps []*hwameistorapi.StoragePool
	for poolName, poolNodeCollection := range storagePoolNodesCollectionMap {
		var sp = &hwameistorapi.StoragePool{}
		sp.PoolName = poolNodeCollection.StoragePool.PoolName
		sp.StorageNodePools = poolNodeCollection.StoragePool.StorageNodePools
		sp.CreateTime = poolNodeCollection.StoragePool.CreateTime
		sp.TotalCapacityBytes = poolNodeCollection.StoragePool.TotalCapacityBytes
		sp.AllocatedCapacityBytes = poolNodeCollection.StoragePool.AllocatedCapacityBytes
		sp.NodeNames = poolNodeCollection.ManagedNodeNames

		if queryPage.PoolName == "" || (queryPage.PoolName != "" && strings.Contains(poolName, queryPage.PoolName)) {
			sps = append(sps, sp)
		}
	}

	return sps, nil
}

// makeStoragePoolNodesCollectionMap
func (lspController *LocalStoragePoolController) makeStoragePoolNodesCollectionMap() (map[string]*hwameistorapi.StoragePoolNodesCollection, error) {

	lsnList := &apisv1alpha1.LocalStorageNodeList{}
	if err := lspController.Client.List(context.TODO(), lsnList); err != nil {
		log.WithError(err).Error("Failed to list LocalStorageNodes")
		return nil, err
	}

	var storagePoolNodesCollectionMap = make(map[string]*hwameistorapi.StoragePoolNodesCollection)
	for _, lsn := range lsnList.Items {
		for _, pool := range lsn.Status.Pools {
			if spnc, exists := storagePoolNodesCollectionMap[pool.Name]; exists {
				spnc.ManagedNodeNames = append(spnc.ManagedNodeNames, lsn.Name)
				var snp hwameistorapi.StorageNodePool
				snp.LocalPool = pool
				snp.NodeName = lsn.Name
				spnc.StoragePool.StorageNodePools = append(spnc.StoragePool.StorageNodePools, snp)
				spnc.StoragePool.TotalCapacityBytes += pool.TotalCapacityBytes
				spnc.StoragePool.AllocatedCapacityBytes += pool.UsedCapacityBytes
				spnc.StoragePool.CreateTime = lsn.CreationTimestamp.Time
				spnc.StoragePool.PoolName = pool.Name
				storagePoolNodesCollectionMap[pool.Name] = spnc
			} else {
				spncnew := &hwameistorapi.StoragePoolNodesCollection{}
				var snp hwameistorapi.StorageNodePool
				snp.LocalPool = pool
				snp.NodeName = lsn.Name
				spncnew.ManagedNodeNames = append(spncnew.ManagedNodeNames, lsn.Name)
				spncnew.StoragePool.StorageNodePools = append(spncnew.StoragePool.StorageNodePools, snp)
				spncnew.StoragePool.TotalCapacityBytes += pool.TotalCapacityBytes
				spncnew.StoragePool.AllocatedCapacityBytes += pool.UsedCapacityBytes
				spncnew.StoragePool.CreateTime = lsn.CreationTimestamp.Time
				spncnew.StoragePool.PoolName = pool.Name
				storagePoolNodesCollectionMap[pool.Name] = spncnew
			}
		}
	}

	return storagePoolNodesCollectionMap, nil
}

// GetStoragePool
func (lspController *LocalStoragePoolController) GetStoragePool(poolName string) (*hwameistorapi.StoragePool, error) {
	var queryPage hwameistorapi.QueryPage
	sps, err := lspController.listLocalStoragePools(queryPage)
	if err != nil {
		log.WithError(err).Error("Failed to listLocalStoragePools")
		return nil, err
	}

	for _, sp := range sps {
		if sp.PoolName == poolName {
			return sp, nil
		}
	}

	return nil, nil
}

// GetStorageNodeByPoolName
func (lspController *LocalStoragePoolController) GetStorageNodeByPoolName(queryPage hwameistorapi.QueryPage) (*hwameistorapi.StorageNodeListByPool, error) {

	snlist, err := lspController.getStorageNodeByPoolName(queryPage)
	if err != nil {
		log.WithError(err).Error("Failed to getStorageNodeByPoolName")
		return nil, err
	}
	var snlistByPool = &hwameistorapi.StorageNodeListByPool{}
	var sns = []*hwameistorapi.StorageNode{}

	snlistByPool.StorageNodes = utils.DataPatination(snlist, queryPage.Page, queryPage.PageSize)
	snlistByPool.StoragePoolName = queryPage.PoolName
	if len(snlist) == 0 {
		snlistByPool.StorageNodes = sns
	}

	var pagination = &hwameistorapi.Pagination{}
	pagination.Page = queryPage.Page
	pagination.PageSize = queryPage.PageSize
	pagination.Total = uint32(len(snlist))
	pagination.Pages = int32(math.Ceil(float64(len(snlist)) / float64(queryPage.PageSize)))
	snlistByPool.Page = pagination

	return snlistByPool, nil
}

// GetStorageNodeByPoolName
func (lspController *LocalStoragePoolController) getStorageNodeByPoolName(queryPage hwameistorapi.QueryPage) ([]*hwameistorapi.StorageNode, error) {
	storagePoolNodesCollectionMap, err := lspController.makeStoragePoolNodesCollectionMap()
	if err != nil {
		log.WithError(err).Error("Failed to makeStoragePoolNodesCollectionMap")
		return nil, err
	}

	var sns []*hwameistorapi.StorageNode
	lsnController := NewLocalStorageNodeController(lspController.Client, lspController.clientset, lspController.EventRecorder)
	if spnc, exists := storagePoolNodesCollectionMap[queryPage.PoolName]; exists {
		for _, nodeName := range spnc.ManagedNodeNames {
			sn, err := lsnController.GetStorageNode(nodeName)
			if err != nil {
				log.WithError(err).Error("Failed to GetStorageNode")
				return nil, err
			}
			fmt.Println("queryPage.NodeState = %v, sn.NodeState = %v", queryPage.NodeState, sn.K8sNodeState)
			if queryPage.NodeName == "" && queryPage.NodeState == hwameistorapi.NodeStateEmpty {
				sns = append(sns, sn)
			} else if (queryPage.NodeName != "" && strings.Contains(sn.LocalStorageNode.Name, queryPage.NodeName)) && (queryPage.NodeState == hwameistorapi.NodeStateEmpty) {
				sns = append(sns, sn)
			} else if (queryPage.NodeName == "") && (queryPage.NodeState != hwameistorapi.NodeStateUnknown && queryPage.NodeState == sn.K8sNodeState) {
				sns = append(sns, sn)
			} else if (queryPage.NodeName != "" && strings.Contains(sn.LocalStorageNode.Name, queryPage.NodeName)) && (queryPage.NodeState != hwameistorapi.NodeStateUnknown && queryPage.NodeState == sn.K8sNodeState) {
				sns = append(sns, sn)
			}
		}
	}

	return sns, nil
}

// StorageNodeDisksGetByPoolName
func (lspController *LocalStoragePoolController) StorageNodeDisksGetByPoolName(queryPage hwameistorapi.QueryPage) (*hwameistorapi.NodeDiskListByPool, error) {
	storagePoolNodesCollectionMap, err := lspController.makeStoragePoolNodesCollectionMap()
	if err != nil {
		log.WithError(err).Error("Failed to makeStoragePoolNodesCollectionMap")
		return nil, err
	}

	var nodeDiskListByPool = &hwameistorapi.NodeDiskListByPool{}
	var lds = []*hwameistorapi.LocalDiskInfo{}
	lsnController := NewLocalStorageNodeController(lspController.Client, lspController.clientset, lspController.EventRecorder)
	if spnc, exists := storagePoolNodesCollectionMap[queryPage.PoolName]; exists {
		for _, nn := range spnc.ManagedNodeNames {
			fmt.Println("StorageNodeDisksGetByPoolName queryPage.NodeName = %v, spnc.ManagedNodeNames = %v", queryPage.NodeName, spnc.ManagedNodeNames)
			if nn == queryPage.NodeName {
				tmplds, err := lsnController.ListStorageNodeDisks(queryPage)
				if err != nil {
					log.WithError(err).Error("Failed to ListStorageNodeDisks")
					return nil, err
				}
				fmt.Println("StorageNodeDisksGetByPoolName tmplds = %v", tmplds)
				for _, ld := range tmplds {
					if ld.LocalStoragePooLName == queryPage.PoolName {
						lds = append(lds, ld)
					}
				}
			}
		}
	}
	nodeDiskListByPool.PoolName = queryPage.PoolName
	nodeDiskListByPool.NodeName = queryPage.NodeName

	nodeDiskListByPool.LocalDisks = utils.DataPatination(lds, queryPage.Page, queryPage.PageSize)
	if len(lds) == 0 {
		nodeDiskListByPool.LocalDisks = lds
	}

	var pagination = &hwameistorapi.Pagination{}
	pagination.Page = queryPage.Page
	pagination.PageSize = queryPage.PageSize
	pagination.Total = uint32(len(lds))
	pagination.Pages = int32(math.Ceil(float64(len(lds)) / float64(queryPage.PageSize)))
	nodeDiskListByPool.Page = pagination

	return nodeDiskListByPool, nil
}

// listClaimedLocalDiskByNode
func (lspController *LocalStoragePoolController) listClaimedLocalDiskByNode(nodeName string) ([]apisv1alpha1.LocalDisk, error) {
	diskList := &apisv1alpha1.LocalDiskList{}
	if err := lspController.Client.List(context.TODO(), diskList); err != nil {
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

// LocalDiskListByNode
//func (lspController *LocalStoragePoolController) LocalDiskListByNode(nodeName string, page, pageSize int32) (*hwameistorapi.LocalDiskListByNode, error) {
//
//	var localDiskList = &hwameistorapi.LocalDiskListByNode{}
//
//	disks, err := lspController.ListStorageNodeDisks(nodeName)
//	if err != nil {
//		log.WithError(err).Error("Failed to ListStorageNodeDisks")
//		return nil, err
//	}
//
//	var pagination = &hwameistorapi.Pagination{}
//	pagination.Page = page
//	pagination.PageSize = pageSize
//	pagination.Total = uint32(len(disks))
//	pagination.Pages = int32(math.Ceil(float64(len(disks)) / float64(pageSize)))
//	localDiskList.Page = pagination
//
//	localDiskList.LocalDisksItemsList.LocalDisks = utils.DataPatination(disks, page, pageSize)
//	localDiskList.NodeName = nodeName
//
//	return localDiskList, nil
//}

// ListStorageNodeDisks
//func (lspController *LocalStoragePoolController) ListStorageNodeDisks(nodeName string) ([]*hwameistorapi.LocalDisk, error) {

//diskList := &apisv1alpha1.LocalDiskList{}
//if err := lspController.Client.List(context.TODO(), diskList); err != nil {
//	log.WithError(err).Error("Failed to list LocalDisks")
//	return nil, err
//}
//
//var disks []*hwameistorapi.LocalDisk
//for i := range diskList.Items {
//	if diskList.Items[i].Spec.NodeName == nodeName {
//		var disk = &hwameistorapi.LocalDisk{}
//		disk.DevPath = diskList.Items[i].Spec.DevicePath
//		disk.State = lspController.convertLocalDiskState(diskList.Items[i].Status.State)
//		if diskList.Items[i].Spec.DiskAttributes.Type == hwameistorapi.DiskClassNameHDD {
//			disk.LocalStoragePooLName = hwameistorapi.PoolNameForHDD
//		} else if diskList.Items[i].Spec.DiskAttributes.Type == hwameistorapi.DiskClassNameSSD {
//			disk.LocalStoragePooLName = hwameistorapi.PoolNameForSSD
//		}
//		disk.Class = diskList.Items[i].Spec.DiskAttributes.Type
//		disk.HasRAID = diskList.Items[i].Spec.HasRAID
//		disk.TotalCapacityBytes = diskList.Items[i].Spec.Capacity
//		availableCapacityBytes := lspController.getAvailableDiskCapacity(nodeName, diskList.Items[i].Spec.DevicePath, diskList.Items[i].Spec.DiskAttributes.Type)
//		disk.AvailableCapacityBytes = availableCapacityBytes
//		disks = append(disks, disk)
//	}
//}

//	return nil, nil
//}
