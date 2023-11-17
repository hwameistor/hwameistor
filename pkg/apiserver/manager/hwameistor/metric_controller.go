package hwameistor

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
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
	hwameistorPrefix   = "hwameistor"
	nodeStorageSortNum = 5
	TimeSort           = "time"
	TypeSort           = "type"
	NameSort           = "name"
	Migrate            = "Migrate"
	Expand             = "Expand"
	Convert            = "Convert"
)

type MetricController struct {
	client.Client
	record.EventRecorder

	nameSpace string
	clientset *kubernetes.Clientset

	storageCapacityCollection *hwameistorapi.StorageCapacityCollection
	storageNodesCollection    *hwameistorapi.StorageNodesCollection
	volumeCollection          *hwameistorapi.VolumeCollection
	diskCollection            *hwameistorapi.DiskCollection
	moduleStatusCollection    *hwameistorapi.ModuleStatusCollection
	storagePoolUseCollection  *hwameistorapi.StoragePoolUseCollection
	nodeStorageUseCollection  *hwameistorapi.NodeStorageUseCollection

	lock sync.Mutex
}

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
		moduleStatusCollection:    &hwameistorapi.ModuleStatusCollection{ModuleStatus: map[string]hwameistorapi.State{}},
		storagePoolUseCollection:  &hwameistorapi.StoragePoolUseCollection{StoragePoolUseMap: map[string]hwameistorapi.StoragePoolUse{}},
		nodeStorageUseCollection:  &hwameistorapi.NodeStorageUseCollection{},
	}
}

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

func (mController *MetricController) GetStoragePoolUseMetric() (*hwameistorapi.StoragePoolUseMetric, error) {
	if err := mController.addStoragePoolUseMetric(); err != nil {
		log.WithError(err).Error("Failed to addStoragePoolUseMetric")
		return nil, err
	}
	storagePoolUseMetric := mController.convertStoragePoolUseMetric()

	return storagePoolUseMetric, nil
}

func (mController *MetricController) GetNodeStorageUseMetric(storagepoolclass string) (*hwameistorapi.NodeStorageUseMetric, error) {
	if err := mController.addNodeStorageUseMetric(storagepoolclass); err != nil {
		log.WithError(err).Error("Failed to addNodeStorageUseMetric")
		return nil, err
	}
	nodeStorageUseMetric := mController.convertNodeStorageUseMetric(storagepoolclass)

	return nodeStorageUseMetric, nil
}

func (mController *MetricController) OperationListMetric(page, pageSize int32, name string) (*hwameistorapi.OperationMetric, error) {
	var operationMetric = &hwameistorapi.OperationMetric{}
	var operationList []hwameistorapi.Operation
	lvmList := apisv1alpha1.LocalVolumeMigrateList{}
	if err := mController.Client.List(context.Background(), &lvmList, &client.ListOptions{}); err != nil {
		return nil, err
	}

	for _, item := range lvmList.Items {
		if name != "" && strings.Contains(item.Name, name) {
			var operation hwameistorapi.Operation
			operation.EventName = item.Name
			operation.EventType = item.Kind
			operation.LocalVolumeName = item.Spec.VolumeName
			operation.Status = hwameistorapi.StateConvert(item.Status.State)
			operation.StartTime = item.CreationTimestamp.Time
			operation.Description = item.Status.Message
			operationList = append(operationList, operation)
		}
	}

	lvcList := apisv1alpha1.LocalVolumeConvertList{}
	if err := mController.Client.List(context.Background(), &lvcList, &client.ListOptions{}); err != nil {
		return nil, err
	}

	for _, item := range lvcList.Items {
		if name != "" && strings.Contains(item.Name, name) {
			var operation hwameistorapi.Operation
			operation.EventName = item.Name
			operation.EventType = item.Kind
			operation.Status = hwameistorapi.StateConvert(item.Status.State)
			operation.StartTime = item.CreationTimestamp.Time
			operation.EndTime = time.Now()
			operation.Description = item.Status.Message
			operationList = append(operationList, operation)
		}
	}

	lveList := apisv1alpha1.LocalVolumeExpandList{}
	if err := mController.Client.List(context.Background(), &lveList, &client.ListOptions{}); err != nil {
		return nil, err
	}

	for _, item := range lveList.Items {
		if name != "" && strings.Contains(item.Name, name) {
			var operation hwameistorapi.Operation
			operation.EventName = item.Name
			operation.EventType = item.Kind
			operation.Status = hwameistorapi.StateConvert(item.Status.State)
			operation.StartTime = item.CreationTimestamp.Time
			operation.EndTime = time.Now()
			operation.Description = item.Status.Message
			operationList = append(operationList, operation)
		}
	}

	var operations []hwameistorapi.Operation
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

func (mController *MetricController) GetOperation(eventName, eventType string) (*hwameistorapi.Operation, error) {

	switch eventType {
	case Migrate:
		lvm := apisv1alpha1.LocalVolumeMigrate{}
		if err := mController.Client.Get(context.Background(), types.NamespacedName{Name: eventName}, &lvm); err != nil {
			return nil, err
		}
		var operation hwameistorapi.Operation
		operation.EventName = lvm.Name
		operation.EventType = lvm.Kind
		operation.LocalVolumeName = lvm.Spec.VolumeName
		operation.Status = hwameistorapi.StateConvert(lvm.Status.State)
		operation.StartTime = lvm.CreationTimestamp.Time
		operation.Description = lvm.Status.Message
		return &operation, nil

	case Convert:
		lvc := apisv1alpha1.LocalVolumeConvert{}
		if err := mController.Client.Get(context.Background(), types.NamespacedName{Name: eventName}, &lvc); err != nil {
			return nil, err
		}
		var operation hwameistorapi.Operation
		operation.EventName = lvc.Name
		operation.EventType = lvc.Kind
		operation.LocalVolumeName = lvc.Spec.VolumeName
		operation.Status = hwameistorapi.StateConvert(lvc.Status.State)
		operation.StartTime = lvc.CreationTimestamp.Time
		operation.Description = lvc.Status.Message
		return &operation, nil

	case Expand:
		lve := apisv1alpha1.LocalVolumeExpand{}
		if err := mController.Client.Get(context.Background(), types.NamespacedName{Name: eventName}, &lve); err != nil {
			return nil, err
		}

		var operation hwameistorapi.Operation
		operation.EventName = lve.Name
		operation.EventType = lve.Kind
		operation.LocalVolumeName = lve.Spec.VolumeName
		operation.Status = hwameistorapi.StateConvert(lve.Status.State)
		operation.StartTime = lve.CreationTimestamp.Time
		operation.Description = lve.Status.Message
		return &operation, nil

	default:
		return nil, fmt.Errorf("Type error, only the following types are supported:Migrate,Expand,Convert")
	}
}

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

func (mController *MetricController) getBaseNodeMetric() error {
	if err := mController.addNodeReservedCapacityMetric(); err != nil {
		return err
	}

	return nil
}

func (mController *MetricController) getHwameistorDaemonsetStatusMetric() error {
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

func (mController *MetricController) getHwameistorDeploymentStatusMetric() error {
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

	if len(clusterList.Items) > 0 {
		moduleStatus.ClusterStatus = clusterList.Items[0].Status
	}

	return moduleStatus, nil
}

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

	sort.Stable(sort.Reverse(nodeStorageUseRatios))
	mController.nodeStorageUseCollection.NodeStorageUseRatios = nodeStorageUseRatios
	return nil
}

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

func (mController *MetricController) EventList(queryPage hwameistorapi.QueryPage) (*hwameistorapi.EventActionList, error) {
	var eventList = &hwameistorapi.EventActionList{}
	evls, err := mController.listEvent(queryPage)
	log.Infof("listEvent vols = %v", evls)
	if err != nil {
		log.WithError(err).Error("Failed to listEvent")
		return nil, err
	}

	if queryPage.Sort == TimeSort {
		a := utils.ByEventTime(evls)
		sort.Sort(a)
	} else if queryPage.Sort == TypeSort {
		a := utils.ByEventType(evls)
		sort.Sort(a)
	} else if queryPage.Sort == NameSort {
		a := utils.ByEventName(evls)
		sort.Sort(a)
	}

	eventList.EventActions = utils.DataPatination(evls, queryPage.Page, queryPage.PageSize)
	if len(evls) == 0 {
		eventList.EventActions = []*hwameistorapi.EventAction{}
	}

	var pagination = &hwameistorapi.Pagination{}
	pagination.Page = queryPage.Page
	pagination.PageSize = queryPage.PageSize
	pagination.Total = uint32(len(evls))
	if len(evls) == 0 {
		pagination.Pages = 0
	} else {
		pagination.Pages = int32(math.Ceil(float64(len(evls)) / float64(queryPage.PageSize)))
	}
	eventList.Page = pagination

	return eventList, nil
}

func (mController *MetricController) listEvent(queryPage hwameistorapi.QueryPage) ([]*hwameistorapi.EventAction, error) {
	evList := &apisv1alpha1.EventList{}
	if err := mController.Client.List(context.TODO(), evList); err != nil {
		log.WithError(err).Error("Failed to list LocalVolumes")
		return nil, err
	}
	log.Infof("listLocalVolume queryPage = %v, queryPage.VolumeState = %v", queryPage, queryPage.VolumeState)

	var eventActions []*hwameistorapi.EventAction
	for _, ev := range evList.Items {
		if (queryPage.ResourceName == "") && (queryPage.ResourceType == "") {
			for _, action := range ev.Spec.Records {
				ea := &hwameistorapi.EventAction{}
				ea.EventRecord = action
				ea.ResourceName = ev.Spec.ResourceName
				ea.ResourceType = ev.Spec.ResourceType
				eventActions = append(eventActions, ea)
			}
		} else if (queryPage.ResourceName != "" && strings.Contains(ev.Spec.ResourceName, queryPage.ResourceName)) && (queryPage.ResourceType == "") {
			for _, action := range ev.Spec.Records {
				ea := &hwameistorapi.EventAction{}
				ea.EventRecord = action
				ea.ResourceName = ev.Spec.ResourceName
				ea.ResourceType = ev.Spec.ResourceType
				eventActions = append(eventActions, ea)
			}
		} else if (queryPage.ResourceName == "") && (queryPage.ResourceType != "" && queryPage.ResourceType == ev.Spec.ResourceType) {
			for _, action := range ev.Spec.Records {
				ea := &hwameistorapi.EventAction{}
				ea.EventRecord = action
				ea.ResourceName = ev.Spec.ResourceName
				ea.ResourceType = ev.Spec.ResourceType
				eventActions = append(eventActions, ea)
			}
		} else if (queryPage.ResourceName != "" && strings.Contains(ev.Spec.ResourceName, queryPage.ResourceName)) && (queryPage.ResourceType != "" && queryPage.ResourceType == ev.Spec.ResourceType) {
			for _, action := range ev.Spec.Records {
				ea := &hwameistorapi.EventAction{}
				ea.EventRecord = action
				ea.ResourceName = ev.Spec.ResourceName
				ea.ResourceType = ev.Spec.ResourceType
				eventActions = append(eventActions, ea)
			}
		}
	}

	if queryPage.Action != "" {
		var eas []*hwameistorapi.EventAction
		for _, action := range eventActions {
			if action.EventRecord.Action == queryPage.Action {
				eas = append(eas, action)
			}
		}
		eventActions = eas
	}

	return eventActions, nil
}

func (mController *MetricController) GetEvent(eventName string) (*hwameistorapi.Event, error) {
	ev := apisv1alpha1.Event{}
	if err := mController.Client.Get(context.Background(), types.NamespacedName{Name: eventName}, &ev); err != nil {
		return nil, err
	}
	var event = &hwameistorapi.Event{}
	event.Event = ev
	return event, nil
}
