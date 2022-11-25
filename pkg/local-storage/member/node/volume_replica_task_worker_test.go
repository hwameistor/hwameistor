package node

import (
	"context"
	"testing"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-storage/common"
	"github.com/hwameistor/hwameistor/pkg/local-storage/member/node/diskmonitor"
	"github.com/hwameistor/hwameistor/pkg/local-storage/member/node/storage"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Test_manager_processVolumeReplica(t *testing.T) {
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
		replicaName string
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
				replicaName: fakeLocalVolumeReplicaName,
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
			if err := m.processVolumeReplica(tt.args.replicaName); (err != nil) != tt.wantErr {
				t.Errorf("processVolumeReplica() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_processVolumeReplicaCheck(t *testing.T) {
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
		//{
		//	args: args{
		//		replica: lvr,
		//	},
		//},
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
			if err := m.processVolumeReplicaCheck(tt.args.replica); (err != nil) != tt.wantErr {
				t.Errorf("processVolumeReplicaCheck() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_processVolumeReplicaCleanup(t *testing.T) {
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
			if err := m.processVolumeReplicaCleanup(tt.args.replica); (err != nil) != tt.wantErr {
				t.Errorf("processVolumeReplicaCleanup() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_processVolumeReplicaCreate(t *testing.T) {
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
		//{
		//	args: args{
		//		replica: lvr,
		//	},
		//},
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
			if err := m.processVolumeReplicaCreate(tt.args.replica); (err != nil) != tt.wantErr {
				t.Errorf("processVolumeReplicaCreate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_processVolumeReplicaDelete(t *testing.T) {
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
		//{
		//	args: args{
		//		replica: lvr,
		//	},
		//},
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
			if err := m.processVolumeReplicaDelete(tt.args.replica); (err != nil) != tt.wantErr {
				t.Errorf("processVolumeReplicaDelete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_processVolumeReplicaSubmit(t *testing.T) {
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
			if err := m.processVolumeReplicaSubmit(tt.args.replica); (err != nil) != tt.wantErr {
				t.Errorf("processVolumeReplicaSubmit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
