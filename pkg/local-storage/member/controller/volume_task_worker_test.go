package controller

import (
	"context"
	"reflect"
	"sync"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	coorv1 "k8s.io/api/coordination/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	v1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-storage/common"
)

var (
	fakeLocalStorageNodeName        = "local-storage-node-example"
	fakeLocalStorageNodeUID         = "local-storage-node-uid"
	fakeLocalVolumeName             = "local-volume-example"
	fakeLocalStorageNodename        = "local-storage-node-example"
	fakeLocalVolumeGroupName        = "local-volume-group-example"
	fakeLocalVolumeReplicaName      = "local-volume-replica-example"
	fakeLocalVolumeConvertName      = "local-volume-convert-example"
	fakeLocalVolumeMigrateName      = "local-volume-migrate-example"
	fakeLocalVolumeGroupMigrateName = "local-volume-group-migrate-example"
	fakeLocalVolumeGroupConvertName = "local-volume-group-convert-example"
	fakeLocalVolumeUID              = "local-volume-uid"
	fakeLocalVolumeGroupUID         = "local-volume-group-uid"
	fakeNamespace                   = "local-volume-test"
	fakeManagerName                 = "manager-example"
	fakeNodenames                   = []string{"10-6-118-10"}
	fakeNodename                    = "10-6-118-10"
	fakeStorageIp                   = "10.6.118.11"
	fakeZone                        = "zone-test"
	fakeRegion                      = "region-test"
	fakeVgType                      = "LocalStorage_PoolHDD"
	fakePersistentVolumeClaimName   = "pvc-name-test"
	fakePoolClass                   = "HDD"
	fakePoolType                    = "REGULAR"
	fakeHolderIdentity              = "fakeHolderIdentity"
	fakeLeaseDurationSeconds        = int32(30)
	fakeLeaseTransitions            = int32(30)
	fakeAcquireTime                 = time.Now()

	fakeTopo                     = v1alpha1.Topology{Region: fakeRegion, Zone: fakeZone}
	fakeVgmNamespacedName        = "local-volume-test/local-volume-group-example"
	fakePods                     = []string{"pod-test1"}
	fakeVolumes                  = []v1alpha1.VolumeInfo{{LocalVolumeName: "local-volume-test1", PersistentVolumeClaimName: "pvc-test1"}}
	fakeTotalCapacityBytes int64 = 10 * 1024 * 1024 * 1024
	fakeFreeCapacityBytes  int64 = 8 * 1024 * 1024 * 1024
	fakeDiskCapacityBytes  int64 = 2 * 1024 * 1024 * 1024

	apiversion                  = "hwameistor.io/v1alpha1"
	LocalVolumeKind             = "LocalVolume"
	LocalVolumeGroupKind        = "LocalVolumeGroup"
	LocalStorageNodeKind        = "LocalStorageNode"
	LeaseKind                   = "Lease"
	LocalVolumeReplicaKind      = "LocalVolumeReplica"
	LocalVolumeConvertKind      = "LocalVolumeConvert"
	LocalVolumeMigrateKind      = "LocalVolumeMigrate"
	LocalVolumeGroupConvertKind = "LocalVolumeGroupConvert"
	LocalVolumeGroupMigrateKind = "LocalVolumeGroupMigrate"
	fakeRecorder                = record.NewFakeRecorder(100)

	defaultDRBDStartPort      = 43001
	defaultHAVolumeTotalCount = 1000
	fakeAcesscibility         = v1alpha1.AccessibilityTopology{Nodes: []string{"test-node1"}}
)

func Test_manager_startVolumeTaskWorker(t *testing.T) {
	type fields struct {
		name                        string
		namespace                   string
		apiClient                   client.Client
		informersCache              cache.Cache
		scheme                      *runtime.Scheme
		volumeScheduler             v1alpha1.VolumeScheduler
		volumeGroupManager          v1alpha1.VolumeGroupManager
		nodeTaskQueue               *common.TaskQueue
		k8sNodeTaskQueue            *common.TaskQueue
		volumeTaskQueue             *common.TaskQueue
		volumeExpandTaskQueue       *common.TaskQueue
		volumeMigrateTaskQueue      *common.TaskQueue
		volumeGroupMigrateTaskQueue *common.TaskQueue
		volumeConvertTaskQueue      *common.TaskQueue
		volumeGroupConvertTaskQueue *common.TaskQueue
		localNodes                  map[string]v1alpha1.State
		logger                      *log.Entry
		lock                        sync.Mutex
	}
	type args struct {
		stopCh <-chan struct{}
	}

	client, _ := CreateFakeClient()

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		//{
		//	args: args{stopCh: stopCh},
		//},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				name:                        fakeManagerName,
				namespace:                   fakeNamespace,
				apiClient:                   client,
				nodeTaskQueue:               common.NewTaskQueue("NodeTask", maxRetries),
				k8sNodeTaskQueue:            common.NewTaskQueue("K8sNodeTask", maxRetries),
				volumeTaskQueue:             common.NewTaskQueue("VolumeTask", maxRetries),
				volumeExpandTaskQueue:       common.NewTaskQueue("VolumeExpandTask", maxRetries),
				volumeMigrateTaskQueue:      common.NewTaskQueue("VolumeMigrateTask", maxRetries),
				volumeGroupMigrateTaskQueue: common.NewTaskQueue("VolumeGroupMigrateTask", maxRetries),
				volumeConvertTaskQueue:      common.NewTaskQueue("VolumeConvertTask", maxRetries),
				volumeGroupConvertTaskQueue: common.NewTaskQueue("VolumeGroupConvertTask", maxRetries),
				localNodes:                  map[string]v1alpha1.State{},
				logger:                      log.WithField("Module", "ControllerManager"),
			}
			m.startVolumeTaskWorker(tt.args.stopCh)
		})
	}
}

// GenFakeLocalVolumeObject Create lv request
func GenFakeLocalVolumeObject() *v1alpha1.LocalVolume {
	lv := &v1alpha1.LocalVolume{}

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

	Spec := v1alpha1.LocalVolumeSpec{
		RequiredCapacityBytes: fakeDiskCapacityBytes,
		ReplicaNumber:         1,
		PoolName:              fakeVgType,
		Delete:                false,
		Convertible:           true,
		Accessibility: v1alpha1.AccessibilityTopology{
			Nodes:   fakeNodenames,
			Regions: []string{fakeRegion},
			Zones:   []string{fakeZone},
		},
		Config: &v1alpha1.VolumeConfig{
			Convertible:           true,
			Initialized:           true,
			ReadyToInitialize:     true,
			RequiredCapacityBytes: fakeDiskCapacityBytes,
			ResourceID:            5,
			Version:               11,
			VolumeName:            fakeLocalVolumeName,
			Replicas: []v1alpha1.VolumeReplica{
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
	lv.Status.State = v1alpha1.VolumeStateCreating
	lv.Status.AllocatedCapacityBytes = fakeTotalCapacityBytes - fakeFreeCapacityBytes
	lv.Status.PublishedNodeName = fakeNodename
	lv.Status.Replicas = []string{fakeLocalVolumeName}

	return lv
}

// GenFakeLocalVolumeObject Create lv request
func GenFakeLocalVolumeReplicaObject() *v1alpha1.LocalVolumeReplica {
	lvr := &v1alpha1.LocalVolumeReplica{}

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

	Spec := v1alpha1.LocalVolumeReplicaSpec{
		RequiredCapacityBytes: fakeDiskCapacityBytes,
		PoolName:              fakeVgType,
		Delete:                false,
		VolumeName:            fakeLocalVolumeName,
		NodeName:              fakeNodename,
	}

	lvr.ObjectMeta = ObjectMata
	lvr.TypeMeta = TypeMeta
	lvr.Spec = Spec
	lvr.Status.State = v1alpha1.VolumeStateCreating
	lvr.Status.AllocatedCapacityBytes = fakeTotalCapacityBytes - fakeFreeCapacityBytes

	return lvr
}

func GenFakeLocalVolumeConvertObject() *v1alpha1.LocalVolumeConvert {
	lvc := &v1alpha1.LocalVolumeConvert{}

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

	Spec := v1alpha1.LocalVolumeConvertSpec{
		ReplicaNumber: 1,
		VolumeName:    fakeLocalVolumeName,
	}

	lvc.ObjectMeta = ObjectMata
	lvc.TypeMeta = TypeMeta
	lvc.Spec = Spec

	return lvc
}

func GenFakeLocalVolumeMigrateObject() *v1alpha1.LocalVolumeMigrate {
	lvm := &v1alpha1.LocalVolumeMigrate{}

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

	Spec := v1alpha1.LocalVolumeMigrateSpec{
		TargetNodesSuggested: fakeNodenames,
		VolumeName:           fakeLocalVolumeName,
	}

	lvm.ObjectMeta = ObjectMata
	lvm.TypeMeta = TypeMeta
	lvm.Spec = Spec

	return lvm
}

func GenFakeLocalVolumeGroupObject() *v1alpha1.LocalVolumeGroup {
	lvc := &v1alpha1.LocalVolumeGroup{}

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

	Spec := v1alpha1.LocalVolumeGroupSpec{
		Volumes:       fakeVolumes,
		Accessibility: fakeAcesscibility,
		Pods:          fakePods,
	}

	lvc.ObjectMeta = ObjectMata
	lvc.TypeMeta = TypeMeta
	lvc.Spec = Spec

	return lvc
}

func GenFakeLocalStorageNodeObject() *v1alpha1.LocalStorageNode {
	lsn := &v1alpha1.LocalStorageNode{}

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

	Spec := v1alpha1.LocalStorageNodeSpec{
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
func CreateFakeClient() (client.Client, *runtime.Scheme) {

	lv := GenFakeLocalVolumeObject()
	lvList := &v1alpha1.LocalVolumeList{
		TypeMeta: metav1.TypeMeta{
			Kind:       LocalVolumeKind,
			APIVersion: apiversion,
		},
	}

	lvc := GenFakeLocalVolumeConvertObject()
	lvcList := &v1alpha1.LocalVolumeConvertList{
		TypeMeta: metav1.TypeMeta{
			Kind:       LocalVolumeConvertKind,
			APIVersion: apiversion,
		},
	}

	lvm := GenFakeLocalVolumeMigrateObject()
	lvmList := &v1alpha1.LocalVolumeMigrateList{
		TypeMeta: metav1.TypeMeta{
			Kind:       LocalVolumeMigrateKind,
			APIVersion: apiversion,
		},
	}

	lvr := GenFakeLocalVolumeReplicaObject()
	lvrList := &v1alpha1.LocalVolumeReplicaList{
		TypeMeta: metav1.TypeMeta{
			Kind:       LocalVolumeReplicaKind,
			APIVersion: apiversion,
		},
	}

	lvg := GenFakeLocalVolumeGroupObject()
	lvgList := &v1alpha1.LocalVolumeGroupList{
		TypeMeta: metav1.TypeMeta{
			Kind:       LocalVolumeGroupKind,
			APIVersion: apiversion,
		},
	}

	lsn := GenFakeLocalStorageNodeObject()
	lsnList := &v1alpha1.LocalStorageNodeList{
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
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, lv)
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, lvList)
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, lvc)
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, lvcList)
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, lvm)
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, lvmList)
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, lvr)
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, lvrList)
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, lvg)
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, lvgList)
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, lsn)
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, lsnList)
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, lease)
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, leaseList)
	return fake.NewClientBuilder().WithScheme(s).Build(), s
}

func Test_manager_processVolume(t *testing.T) {
	type fields struct {
		name                        string
		namespace                   string
		apiClient                   client.Client
		informersCache              cache.Cache
		scheme                      *runtime.Scheme
		volumeScheduler             v1alpha1.VolumeScheduler
		volumeGroupManager          v1alpha1.VolumeGroupManager
		nodeTaskQueue               *common.TaskQueue
		k8sNodeTaskQueue            *common.TaskQueue
		volumeTaskQueue             *common.TaskQueue
		volumeExpandTaskQueue       *common.TaskQueue
		volumeMigrateTaskQueue      *common.TaskQueue
		volumeGroupMigrateTaskQueue *common.TaskQueue
		volumeConvertTaskQueue      *common.TaskQueue
		volumeGroupConvertTaskQueue *common.TaskQueue
		localNodes                  map[string]v1alpha1.State
		logger                      *log.Entry
		lock                        sync.Mutex
	}

	client, _ := CreateFakeClient()
	type args struct {
		volName string
	}
	var volName = fakeLocalVolumeName
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args: args{volName: volName},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				name:                        fakeManagerName,
				namespace:                   fakeNamespace,
				apiClient:                   client,
				nodeTaskQueue:               common.NewTaskQueue("NodeTask", maxRetries),
				k8sNodeTaskQueue:            common.NewTaskQueue("K8sNodeTask", maxRetries),
				volumeTaskQueue:             common.NewTaskQueue("VolumeTask", maxRetries),
				volumeExpandTaskQueue:       common.NewTaskQueue("VolumeExpandTask", maxRetries),
				volumeMigrateTaskQueue:      common.NewTaskQueue("VolumeMigrateTask", maxRetries),
				volumeGroupMigrateTaskQueue: common.NewTaskQueue("VolumeGroupMigrateTask", maxRetries),
				volumeConvertTaskQueue:      common.NewTaskQueue("VolumeConvertTask", maxRetries),
				volumeGroupConvertTaskQueue: common.NewTaskQueue("VolumeGroupConvertTask", maxRetries),
				localNodes:                  map[string]v1alpha1.State{},
				logger:                      log.WithField("Module", "ControllerManager"),
			}
			if err := m.processVolume(tt.args.volName); (err != nil) != tt.wantErr {
				t.Errorf("processVolume() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_processVolumeSubmit(t *testing.T) {
	type fields struct {
		name                        string
		namespace                   string
		apiClient                   client.Client
		informersCache              cache.Cache
		scheme                      *runtime.Scheme
		volumeScheduler             v1alpha1.VolumeScheduler
		volumeGroupManager          v1alpha1.VolumeGroupManager
		nodeTaskQueue               *common.TaskQueue
		k8sNodeTaskQueue            *common.TaskQueue
		volumeTaskQueue             *common.TaskQueue
		volumeExpandTaskQueue       *common.TaskQueue
		volumeMigrateTaskQueue      *common.TaskQueue
		volumeGroupMigrateTaskQueue *common.TaskQueue
		volumeConvertTaskQueue      *common.TaskQueue
		volumeGroupConvertTaskQueue *common.TaskQueue
		localNodes                  map[string]v1alpha1.State
		logger                      *log.Entry
		lock                        sync.Mutex
	}
	type args struct {
		vol *v1alpha1.LocalVolume
	}

	client, _ := CreateFakeClient()

	// Create LocalVolume
	lv := GenFakeLocalVolumeObject()
	lv.Name = fakeLocalVolumeName
	lv.Spec.VolumeGroup = fakeLocalVolumeGroupName
	lv.Spec.PersistentVolumeClaimName = fakePersistentVolumeClaimName
	lv.Spec.ReplicaNumber = 1
	lv.Spec.Convertible = true
	lv.Spec.PersistentVolumeClaimNamespace = fakeNamespace

	err := client.Create(context.Background(), lv)
	if err != nil {
		t.Errorf("Create LocalVolume fail %v", err)
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args:    args{vol: lv},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				name:                        fakeManagerName,
				namespace:                   fakeNamespace,
				apiClient:                   client,
				nodeTaskQueue:               common.NewTaskQueue("NodeTask", maxRetries),
				k8sNodeTaskQueue:            common.NewTaskQueue("K8sNodeTask", maxRetries),
				volumeTaskQueue:             common.NewTaskQueue("VolumeTask", maxRetries),
				volumeExpandTaskQueue:       common.NewTaskQueue("VolumeExpandTask", maxRetries),
				volumeMigrateTaskQueue:      common.NewTaskQueue("VolumeMigrateTask", maxRetries),
				volumeGroupMigrateTaskQueue: common.NewTaskQueue("VolumeGroupMigrateTask", maxRetries),
				volumeConvertTaskQueue:      common.NewTaskQueue("VolumeConvertTask", maxRetries),
				volumeGroupConvertTaskQueue: common.NewTaskQueue("VolumeGroupConvertTask", maxRetries),
				localNodes:                  map[string]v1alpha1.State{},
				logger:                      log.WithField("Module", "ControllerManager"),
			}
			if err := m.processVolumeSubmit(tt.args.vol); (err != nil) != tt.wantErr {
				t.Errorf("processVolumeSubmit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_getReplicasForVolume(t *testing.T) {
	type fields struct {
		name                        string
		namespace                   string
		apiClient                   client.Client
		informersCache              cache.Cache
		scheme                      *runtime.Scheme
		volumeScheduler             v1alpha1.VolumeScheduler
		volumeGroupManager          v1alpha1.VolumeGroupManager
		nodeTaskQueue               *common.TaskQueue
		k8sNodeTaskQueue            *common.TaskQueue
		volumeTaskQueue             *common.TaskQueue
		volumeExpandTaskQueue       *common.TaskQueue
		volumeMigrateTaskQueue      *common.TaskQueue
		volumeGroupMigrateTaskQueue *common.TaskQueue
		volumeConvertTaskQueue      *common.TaskQueue
		volumeGroupConvertTaskQueue *common.TaskQueue
		localNodes                  map[string]v1alpha1.State
		logger                      *log.Entry
		lock                        sync.Mutex
	}
	type args struct {
		volName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*v1alpha1.LocalVolumeReplica
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				name:                        tt.fields.name,
				namespace:                   tt.fields.namespace,
				apiClient:                   tt.fields.apiClient,
				informersCache:              tt.fields.informersCache,
				scheme:                      tt.fields.scheme,
				volumeScheduler:             tt.fields.volumeScheduler,
				volumeGroupManager:          tt.fields.volumeGroupManager,
				nodeTaskQueue:               tt.fields.nodeTaskQueue,
				k8sNodeTaskQueue:            tt.fields.k8sNodeTaskQueue,
				volumeTaskQueue:             tt.fields.volumeTaskQueue,
				volumeExpandTaskQueue:       tt.fields.volumeExpandTaskQueue,
				volumeMigrateTaskQueue:      tt.fields.volumeMigrateTaskQueue,
				volumeGroupMigrateTaskQueue: tt.fields.volumeGroupMigrateTaskQueue,
				volumeConvertTaskQueue:      tt.fields.volumeConvertTaskQueue,
				volumeGroupConvertTaskQueue: tt.fields.volumeGroupConvertTaskQueue,
				localNodes:                  tt.fields.localNodes,
				logger:                      tt.fields.logger,
				lock:                        tt.fields.lock,
			}
			got, err := m.getReplicasForVolume(tt.args.volName)
			if (err != nil) != tt.wantErr {
				t.Errorf("getReplicasForVolume() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getReplicasForVolume() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_manager_processVolumeCreate(t *testing.T) {
	type fields struct {
		name                        string
		namespace                   string
		apiClient                   client.Client
		informersCache              cache.Cache
		scheme                      *runtime.Scheme
		volumeScheduler             v1alpha1.VolumeScheduler
		volumeGroupManager          v1alpha1.VolumeGroupManager
		nodeTaskQueue               *common.TaskQueue
		k8sNodeTaskQueue            *common.TaskQueue
		volumeTaskQueue             *common.TaskQueue
		volumeExpandTaskQueue       *common.TaskQueue
		volumeMigrateTaskQueue      *common.TaskQueue
		volumeGroupMigrateTaskQueue *common.TaskQueue
		volumeConvertTaskQueue      *common.TaskQueue
		volumeGroupConvertTaskQueue *common.TaskQueue
		localNodes                  map[string]v1alpha1.State
		logger                      *log.Entry
		lock                        sync.Mutex
	}
	type args struct {
		vol *v1alpha1.LocalVolume
	}
	client, _ := CreateFakeClient()

	// Create LocalVolume
	lv := GenFakeLocalVolumeObject()
	lv.Name = fakeLocalVolumeName
	lv.Spec.VolumeGroup = fakeLocalVolumeGroupName
	lv.Spec.PersistentVolumeClaimName = fakePersistentVolumeClaimName
	lv.Spec.ReplicaNumber = 1
	lv.Spec.Convertible = true
	lv.Spec.PersistentVolumeClaimNamespace = fakeNamespace

	err := client.Create(context.Background(), lv)
	if err != nil {
		t.Errorf("Create LocalVolume fail %v", err)
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		//{
		//	args: args{vol: vol},
		//},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				name:                        fakeManagerName,
				namespace:                   fakeNamespace,
				apiClient:                   client,
				nodeTaskQueue:               common.NewTaskQueue("NodeTask", maxRetries),
				k8sNodeTaskQueue:            common.NewTaskQueue("K8sNodeTask", maxRetries),
				volumeTaskQueue:             common.NewTaskQueue("VolumeTask", maxRetries),
				volumeExpandTaskQueue:       common.NewTaskQueue("VolumeExpandTask", maxRetries),
				volumeMigrateTaskQueue:      common.NewTaskQueue("VolumeMigrateTask", maxRetries),
				volumeGroupMigrateTaskQueue: common.NewTaskQueue("VolumeGroupMigrateTask", maxRetries),
				volumeConvertTaskQueue:      common.NewTaskQueue("VolumeConvertTask", maxRetries),
				volumeGroupConvertTaskQueue: common.NewTaskQueue("VolumeGroupConvertTask", maxRetries),
				localNodes:                  map[string]v1alpha1.State{},
				logger:                      log.WithField("Module", "ControllerManager"),
			}
			if err := m.processVolumeCreate(tt.args.vol); (err != nil) != tt.wantErr {
				t.Errorf("processVolumeCreate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_processVolumeReadyAndNotReady(t *testing.T) {
	type fields struct {
		name                        string
		namespace                   string
		apiClient                   client.Client
		informersCache              cache.Cache
		scheme                      *runtime.Scheme
		volumeScheduler             v1alpha1.VolumeScheduler
		volumeGroupManager          v1alpha1.VolumeGroupManager
		nodeTaskQueue               *common.TaskQueue
		k8sNodeTaskQueue            *common.TaskQueue
		volumeTaskQueue             *common.TaskQueue
		volumeExpandTaskQueue       *common.TaskQueue
		volumeMigrateTaskQueue      *common.TaskQueue
		volumeGroupMigrateTaskQueue *common.TaskQueue
		volumeConvertTaskQueue      *common.TaskQueue
		volumeGroupConvertTaskQueue *common.TaskQueue
		localNodes                  map[string]v1alpha1.State
		logger                      *log.Entry
		lock                        sync.Mutex
	}
	type args struct {
		vol *v1alpha1.LocalVolume
	}

	client, _ := CreateFakeClient()

	// Create LocalVolume
	lv := GenFakeLocalVolumeObject()
	lv.Name = fakeLocalVolumeName
	lv.Spec.VolumeGroup = fakeLocalVolumeGroupName
	lv.Spec.PersistentVolumeClaimName = fakePersistentVolumeClaimName
	lv.Spec.ReplicaNumber = 1
	lv.Spec.Convertible = true
	lv.Spec.PersistentVolumeClaimNamespace = fakeNamespace

	err := client.Create(context.Background(), lv)
	if err != nil {
		t.Errorf("Create LocalVolume fail %v", err)
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args:    args{vol: lv},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				name:                        fakeManagerName,
				namespace:                   fakeNamespace,
				apiClient:                   client,
				nodeTaskQueue:               common.NewTaskQueue("NodeTask", maxRetries),
				k8sNodeTaskQueue:            common.NewTaskQueue("K8sNodeTask", maxRetries),
				volumeTaskQueue:             common.NewTaskQueue("VolumeTask", maxRetries),
				volumeExpandTaskQueue:       common.NewTaskQueue("VolumeExpandTask", maxRetries),
				volumeMigrateTaskQueue:      common.NewTaskQueue("VolumeMigrateTask", maxRetries),
				volumeGroupMigrateTaskQueue: common.NewTaskQueue("VolumeGroupMigrateTask", maxRetries),
				volumeConvertTaskQueue:      common.NewTaskQueue("VolumeConvertTask", maxRetries),
				volumeGroupConvertTaskQueue: common.NewTaskQueue("VolumeGroupConvertTask", maxRetries),
				localNodes:                  map[string]v1alpha1.State{},
				logger:                      log.WithField("Module", "ControllerManager"),
			}
			if err := m.processVolumeReadyAndNotReady(tt.args.vol); (err != nil) != tt.wantErr {
				t.Errorf("processVolumeReadyAndNotReady() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_isVolumeReplicaUp(t *testing.T) {
	type args struct {
		replica *v1alpha1.LocalVolumeReplica
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isVolumeReplicaUp(tt.args.replica); got != tt.want {
				t.Errorf("isVolumeReplicaUp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_manager_processVolumeDelete(t *testing.T) {
	type fields struct {
		name                        string
		namespace                   string
		apiClient                   client.Client
		informersCache              cache.Cache
		scheme                      *runtime.Scheme
		volumeScheduler             v1alpha1.VolumeScheduler
		volumeGroupManager          v1alpha1.VolumeGroupManager
		nodeTaskQueue               *common.TaskQueue
		k8sNodeTaskQueue            *common.TaskQueue
		volumeTaskQueue             *common.TaskQueue
		volumeExpandTaskQueue       *common.TaskQueue
		volumeMigrateTaskQueue      *common.TaskQueue
		volumeGroupMigrateTaskQueue *common.TaskQueue
		volumeConvertTaskQueue      *common.TaskQueue
		volumeGroupConvertTaskQueue *common.TaskQueue
		localNodes                  map[string]v1alpha1.State
		logger                      *log.Entry
		lock                        sync.Mutex
	}
	type args struct {
		vol *v1alpha1.LocalVolume
	}

	client, _ := CreateFakeClient()

	// Create LocalVolume
	lv := GenFakeLocalVolumeObject()
	lv.Name = fakeLocalVolumeName
	lv.Spec.VolumeGroup = fakeLocalVolumeGroupName
	lv.Spec.PersistentVolumeClaimName = fakePersistentVolumeClaimName
	lv.Spec.ReplicaNumber = 1
	lv.Spec.Convertible = true
	lv.Spec.PersistentVolumeClaimNamespace = fakeNamespace

	err := client.Create(context.Background(), lv)
	if err != nil {
		t.Errorf("Create LocalVolume fail %v", err)
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args:    args{vol: lv},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				name:                        fakeManagerName,
				namespace:                   fakeNamespace,
				apiClient:                   client,
				nodeTaskQueue:               common.NewTaskQueue("NodeTask", maxRetries),
				k8sNodeTaskQueue:            common.NewTaskQueue("K8sNodeTask", maxRetries),
				volumeTaskQueue:             common.NewTaskQueue("VolumeTask", maxRetries),
				volumeExpandTaskQueue:       common.NewTaskQueue("VolumeExpandTask", maxRetries),
				volumeMigrateTaskQueue:      common.NewTaskQueue("VolumeMigrateTask", maxRetries),
				volumeGroupMigrateTaskQueue: common.NewTaskQueue("VolumeGroupMigrateTask", maxRetries),
				volumeConvertTaskQueue:      common.NewTaskQueue("VolumeConvertTask", maxRetries),
				volumeGroupConvertTaskQueue: common.NewTaskQueue("VolumeGroupConvertTask", maxRetries),
				localNodes:                  map[string]v1alpha1.State{},
				logger:                      log.WithField("Module", "ControllerManager"),
			}
			if err := m.processVolumeDelete(tt.args.vol); (err != nil) != tt.wantErr {
				t.Errorf("processVolumeDelete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_processVolumeCleanup(t *testing.T) {
	type fields struct {
		name                        string
		namespace                   string
		apiClient                   client.Client
		informersCache              cache.Cache
		scheme                      *runtime.Scheme
		volumeScheduler             v1alpha1.VolumeScheduler
		volumeGroupManager          v1alpha1.VolumeGroupManager
		nodeTaskQueue               *common.TaskQueue
		k8sNodeTaskQueue            *common.TaskQueue
		volumeTaskQueue             *common.TaskQueue
		volumeExpandTaskQueue       *common.TaskQueue
		volumeMigrateTaskQueue      *common.TaskQueue
		volumeGroupMigrateTaskQueue *common.TaskQueue
		volumeConvertTaskQueue      *common.TaskQueue
		volumeGroupConvertTaskQueue *common.TaskQueue
		localNodes                  map[string]v1alpha1.State
		logger                      *log.Entry
		lock                        sync.Mutex
	}
	type args struct {
		vol *v1alpha1.LocalVolume
	}

	client, _ := CreateFakeClient()

	// Create LocalVolume
	lv := GenFakeLocalVolumeObject()
	lv.Name = fakeLocalVolumeName
	lv.Spec.VolumeGroup = fakeLocalVolumeGroupName
	lv.Spec.PersistentVolumeClaimName = fakePersistentVolumeClaimName
	lv.Spec.ReplicaNumber = 1
	lv.Spec.Convertible = true
	lv.Spec.PersistentVolumeClaimNamespace = fakeNamespace

	err := client.Create(context.Background(), lv)
	if err != nil {
		t.Errorf("Create LocalVolume fail %v", err)
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args:    args{vol: lv},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				name:                        fakeManagerName,
				namespace:                   fakeNamespace,
				apiClient:                   client,
				nodeTaskQueue:               common.NewTaskQueue("NodeTask", maxRetries),
				k8sNodeTaskQueue:            common.NewTaskQueue("K8sNodeTask", maxRetries),
				volumeTaskQueue:             common.NewTaskQueue("VolumeTask", maxRetries),
				volumeExpandTaskQueue:       common.NewTaskQueue("VolumeExpandTask", maxRetries),
				volumeMigrateTaskQueue:      common.NewTaskQueue("VolumeMigrateTask", maxRetries),
				volumeGroupMigrateTaskQueue: common.NewTaskQueue("VolumeGroupMigrateTask", maxRetries),
				volumeConvertTaskQueue:      common.NewTaskQueue("VolumeConvertTask", maxRetries),
				volumeGroupConvertTaskQueue: common.NewTaskQueue("VolumeGroupConvertTask", maxRetries),
				localNodes:                  map[string]v1alpha1.State{},
				logger:                      log.WithField("Module", "ControllerManager"),
			}
			if err := m.processVolumeCleanup(tt.args.vol); (err != nil) != tt.wantErr {
				t.Errorf("processVolumeCleanup() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
