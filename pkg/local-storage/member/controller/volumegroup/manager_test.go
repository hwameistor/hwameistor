package volumegroup

import (
	"context"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	coorv1 "k8s.io/api/coordination/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	crmgr "sigs.k8s.io/controller-runtime/pkg/manager"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-storage/common"
)

// SystemMode of HA module
type SystemMode string

// Infinitely retry
const maxRetries = 0

var (
	fakeLocalStorageNodeName              = "local-storage-node-example"
	fakeLocalVolumeGroupName              = "local-volume-group-example"
	fakeLocalStorageNodeUID               = "local-storage-node-uid"
	fakeLocalStorageNodename              = "local-storage-node-example"
	fakeLocalVolumeReplicaName            = "local-volume-replica-example"
	fakeLocalVolumeConvertName            = "local-volume-convert-example"
	fakeLocalVolumeMigrateName            = "local-volume-migrate-example"
	fakeLocalVolumeGroupMigrateName       = "local-volume-group-migrate-example"
	fakeLocalVolumeGroupConvertName       = "local-volume-group-convert-example"
	fakeNamespace                         = "local-volume-group-test"
	fakeName                              = "name-test"
	fakeNodename                          = "10-6-118-10"
	fakeStorageIp                         = "10.6.118.11"
	fakeZone                              = "zone-test"
	fakeRegion                            = "region-test"
	fakeVgType                            = "LocalStorage_PoolHDD"
	fakeVgName                            = "vg-test"
	fakeTopo                              = apisv1alpha1.Topology{Region: fakeRegion, Zone: fakeZone}
	fakeNodenames                         = []string{"10-6-118-10"}
	fakePersistentPvcName                 = "pvc-test"
	fakePodName                           = "pod-test"
	fakeContainerName                     = "container-test"
	fakePoolClass                         = "HDD"
	fakePoolType                          = "REGULAR"
	fakeLocalVolumeUID                    = "local-volume-uid"
	fakeTotalCapacityBytes          int64 = 10 * 1024 * 1024 * 1024
	fakeFreeCapacityBytes           int64 = 8 * 1024 * 1024 * 1024
	fakeDiskCapacityBytes           int64 = 2 * 1024 * 1024 * 1024
	fakeHolderIdentity                    = "fakeHolderIdentity"
	fakeLeaseDurationSeconds              = int32(30)
	fakeLeaseTransitions                  = int32(30)
	fakeAcquireTime                       = time.Now()
	fakeStorageClassName                  = "sc-test"

	LocalVolumeKind             = "LocalVolume"
	LocalStorageNodeKind        = "LocalStorageNode"
	LeaseKind                   = "Lease"
	LocalVolumeReplicaKind      = "LocalVolumeReplica"
	LocalVolumeConvertKind      = "LocalVolumeConvert"
	LocalVolumeMigrateKind      = "LocalVolumeMigrate"
	LocalVolumeGroupConvertKind = "LocalVolumeGroupConvert"
	LocalVolumeGroupMigrateKind = "LocalVolumeGroupMigrate"

	fakePods                             = []string{"pod-test1"}
	fakeAcesscibility                    = apisv1alpha1.AccessibilityTopology{Nodes: []string{"test-node1"}}
	fakeLocalVolumeName                  = "local-volume-test1"
	fakeVolumes                          = []apisv1alpha1.VolumeInfo{{LocalVolumeName: fakeLocalVolumeName, PersistentVolumeClaimName: fakePersistentPvcName}}
	apiversion                           = "hwameistor.io/v1alpha1"
	LocalVolumeGroupKind                 = "LocalVolumeGroup"
	fakeRecorder                         = record.NewFakeRecorder(100)
	SystemModeDRBD            SystemMode = "drbd"
	defaultDRBDStartPort                 = 43001
	defaultHAVolumeTotalCount            = 1000
)

//func TestNewManager(t *testing.T) {
//	// Set default manager options
//	options := crmgr.Options{
//		Namespace: "", // watch all namespaces
//	}
//
//	// Get a config to talk to the apiserver
//	cfg, err := config.GetConfig()
//	if err != nil {
//		log.Error(err, "")
//		os.Exit(1)
//	}
//
//	// Create a new manager to provide shared dependencies and start components
//	mgr, err := crmgr.New(cfg, options)
//	if err != nil {
//		log.Error(err, "")
//		os.Exit(1)
//	}
//
//	type args struct {
//		cli            client.Client
//		informersCache cache.Cache
//	}
//	tests := []struct {
//		name string
//		args args
//		want v1alpha1.VolumeGroupManager
//	}{
//		// TODO: Add test cases.
//		{
//			args: args{
//				cli:            mgr.GetClient(),
//				informersCache: mgr.GetCache(),
//			},
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if got := NewManager(tt.args.cli, tt.args.informersCache); !reflect.DeepEqual(got, tt.want) {
//				//t.Logf("NewManager() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}

// GenFakeLocalVolumeObject Create lv request
func GenFakeLocalVolumeObject() *apisv1alpha1.LocalVolume {
	lv := &apisv1alpha1.LocalVolume{}

	TypeMeta := metav1.TypeMeta{
		Kind:       LocalVolumeKind,
		APIVersion: apiversion,
	}

	ObjectMata := metav1.ObjectMeta{
		Name:              fakeLocalVolumeName,
		ResourceVersion:   "",
		UID:               types.UID(fakeLocalVolumeUID),
		CreationTimestamp: metav1.Time{Time: time.Now()},
	}

	Spec := apisv1alpha1.LocalVolumeSpec{
		RequiredCapacityBytes: fakeDiskCapacityBytes,
		ReplicaNumber:         1,
		PoolName:              fakeVgType,
		Delete:                false,
		Convertible:           true,
		Accessibility: apisv1alpha1.AccessibilityTopology{
			Nodes:   fakeNodenames,
			Regions: []string{fakeRegion},
			Zones:   []string{fakeZone},
		},
		Config: &apisv1alpha1.VolumeConfig{
			Convertible:           true,
			Initialized:           true,
			ReadyToInitialize:     true,
			RequiredCapacityBytes: fakeDiskCapacityBytes,
			ResourceID:            5,
			Version:               11,
			VolumeName:            fakeLocalVolumeName,
			Replicas: []apisv1alpha1.VolumeReplica{
				{
					Hostname: fakeNodename,
					ID:       1,
					IP:       fakeStorageIp,
					Primary:  true,
				},
			},
		},
	}

	lv.ObjectMeta = ObjectMata
	lv.TypeMeta = TypeMeta
	lv.Spec = Spec
	lv.Status.State = apisv1alpha1.VolumeStateCreating
	lv.Status.AllocatedCapacityBytes = fakeTotalCapacityBytes - fakeFreeCapacityBytes
	lv.Status.PublishedNodeName = fakeNodename
	lv.Status.Replicas = []string{fakeLocalVolumeName}

	return lv
}

// GenFakeLocalVolumeObject Create lv request
func GenFakeLocalVolumeReplicaObject() *apisv1alpha1.LocalVolumeReplica {
	lvr := &apisv1alpha1.LocalVolumeReplica{}

	TypeMeta := metav1.TypeMeta{
		Kind:       LocalVolumeReplicaKind,
		APIVersion: apiversion,
	}

	ObjectMata := metav1.ObjectMeta{
		Name:              fakeLocalVolumeName,
		Namespace:         fakeNamespace,
		ResourceVersion:   "",
		UID:               types.UID(fakeLocalVolumeUID),
		CreationTimestamp: metav1.Time{Time: time.Now()},
	}

	Spec := apisv1alpha1.LocalVolumeReplicaSpec{
		RequiredCapacityBytes: fakeDiskCapacityBytes,
		PoolName:              fakeVgType,
		Delete:                false,
		VolumeName:            fakeLocalVolumeName,
		NodeName:              fakeNodename,
	}

	lvr.ObjectMeta = ObjectMata
	lvr.TypeMeta = TypeMeta
	lvr.Spec = Spec
	lvr.Status.State = apisv1alpha1.VolumeStateCreating
	lvr.Status.AllocatedCapacityBytes = fakeTotalCapacityBytes - fakeFreeCapacityBytes

	return lvr
}

func GenFakeLocalVolumeConvertObject() *apisv1alpha1.LocalVolumeConvert {
	lvc := &apisv1alpha1.LocalVolumeConvert{}

	TypeMeta := metav1.TypeMeta{
		Kind:       LocalVolumeConvertKind,
		APIVersion: apiversion,
	}

	ObjectMata := metav1.ObjectMeta{
		Name:              fakeLocalVolumeConvertName,
		ResourceVersion:   "",
		UID:               types.UID(fakeLocalVolumeUID),
		CreationTimestamp: metav1.Time{Time: time.Now()},
	}

	Spec := apisv1alpha1.LocalVolumeConvertSpec{
		ReplicaNumber: 1,
		VolumeName:    fakeLocalVolumeName,
	}

	lvc.ObjectMeta = ObjectMata
	lvc.TypeMeta = TypeMeta
	lvc.Spec = Spec

	return lvc
}

func GenFakeLocalVolumeMigrateObject() *apisv1alpha1.LocalVolumeMigrate {
	lvm := &apisv1alpha1.LocalVolumeMigrate{}

	TypeMeta := metav1.TypeMeta{
		Kind:       LocalVolumeMigrateKind,
		APIVersion: apiversion,
	}

	ObjectMata := metav1.ObjectMeta{
		Name:              fakeLocalVolumeConvertName,
		ResourceVersion:   "",
		UID:               types.UID(fakeLocalVolumeUID),
		CreationTimestamp: metav1.Time{Time: time.Now()},
	}

	Spec := apisv1alpha1.LocalVolumeMigrateSpec{
		TargetNodesSuggested: fakeNodenames,
		VolumeName:           fakeLocalVolumeName,
	}

	lvm.ObjectMeta = ObjectMata
	lvm.TypeMeta = TypeMeta
	lvm.Spec = Spec

	return lvm
}

func GenFakePVCObject() *corev1.PersistentVolumeClaim {
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fakePersistentPvcName,
			Namespace: fakeNamespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			StorageClassName: &fakeStorageClassName,
		},
	}
	return pvc
}

func GenFakePodObject() *corev1.Pod {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fakePodName,
			Namespace: fakeNamespace,
		},
		Spec: corev1.PodSpec{
			RestartPolicy: "Never",
			Containers: []corev1.Container{
				{
					Name: fakeContainerName,
				},
			},
		},
	}
	return pod
}

func GenFakeLocalVolumeGroupObject() *apisv1alpha1.LocalVolumeGroup {
	lvg := &apisv1alpha1.LocalVolumeGroup{}

	TypeMeta := metav1.TypeMeta{
		Kind:       LocalVolumeGroupKind,
		APIVersion: apiversion,
	}

	ObjectMata := metav1.ObjectMeta{
		Name:              fakeLocalVolumeGroupName,
		Namespace:         fakeNamespace,
		ResourceVersion:   "",
		UID:               types.UID(fakeLocalVolumeUID),
		CreationTimestamp: metav1.Time{Time: time.Now()},
	}

	Spec := apisv1alpha1.LocalVolumeGroupSpec{
		Volumes:       fakeVolumes,
		Accessibility: fakeAcesscibility,
		Pods:          fakePods,
	}

	lvg.ObjectMeta = ObjectMata
	lvg.TypeMeta = TypeMeta
	lvg.Spec = Spec

	return lvg
}

func GenFakeLocalStorageNodeObject() *apisv1alpha1.LocalStorageNode {
	lsn := &apisv1alpha1.LocalStorageNode{}

	TypeMeta := metav1.TypeMeta{
		Kind:       LocalVolumeGroupKind,
		APIVersion: apiversion,
	}

	ObjectMata := metav1.ObjectMeta{
		Name:              fakeLocalVolumeGroupName,
		Namespace:         fakeNamespace,
		ResourceVersion:   "",
		UID:               types.UID(fakeLocalVolumeUID),
		CreationTimestamp: metav1.Time{Time: time.Now()},
	}

	Spec := apisv1alpha1.LocalStorageNodeSpec{
		HostName:  fakeNodename,
		StorageIP: fakeStorageIp,
		Topo:      fakeTopo,
	}

	lsn.ObjectMeta = ObjectMata
	lsn.TypeMeta = TypeMeta
	lsn.Spec = Spec

	return lsn
}

func GenFakeLeaseObject() *coorv1.Lease {
	lease := &coorv1.Lease{}

	TypeMeta := metav1.TypeMeta{
		Kind:       LocalVolumeGroupKind,
		APIVersion: apiversion,
	}

	ObjectMata := metav1.ObjectMeta{
		Name:              fakeLocalVolumeGroupName,
		Namespace:         fakeNamespace,
		ResourceVersion:   "",
		UID:               types.UID(fakeLocalVolumeUID),
		CreationTimestamp: metav1.Time{Time: time.Now()},
	}

	Spec := coorv1.LeaseSpec{
		HolderIdentity:       &fakeHolderIdentity,
		LeaseDurationSeconds: &fakeLeaseDurationSeconds,
		LeaseTransitions:     &fakeLeaseTransitions,
	}

	lease.ObjectMeta = ObjectMata
	lease.TypeMeta = TypeMeta
	lease.Spec = Spec

	return lease
}

// CreateFakeClient Create LocalVolume resource
func CreateFakeMgr() crmgr.Manager {
	// Set default manager options
	options := crmgr.Options{
		Namespace: "", // watch all namespaces
	}

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Create a new manager to provide shared dependencies and start components
	mgr, err := crmgr.New(cfg, options)
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}
	return mgr
}

// CreateFakeClient Create LocalVolume resource
func CreateFakeClient() (client.Client, *runtime.Scheme) {

	lv := GenFakeLocalVolumeObject()
	lvList := &apisv1alpha1.LocalVolumeList{
		TypeMeta: metav1.TypeMeta{
			Kind:       LocalVolumeKind,
			APIVersion: apiversion,
		},
	}

	lvc := GenFakeLocalVolumeConvertObject()
	lvcList := &apisv1alpha1.LocalVolumeConvertList{
		TypeMeta: metav1.TypeMeta{
			Kind:       LocalVolumeConvertKind,
			APIVersion: apiversion,
		},
	}

	lvm := GenFakeLocalVolumeMigrateObject()
	lvmList := &apisv1alpha1.LocalVolumeMigrateList{
		TypeMeta: metav1.TypeMeta{
			Kind:       LocalVolumeMigrateKind,
			APIVersion: apiversion,
		},
	}

	lvr := GenFakeLocalVolumeReplicaObject()
	lvrList := &apisv1alpha1.LocalVolumeReplicaList{
		TypeMeta: metav1.TypeMeta{
			Kind:       LocalVolumeReplicaKind,
			APIVersion: apiversion,
		},
	}

	lvg := GenFakeLocalVolumeGroupObject()
	lvgList := &apisv1alpha1.LocalVolumeGroupList{
		TypeMeta: metav1.TypeMeta{
			Kind:       LocalVolumeGroupKind,
			APIVersion: apiversion,
		},
	}

	lsn := GenFakeLocalStorageNodeObject()
	lsnList := &apisv1alpha1.LocalStorageNodeList{
		TypeMeta: metav1.TypeMeta{
			Kind:       LocalStorageNodeKind,
			APIVersion: apiversion,
		},
	}

	lease := GenFakeLeaseObject()
	leaseList := &coorv1.LeaseList{
		TypeMeta: metav1.TypeMeta{
			Kind:       LeaseKind,
			APIVersion: apiversion,
		},
	}

	s := scheme.Scheme
	s.AddKnownTypes(apisv1alpha1.SchemeGroupVersion, lv)
	s.AddKnownTypes(apisv1alpha1.SchemeGroupVersion, lvList)
	s.AddKnownTypes(apisv1alpha1.SchemeGroupVersion, lvc)
	s.AddKnownTypes(apisv1alpha1.SchemeGroupVersion, lvcList)
	s.AddKnownTypes(apisv1alpha1.SchemeGroupVersion, lvm)
	s.AddKnownTypes(apisv1alpha1.SchemeGroupVersion, lvmList)
	s.AddKnownTypes(apisv1alpha1.SchemeGroupVersion, lvr)
	s.AddKnownTypes(apisv1alpha1.SchemeGroupVersion, lvrList)
	s.AddKnownTypes(apisv1alpha1.SchemeGroupVersion, lvg)
	s.AddKnownTypes(apisv1alpha1.SchemeGroupVersion, lvgList)
	s.AddKnownTypes(apisv1alpha1.SchemeGroupVersion, lsn)
	s.AddKnownTypes(apisv1alpha1.SchemeGroupVersion, lsnList)
	s.AddKnownTypes(apisv1alpha1.SchemeGroupVersion, lease)
	s.AddKnownTypes(apisv1alpha1.SchemeGroupVersion, leaseList)
	return fake.NewFakeClientWithScheme(s), s
}

//func Test_manager_GetLocalVolumeGroupByLocalVolume(t *testing.T) {
//	type fields struct {
//		apiClient                 client.Client
//		informersCache            cache.Cache
//		logger                    *log.Entry
//		nameSpace                 string
//		lock                      sync.Mutex
//		localVolumeGroupQueue     *common.TaskQueue
//		localVolumeQueue          *common.TaskQueue
//		pvcQueue                  *common.TaskQueue
//		podQueue                  *common.TaskQueue
//		localVolumeToVolumeGroups map[string]string
//		pvcToVolumeGroups         map[string]string
//		podToVolumeGroups         map[string]string
//	}
//
//	client, _ := CreateFakeClient()
//
//	// Create LocalVolumeGroup
//	lvg := GenFakeLocalVolumeGroupObject()
//	lvg.Name = fakeLocalVolumeGroupName
//	lvg.Namespace = fakeNamespace
//	err := client.Create(context.Background(), lvg)
//	if err != nil {
//		t.Logf("Create LocalVolumeGroup fail %v", err)
//	}
//
//	var tmplvg = &apisv1alpha1.LocalVolumeGroup{}
//
//	type args struct {
//		ns     string
//		lvName string
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		want    *apisv1alpha1.LocalVolumeGroup
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//		{
//			args: args{
//				ns:     fakeNamespace,
//				lvName: fakeLocalVolumeName,
//			},
//			wantErr: true,
//			want:    tmplvg,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			m := &manager{
//				nameSpace:                 fakeNamespace,
//				apiClient:                 client,
//				localVolumeGroupQueue:     common.NewTaskQueue("localVolumeGroup", maxRetries),
//				localVolumeQueue:          common.NewTaskQueue("localVolume", maxRetries),
//				pvcQueue:                  common.NewTaskQueue("pvc", maxRetries),
//				podQueue:                  common.NewTaskQueue("pod", maxRetries),
//				localVolumeToVolumeGroups: make(map[string]string),
//				pvcToVolumeGroups:         make(map[string]string),
//				podToVolumeGroups:         make(map[string]string),
//				logger:                    log.WithField("Module", "ControllerManager"),
//			}
//			got, err := m.GetLocalVolumeGroupByLocalVolume(tt.args.ns, tt.args.lvName)
//			if (err != nil) != tt.wantErr {
//				t.Logf("GetLocalVolumeGroupByLocalVolume() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(got, tt.want) {
//				t.Logf("GetLocalVolumeGroupByLocalVolume() got = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func Test_manager_GetLocalVolumeGroupByName(t *testing.T) {
//	type fields struct {
//		apiClient                 client.Client
//		informersCache            cache.Cache
//		logger                    *log.Entry
//		nameSpace                 string
//		lock                      sync.Mutex
//		localVolumeGroupQueue     *common.TaskQueue
//		localVolumeQueue          *common.TaskQueue
//		pvcQueue                  *common.TaskQueue
//		podQueue                  *common.TaskQueue
//		localVolumeToVolumeGroups map[string]string
//		pvcToVolumeGroups         map[string]string
//		podToVolumeGroups         map[string]string
//	}
//
//	client, _ := CreateFakeClient()
//
//	// Create LocalVolumeGroup
//	lvg := GenFakeLocalVolumeGroupObject()
//	lvg.Name = fakeLocalVolumeGroupName
//	lvg.Namespace = fakeNamespace
//	err := client.Create(context.Background(), lvg)
//	if err != nil {
//		t.Logf("Create LocalVolumeGroup fail %v", err)
//	}
//
//	type args struct {
//		ns      string
//		lvgName string
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		want    *apisv1alpha1.LocalVolumeGroup
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//		{
//			args: args{
//				ns:      fakeNamespace,
//				lvgName: fakeLocalVolumeGroupName,
//			},
//			wantErr: false,
//			want:    lvg,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			m := &manager{
//				nameSpace:                 fakeNamespace,
//				apiClient:                 client,
//				localVolumeGroupQueue:     common.NewTaskQueue("localVolumeGroup", maxRetries),
//				localVolumeQueue:          common.NewTaskQueue("localVolume", maxRetries),
//				pvcQueue:                  common.NewTaskQueue("pvc", maxRetries),
//				podQueue:                  common.NewTaskQueue("pod", maxRetries),
//				localVolumeToVolumeGroups: make(map[string]string),
//				pvcToVolumeGroups:         make(map[string]string),
//				podToVolumeGroups:         make(map[string]string),
//				logger:                    log.WithField("Module", "ControllerManager"),
//			}
//			got, err := m.GetLocalVolumeGroupByName(tt.args.ns, tt.args.lvgName)
//			if (err != nil) != tt.wantErr {
//				t.Logf("GetLocalVolumeGroupByName() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(got.Name, tt.want.Name) {
//				t.Logf("GetLocalVolumeGroupByName() got = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func Test_manager_GetLocalVolumeGroupByPVC(t *testing.T) {
//	type fields struct {
//		apiClient                 client.Client
//		informersCache            cache.Cache
//		logger                    *log.Entry
//		nameSpace                 string
//		lock                      sync.Mutex
//		localVolumeGroupQueue     *common.TaskQueue
//		localVolumeQueue          *common.TaskQueue
//		pvcQueue                  *common.TaskQueue
//		podQueue                  *common.TaskQueue
//		localVolumeToVolumeGroups map[string]string
//		pvcToVolumeGroups         map[string]string
//		podToVolumeGroups         map[string]string
//	}
//
//	client, _ := CreateFakeClient()
//
//	// Create LocalVolumeGroup
//	lvg := GenFakeLocalVolumeGroupObject()
//	lvg.Name = fakeLocalVolumeGroupName
//	lvg.Namespace = fakeNamespace
//	err := client.Create(context.Background(), lvg)
//	if err != nil {
//		t.Logf("Create LocalVolumeGroup fail %v", err)
//	}
//
//	type args struct {
//		pvcNamespace string
//		pvcName      string
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		want    *apisv1alpha1.LocalVolumeGroup
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//		{
//			args: args{
//				pvcNamespace: fakeNamespace,
//				pvcName:      fakePersistentPvcName,
//			},
//			wantErr: true,
//			want:    nil,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			m := &manager{
//				nameSpace:                 fakeNamespace,
//				apiClient:                 client,
//				localVolumeGroupQueue:     common.NewTaskQueue("localVolumeGroup", maxRetries),
//				localVolumeQueue:          common.NewTaskQueue("localVolume", maxRetries),
//				pvcQueue:                  common.NewTaskQueue("pvc", maxRetries),
//				podQueue:                  common.NewTaskQueue("pod", maxRetries),
//				localVolumeToVolumeGroups: make(map[string]string),
//				pvcToVolumeGroups:         make(map[string]string),
//				podToVolumeGroups:         make(map[string]string),
//				logger:                    log.WithField("Module", "ControllerManager"),
//			}
//			got, err := m.GetLocalVolumeGroupByPVC(tt.args.pvcNamespace, tt.args.pvcName)
//			if (err != nil) != tt.wantErr {
//				t.Logf("GetLocalVolumeGroupByPVC() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(got, tt.want) {
//				t.Logf("GetLocalVolumeGroupByPVC() got = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}

func Test_manager_Init(t *testing.T) {
	type fields struct {
		apiClient                 client.Client
		informersCache            cache.Cache
		logger                    *log.Entry
		nameSpace                 string
		lock                      sync.Mutex
		localVolumeGroupQueue     *common.TaskQueue
		localVolumeQueue          *common.TaskQueue
		pvcQueue                  *common.TaskQueue
		podQueue                  *common.TaskQueue
		localVolumeToVolumeGroups map[string]string
		pvcToVolumeGroups         map[string]string
		podToVolumeGroups         map[string]string
	}
	var stopCh <-chan struct{}

	client, _ := CreateFakeClient()

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = fakeNamespace
	err := client.Create(context.Background(), lvg)
	if err != nil {
		t.Logf("Create LocalVolumeGroup fail %v", err)
	}

	mgr := CreateFakeMgr()

	type args struct {
		stopCh <-chan struct{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{
			args: args{
				stopCh: stopCh,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = &manager{
				nameSpace:                 fakeNamespace,
				apiClient:                 client,
				informersCache:            mgr.GetCache(),
				localVolumeGroupQueue:     common.NewTaskQueue("localVolumeGroup", maxRetries),
				localVolumeQueue:          common.NewTaskQueue("localVolume", maxRetries),
				pvcQueue:                  common.NewTaskQueue("pvc", maxRetries),
				podQueue:                  common.NewTaskQueue("pod", maxRetries),
				localVolumeToVolumeGroups: make(map[string]string),
				pvcToVolumeGroups:         make(map[string]string),
				podToVolumeGroups:         make(map[string]string),
				logger:                    log.WithField("Module", "ControllerManager"),
			}
			t.Logf("Init Debug ()")

			//m.Init(tt.args.stopCh)
		})
	}
}

//func Test_manager_ReconcileVolumeGroup(t *testing.T) {
//	type fields struct {
//		apiClient                 client.Client
//		informersCache            cache.Cache
//		logger                    *log.Entry
//		nameSpace                 string
//		lock                      sync.Mutex
//		localVolumeGroupQueue     *common.TaskQueue
//		localVolumeQueue          *common.TaskQueue
//		pvcQueue                  *common.TaskQueue
//		podQueue                  *common.TaskQueue
//		localVolumeToVolumeGroups map[string]string
//		pvcToVolumeGroups         map[string]string
//		podToVolumeGroups         map[string]string
//	}
//
//	client, _ := CreateFakeClient()
//
//	// Create LocalVolumeGroup
//	lvg := GenFakeLocalVolumeGroupObject()
//	lvg.Name = fakeLocalVolumeGroupName
//	lvg.Namespace = fakeNamespace
//	err := client.Create(context.Background(), lvg)
//	if err != nil {
//		t.Logf("Create LocalVolumeGroup fail %v", err)
//	}
//
//	type args struct {
//		lvg *apisv1alpha1.LocalVolumeGroup
//	}
//	tests := []struct {
//		name   string
//		fields fields
//		args   args
//	}{
//		// TODO: Add test cases.
//		{
//			args: args{
//				lvg: lvg,
//			},
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			_ = &manager{
//				nameSpace:                 fakeNamespace,
//				apiClient:                 client,
//				localVolumeGroupQueue:     common.NewTaskQueue("localVolumeGroup", maxRetries),
//				localVolumeQueue:          common.NewTaskQueue("localVolume", maxRetries),
//				pvcQueue:                  common.NewTaskQueue("pvc", maxRetries),
//				podQueue:                  common.NewTaskQueue("pod", maxRetries),
//				localVolumeToVolumeGroups: make(map[string]string),
//				pvcToVolumeGroups:         make(map[string]string),
//				podToVolumeGroups:         make(map[string]string),
//				logger:                    log.WithField("Module", "ControllerManager"),
//			}
//			//m.ReconcileVolumeGroup(tt.args.lvg)
//		})
//	}
//}

//func Test_manager_addLocalVolume(t *testing.T) {
//	type fields struct {
//		apiClient                 client.Client
//		informersCache            cache.Cache
//		logger                    *log.Entry
//		nameSpace                 string
//		lock                      sync.Mutex
//		localVolumeGroupQueue     *common.TaskQueue
//		localVolumeQueue          *common.TaskQueue
//		pvcQueue                  *common.TaskQueue
//		podQueue                  *common.TaskQueue
//		localVolumeToVolumeGroups map[string]string
//		pvcToVolumeGroups         map[string]string
//		podToVolumeGroups         map[string]string
//	}
//
//	client, _ := CreateFakeClient()
//
//	// Create LocalVolumeGroup
//	lvg := GenFakeLocalVolumeGroupObject()
//	lvg.Name = fakeLocalVolumeGroupName
//	lvg.Namespace = fakeNamespace
//	err := client.Create(context.Background(), lvg)
//	if err != nil {
//		t.Logf("Create LocalVolumeGroup fail %v", err)
//	}
//
//	mgr := CreateFakeMgr()
//
//	lv := GenFakeLocalVolumeObject()
//
//	type args struct {
//		lv *apisv1alpha1.LocalVolume
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//		{
//			args: args{
//				lv: lv,
//			},
//			wantErr: true,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			m := &manager{
//				nameSpace:                 fakeNamespace,
//				apiClient:                 client,
//				informersCache:            mgr.GetCache(),
//				localVolumeGroupQueue:     common.NewTaskQueue("localVolumeGroup", maxRetries),
//				localVolumeQueue:          common.NewTaskQueue("localVolume", maxRetries),
//				pvcQueue:                  common.NewTaskQueue("pvc", maxRetries),
//				podQueue:                  common.NewTaskQueue("pod", maxRetries),
//				localVolumeToVolumeGroups: make(map[string]string),
//				pvcToVolumeGroups:         make(map[string]string),
//				podToVolumeGroups:         make(map[string]string),
//				logger:                    log.WithField("Module", "ControllerManager"),
//			}
//			if err := m.addLocalVolume(tt.args.lv); (err != nil) != tt.wantErr {
//				t.Logf("addLocalVolume() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
//}

//
//func Test_manager_addLocalVolumeGroup(t *testing.T) {
//	type fields struct {
//		apiClient                 client.Client
//		informersCache            cache.Cache
//		logger                    *log.Entry
//		nameSpace                 string
//		lock                      sync.Mutex
//		localVolumeGroupQueue     *common.TaskQueue
//		localVolumeQueue          *common.TaskQueue
//		pvcQueue                  *common.TaskQueue
//		podQueue                  *common.TaskQueue
//		localVolumeToVolumeGroups map[string]string
//		pvcToVolumeGroups         map[string]string
//		podToVolumeGroups         map[string]string
//	}
//
//	client, _ := CreateFakeClient()
//
//	// Create LocalVolumeGroup
//	lvg := GenFakeLocalVolumeGroupObject()
//	lvg.Name = fakeLocalVolumeGroupName
//	lvg.Namespace = fakeNamespace
//	err := client.Create(context.Background(), lvg)
//	if err != nil {
//		t.Logf("Create LocalVolumeGroup fail %v", err)
//	}
//
//	type args struct {
//		lvg *apisv1alpha1.LocalVolumeGroup
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//		{
//			args: args{
//				lvg: lvg,
//			},
//			wantErr: false,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			m := &manager{
//				nameSpace:                 fakeNamespace,
//				apiClient:                 client,
//				localVolumeGroupQueue:     common.NewTaskQueue("localVolumeGroup", maxRetries),
//				localVolumeQueue:          common.NewTaskQueue("localVolume", maxRetries),
//				pvcQueue:                  common.NewTaskQueue("pvc", maxRetries),
//				podQueue:                  common.NewTaskQueue("pod", maxRetries),
//				localVolumeToVolumeGroups: make(map[string]string),
//				pvcToVolumeGroups:         make(map[string]string),
//				podToVolumeGroups:         make(map[string]string),
//				logger:                    log.WithField("Module", "ControllerManager"),
//			}
//			if err := m.addLocalVolumeGroup(tt.args.lvg); (err != nil) != tt.wantErr {
//				t.Logf("addLocalVolumeGroup() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
//}

func Test_manager_addPVC(t *testing.T) {
	type fields struct {
		apiClient                 client.Client
		informersCache            cache.Cache
		logger                    *log.Entry
		nameSpace                 string
		lock                      sync.Mutex
		localVolumeGroupQueue     *common.TaskQueue
		localVolumeQueue          *common.TaskQueue
		pvcQueue                  *common.TaskQueue
		podQueue                  *common.TaskQueue
		localVolumeToVolumeGroups map[string]string
		pvcToVolumeGroups         map[string]string
		podToVolumeGroups         map[string]string
	}
	type args struct {
		pvc *corev1.PersistentVolumeClaim
	}

	client, _ := CreateFakeClient()

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = fakeNamespace
	err := client.Create(context.Background(), lvg)
	if err != nil {
		t.Logf("Create LocalVolumeGroup fail %v", err)
	}

	pvc := GenFakePVCObject()

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args: args{
				pvc: pvc,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				nameSpace:                 fakeNamespace,
				apiClient:                 client,
				localVolumeGroupQueue:     common.NewTaskQueue("localVolumeGroup", maxRetries),
				localVolumeQueue:          common.NewTaskQueue("localVolume", maxRetries),
				pvcQueue:                  common.NewTaskQueue("pvc", maxRetries),
				podQueue:                  common.NewTaskQueue("pod", maxRetries),
				localVolumeToVolumeGroups: make(map[string]string),
				pvcToVolumeGroups:         make(map[string]string),
				podToVolumeGroups:         make(map[string]string),
				logger:                    log.WithField("Module", "ControllerManager"),
			}
			if err := m.addPVC(tt.args.pvc); (err != nil) != tt.wantErr {
				t.Logf("addPVC() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_addPod(t *testing.T) {
	type fields struct {
		apiClient                 client.Client
		informersCache            cache.Cache
		logger                    *log.Entry
		nameSpace                 string
		lock                      sync.Mutex
		localVolumeGroupQueue     *common.TaskQueue
		localVolumeQueue          *common.TaskQueue
		pvcQueue                  *common.TaskQueue
		podQueue                  *common.TaskQueue
		localVolumeToVolumeGroups map[string]string
		pvcToVolumeGroups         map[string]string
		podToVolumeGroups         map[string]string
	}
	type args struct {
		pod *corev1.Pod
	}

	client, _ := CreateFakeClient()

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = fakeNamespace
	err := client.Create(context.Background(), lvg)
	if err != nil {
		t.Logf("Create LocalVolumeGroup fail %v", err)
	}

	pod := GenFakePodObject()

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args: args{
				pod: pod,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				nameSpace:                 fakeNamespace,
				apiClient:                 client,
				localVolumeGroupQueue:     common.NewTaskQueue("localVolumeGroup", maxRetries),
				localVolumeQueue:          common.NewTaskQueue("localVolume", maxRetries),
				pvcQueue:                  common.NewTaskQueue("pvc", maxRetries),
				podQueue:                  common.NewTaskQueue("pod", maxRetries),
				localVolumeToVolumeGroups: make(map[string]string),
				pvcToVolumeGroups:         make(map[string]string),
				podToVolumeGroups:         make(map[string]string),
				logger:                    log.WithField("Module", "ControllerManager"),
			}
			if err := m.addPod(tt.args.pod); (err != nil) != tt.wantErr {
				t.Logf("addPod() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_cleanCacheForLocalVolume(t *testing.T) {
	type fields struct {
		apiClient                 client.Client
		informersCache            cache.Cache
		logger                    *log.Entry
		nameSpace                 string
		lock                      sync.Mutex
		localVolumeGroupQueue     *common.TaskQueue
		localVolumeQueue          *common.TaskQueue
		pvcQueue                  *common.TaskQueue
		podQueue                  *common.TaskQueue
		localVolumeToVolumeGroups map[string]string
		pvcToVolumeGroups         map[string]string
		podToVolumeGroups         map[string]string
	}
	type args struct {
		name string
	}

	client, _ := CreateFakeClient()

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = fakeNamespace
	err := client.Create(context.Background(), lvg)
	if err != nil {
		t.Logf("Create LocalVolumeGroup fail %v", err)
	}

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{
			args: args{
				name: fakeLocalVolumeName,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				nameSpace:                 fakeNamespace,
				apiClient:                 client,
				localVolumeGroupQueue:     common.NewTaskQueue("localVolumeGroup", maxRetries),
				localVolumeQueue:          common.NewTaskQueue("localVolume", maxRetries),
				pvcQueue:                  common.NewTaskQueue("pvc", maxRetries),
				podQueue:                  common.NewTaskQueue("pod", maxRetries),
				localVolumeToVolumeGroups: make(map[string]string),
				pvcToVolumeGroups:         make(map[string]string),
				podToVolumeGroups:         make(map[string]string),
				logger:                    log.WithField("Module", "ControllerManager"),
			}
			m.cleanCacheForLocalVolume(tt.args.name)
		})
	}
}

func Test_manager_cleanCacheForLocalVolumeGroup(t *testing.T) {
	type fields struct {
		apiClient                 client.Client
		informersCache            cache.Cache
		logger                    *log.Entry
		nameSpace                 string
		lock                      sync.Mutex
		localVolumeGroupQueue     *common.TaskQueue
		localVolumeQueue          *common.TaskQueue
		pvcQueue                  *common.TaskQueue
		podQueue                  *common.TaskQueue
		localVolumeToVolumeGroups map[string]string
		pvcToVolumeGroups         map[string]string
		podToVolumeGroups         map[string]string
	}
	type args struct {
		name string
	}

	client, _ := CreateFakeClient()

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = fakeNamespace
	err := client.Create(context.Background(), lvg)
	if err != nil {
		t.Logf("Create LocalVolumeGroup fail %v", err)
	}

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{
			args: args{
				name: fakeLocalVolumeGroupName,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				nameSpace:                 fakeNamespace,
				apiClient:                 client,
				localVolumeGroupQueue:     common.NewTaskQueue("localVolumeGroup", maxRetries),
				localVolumeQueue:          common.NewTaskQueue("localVolume", maxRetries),
				pvcQueue:                  common.NewTaskQueue("pvc", maxRetries),
				podQueue:                  common.NewTaskQueue("pod", maxRetries),
				localVolumeToVolumeGroups: make(map[string]string),
				pvcToVolumeGroups:         make(map[string]string),
				podToVolumeGroups:         make(map[string]string),
				logger:                    log.WithField("Module", "ControllerManager"),
			}
			m.cleanCacheForLocalVolumeGroup(tt.args.name)
		})
	}
}

func Test_manager_cleanCacheForPVC(t *testing.T) {
	type fields struct {
		apiClient                 client.Client
		informersCache            cache.Cache
		logger                    *log.Entry
		nameSpace                 string
		lock                      sync.Mutex
		localVolumeGroupQueue     *common.TaskQueue
		localVolumeQueue          *common.TaskQueue
		pvcQueue                  *common.TaskQueue
		podQueue                  *common.TaskQueue
		localVolumeToVolumeGroups map[string]string
		pvcToVolumeGroups         map[string]string
		podToVolumeGroups         map[string]string
	}
	type args struct {
		namespace string
		name      string
	}

	client, _ := CreateFakeClient()

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = fakeNamespace
	err := client.Create(context.Background(), lvg)
	if err != nil {
		t.Logf("Create LocalVolumeGroup fail %v", err)
	}

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{
			args: args{
				namespace: fakeNamespace,
				name:      fakePersistentPvcName,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				nameSpace:                 fakeNamespace,
				apiClient:                 client,
				localVolumeGroupQueue:     common.NewTaskQueue("localVolumeGroup", maxRetries),
				localVolumeQueue:          common.NewTaskQueue("localVolume", maxRetries),
				pvcQueue:                  common.NewTaskQueue("pvc", maxRetries),
				podQueue:                  common.NewTaskQueue("pod", maxRetries),
				localVolumeToVolumeGroups: make(map[string]string),
				pvcToVolumeGroups:         make(map[string]string),
				podToVolumeGroups:         make(map[string]string),
				logger:                    log.WithField("Module", "ControllerManager"),
			}
			m.cleanCacheForPVC(tt.args.namespace, tt.args.name)
		})
	}
}

func Test_manager_debug(t *testing.T) {
	type fields struct {
		apiClient                 client.Client
		informersCache            cache.Cache
		logger                    *log.Entry
		nameSpace                 string
		lock                      sync.Mutex
		localVolumeGroupQueue     *common.TaskQueue
		localVolumeQueue          *common.TaskQueue
		pvcQueue                  *common.TaskQueue
		podQueue                  *common.TaskQueue
		localVolumeToVolumeGroups map[string]string
		pvcToVolumeGroups         map[string]string
		podToVolumeGroups         map[string]string
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				apiClient:                 tt.fields.apiClient,
				informersCache:            tt.fields.informersCache,
				logger:                    tt.fields.logger,
				nameSpace:                 tt.fields.nameSpace,
				lock:                      tt.fields.lock,
				localVolumeGroupQueue:     tt.fields.localVolumeGroupQueue,
				localVolumeQueue:          tt.fields.localVolumeQueue,
				pvcQueue:                  tt.fields.pvcQueue,
				podQueue:                  tt.fields.podQueue,
				localVolumeToVolumeGroups: tt.fields.localVolumeToVolumeGroups,
				pvcToVolumeGroups:         tt.fields.pvcToVolumeGroups,
				podToVolumeGroups:         tt.fields.podToVolumeGroups,
			}
			m.debug()
		})
	}
}

func Test_manager_deleteLocalVolume(t *testing.T) {
	type fields struct {
		apiClient                 client.Client
		informersCache            cache.Cache
		logger                    *log.Entry
		nameSpace                 string
		lock                      sync.Mutex
		localVolumeGroupQueue     *common.TaskQueue
		localVolumeQueue          *common.TaskQueue
		pvcQueue                  *common.TaskQueue
		podQueue                  *common.TaskQueue
		localVolumeToVolumeGroups map[string]string
		pvcToVolumeGroups         map[string]string
		podToVolumeGroups         map[string]string
	}
	type args struct {
		lvName string
	}

	client, _ := CreateFakeClient()

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = fakeNamespace
	err := client.Create(context.Background(), lvg)
	if err != nil {
		t.Logf("Create LocalVolumeGroup fail %v", err)
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args: args{
				lvName: fakeLocalVolumeName,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				nameSpace:                 fakeNamespace,
				apiClient:                 client,
				localVolumeGroupQueue:     common.NewTaskQueue("localVolumeGroup", maxRetries),
				localVolumeQueue:          common.NewTaskQueue("localVolume", maxRetries),
				pvcQueue:                  common.NewTaskQueue("pvc", maxRetries),
				podQueue:                  common.NewTaskQueue("pod", maxRetries),
				localVolumeToVolumeGroups: make(map[string]string),
				pvcToVolumeGroups:         make(map[string]string),
				podToVolumeGroups:         make(map[string]string),
				logger:                    log.WithField("Module", "ControllerManager"),
			}
			if err := m.deleteLocalVolume(tt.args.lvName); (err != nil) != tt.wantErr {
				t.Logf("deleteLocalVolume() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_deleteLocalVolumeGroup(t *testing.T) {
	type fields struct {
		apiClient                 client.Client
		informersCache            cache.Cache
		logger                    *log.Entry
		nameSpace                 string
		lock                      sync.Mutex
		localVolumeGroupQueue     *common.TaskQueue
		localVolumeQueue          *common.TaskQueue
		pvcQueue                  *common.TaskQueue
		podQueue                  *common.TaskQueue
		localVolumeToVolumeGroups map[string]string
		pvcToVolumeGroups         map[string]string
		podToVolumeGroups         map[string]string
	}
	type args struct {
		lvg *apisv1alpha1.LocalVolumeGroup
	}

	client, _ := CreateFakeClient()

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = fakeNamespace
	err := client.Create(context.Background(), lvg)
	if err != nil {
		t.Logf("Create LocalVolumeGroup fail %v", err)
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args: args{
				lvg: lvg,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				nameSpace:                 fakeNamespace,
				apiClient:                 client,
				localVolumeGroupQueue:     common.NewTaskQueue("localVolumeGroup", maxRetries),
				localVolumeQueue:          common.NewTaskQueue("localVolume", maxRetries),
				pvcQueue:                  common.NewTaskQueue("pvc", maxRetries),
				podQueue:                  common.NewTaskQueue("pod", maxRetries),
				localVolumeToVolumeGroups: make(map[string]string),
				pvcToVolumeGroups:         make(map[string]string),
				podToVolumeGroups:         make(map[string]string),
				logger:                    log.WithField("Module", "ControllerManager"),
			}
			if err := m.deleteLocalVolumeGroup(tt.args.lvg); (err != nil) != tt.wantErr {
				t.Logf("deleteLocalVolumeGroup() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_deletePVC(t *testing.T) {
	type fields struct {
		apiClient                 client.Client
		informersCache            cache.Cache
		logger                    *log.Entry
		nameSpace                 string
		lock                      sync.Mutex
		localVolumeGroupQueue     *common.TaskQueue
		localVolumeQueue          *common.TaskQueue
		pvcQueue                  *common.TaskQueue
		podQueue                  *common.TaskQueue
		localVolumeToVolumeGroups map[string]string
		pvcToVolumeGroups         map[string]string
		podToVolumeGroups         map[string]string
	}
	type args struct {
		namespace string
		name      string
	}

	client, _ := CreateFakeClient()

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = fakeNamespace
	err := client.Create(context.Background(), lvg)
	if err != nil {
		t.Logf("Create LocalVolumeGroup fail %v", err)
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args: args{
				namespace: fakeNamespace,
				name:      fakePersistentPvcName,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				nameSpace:                 fakeNamespace,
				apiClient:                 client,
				localVolumeGroupQueue:     common.NewTaskQueue("localVolumeGroup", maxRetries),
				localVolumeQueue:          common.NewTaskQueue("localVolume", maxRetries),
				pvcQueue:                  common.NewTaskQueue("pvc", maxRetries),
				podQueue:                  common.NewTaskQueue("pod", maxRetries),
				localVolumeToVolumeGroups: make(map[string]string),
				pvcToVolumeGroups:         make(map[string]string),
				podToVolumeGroups:         make(map[string]string),
				logger:                    log.WithField("Module", "ControllerManager"),
			}
			if err := m.deletePVC(tt.args.namespace, tt.args.name); (err != nil) != tt.wantErr {
				t.Logf("deletePVC() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_deletePod(t *testing.T) {
	type fields struct {
		apiClient                 client.Client
		informersCache            cache.Cache
		logger                    *log.Entry
		nameSpace                 string
		lock                      sync.Mutex
		localVolumeGroupQueue     *common.TaskQueue
		localVolumeQueue          *common.TaskQueue
		pvcQueue                  *common.TaskQueue
		podQueue                  *common.TaskQueue
		localVolumeToVolumeGroups map[string]string
		pvcToVolumeGroups         map[string]string
		podToVolumeGroups         map[string]string
	}
	type args struct {
		namespace string
		name      string
	}

	client, _ := CreateFakeClient()

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = fakeNamespace
	err := client.Create(context.Background(), lvg)
	if err != nil {
		t.Logf("Create LocalVolumeGroup fail %v", err)
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args: args{
				namespace: fakeNamespace,
				name:      fakePodName,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				nameSpace:                 fakeNamespace,
				apiClient:                 client,
				localVolumeGroupQueue:     common.NewTaskQueue("localVolumeGroup", maxRetries),
				localVolumeQueue:          common.NewTaskQueue("localVolume", maxRetries),
				pvcQueue:                  common.NewTaskQueue("pvc", maxRetries),
				podQueue:                  common.NewTaskQueue("pod", maxRetries),
				localVolumeToVolumeGroups: make(map[string]string),
				pvcToVolumeGroups:         make(map[string]string),
				podToVolumeGroups:         make(map[string]string),
				logger:                    log.WithField("Module", "ControllerManager"),
			}
			if err := m.deletePod(tt.args.namespace, tt.args.name); (err != nil) != tt.wantErr {
				t.Logf("deletePod() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_handleLocalVolumeEventAdd(t *testing.T) {
	type fields struct {
		apiClient                 client.Client
		informersCache            cache.Cache
		logger                    *log.Entry
		nameSpace                 string
		lock                      sync.Mutex
		localVolumeGroupQueue     *common.TaskQueue
		localVolumeQueue          *common.TaskQueue
		pvcQueue                  *common.TaskQueue
		podQueue                  *common.TaskQueue
		localVolumeToVolumeGroups map[string]string
		pvcToVolumeGroups         map[string]string
		podToVolumeGroups         map[string]string
	}
	type args struct {
		obj interface{}
	}

	client, _ := CreateFakeClient()

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = fakeNamespace
	err := client.Create(context.Background(), lvg)
	if err != nil {
		t.Logf("Create LocalVolumeGroup fail %v", err)
	}

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{
			args: args{
				obj: GenFakeLocalVolumeObject(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				nameSpace:                 fakeNamespace,
				apiClient:                 client,
				localVolumeGroupQueue:     common.NewTaskQueue("localVolumeGroup", maxRetries),
				localVolumeQueue:          common.NewTaskQueue("localVolume", maxRetries),
				pvcQueue:                  common.NewTaskQueue("pvc", maxRetries),
				podQueue:                  common.NewTaskQueue("pod", maxRetries),
				localVolumeToVolumeGroups: make(map[string]string),
				pvcToVolumeGroups:         make(map[string]string),
				podToVolumeGroups:         make(map[string]string),
				logger:                    log.WithField("Module", "ControllerManager"),
			}
			m.handleLocalVolumeEventAdd(tt.args.obj)
		})
	}
}

func Test_manager_handleLocalVolumeEventDelete(t *testing.T) {
	type fields struct {
		apiClient                 client.Client
		informersCache            cache.Cache
		logger                    *log.Entry
		nameSpace                 string
		lock                      sync.Mutex
		localVolumeGroupQueue     *common.TaskQueue
		localVolumeQueue          *common.TaskQueue
		pvcQueue                  *common.TaskQueue
		podQueue                  *common.TaskQueue
		localVolumeToVolumeGroups map[string]string
		pvcToVolumeGroups         map[string]string
		podToVolumeGroups         map[string]string
	}
	type args struct {
		obj interface{}
	}

	client, _ := CreateFakeClient()

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = fakeNamespace
	err := client.Create(context.Background(), lvg)
	if err != nil {
		t.Logf("Create LocalVolumeGroup fail %v", err)
	}

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{
			args: args{
				obj: GenFakeLocalVolumeObject(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				nameSpace:                 fakeNamespace,
				apiClient:                 client,
				localVolumeGroupQueue:     common.NewTaskQueue("localVolumeGroup", maxRetries),
				localVolumeQueue:          common.NewTaskQueue("localVolume", maxRetries),
				pvcQueue:                  common.NewTaskQueue("pvc", maxRetries),
				podQueue:                  common.NewTaskQueue("pod", maxRetries),
				localVolumeToVolumeGroups: make(map[string]string),
				pvcToVolumeGroups:         make(map[string]string),
				podToVolumeGroups:         make(map[string]string),
				logger:                    log.WithField("Module", "ControllerManager"),
			}
			m.handleLocalVolumeEventDelete(tt.args.obj)
		})
	}
}

func Test_manager_handleLocalVolumeEventUpdate(t *testing.T) {
	type fields struct {
		apiClient                 client.Client
		informersCache            cache.Cache
		logger                    *log.Entry
		nameSpace                 string
		lock                      sync.Mutex
		localVolumeGroupQueue     *common.TaskQueue
		localVolumeQueue          *common.TaskQueue
		pvcQueue                  *common.TaskQueue
		podQueue                  *common.TaskQueue
		localVolumeToVolumeGroups map[string]string
		pvcToVolumeGroups         map[string]string
		podToVolumeGroups         map[string]string
	}
	type args struct {
		oldObj interface{}
		newObj interface{}
	}

	client, _ := CreateFakeClient()

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = fakeNamespace
	err := client.Create(context.Background(), lvg)
	if err != nil {
		t.Logf("Create LocalVolumeGroup fail %v", err)
	}

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{
			args: args{
				oldObj: GenFakeLocalVolumeObject(),
				newObj: GenFakeLocalVolumeObject(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				nameSpace:                 fakeNamespace,
				apiClient:                 client,
				localVolumeGroupQueue:     common.NewTaskQueue("localVolumeGroup", maxRetries),
				localVolumeQueue:          common.NewTaskQueue("localVolume", maxRetries),
				pvcQueue:                  common.NewTaskQueue("pvc", maxRetries),
				podQueue:                  common.NewTaskQueue("pod", maxRetries),
				localVolumeToVolumeGroups: make(map[string]string),
				pvcToVolumeGroups:         make(map[string]string),
				podToVolumeGroups:         make(map[string]string),
				logger:                    log.WithField("Module", "ControllerManager"),
			}
			m.handleLocalVolumeEventUpdate(tt.args.oldObj, tt.args.newObj)
		})
	}
}

func Test_manager_handlePVCEventAdd(t *testing.T) {
	type fields struct {
		apiClient                 client.Client
		informersCache            cache.Cache
		logger                    *log.Entry
		nameSpace                 string
		lock                      sync.Mutex
		localVolumeGroupQueue     *common.TaskQueue
		localVolumeQueue          *common.TaskQueue
		pvcQueue                  *common.TaskQueue
		podQueue                  *common.TaskQueue
		localVolumeToVolumeGroups map[string]string
		pvcToVolumeGroups         map[string]string
		podToVolumeGroups         map[string]string
	}
	type args struct {
		obj interface{}
	}

	client, _ := CreateFakeClient()

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = fakeNamespace
	err := client.Create(context.Background(), lvg)
	if err != nil {
		t.Logf("Create LocalVolumeGroup fail %v", err)
	}

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{
			args: args{
				obj: GenFakePVCObject(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				nameSpace:                 fakeNamespace,
				apiClient:                 client,
				localVolumeGroupQueue:     common.NewTaskQueue("localVolumeGroup", maxRetries),
				localVolumeQueue:          common.NewTaskQueue("localVolume", maxRetries),
				pvcQueue:                  common.NewTaskQueue("pvc", maxRetries),
				podQueue:                  common.NewTaskQueue("pod", maxRetries),
				localVolumeToVolumeGroups: make(map[string]string),
				pvcToVolumeGroups:         make(map[string]string),
				podToVolumeGroups:         make(map[string]string),
				logger:                    log.WithField("Module", "ControllerManager"),
			}
			m.handlePVCEventAdd(tt.args.obj)
		})
	}
}

func Test_manager_handlePVCEventDelete(t *testing.T) {
	type fields struct {
		apiClient                 client.Client
		informersCache            cache.Cache
		logger                    *log.Entry
		nameSpace                 string
		lock                      sync.Mutex
		localVolumeGroupQueue     *common.TaskQueue
		localVolumeQueue          *common.TaskQueue
		pvcQueue                  *common.TaskQueue
		podQueue                  *common.TaskQueue
		localVolumeToVolumeGroups map[string]string
		pvcToVolumeGroups         map[string]string
		podToVolumeGroups         map[string]string
	}
	type args struct {
		obj interface{}
	}

	client, _ := CreateFakeClient()

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = fakeNamespace
	err := client.Create(context.Background(), lvg)
	if err != nil {
		t.Logf("Create LocalVolumeGroup fail %v", err)
	}

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{
			args: args{
				obj: GenFakePVCObject(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				nameSpace:                 fakeNamespace,
				apiClient:                 client,
				localVolumeGroupQueue:     common.NewTaskQueue("localVolumeGroup", maxRetries),
				localVolumeQueue:          common.NewTaskQueue("localVolume", maxRetries),
				pvcQueue:                  common.NewTaskQueue("pvc", maxRetries),
				podQueue:                  common.NewTaskQueue("pod", maxRetries),
				localVolumeToVolumeGroups: make(map[string]string),
				pvcToVolumeGroups:         make(map[string]string),
				podToVolumeGroups:         make(map[string]string),
				logger:                    log.WithField("Module", "ControllerManager"),
			}
			m.handlePVCEventDelete(tt.args.obj)
		})
	}
}

func Test_manager_handlePodEventAdd(t *testing.T) {
	type fields struct {
		apiClient                 client.Client
		informersCache            cache.Cache
		logger                    *log.Entry
		nameSpace                 string
		lock                      sync.Mutex
		localVolumeGroupQueue     *common.TaskQueue
		localVolumeQueue          *common.TaskQueue
		pvcQueue                  *common.TaskQueue
		podQueue                  *common.TaskQueue
		localVolumeToVolumeGroups map[string]string
		pvcToVolumeGroups         map[string]string
		podToVolumeGroups         map[string]string
	}
	type args struct {
		obj interface{}
	}

	client, _ := CreateFakeClient()

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = fakeNamespace
	err := client.Create(context.Background(), lvg)
	if err != nil {
		t.Logf("Create LocalVolumeGroup fail %v", err)
	}

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{
			args: args{
				obj: GenFakePodObject(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				nameSpace:                 fakeNamespace,
				apiClient:                 client,
				localVolumeGroupQueue:     common.NewTaskQueue("localVolumeGroup", maxRetries),
				localVolumeQueue:          common.NewTaskQueue("localVolume", maxRetries),
				pvcQueue:                  common.NewTaskQueue("pvc", maxRetries),
				podQueue:                  common.NewTaskQueue("pod", maxRetries),
				localVolumeToVolumeGroups: make(map[string]string),
				pvcToVolumeGroups:         make(map[string]string),
				podToVolumeGroups:         make(map[string]string),
				logger:                    log.WithField("Module", "ControllerManager"),
			}
			m.handlePodEventAdd(tt.args.obj)
		})
	}
}

func Test_manager_handlePodEventDelete(t *testing.T) {
	type fields struct {
		apiClient                 client.Client
		informersCache            cache.Cache
		logger                    *log.Entry
		nameSpace                 string
		lock                      sync.Mutex
		localVolumeGroupQueue     *common.TaskQueue
		localVolumeQueue          *common.TaskQueue
		pvcQueue                  *common.TaskQueue
		podQueue                  *common.TaskQueue
		localVolumeToVolumeGroups map[string]string
		pvcToVolumeGroups         map[string]string
		podToVolumeGroups         map[string]string
	}
	type args struct {
		obj interface{}
	}

	client, _ := CreateFakeClient()

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = fakeNamespace
	err := client.Create(context.Background(), lvg)
	if err != nil {
		t.Logf("Create LocalVolumeGroup fail %v", err)
	}

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{
			args: args{
				obj: GenFakePodObject(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				nameSpace:                 fakeNamespace,
				apiClient:                 client,
				localVolumeGroupQueue:     common.NewTaskQueue("localVolumeGroup", maxRetries),
				localVolumeQueue:          common.NewTaskQueue("localVolume", maxRetries),
				pvcQueue:                  common.NewTaskQueue("pvc", maxRetries),
				podQueue:                  common.NewTaskQueue("pod", maxRetries),
				localVolumeToVolumeGroups: make(map[string]string),
				pvcToVolumeGroups:         make(map[string]string),
				podToVolumeGroups:         make(map[string]string),
				logger:                    log.WithField("Module", "ControllerManager"),
			}
			m.handlePodEventDelete(tt.args.obj)
		})
	}
}

func Test_manager_isHwameiStorPVC(t *testing.T) {
	type fields struct {
		apiClient                 client.Client
		informersCache            cache.Cache
		logger                    *log.Entry
		nameSpace                 string
		lock                      sync.Mutex
		localVolumeGroupQueue     *common.TaskQueue
		localVolumeQueue          *common.TaskQueue
		pvcQueue                  *common.TaskQueue
		podQueue                  *common.TaskQueue
		localVolumeToVolumeGroups map[string]string
		pvcToVolumeGroups         map[string]string
		podToVolumeGroups         map[string]string
	}
	type args struct {
		pvc *corev1.PersistentVolumeClaim
	}

	client, _ := CreateFakeClient()

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = fakeNamespace
	err := client.Create(context.Background(), lvg)
	if err != nil {
		t.Logf("Create LocalVolumeGroup fail %v", err)
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
		{
			args: args{
				pvc: GenFakePVCObject(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				nameSpace:                 fakeNamespace,
				apiClient:                 client,
				localVolumeGroupQueue:     common.NewTaskQueue("localVolumeGroup", maxRetries),
				localVolumeQueue:          common.NewTaskQueue("localVolume", maxRetries),
				pvcQueue:                  common.NewTaskQueue("pvc", maxRetries),
				podQueue:                  common.NewTaskQueue("pod", maxRetries),
				localVolumeToVolumeGroups: make(map[string]string),
				pvcToVolumeGroups:         make(map[string]string),
				podToVolumeGroups:         make(map[string]string),
				logger:                    log.WithField("Module", "ControllerManager"),
			}
			if got := m.isHwameiStorPVC(tt.args.pvc); got != tt.want {
				t.Logf("isHwameiStorPVC() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_manager_isHwameiStorPod(t *testing.T) {
	type fields struct {
		apiClient                 client.Client
		informersCache            cache.Cache
		logger                    *log.Entry
		nameSpace                 string
		lock                      sync.Mutex
		localVolumeGroupQueue     *common.TaskQueue
		localVolumeQueue          *common.TaskQueue
		pvcQueue                  *common.TaskQueue
		podQueue                  *common.TaskQueue
		localVolumeToVolumeGroups map[string]string
		pvcToVolumeGroups         map[string]string
		podToVolumeGroups         map[string]string
	}
	type args struct {
		pod *corev1.Pod
	}

	client, _ := CreateFakeClient()

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = fakeNamespace
	err := client.Create(context.Background(), lvg)
	if err != nil {
		t.Logf("Create LocalVolumeGroup fail %v", err)
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
		{
			args: args{
				pod: GenFakePodObject(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				nameSpace:                 fakeNamespace,
				apiClient:                 client,
				localVolumeGroupQueue:     common.NewTaskQueue("localVolumeGroup", maxRetries),
				localVolumeQueue:          common.NewTaskQueue("localVolume", maxRetries),
				pvcQueue:                  common.NewTaskQueue("pvc", maxRetries),
				podQueue:                  common.NewTaskQueue("pod", maxRetries),
				localVolumeToVolumeGroups: make(map[string]string),
				pvcToVolumeGroups:         make(map[string]string),
				podToVolumeGroups:         make(map[string]string),
				logger:                    log.WithField("Module", "ControllerManager"),
			}
			if got := m.isHwameiStorPod(tt.args.pod); got != tt.want {
				t.Logf("isHwameiStorPod() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_manager_processLocalVolume(t *testing.T) {
	type fields struct {
		apiClient                 client.Client
		informersCache            cache.Cache
		logger                    *log.Entry
		nameSpace                 string
		lock                      sync.Mutex
		localVolumeGroupQueue     *common.TaskQueue
		localVolumeQueue          *common.TaskQueue
		pvcQueue                  *common.TaskQueue
		podQueue                  *common.TaskQueue
		localVolumeToVolumeGroups map[string]string
		pvcToVolumeGroups         map[string]string
		podToVolumeGroups         map[string]string
	}
	type args struct {
		lvName string
	}

	client, _ := CreateFakeClient()

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = fakeNamespace
	err := client.Create(context.Background(), lvg)
	if err != nil {
		t.Logf("Create LocalVolumeGroup fail %v", err)
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args: args{
				lvName: fakeLocalVolumeName,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				nameSpace:                 fakeNamespace,
				apiClient:                 client,
				localVolumeGroupQueue:     common.NewTaskQueue("localVolumeGroup", maxRetries),
				localVolumeQueue:          common.NewTaskQueue("localVolume", maxRetries),
				pvcQueue:                  common.NewTaskQueue("pvc", maxRetries),
				podQueue:                  common.NewTaskQueue("pod", maxRetries),
				localVolumeToVolumeGroups: make(map[string]string),
				pvcToVolumeGroups:         make(map[string]string),
				podToVolumeGroups:         make(map[string]string),
				logger:                    log.WithField("Module", "ControllerManager"),
			}
			if err := m.processLocalVolume(tt.args.lvName); (err != nil) != tt.wantErr {
				t.Logf("processLocalVolume() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_processLocalVolumeGroup(t *testing.T) {
	type fields struct {
		apiClient                 client.Client
		informersCache            cache.Cache
		logger                    *log.Entry
		nameSpace                 string
		lock                      sync.Mutex
		localVolumeGroupQueue     *common.TaskQueue
		localVolumeQueue          *common.TaskQueue
		pvcQueue                  *common.TaskQueue
		podQueue                  *common.TaskQueue
		localVolumeToVolumeGroups map[string]string
		pvcToVolumeGroups         map[string]string
		podToVolumeGroups         map[string]string
	}
	type args struct {
		lvgNamespacedName string
	}

	client, _ := CreateFakeClient()

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = fakeNamespace
	err := client.Create(context.Background(), lvg)
	if err != nil {
		t.Logf("Create LocalVolumeGroup fail %v", err)
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args: args{
				lvgNamespacedName: fakeLocalVolumeGroupName,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				nameSpace:                 fakeNamespace,
				apiClient:                 client,
				localVolumeGroupQueue:     common.NewTaskQueue("localVolumeGroup", maxRetries),
				localVolumeQueue:          common.NewTaskQueue("localVolume", maxRetries),
				pvcQueue:                  common.NewTaskQueue("pvc", maxRetries),
				podQueue:                  common.NewTaskQueue("pod", maxRetries),
				localVolumeToVolumeGroups: make(map[string]string),
				pvcToVolumeGroups:         make(map[string]string),
				podToVolumeGroups:         make(map[string]string),
				logger:                    log.WithField("Module", "ControllerManager"),
			}
			if err := m.processLocalVolumeGroup(tt.args.lvgNamespacedName); (err != nil) != tt.wantErr {
				t.Logf("processLocalVolumeGroup() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_processPVC(t *testing.T) {
	type fields struct {
		apiClient                 client.Client
		informersCache            cache.Cache
		logger                    *log.Entry
		nameSpace                 string
		lock                      sync.Mutex
		localVolumeGroupQueue     *common.TaskQueue
		localVolumeQueue          *common.TaskQueue
		pvcQueue                  *common.TaskQueue
		podQueue                  *common.TaskQueue
		localVolumeToVolumeGroups map[string]string
		pvcToVolumeGroups         map[string]string
		podToVolumeGroups         map[string]string
	}
	type args struct {
		nn string
	}

	client, _ := CreateFakeClient()

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = fakeNamespace
	err := client.Create(context.Background(), lvg)
	if err != nil {
		t.Logf("Create LocalVolumeGroup fail %v", err)
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args: args{
				nn: fakeNamespace + "/" + fakePersistentPvcName,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				nameSpace:                 fakeNamespace,
				apiClient:                 client,
				localVolumeGroupQueue:     common.NewTaskQueue("localVolumeGroup", maxRetries),
				localVolumeQueue:          common.NewTaskQueue("localVolume", maxRetries),
				pvcQueue:                  common.NewTaskQueue("pvc", maxRetries),
				podQueue:                  common.NewTaskQueue("pod", maxRetries),
				localVolumeToVolumeGroups: make(map[string]string),
				pvcToVolumeGroups:         make(map[string]string),
				podToVolumeGroups:         make(map[string]string),
				logger:                    log.WithField("Module", "ControllerManager"),
			}
			if err := m.processPVC(tt.args.nn); (err != nil) != tt.wantErr {
				t.Logf("processPVC() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_processPod(t *testing.T) {
	type fields struct {
		apiClient                 client.Client
		informersCache            cache.Cache
		logger                    *log.Entry
		nameSpace                 string
		lock                      sync.Mutex
		localVolumeGroupQueue     *common.TaskQueue
		localVolumeQueue          *common.TaskQueue
		pvcQueue                  *common.TaskQueue
		podQueue                  *common.TaskQueue
		localVolumeToVolumeGroups map[string]string
		pvcToVolumeGroups         map[string]string
		podToVolumeGroups         map[string]string
	}
	type args struct {
		nn string
	}

	client, _ := CreateFakeClient()

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = fakeNamespace
	err := client.Create(context.Background(), lvg)
	if err != nil {
		t.Logf("Create LocalVolumeGroup fail %v", err)
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args: args{
				nn: fakeNamespace + "/" + fakePodName,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				nameSpace:                 fakeNamespace,
				apiClient:                 client,
				localVolumeGroupQueue:     common.NewTaskQueue("localVolumeGroup", maxRetries),
				localVolumeQueue:          common.NewTaskQueue("localVolume", maxRetries),
				pvcQueue:                  common.NewTaskQueue("pvc", maxRetries),
				podQueue:                  common.NewTaskQueue("pod", maxRetries),
				localVolumeToVolumeGroups: make(map[string]string),
				pvcToVolumeGroups:         make(map[string]string),
				podToVolumeGroups:         make(map[string]string),
				logger:                    log.WithField("Module", "ControllerManager"),
			}
			if err := m.processPod(tt.args.nn); (err != nil) != tt.wantErr {
				t.Logf("processPod() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_releaseLocalVolumeGroup(t *testing.T) {
	type fields struct {
		apiClient                 client.Client
		informersCache            cache.Cache
		logger                    *log.Entry
		nameSpace                 string
		lock                      sync.Mutex
		localVolumeGroupQueue     *common.TaskQueue
		localVolumeQueue          *common.TaskQueue
		pvcQueue                  *common.TaskQueue
		podQueue                  *common.TaskQueue
		localVolumeToVolumeGroups map[string]string
		pvcToVolumeGroups         map[string]string
		podToVolumeGroups         map[string]string
	}
	type args struct {
		lvg *apisv1alpha1.LocalVolumeGroup
	}

	client, _ := CreateFakeClient()

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = fakeNamespace
	err := client.Create(context.Background(), lvg)
	if err != nil {
		t.Logf("Create LocalVolumeGroup fail %v", err)
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args: args{
				lvg: lvg,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				nameSpace:                 fakeNamespace,
				apiClient:                 client,
				localVolumeGroupQueue:     common.NewTaskQueue("localVolumeGroup", maxRetries),
				localVolumeQueue:          common.NewTaskQueue("localVolume", maxRetries),
				pvcQueue:                  common.NewTaskQueue("pvc", maxRetries),
				podQueue:                  common.NewTaskQueue("pod", maxRetries),
				localVolumeToVolumeGroups: make(map[string]string),
				pvcToVolumeGroups:         make(map[string]string),
				podToVolumeGroups:         make(map[string]string),
				logger:                    log.WithField("Module", "ControllerManager"),
			}
			if err := m.releaseLocalVolumeGroup(tt.args.lvg); (err != nil) != tt.wantErr {
				t.Logf("releaseLocalVolumeGroup() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_startLocalVolumeGroupWorker(t *testing.T) {
	type fields struct {
		apiClient                 client.Client
		informersCache            cache.Cache
		logger                    *log.Entry
		nameSpace                 string
		lock                      sync.Mutex
		localVolumeGroupQueue     *common.TaskQueue
		localVolumeQueue          *common.TaskQueue
		pvcQueue                  *common.TaskQueue
		podQueue                  *common.TaskQueue
		localVolumeToVolumeGroups map[string]string
		pvcToVolumeGroups         map[string]string
		podToVolumeGroups         map[string]string
	}
	type args struct {
		stopCh <-chan struct{}
	}
	//var stopCh <-chan struct{}

	client, _ := CreateFakeClient()

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = fakeNamespace
	err := client.Create(context.Background(), lvg)
	if err != nil {
		t.Logf("Create LocalVolumeGroup fail %v", err)
	}

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		//{
		//	args: args{
		//		stopCh: stopCh,
		//	},
		//},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				nameSpace:                 fakeNamespace,
				apiClient:                 client,
				localVolumeGroupQueue:     common.NewTaskQueue("localVolumeGroup", maxRetries),
				localVolumeQueue:          common.NewTaskQueue("localVolume", maxRetries),
				pvcQueue:                  common.NewTaskQueue("pvc", maxRetries),
				podQueue:                  common.NewTaskQueue("pod", maxRetries),
				localVolumeToVolumeGroups: make(map[string]string),
				pvcToVolumeGroups:         make(map[string]string),
				podToVolumeGroups:         make(map[string]string),
				logger:                    log.WithField("Module", "ControllerManager"),
			}
			m.startLocalVolumeGroupWorker(tt.args.stopCh)
		})
	}
}

func Test_manager_startLocalVolumeWorker(t *testing.T) {
	type fields struct {
		apiClient                 client.Client
		informersCache            cache.Cache
		logger                    *log.Entry
		nameSpace                 string
		lock                      sync.Mutex
		localVolumeGroupQueue     *common.TaskQueue
		localVolumeQueue          *common.TaskQueue
		pvcQueue                  *common.TaskQueue
		podQueue                  *common.TaskQueue
		localVolumeToVolumeGroups map[string]string
		pvcToVolumeGroups         map[string]string
		podToVolumeGroups         map[string]string
	}
	type args struct {
		stopCh <-chan struct{}
	}

	//var stopCh <-chan struct{}

	client, _ := CreateFakeClient()

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = fakeNamespace
	err := client.Create(context.Background(), lvg)
	if err != nil {
		t.Logf("Create LocalVolumeGroup fail %v", err)
	}

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		//{
		//	args: args{
		//		stopCh: stopCh,
		//	},
		//},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				nameSpace:                 fakeNamespace,
				apiClient:                 client,
				localVolumeGroupQueue:     common.NewTaskQueue("localVolumeGroup", maxRetries),
				localVolumeQueue:          common.NewTaskQueue("localVolume", maxRetries),
				pvcQueue:                  common.NewTaskQueue("pvc", maxRetries),
				podQueue:                  common.NewTaskQueue("pod", maxRetries),
				localVolumeToVolumeGroups: make(map[string]string),
				pvcToVolumeGroups:         make(map[string]string),
				podToVolumeGroups:         make(map[string]string),
				logger:                    log.WithField("Module", "ControllerManager"),
			}
			m.startLocalVolumeWorker(tt.args.stopCh)
		})
	}
}

func Test_manager_startPVCWorker(t *testing.T) {
	type fields struct {
		apiClient                 client.Client
		informersCache            cache.Cache
		logger                    *log.Entry
		nameSpace                 string
		lock                      sync.Mutex
		localVolumeGroupQueue     *common.TaskQueue
		localVolumeQueue          *common.TaskQueue
		pvcQueue                  *common.TaskQueue
		podQueue                  *common.TaskQueue
		localVolumeToVolumeGroups map[string]string
		pvcToVolumeGroups         map[string]string
		podToVolumeGroups         map[string]string
	}
	type args struct {
		stopCh <-chan struct{}
	}

	//var stopCh <-chan struct{}

	client, _ := CreateFakeClient()

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = fakeNamespace
	err := client.Create(context.Background(), lvg)
	if err != nil {
		t.Logf("Create LocalVolumeGroup fail %v", err)
	}

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		//{
		//	args: args{
		//		stopCh: stopCh,
		//	},
		//},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				nameSpace:                 fakeNamespace,
				apiClient:                 client,
				localVolumeGroupQueue:     common.NewTaskQueue("localVolumeGroup", maxRetries),
				localVolumeQueue:          common.NewTaskQueue("localVolume", maxRetries),
				pvcQueue:                  common.NewTaskQueue("pvc", maxRetries),
				podQueue:                  common.NewTaskQueue("pod", maxRetries),
				localVolumeToVolumeGroups: make(map[string]string),
				pvcToVolumeGroups:         make(map[string]string),
				podToVolumeGroups:         make(map[string]string),
				logger:                    log.WithField("Module", "ControllerManager"),
			}
			m.startPVCWorker(tt.args.stopCh)
		})
	}
}

func Test_manager_startPodWorker(t *testing.T) {
	type fields struct {
		apiClient                 client.Client
		informersCache            cache.Cache
		logger                    *log.Entry
		nameSpace                 string
		lock                      sync.Mutex
		localVolumeGroupQueue     *common.TaskQueue
		localVolumeQueue          *common.TaskQueue
		pvcQueue                  *common.TaskQueue
		podQueue                  *common.TaskQueue
		localVolumeToVolumeGroups map[string]string
		pvcToVolumeGroups         map[string]string
		podToVolumeGroups         map[string]string
	}
	type args struct {
		stopCh <-chan struct{}
	}

	//var stopCh <-chan struct{}

	client, _ := CreateFakeClient()

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = fakeNamespace
	err := client.Create(context.Background(), lvg)
	if err != nil {
		t.Logf("Create LocalVolumeGroup fail %v", err)
	}

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		//{
		//	args: args{
		//		stopCh: stopCh,
		//	},
		//},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				nameSpace:                 fakeNamespace,
				apiClient:                 client,
				localVolumeGroupQueue:     common.NewTaskQueue("localVolumeGroup", maxRetries),
				localVolumeQueue:          common.NewTaskQueue("localVolume", maxRetries),
				pvcQueue:                  common.NewTaskQueue("pvc", maxRetries),
				podQueue:                  common.NewTaskQueue("pod", maxRetries),
				localVolumeToVolumeGroups: make(map[string]string),
				pvcToVolumeGroups:         make(map[string]string),
				podToVolumeGroups:         make(map[string]string),
				logger:                    log.WithField("Module", "ControllerManager"),
			}
			m.startPodWorker(tt.args.stopCh)
		})
	}
}

func Test_namespacedName(t *testing.T) {
	type args struct {
		namespace string
		name      string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{
			args: args{
				namespace: fakeNamespace,
				name:      fakeName,
			},
			want: fakeNamespace + "/" + fakeName,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := namespacedName(tt.args.namespace, tt.args.name); got != tt.want {
				t.Logf("namespacedName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseNamespacedName(t *testing.T) {
	type args struct {
		nn string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
	}{
		// TODO: Add test cases.
		{
			args: args{
				nn: fakeNamespace + "/" + fakeName,
			},
			want:  fakeNamespace,
			want1: fakeName,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := parseNamespacedName(tt.args.nn)
			if got != tt.want {
				t.Logf("parseNamespacedName() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Logf("parseNamespacedName() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_updateLocalVolumeGroupAccessibility(t *testing.T) {
	tests := []struct {
		Description         string
		Lvg                 *apisv1alpha1.LocalVolumeGroup
		Lv                  *apisv1alpha1.LocalVolume
		ExpectLvgAccessNode []string
	}{
		{
			Description:         "It is an update LocalVolumeGroup access node test,for LocalVolumeGroup keep a sync with LocalVolume access node",
			Lvg:                 GenFakeLocalVolumeGroupObject(),
			Lv:                  GenFakeLocalVolumeObject(),
			ExpectLvgAccessNode: fakeNodenames,
		},
	}
	for _, tt := range tests {
		t.Run(tt.Description, func(t *testing.T) {
			fakeClient, _ := CreateFakeClient()
			tt.Lv.Spec.VolumeGroup = tt.Lvg.Name
			err := fakeClient.Create(context.Background(), tt.Lv)
			if err != nil {
				t.Fatalf("Create LocalVolume fail %v", err)
			}
			err = fakeClient.Create(context.Background(), tt.Lvg)
			if err != nil {
				t.Fatalf("Create LocalVolumeGroup fail %v", err)
			}
			m := &manager{
				nameSpace:                 fakeNamespace,
				apiClient:                 fakeClient,
				localVolumeGroupQueue:     common.NewTaskQueue("localVolumeGroup", maxRetries),
				localVolumeQueue:          common.NewTaskQueue("localVolume", maxRetries),
				pvcQueue:                  common.NewTaskQueue("pvc", maxRetries),
				podQueue:                  common.NewTaskQueue("pod", maxRetries),
				localVolumeToVolumeGroups: make(map[string]string),
				pvcToVolumeGroups:         make(map[string]string),
				podToVolumeGroups:         make(map[string]string),
				logger:                    log.WithField("Module", "ControllerManager"),
			}
			m.updateLocalVolumeGroupAccessibility(tt.Lvg)
			var newLvg = &apisv1alpha1.LocalVolumeGroup{}
			if err = fakeClient.Get(context.Background(), client.ObjectKey{Namespace: tt.Lvg.Namespace, Name: tt.Lvg.Name}, newLvg); err == nil {
				if !reflect.DeepEqual(tt.ExpectLvgAccessNode, newLvg.Spec.Accessibility.Nodes) {
					t.Fatal("test updateLocalVolumeGroupAccessibility failed")
				}
			}
		})
	}
}
