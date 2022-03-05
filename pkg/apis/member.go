package apis

import (
	localstoragev1alpha1 "github.com/hwameistor/local-storage/pkg/apis/localstorage/v1alpha1"
	nodestorage "github.com/hwameistor/local-storage/pkg/member/node/storage"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// consts
const (
	Version = "1.0.0"

	NodeLeaseNamePrefix = "hwameistor-local-storage-worker"

	ControllerLeaseName = "hwameistor-local-storage-master"

	TopologyNodeKey = "topology.lvm.hwameistor.io/node"
)

// LocalStorageMember interface
type LocalStorageMember interface {
	Run(stopCh <-chan struct{})

	// ******  configuration ******* //
	ConfigureBase(name string, namespace string, haSystemConfig localstoragev1alpha1.SystemConfig, cli client.Client, informersCache cache.Cache) LocalStorageMember

	ConfigureNode() LocalStorageMember

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
type ControllerManager interface {
	Run(stopCh <-chan struct{})

	ReconcileNode(node *localstoragev1alpha1.LocalStorageNode)

	ReconcileVolume(vol *localstoragev1alpha1.LocalVolume)

	ReconcileVolumeExpand(vol *localstoragev1alpha1.LocalVolumeExpand)

	ReconcileVolumeMigrate(vol *localstoragev1alpha1.LocalVolumeMigrate)

	ReconcileVolumeConvert(vol *localstoragev1alpha1.LocalVolumeConvert)
}

// NodeManager interface
type NodeManager interface {
	Run(stopCh <-chan struct{})

	Storage() *nodestorage.LocalManager

	TakeVolumeReplicaTaskAssignment(vol *localstoragev1alpha1.LocalVolume)

	ReconcileVolumeReplica(replica *localstoragev1alpha1.LocalVolumeReplica)
}
