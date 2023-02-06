package hwameistor

import (
	"context"
	"k8s.io/apimachinery/pkg/api/errors"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	hoapisv1alpha1 "github.com/hwameistor/hwameistor-operator/api/v1alpha1"
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	hwameistorapi "github.com/hwameistor/hwameistor/pkg/apiserver/api"
	utils "github.com/hwameistor/hwameistor/pkg/apiserver/util"
)

const (
	hwameistorPrefix    = "hwameistor"
	nodeStorageSortNum  = 5
	OperatorClusterName = "cluster-sample"
)

// MetricController
type MetricController struct {
	client.Client
	record.EventRecorder

	nameSpace string

	clientset *kubernetes.Clientset

	storageCapacityCollection *hwameistorapi.StorageCapacityCollection

	storageNodesCollection *hwameistorapi.StorageNodesCollection

	volumeCollection *hwameistorapi.VolumeCollection

	diskCollection *hwameistorapi.DiskCollection

	moduleStatusCollection *hwameistorapi.ModuleStatusCollection

	storagePoolUseCollection *hwameistorapi.StoragePoolUseCollection

	nodeStorageUseCollection *hwameistorapi.NodeStorageUseCollection

	lock sync.Mutex
}

// NewMetricController
func NewMetricController(client client.Client, clientset *kubernetes.Clientset, recorder record.EventRecorder) *MetricController {
	return &MetricController{
		Client:                    client,
		EventRecorder:             recorder,
		clientset:                 clientset,
		nameSpace:                 utils.GetNamespace(),
		storageNodesCollection:    &hwameistorapi.StorageNodesCollection{},
		storageCapacityCollection: &hwameistorapi.StorageCapacityCollection{},
		volumeCollection:          &hwameistorapi.VolumeCollection{},
		diskCollection:            &hwameistorapi.DiskCollection{},
		moduleStatusCollection:    &hwameistorapi.ModuleStatusCollection{map[string]hwameistorapi.State{}},
		storagePoolUseCollection:  &hwameistorapi.StoragePoolUseCollection{map[string]hwameistorapi.StoragePoolUse{}},
		nodeStorageUseCollection:  &hwameistorapi.NodeStorageUseCollection{},
	}
}

// GetBaseMetric
func (mController *MetricController) GetBaseMetric() (*hwameistorapi.BaseMetric, error) {

	if err := mController.getBaseCapacityMetric(); err != nil {
		log.WithError(err).Error("Failed to getBaseCapacityMetric")
		return nil, err
	}
	if err := mController.getBaseVolumeMetric(); err != nil {
		log.WithError(err).Error("Failed to getBaseVolumeMetric")
		return nil, err
	}
	if err := mController.getBaseDiskMetric(); err != nil {
		log.WithError(err).Error("Failed to getBaseDiskMetric")
		return nil, err
	}
	if err := mController.getBaseNodeMetric(); err != nil {
		log.WithError(err).Error("Failed to getBaseNodeMetric")
		return nil, err
	}

	basemetric := mController.convertBaseMetric()

	return basemetric, nil
}

// GetModuleStatus
func (mController *MetricController) GetModuleStatus() (*hwameistorapi.ModuleStatus, error) {

	if err := mController.getHwameistorDaemonsetStatusMetric(); err != nil {
		log.WithError(err).Error("Failed to getHwameistorDaemonsetStatusMetric")
		return nil, err
	}

	if err := mController.getHwameistorDeploymentStatusMetric(); err != nil {
		log.WithError(err).Error("Failed to getHwameistorDeploymentStatusMetric")
		return nil, err
	}

	moduleStatus := mController.convertModuleStatus()

	operatorModuleStatus, err := mController.getHwameistorOperatorStatusMetric()
	if err != nil {
		log.WithError(err).Error("Failed to getHwameistorOperatorStatusMetric")
		return moduleStatus, err
	}

	return operatorModuleStatus, nil
}

// GetStoragePoolUseMetric
func (mController *MetricController) GetStoragePoolUseMetric() (*hwameistorapi.StoragePoolUseMetric, error) {

	if err := mController.addStoragePoolUseMetric(); err != nil {
		log.WithError(err).Error("Failed to addStoragePoolUseMetric")
		return nil, err
	}
	storagePoolUseMetric := mController.convertStoragePoolUseMetric()

	return storagePoolUseMetric, nil
}

// GetNodeStorageUseMetric
func (mController *MetricController) GetNodeStorageUseMetric(storagepoolclass string) (*hwameistorapi.NodeStorageUseMetric, error) {

	if err := mController.addNodeStorageUseMetric(storagepoolclass); err != nil {
		log.WithError(err).Error("Failed to addNodeStorageUseMetric")
		return nil, err
	}
	nodeStorageUseMetric := mController.convertNodeStorageUseMetric(storagepoolclass)

	return nodeStorageUseMetric, nil
}

// OperationListMetric
func (mController *MetricController) OperationListMetric(page, pageSize int32) (*hwameistorapi.OperationMetric, error) {

	var operationMetric = &hwameistorapi.OperationMetric{}
	var operationList []hwameistorapi.Operation
	lvmList := apisv1alpha1.LocalVolumeMigrateList{}
	if err := mController.Client.List(context.Background(), &lvmList, &client.ListOptions{}); err != nil {
		return nil, err
	}

	for _, item := range lvmList.Items {
		var operation hwameistorapi.Operation
		operation.EventName = item.Name
		operation.EventType = item.Kind
		operation.LocalVolumeName = item.Spec.VolumeName
		operation.Status = hwameistorapi.StateConvert(item.Status.State)
		operation.StartTime = item.CreationTimestamp.Time
		operation.Description = item.Status.Message
		operationList = append(operationList, operation)
	}

	lvcList := apisv1alpha1.LocalVolumeConvertList{}
	if err := mController.Client.List(context.Background(), &lvcList, &client.ListOptions{}); err != nil {
		return nil, err
	}

	for _, item := range lvcList.Items {
		var operation hwameistorapi.Operation
		operation.EventName = item.Name
		operation.EventType = item.Kind
		operation.Status = hwameistorapi.StateConvert(item.Status.State)
		operation.StartTime = item.CreationTimestamp.Time
		operation.EndTime = time.Now()
		operation.Description = item.Status.Message
		operationList = append(operationList, operation)
	}

	lveList := apisv1alpha1.LocalVolumeExpandList{}
	if err := mController.Client.List(context.Background(), &lveList, &client.ListOptions{}); err != nil {
		return nil, err
	}

	for _, item := range lveList.Items {
		var operation hwameistorapi.Operation
		operation.EventName = item.Name
		operation.EventType = item.Kind
		operation.Status = hwameistorapi.StateConvert(item.Status.State)
		operation.StartTime = item.CreationTimestamp.Time
		operation.EndTime = time.Now()
		operation.Description = item.Status.Message
		operationList = append(operationList, operation)
	}

	var operations = []hwameistorapi.Operation{}
	operationMetric.OperationList = utils.DataPatination(operationList, page, pageSize)
	if len(operationList) == 0 {
		operationMetric.OperationList = operations
	}

	var pagination = &hwameistorapi.Pagination{}
	pagination.Page = page
	pagination.PageSize = pageSize
	pagination.Total = uint32(len(operationList))
	if len(operationList) == 0 {
		pagination.Pages = 0
	} else {
		pagination.Pages = int32(math.Ceil(float64(len(operationList)) / float64(pageSize)))
	}
	operationMetric.Page = pagination

	return operationMetric, nil
}

// getBaseCapacityMetric
func (mController *MetricController) getBaseCapacityMetric() error {

	mController.resetNodeResourceMetric()

	mController.addK8sNodeMetric()

	lsnList := &apisv1alpha1.LocalStorageNodeList{}
	if err := mController.Client.List(context.TODO(), lsnList); err != nil {
		log.WithError(err).Error("Failed to list LocalStorageNodes")
		return err
	}

	for i := range lsnList.Items {
		if lsnList.Items[i].Status.State == apisv1alpha1.NodeStateReady {
			mController.addNodeResourceMetric(&lsnList.Items[i])
		}
	}

	return nil
}

// getBaseVolumeMetric
func (mController *MetricController) getBaseVolumeMetric() error {

	volList := &apisv1alpha1.LocalVolumeList{}
	if err := mController.Client.List(context.TODO(), volList); err != nil {
		log.WithError(err).Error("Failed to list LocalVolumes")
		return err
	}

	// 高可用 = convertible + replicas >= 2
	// 非高可用 = unconvertible + (convertible & replicas == 1)
	for i := range volList.Items {
		if volList.Items[i].Spec.Convertible == true && volList.Items[i].Spec.ReplicaNumber >= 2 {
			mController.volumeCollection.HAVolumeNum++
		} else {
			mController.volumeCollection.NonHAVolumeNum++
		}
	}
	mController.volumeCollection.TotalVolumesNum = int64(len(volList.Items))

	return nil
}

// getBaseDiskMetric
func (mController *MetricController) getBaseDiskMetric() error {
	diskList := &apisv1alpha1.LocalDiskList{}
	if err := mController.Client.List(context.TODO(), diskList); err != nil {
		log.WithError(err).Error("Failed to list LocalDisks")
		return err
	}

	for i := range diskList.Items {
		if diskList.Items[i].Spec.State == apisv1alpha1.LocalDiskActive {
			mController.diskCollection.HealthyDiskNum++
		} else {
			mController.diskCollection.ErrorDiskNum++
		}
	}
	mController.diskCollection.TotalDisksNum = int64(len(diskList.Items))
	mController.diskCollection.BoundedDiskNum = int64(len(diskList.Items))

	return nil
}

// convertBaseMetric
func (mController *MetricController) convertBaseMetric() *hwameistorapi.BaseMetric {
	basemetric := &hwameistorapi.BaseMetric{}

	basemetric.ReservedCapacityBytes = mController.storageCapacityCollection.ReservedCapacityBytes
	basemetric.TotalCapacityBytes = mController.storageCapacityCollection.TotalCapacityBytes
	basemetric.AllocatedCapacityBytes = mController.storageCapacityCollection.AllocatedCapacityBytes
	basemetric.FreeCapacityBytes = mController.storageCapacityCollection.FreeCapacityBytes
	basemetric.TotalNodeNum = mController.storageNodesCollection.TotalNodesNum
	basemetric.ClaimedNodeNum = mController.storageNodesCollection.ManagedNodesNum

	basemetric.HighAvailableVolumeNum = mController.volumeCollection.HAVolumeNum
	basemetric.NonHighAvailableVolumeNum = mController.volumeCollection.NonHAVolumeNum
	basemetric.LocalVolumeNum = mController.volumeCollection.TotalVolumesNum

	basemetric.HealthyDiskNum = mController.diskCollection.HealthyDiskNum
	basemetric.BoundedDiskNum = mController.diskCollection.BoundedDiskNum
	basemetric.TotalDiskNum = mController.diskCollection.TotalDisksNum
	basemetric.UnHealthyDiskNum = mController.diskCollection.ErrorDiskNum

	return basemetric
}

// convertModuleStatus
func (mController *MetricController) convertModuleStatus() *hwameistorapi.ModuleStatus {
	ModuleStatus := &hwameistorapi.ModuleStatus{}

	if mController.moduleStatusCollection != nil {
		for name, state := range mController.moduleStatusCollection.ModuleStatus {
			moduleState := hwameistorapi.ModuleState{}
			moduleState.Name = name
			moduleState.State = state
			ModuleStatus.ModulesStatus = append(ModuleStatus.ModulesStatus, moduleState)
		}
	}

	return ModuleStatus
}

// convertStoragePoolUseMetric
func (mController *MetricController) convertStoragePoolUseMetric() *hwameistorapi.StoragePoolUseMetric {
	storagePoolUseMetric := &hwameistorapi.StoragePoolUseMetric{}

	if mController.storagePoolUseCollection != nil {
		for name, poolUse := range mController.storagePoolUseCollection.StoragePoolUseMap {
			storagePoolUse := hwameistorapi.StoragePoolUse{}
			storagePoolUse.Name = name
			storagePoolUse.AllocatedCapacityBytes = poolUse.AllocatedCapacityBytes
			storagePoolUse.TotalCapacityBytes = poolUse.TotalCapacityBytes
			storagePoolUseMetric.StoragePoolsUse = append(storagePoolUseMetric.StoragePoolsUse, storagePoolUse)
		}
	}

	return storagePoolUseMetric
}

// resetNodeResourceMetric
func (mController *MetricController) resetNodeResourceMetric() {
	mController.storageCapacityCollection.ReservedCapacityBytes = 0
	mController.storageCapacityCollection.AllocatedCapacityBytes = 0
	mController.storageCapacityCollection.TotalCapacityBytes = 0
	mController.storageCapacityCollection.FreeCapacityBytes = 0

	mController.storageNodesCollection.TotalNodesNum = 0
	mController.storageNodesCollection.ManagedNodesNum = 0

	mController.volumeCollection.HAVolumeNum = 0
	mController.volumeCollection.NonHAVolumeNum = 0
	mController.volumeCollection.TotalVolumesNum = 0

	mController.diskCollection.TotalDisksNum = 0
	mController.diskCollection.BoundedDiskNum = 0
	mController.diskCollection.ErrorDiskNum = 0
	mController.diskCollection.HealthyDiskNum = 0
}

// addNodeResourceMetric
func (mController *MetricController) addNodeResourceMetric(node *apisv1alpha1.LocalStorageNode) {
	mController.lock.Lock()
	defer mController.lock.Unlock()

	for _, pool := range node.Status.Pools {
		mController.storageCapacityCollection.TotalCapacityBytes += pool.TotalCapacityBytes
		mController.storageCapacityCollection.AllocatedCapacityBytes += pool.UsedCapacityBytes
		mController.storageCapacityCollection.FreeCapacityBytes += pool.FreeCapacityBytes
	}
	mController.storageNodesCollection.ManagedNodesNum++
}

// addK8sNodeMetric
func (mController *MetricController) addK8sNodeMetric() {
	mController.lock.Lock()
	defer mController.lock.Unlock()

	// list K8S nodes
	nodes, err := mController.clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	mController.storageNodesCollection.TotalNodesNum = int64(len(nodes.Items))
}

// addNodeReservedCapacityMetric
func (mController *MetricController) addNodeReservedCapacityMetric() error {
	mController.lock.Lock()
	defer mController.lock.Unlock()

	diskList := &apisv1alpha1.LocalDiskList{}
	if err := mController.Client.List(context.TODO(), diskList); err != nil {
		log.WithError(err).Error("Failed to list LocalDisks")
		return err
	}
	for i := range diskList.Items {
		if diskList.Items[i].Spec.Reserved == true {
			mController.storageCapacityCollection.ReservedCapacityBytes += diskList.Items[i].Spec.Capacity
		}
	}
	return nil
}

// getBaseNodeMetric
func (mController *MetricController) getBaseNodeMetric() error {

	if err := mController.addNodeReservedCapacityMetric(); err != nil {
		return err
	}

	return nil
}

// getHwameistorDaemonsetStatusMetric
func (mController *MetricController) getHwameistorDaemonsetStatusMetric() error {

	// 获取daemonset的资源名字
	daemonsets, err := mController.clientset.AppsV1().DaemonSets(mController.nameSpace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.WithError(err).Error("Failed to list daemonsets")
		return err
	}
	for _, ds := range daemonsets.Items {
		for _, label := range ds.Spec.Selector.MatchLabels {
			if !strings.Contains(label, hwameistorPrefix) {
				continue
			}
			if ds.Status.CurrentNumberScheduled != 0 && ds.Status.CurrentNumberScheduled == ds.Status.DesiredNumberScheduled && ds.Status.CurrentNumberScheduled == ds.Status.NumberReady {
				mController.moduleStatusCollection.ModuleStatus[ds.Name] = hwameistorapi.ModuleStatusRunning
			} else {
				mController.moduleStatusCollection.ModuleStatus[ds.Name] = hwameistorapi.ModuleStatusNotReady
			}
		}
	}

	return nil
}

// getHwameistorDeploymentStatusMetric
func (mController *MetricController) getHwameistorDeploymentStatusMetric() error {

	// 获取deployments的资源名字
	deployments, err := mController.clientset.AppsV1().Deployments(mController.nameSpace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.WithError(err).Error("Failed to list daemonsets")
		return err
	}
	for _, deployment := range deployments.Items {
		for _, label := range deployment.Spec.Selector.MatchLabels {
			if !strings.Contains(label, hwameistorPrefix) {
				continue
			}
			if deployment.Status.ReadyReplicas != 0 && deployment.Status.ReadyReplicas == deployment.Status.AvailableReplicas {
				mController.moduleStatusCollection.ModuleStatus[deployment.Name] = hwameistorapi.ModuleStatusRunning
			} else {
				mController.moduleStatusCollection.ModuleStatus[deployment.Name] = hwameistorapi.ModuleStatusNotReady
			}
		}
	}

	return nil
}

// getHwameistorOperatorStatusMetric
func (mController *MetricController) getHwameistorOperatorStatusMetric() (*hwameistorapi.ModuleStatus, error) {

	var moduleStatus = &hwameistorapi.ModuleStatus{}
	clusterList := &hoapisv1alpha1.ClusterList{}
	if err := mController.Client.List(context.TODO(), clusterList); err != nil {
		if !errors.IsNotFound(err) {
			log.WithError(err).Error("Failed to list clusterList")
		} else {
			log.Info("Not found the clusterList")
		}
		return moduleStatus, err
	}

	for _, cluster := range clusterList.Items {
		if cluster.Name == OperatorClusterName {
			moduleStatus.ClusterStatus = cluster.Status
		}
	}
	return moduleStatus, nil
}

// addStoragePoolUseMetric
func (mController *MetricController) addStoragePoolUseMetric() error {
	mController.lock.Lock()
	defer mController.lock.Unlock()

	lsnList := &apisv1alpha1.LocalStorageNodeList{}
	if err := mController.Client.List(context.TODO(), lsnList); err != nil {
		log.WithError(err).Error("Failed to list LocalStorageNodes")
		return err
	}

	for i := range lsnList.Items {
		if lsnList.Items[i].Status.State == apisv1alpha1.NodeStateReady {
			for _, pool := range lsnList.Items[i].Status.Pools {
				if value, exists := mController.storagePoolUseCollection.StoragePoolUseMap[pool.Name]; exists {
					value.TotalCapacityBytes = value.TotalCapacityBytes + pool.TotalCapacityBytes
					value.AllocatedCapacityBytes = value.AllocatedCapacityBytes + pool.UsedCapacityBytes
					mController.storagePoolUseCollection.StoragePoolUseMap[pool.Name] = value
				} else {
					storagePoolUse := hwameistorapi.StoragePoolUse{}
					storagePoolUse.TotalCapacityBytes = pool.TotalCapacityBytes
					storagePoolUse.AllocatedCapacityBytes = pool.UsedCapacityBytes
					mController.storagePoolUseCollection.StoragePoolUseMap[pool.Name] = storagePoolUse
				}
			}
		}
	}
	return nil
}

// addNodeStorageUseMetric
func (mController *MetricController) addNodeStorageUseMetric(storagepoolclass string) error {
	mController.lock.Lock()
	defer mController.lock.Unlock()

	lsnList := &apisv1alpha1.LocalStorageNodeList{}
	if err := mController.Client.List(context.TODO(), lsnList); err != nil {
		log.WithError(err).Error("Failed to list LocalStorageNodes")
		return err
	}

	var nodeStorageUseRatios hwameistorapi.NodeStorageUseRatios
	for i := range lsnList.Items {
		if lsnList.Items[i].Status.State == apisv1alpha1.NodeStateReady {
			var nodeStorageUseRatio = &hwameistorapi.NodeStorageUseRatio{}
			for _, pool := range lsnList.Items[i].Status.Pools {
				if pool.Class == storagepoolclass {
					nodeStorageUseRatio.Name = lsnList.Items[i].Name
					nodeStorageUseRatio.TotalCapacityBytes = pool.TotalCapacityBytes
					nodeStorageUseRatio.AllocatedCapacityBytes = pool.UsedCapacityBytes
					capacityRatio, _ := utils.DivideOperate(pool.UsedCapacityBytes, pool.TotalCapacityBytes)
					nodeStorageUseRatio.CapacityBytesRatio = int64(capacityRatio * 100)
					nodeStorageUseRatios = append(nodeStorageUseRatios, nodeStorageUseRatio)
				}
			}
		}
	}

	// 使用sort包进行排序
	sort.Stable(sort.Reverse(nodeStorageUseRatios))
	mController.nodeStorageUseCollection.NodeStorageUseRatios = nodeStorageUseRatios

	return nil
}

// convertNodeStorageUseMetric
func (mController *MetricController) convertNodeStorageUseMetric(storagepoolclass string) *hwameistorapi.NodeStorageUseMetric {
	nodeStorageUseMetric := &hwameistorapi.NodeStorageUseMetric{}

	nodeStorageUseMetric.StoragePoolClass = storagepoolclass
	if mController.nodeStorageUseCollection != nil {
		for i, ratio := range mController.nodeStorageUseCollection.NodeStorageUseRatios {
			if i < nodeStorageSortNum {
				nodeStorageUse := hwameistorapi.NodeStorageUse{}
				nodeStorageUse.Name = ratio.Name
				nodeStorageUse.AllocatedCapacityBytes = ratio.AllocatedCapacityBytes
				nodeStorageUse.TotalCapacityBytes = ratio.TotalCapacityBytes
				nodeStorageUseMetric.NodeStoragesUse = append(nodeStorageUseMetric.NodeStoragesUse, nodeStorageUse)
			}
		}
	}

	return nodeStorageUseMetric
}
