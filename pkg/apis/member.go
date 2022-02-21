package apis

import (
	udsv1alpha1 "github.com/HwameiStor/local-storage/pkg/apis/uds/v1alpha1"
	nodestorage "github.com/HwameiStor/local-storage/pkg/member/node/storage"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// consts
const (
	Version = "1.0.0"

	NodeLeaseNamePrefix = "dce-uds-local-storage-worker"

	ControllerLeaseName = "dce-uds-local-storage-master"

	TopologyNodeKey = "uds.dce.daocloud.io/local-storage-topology-node"
)

// LocalStorageMember interface
type LocalStorageMember interface {
	Run(stopCh <-chan struct{})

	// ******  configuration ******* //
	ConfigureBase(name string, namespace string, haSystemConfig udsv1alpha1.SystemConfig, cli client.Client, informersCache cache.Cache) LocalStorageMember

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

	ReconcileNode(node *udsv1alpha1.LocalStorageNode)

	ReconcileVolume(vol *udsv1alpha1.LocalVolume)

	ReconcileVolumeExpand(vol *udsv1alpha1.LocalVolumeExpand)

	ReconcileVolumeMigrate(vol *udsv1alpha1.LocalVolumeMigrate)

	ReconcileVolumeConvert(vol *udsv1alpha1.LocalVolumeConvert)
}

// NodeManager interface
type NodeManager interface {
	Run(stopCh <-chan struct{})

	Storage() *nodestorage.LocalManager

	TakeVolumeReplicaTaskAssignment(vol *udsv1alpha1.LocalVolume)

	ReconcileVolumeReplica(replica *udsv1alpha1.LocalVolumeReplica)
}
