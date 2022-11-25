package node

import (
	"context"
	"reflect"
	"testing"
	"time"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-storage/common"
	"github.com/hwameistor/hwameistor/pkg/local-storage/member/node/diskmonitor"
	"github.com/hwameistor/hwameistor/pkg/local-storage/member/node/storage"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	fakeLocalStorageNodeName        = "local-storage-node-example"
	fakeLocalStorageNodeUID         = "local-storage-node-uid"
	fakeLocalVolumeName             = "local-volume-example"
	fakeLocalDiskName               = "local-disk-example"
	fakeLocalDiskNames              = []string{"local-disk-example"}
	fakelocalDiskNameSpacedName     = "fakeNameSpace/fakeName"
	fakeLocalDiskClaimName          = "local-disk-claim-example"
	fakeLocalStorageNodename        = "local-storage-node-example"
	fakeLocalVolumeGroupName        = "local-volume-group-example"
	fakeLocalVolumeReplicaName      = "local-volume-replica-example"
	fakeLocalVolumeConvertName      = "local-volume-convert-example"
	fakeLocalVolumeMigrateName      = "local-volume-migrate-example"
	fakeLocalVolumeGroupMigrateName = "local-volume-group-migrate-example"
	fakeLocalVolumeGroupConvertName = "local-volume-group-convert-example"
	fakeManagerName                 = "manager-example"
	fakeLocalVolumeUID              = "local-volume-uid"
	fakeLocalDiskUID                = "local-disk-uid"
	fakeLocalDiskClaimUID           = "local-disk-claim-uid"
	fakeDevicePath                  = "/dev/test"
	fakeLocalVolumeGroupUID         = "local-volume-group-uid"
	fakeNamespace                   = "local-volume-test"
	fakeNodenames                   = []string{"10-6-118-10"}
	fakeNodename                    = "10-6-118-10"
	fakeStorageIp                   = "10.6.118.11"
	fakeZone                        = "zone-test"
	fakeRegion                      = "region-test"
	fakeVgType                      = "LocalStorage_PoolHDD"
	fakePartitionInfo               = []apisv1alpha1.PartitionInfo{{Path: "test", HasFileSystem: true, FileSystem: apisv1alpha1.FileSystemInfo{Type: "test", Mountpoint: "test"}}}
	fakeRaidInfo                    = apisv1alpha1.RAIDInfo{RAIDMaster: "test"}
	fakeSmartInfo                   = apisv1alpha1.SmartInfo{OverallHealth: apisv1alpha1.SmartAssessResult("test")}
	fakeDiskAttributes              = apisv1alpha1.DiskAttributes{Type: "test"}
	fakeDescription                 = "fakeDescription"
	fakePersistentVolumeClaimName   = "pvc-name-test"
	fakePoolClass                   = "HDD"
	fakePoolType                    = "REGULAR"
	fakeHolderIdentity              = "fakeHolderIdentity"
	fakeLeaseDurationSeconds        = int32(30)
	fakeLeaseTransitions            = int32(30)
	fakeAcquireTime                 = time.Now()

	fakeTopo                     = apisv1alpha1.Topology{Region: fakeRegion, Zone: fakeZone}
	fakeVgmNamespacedName        = "local-volume-test/local-volume-group-example"
	fakePods                     = []string{"pod-test1"}
	fakeVolumes                  = []apisv1alpha1.VolumeInfo{{LocalVolumeName: "local-volume-test1", PersistentVolumeClaimName: "pvc-test1"}}
	fakeTotalCapacityBytes int64 = 10 * 1024 * 1024 * 1024
	fakeFreeCapacityBytes  int64 = 8 * 1024 * 1024 * 1024
	fakeDiskCapacityBytes  int64 = 2 * 1024 * 1024 * 1024

	apiversion                  = "hwameistor.io/v1alpha1"
	LocalVolumeKind             = "LocalVolume"
	LocalDiskKind               = "localDisk"
	LocalDiskClaimKind          = "LocalDiskClaim"
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
	fakeAcesscibility         = apisv1alpha1.AccessibilityTopology{Nodes: []string{"test-node1"}}
)

func Test_manager_cleanupVolumeReplica(t *testing.T) {
	type fields struct {
		name                    string
		namespace               string
		apiClient               client.Client
		informersCache          cache.Cache
		replicaRecords          map[string]string
		storageMgr              *storage.LocalManager
		diskEventQueue          *diskmonitor.EventQueue
		volumeTaskQueue         *common.TaskQueue
		volumeReplicaTaskQueue  *common.TaskQueue
		localDiskClaimTaskQueue *common.TaskQueue
		localDiskTaskQueue      *common.TaskQueue
		configManager           *configManager
		logger                  *log.Entry
	}
	type args struct {
		volName string
	}

	client, _ := CreateFakeClient()

	// Create LocalVolumeReplica
	lvr := GenFakeLocalVolumeReplicaObject()
	lvr.Name = fakeLocalVolumeReplicaName
	lvr.Spec.RequiredCapacityBytes = fakeDiskCapacityBytes
	err := client.Create(context.Background(), lvr)
	if err != nil {
		t.Errorf("Create LocalVolumeConvert fail %v", err)
	}

	// Create LocalVolume
	lv := GenFakeLocalVolumeObject()
	lv.Name = fakeLocalVolumeName
	lv.Spec.RequiredCapacityBytes = fakeDiskCapacityBytes
	err = client.Create(context.Background(), lv)
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
			args: args{
				volName: fakeLocalVolumeName,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				name:                    fakeManagerName,
				namespace:               fakeNamespace,
				apiClient:               client,
				replicaRecords:          map[string]string{},
				volumeTaskQueue:         common.NewTaskQueue("VolumeTask", maxRetries),
				volumeReplicaTaskQueue:  common.NewTaskQueue("VolumeReplicaTask", maxRetries),
				localDiskClaimTaskQueue: common.NewTaskQueue("LocalDiskClaim", maxRetries),
				localDiskTaskQueue:      common.NewTaskQueue("localDisk", maxRetries),
				// healthCheckQueue:        common.NewTaskQueue("HealthCheckTask", maxRetries),
				diskEventQueue: diskmonitor.NewEventQueue("DiskEvents"),
				logger:         log.WithField("Module", "NodeManager"),
			}
			if err := m.cleanupVolumeReplica(tt.args.volName); (err != nil) != tt.wantErr {
				t.Errorf("cleanupVolumeReplica() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_createVolumeReplica(t *testing.T) {
	type fields struct {
		name                    string
		namespace               string
		apiClient               client.Client
		informersCache          cache.Cache
		replicaRecords          map[string]string
		storageMgr              *storage.LocalManager
		diskEventQueue          *diskmonitor.EventQueue
		volumeTaskQueue         *common.TaskQueue
		volumeReplicaTaskQueue  *common.TaskQueue
		localDiskClaimTaskQueue *common.TaskQueue
		localDiskTaskQueue      *common.TaskQueue
		configManager           *configManager
		logger                  *log.Entry
	}
	type args struct {
		vol *apisv1alpha1.LocalVolume
	}

	client, _ := CreateFakeClient()

	// Create LocalVolumeReplica
	lvr := GenFakeLocalVolumeReplicaObject()
	lvr.Name = fakeLocalVolumeReplicaName
	lvr.Spec.RequiredCapacityBytes = fakeDiskCapacityBytes
	err := client.Create(context.Background(), lvr)
	if err != nil {
		t.Errorf("Create LocalVolumeConvert fail %v", err)
	}

	// Create LocalVolume
	lv := GenFakeLocalVolumeObject()
	lv.Name = fakeLocalVolumeName
	lv.Spec.RequiredCapacityBytes = fakeDiskCapacityBytes
	err = client.Create(context.Background(), lv)
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
			args: args{
				vol: lv,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				name:                    fakeManagerName,
				namespace:               fakeNamespace,
				apiClient:               client,
				replicaRecords:          map[string]string{},
				volumeTaskQueue:         common.NewTaskQueue("VolumeTask", maxRetries),
				volumeReplicaTaskQueue:  common.NewTaskQueue("VolumeReplicaTask", maxRetries),
				localDiskClaimTaskQueue: common.NewTaskQueue("LocalDiskClaim", maxRetries),
				localDiskTaskQueue:      common.NewTaskQueue("localDisk", maxRetries),
				// healthCheckQueue:        common.NewTaskQueue("HealthCheckTask", maxRetries),
				diskEventQueue: diskmonitor.NewEventQueue("DiskEvents"),
				logger:         log.WithField("Module", "NodeManager"),
			}
			if err := m.createVolumeReplica(tt.args.vol); (err != nil) != tt.wantErr {
				t.Errorf("createVolumeReplica() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_deleteVolumeReplica(t *testing.T) {
	type fields struct {
		name                    string
		namespace               string
		apiClient               client.Client
		informersCache          cache.Cache
		replicaRecords          map[string]string
		storageMgr              *storage.LocalManager
		diskEventQueue          *diskmonitor.EventQueue
		volumeTaskQueue         *common.TaskQueue
		volumeReplicaTaskQueue  *common.TaskQueue
		localDiskClaimTaskQueue *common.TaskQueue
		localDiskTaskQueue      *common.TaskQueue
		configManager           *configManager
		logger                  *log.Entry
	}
	type args struct {
		replica *apisv1alpha1.LocalVolumeReplica
	}

	client, _ := CreateFakeClient()

	// Create LocalVolumeReplica
	lvr := GenFakeLocalVolumeReplicaObject()
	lvr.Name = fakeLocalVolumeReplicaName
	lvr.Spec.RequiredCapacityBytes = fakeDiskCapacityBytes
	err := client.Create(context.Background(), lvr)
	if err != nil {
		t.Errorf("Create LocalVolumeConvert fail %v", err)
	}

	// Create LocalVolume
	lv := GenFakeLocalVolumeObject()
	lv.Name = fakeLocalVolumeName
	lv.Spec.RequiredCapacityBytes = fakeDiskCapacityBytes
	err = client.Create(context.Background(), lv)
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
			args: args{
				replica: lvr,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				name:                    fakeManagerName,
				namespace:               fakeNamespace,
				apiClient:               client,
				replicaRecords:          map[string]string{},
				volumeTaskQueue:         common.NewTaskQueue("VolumeTask", maxRetries),
				volumeReplicaTaskQueue:  common.NewTaskQueue("VolumeReplicaTask", maxRetries),
				localDiskClaimTaskQueue: common.NewTaskQueue("LocalDiskClaim", maxRetries),
				localDiskTaskQueue:      common.NewTaskQueue("localDisk", maxRetries),
				// healthCheckQueue:        common.NewTaskQueue("HealthCheckTask", maxRetries),
				diskEventQueue: diskmonitor.NewEventQueue("DiskEvents"),
				logger:         log.WithField("Module", "NodeManager"),
			}
			if err := m.deleteVolumeReplica(tt.args.replica); (err != nil) != tt.wantErr {
				t.Errorf("deleteVolumeReplica() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_getMyVolumeReplica(t *testing.T) {
	type fields struct {
		name                    string
		namespace               string
		apiClient               client.Client
		informersCache          cache.Cache
		replicaRecords          map[string]string
		storageMgr              *storage.LocalManager
		diskEventQueue          *diskmonitor.EventQueue
		volumeTaskQueue         *common.TaskQueue
		volumeReplicaTaskQueue  *common.TaskQueue
		localDiskClaimTaskQueue *common.TaskQueue
		localDiskTaskQueue      *common.TaskQueue
		configManager           *configManager
		logger                  *log.Entry
	}
	type args struct {
		volName string
	}

	client, _ := CreateFakeClient()

	// Create LocalVolumeReplica
	lvr := GenFakeLocalVolumeReplicaObject()
	lvr.Name = fakeLocalVolumeReplicaName
	lvr.Spec.RequiredCapacityBytes = fakeDiskCapacityBytes
	err := client.Create(context.Background(), lvr)
	if err != nil {
		t.Errorf("Create LocalVolumeConvert fail %v", err)
	}

	// Create LocalVolume
	lv := GenFakeLocalVolumeObject()
	lv.Name = fakeLocalVolumeName
	lv.Spec.RequiredCapacityBytes = fakeDiskCapacityBytes
	err = client.Create(context.Background(), lv)
	if err != nil {
		t.Errorf("Create LocalVolume fail %v", err)
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *apisv1alpha1.LocalVolumeReplica
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args: args{
				volName: fakeLocalVolumeName,
			},
			want:    lvr,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				name:                    fakeManagerName,
				namespace:               fakeNamespace,
				apiClient:               client,
				replicaRecords:          map[string]string{},
				volumeTaskQueue:         common.NewTaskQueue("VolumeTask", maxRetries),
				volumeReplicaTaskQueue:  common.NewTaskQueue("VolumeReplicaTask", maxRetries),
				localDiskClaimTaskQueue: common.NewTaskQueue("LocalDiskClaim", maxRetries),
				localDiskTaskQueue:      common.NewTaskQueue("localDisk", maxRetries),
				// healthCheckQueue:        common.NewTaskQueue("HealthCheckTask", maxRetries),
				diskEventQueue: diskmonitor.NewEventQueue("DiskEvents"),
				logger:         log.WithField("Module", "NodeManager"),
			}
			m.createVolumeReplica(lv)
			got, err := m.getMyVolumeReplica(tt.args.volName)
			if (err != nil) != tt.wantErr {
				t.Errorf("getMyVolumeReplica() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.DeepEqual(got, tt.want) {
				t.Errorf("getMyVolumeReplica() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_manager_processVolumeReplicaTaskAssignment(t *testing.T) {
	type fields struct {
		name                    string
		namespace               string
		apiClient               client.Client
		informersCache          cache.Cache
		replicaRecords          map[string]string
		storageMgr              *storage.LocalManager
		diskEventQueue          *diskmonitor.EventQueue
		volumeTaskQueue         *common.TaskQueue
		volumeReplicaTaskQueue  *common.TaskQueue
		localDiskClaimTaskQueue *common.TaskQueue
		localDiskTaskQueue      *common.TaskQueue
		configManager           *configManager
		logger                  *log.Entry
	}
	type args struct {
		volName string
	}
	client, _ := CreateFakeClient()

	// Create LocalVolumeReplica
	lvr := GenFakeLocalVolumeReplicaObject()
	lvr.Name = fakeLocalVolumeReplicaName
	lvr.Spec.RequiredCapacityBytes = fakeDiskCapacityBytes
	err := client.Create(context.Background(), lvr)
	if err != nil {
		t.Errorf("Create LocalVolumeConvert fail %v", err)
	}

	// Create LocalVolume
	lv := GenFakeLocalVolumeObject()
	lv.Name = fakeLocalVolumeName
	lv.Spec.RequiredCapacityBytes = fakeDiskCapacityBytes
	err = client.Create(context.Background(), lv)
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
			args: args{
				volName: fakeLocalVolumeName,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				name:                    fakeManagerName,
				namespace:               fakeNamespace,
				apiClient:               client,
				replicaRecords:          map[string]string{},
				volumeTaskQueue:         common.NewTaskQueue("VolumeTask", maxRetries),
				volumeReplicaTaskQueue:  common.NewTaskQueue("VolumeReplicaTask", maxRetries),
				localDiskClaimTaskQueue: common.NewTaskQueue("LocalDiskClaim", maxRetries),
				localDiskTaskQueue:      common.NewTaskQueue("localDisk", maxRetries),
				// healthCheckQueue:        common.NewTaskQueue("HealthCheckTask", maxRetries),
				diskEventQueue: diskmonitor.NewEventQueue("DiskEvents"),
				logger:         log.WithField("Module", "NodeManager"),
			}
			if err := m.processVolumeReplicaTaskAssignment(tt.args.volName); (err != nil) != tt.wantErr {
				t.Errorf("processVolumeReplicaTaskAssignment() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_updateVolumeReplica(t *testing.T) {
	type fields struct {
		name                    string
		namespace               string
		apiClient               client.Client
		informersCache          cache.Cache
		replicaRecords          map[string]string
		storageMgr              *storage.LocalManager
		diskEventQueue          *diskmonitor.EventQueue
		volumeTaskQueue         *common.TaskQueue
		volumeReplicaTaskQueue  *common.TaskQueue
		localDiskClaimTaskQueue *common.TaskQueue
		localDiskTaskQueue      *common.TaskQueue
		configManager           *configManager
		logger                  *log.Entry
	}
	type args struct {
		replica *apisv1alpha1.LocalVolumeReplica
		vol     *apisv1alpha1.LocalVolume
	}

	client, _ := CreateFakeClient()

	// Create LocalVolumeReplica
	lvr := GenFakeLocalVolumeReplicaObject()
	lvr.Name = fakeLocalVolumeReplicaName
	lvr.Spec.RequiredCapacityBytes = fakeDiskCapacityBytes
	err := client.Create(context.Background(), lvr)
	if err != nil {
		t.Errorf("Create LocalVolumeConvert fail %v", err)
	}

	// Create LocalVolume
	lv := GenFakeLocalVolumeObject()
	lv.Name = fakeLocalVolumeName
	lv.Spec.RequiredCapacityBytes = fakeDiskCapacityBytes
	err = client.Create(context.Background(), lv)
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
			args: args{
				replica: lvr,
				vol:     lv,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				name:                    fakeManagerName,
				namespace:               fakeNamespace,
				apiClient:               client,
				replicaRecords:          map[string]string{},
				volumeTaskQueue:         common.NewTaskQueue("VolumeTask", maxRetries),
				volumeReplicaTaskQueue:  common.NewTaskQueue("VolumeReplicaTask", maxRetries),
				localDiskClaimTaskQueue: common.NewTaskQueue("LocalDiskClaim", maxRetries),
				localDiskTaskQueue:      common.NewTaskQueue("localDisk", maxRetries),
				// healthCheckQueue:        common.NewTaskQueue("HealthCheckTask", maxRetries),
				diskEventQueue: diskmonitor.NewEventQueue("DiskEvents"),
				logger:         log.WithField("Module", "NodeManager"),
			}
			if err := m.updateVolumeReplica(tt.args.replica, tt.args.vol); (err != nil) != tt.wantErr {
				t.Errorf("updateVolumeReplica() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// GenFakeLocalDiskClaimObject Create ldc request
func GenFakeLocalDiskClaimObject() *apisv1alpha1.LocalDiskClaim {
	ldc := &apisv1alpha1.LocalDiskClaim{}

	TypeMeta := metav1.TypeMeta{
		Kind:       LocalDiskClaimKind,
		APIVersion: apiversion,
	}

	ObjectMata := metav1.ObjectMeta{
		Name:              fakeLocalDiskName,
		ResourceVersion:   "",
		UID:               types.UID(fakeLocalDiskUID),
		CreationTimestamp: metav1.Time{Time: time.Now()},
	}

	Spec := apisv1alpha1.LocalDiskClaimSpec{
		NodeName: fakeNodename,
		Description: apisv1alpha1.DiskClaimDescription{
			DiskType: "test",
			Capacity: fakeDiskCapacityBytes,
		},
	}

	ldc.ObjectMeta = ObjectMata
	ldc.TypeMeta = TypeMeta
	ldc.Spec = Spec
	ldc.Status.Status = apisv1alpha1.DiskClaimStatusEmpty

	return ldc
}

// GenFakeLocalDiskObject Create ld request
func GenFakeLocalDiskObject() *apisv1alpha1.LocalDisk {
	ld := &apisv1alpha1.LocalDisk{}

	TypeMeta := metav1.TypeMeta{
		Kind:       LocalDiskKind,
		APIVersion: apiversion,
	}

	ObjectMata := metav1.ObjectMeta{
		Name:              fakeLocalDiskName,
		ResourceVersion:   "",
		UID:               types.UID(fakeLocalDiskUID),
		CreationTimestamp: metav1.Time{Time: time.Now()},
	}

	Spec := apisv1alpha1.LocalDiskSpec{
		NodeName:       fakeNodename,
		UUID:           fakeLocalDiskUID,
		DevicePath:     fakeDevicePath,
		Capacity:       fakeDiskCapacityBytes,
		HasPartition:   true,
		PartitionInfo:  fakePartitionInfo,
		HasRAID:        true,
		RAIDInfo:       fakeRaidInfo,
		HasSmartInfo:   true,
		SmartInfo:      fakeSmartInfo,
		DiskAttributes: fakeDiskAttributes,
	}

	ld.ObjectMeta = ObjectMata
	ld.TypeMeta = TypeMeta
	ld.Spec = Spec
	ld.Status.State = apisv1alpha1.LocalDiskAvailable

	return ld
}

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

// CreateFakeClient Create LocalVolume resource
func CreateFakeClient() (client.Client, *runtime.Scheme) {

	lv := GenFakeLocalVolumeObject()
	lvList := &apisv1alpha1.LocalVolumeList{
		TypeMeta: metav1.TypeMeta{
			Kind:       LocalVolumeKind,
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

	ld := GenFakeLocalDiskObject()
	ldList := &apisv1alpha1.LocalDiskList{
		TypeMeta: metav1.TypeMeta{
			Kind:       LocalDiskKind,
			APIVersion: apiversion,
		},
	}

	ldc := GenFakeLocalDiskClaimObject()
	ldcList := &apisv1alpha1.LocalDiskClaimList{
		TypeMeta: metav1.TypeMeta{
			Kind:       LocalDiskClaimKind,
			APIVersion: apiversion,
		},
	}

	s := scheme.Scheme
	s.AddKnownTypes(apisv1alpha1.SchemeGroupVersion, lv)
	s.AddKnownTypes(apisv1alpha1.SchemeGroupVersion, lvList)
	s.AddKnownTypes(apisv1alpha1.SchemeGroupVersion, lvr)
	s.AddKnownTypes(apisv1alpha1.SchemeGroupVersion, lvrList)
	s.AddKnownTypes(apisv1alpha1.SchemeGroupVersion, ld)
	s.AddKnownTypes(apisv1alpha1.SchemeGroupVersion, ldList)
	s.AddKnownTypes(apisv1alpha1.SchemeGroupVersion, ldc)
	s.AddKnownTypes(apisv1alpha1.SchemeGroupVersion, ldcList)

	return fake.NewFakeClientWithScheme(s), s
}
