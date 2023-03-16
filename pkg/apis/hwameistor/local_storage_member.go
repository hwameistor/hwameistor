package apis

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	nodestorage "github.com/hwameistor/hwameistor/pkg/local-storage/member/node/storage"
)

// consts
const (
	Version = "1.0.0"

	NodeLeaseNamePrefix = "hwameistor-local-storage-worker"

	ControllerLeaseName = "hwameistor-local-storage-master"

	TopologyNodeKey = "topology.lvm.hwameistor.io/node"
)

// LocalStorageMember interface
// //go:generate mockgen -source=member.go -destination=../member/member_mock.go  -package=member
type LocalStorageMember interface {
	Run(stopCh <-chan struct{})

	// ******  configuration ******* //
	ConfigureBase(name string, namespace string, haSystemConfig apisv1alpha1.SystemConfig, cli client.Client, informersCache cache.Cache, recorder record.EventRecorder) LocalStorageMember

	ConfigureNode(scheme *runtime.Scheme) LocalStorageMember

	ConfigureController(scheme *runtime.Scheme) LocalStorageMember

	ConfigureCSIDriver(driverName string, sockAddr string) LocalStorageMember

	ConfigureRESTServer(httpPort int) LocalStorageMember

	// access the modules
	Controller() ControllerManager

	Node() NodeManager

	Name() string

	Version() string

	DriverName() string
}

// ControllerManager interface
//
//go:generate mockgen -source=member.go -destination=../member/controller/manager_mock.go  -package=controller
type ControllerManager interface {
	Run(stopCh <-chan struct{})

	VolumeScheduler() apisv1alpha1.VolumeScheduler

	VolumeGroupManager() apisv1alpha1.VolumeGroupManager

	ReconcileNode(node *apisv1alpha1.LocalStorageNode)

	ReconcileVolume(vol *apisv1alpha1.LocalVolume)

	ReconcileVolumeGroup(volGroup *apisv1alpha1.LocalVolumeGroup)

	ReconcileVolumeExpand(expand *apisv1alpha1.LocalVolumeExpand)

	ReconcileVolumeMigrate(migrate *apisv1alpha1.LocalVolumeMigrate)

	ReconcileVolumeConvert(convert *apisv1alpha1.LocalVolumeConvert)
}

// NodeManager interface
// //go:generate mockgen -source=member.go -destination=../member/node/manager_mock.go  -package=node
type NodeManager interface {
	Run(stopCh <-chan struct{})

	Storage() *nodestorage.LocalManager

	TakeVolumeReplicaTaskAssignment(vol *apisv1alpha1.LocalVolume)

	ReconcileVolumeReplica(replica *apisv1alpha1.LocalVolumeReplica)
}
