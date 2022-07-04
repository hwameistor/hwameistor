package controller

import (
	"context"
	"github.com/hwameistor/local-storage/pkg/apis/hwameistor/v1alpha1"
	apisv1alpha1 "github.com/hwameistor/local-storage/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/local-storage/pkg/common"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
	"testing"
)

func Test_manager_VolumeGroupMigrateAbort(t *testing.T) {
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
		localNodes                  map[string]apisv1alpha1.State
		logger                      *log.Entry
		lock                        sync.Mutex
	}
	type args struct {
		migrate *apisv1alpha1.LocalVolumeGroupMigrate
	}

	client, _ := CreateFakeClient()

	// Create LocalVolumeConvert
	lvc := GenFakeLocalVolumeConvertObject()
	lvc.Name = fakeLocalVolumeConvertName
	err := client.Create(context.Background(), lvc)
	if err != nil {
		t.Errorf("Create LocalVolumeConvert fail %v", err)
	}

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = fakeNamespace
	err = client.Create(context.Background(), lvg)
	if err != nil {
		t.Errorf("Create LocalVolumeGroup fail %v", err)
	}

	// Create LocalVolumeGroupMigrate
	lvgm := GenFakeLocalVolumeGroupMigrateObject()
	lvgm.Name = fakeLocalVolumeGroupMigrateName
	lvgm.Namespace = fakeNamespace
	lvgm.Spec.LocalVolumeGroupName = fakeLocalVolumeGroupName
	lvgm.Spec.SourceNodesNames = fakeNodenames
	lvgm.Spec.TargetNodesNames = fakeNodenames
	err = client.Create(context.Background(), lvgm)
	if err != nil {
		t.Errorf("Create LocalVolumeGroupMigrate fail %v", err)
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
				migrate: lvgm,
			},
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
				localNodes:                  map[string]apisv1alpha1.State{},
				logger:                      log.WithField("Module", "ControllerManager"),
			}
			if err := m.VolumeGroupMigrateAbort(tt.args.migrate); (err != nil) != tt.wantErr {
				t.Errorf("VolumeGroupMigrateAbort() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_VolumeGroupMigrateCleanup(t *testing.T) {
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
		localNodes                  map[string]apisv1alpha1.State
		logger                      *log.Entry
		lock                        sync.Mutex
	}
	type args struct {
		migrate *apisv1alpha1.LocalVolumeGroupMigrate
	}
	client, _ := CreateFakeClient()

	// Create LocalVolumeGroupMigrate
	lvgm := GenFakeLocalVolumeGroupMigrateObject()
	lvgm.Name = fakeLocalVolumeGroupMigrateName
	lvgm.Namespace = fakeNamespace
	err := client.Create(context.Background(), lvgm)
	if err != nil {
		t.Errorf("Create LocalVolumeGroupMigrate fail %v", err)
	}

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = fakeNamespace
	err = client.Create(context.Background(), lvg)
	if err != nil {
		t.Errorf("Create LocalVolumeGroup fail %v", err)
	}

	var migrate = &apisv1alpha1.LocalVolumeGroupMigrate{}
	migrate.Name = fakeLocalVolumeGroupMigrateName
	migrate.Namespace = fakeNamespace
	migrate.Spec.LocalVolumeGroupName = fakeLocalVolumeGroupName
	migrate.Spec.SourceNodesNames = fakeNodenames
	migrate.Spec.TargetNodesNames = fakeNodenames

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args: args{
				migrate: migrate,
			},
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
				localNodes:                  map[string]apisv1alpha1.State{},
				logger:                      log.WithField("Module", "ControllerManager"),
			}
			if err := m.VolumeGroupMigrateCleanup(tt.args.migrate); (err != nil) != tt.wantErr {
				t.Errorf("VolumeGroupMigrateCleanup() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_VolumeGroupMigrateInProgress(t *testing.T) {
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
		localNodes                  map[string]apisv1alpha1.State
		logger                      *log.Entry
		lock                        sync.Mutex
	}
	type args struct {
		migrate *apisv1alpha1.LocalVolumeGroupMigrate
	}

	client, _ := CreateFakeClient()

	// Create LocalVolumeGroupMigrate
	lvgm := GenFakeLocalVolumeGroupMigrateObject()
	lvgm.Name = fakeLocalVolumeGroupMigrateName
	lvgm.Namespace = fakeNamespace
	err := client.Create(context.Background(), lvgm)
	if err != nil {
		t.Errorf("Create LocalVolumeGroupMigrate fail %v", err)
	}

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = fakeNamespace
	err = client.Create(context.Background(), lvg)
	if err != nil {
		t.Errorf("Create LocalVolumeGroup fail %v", err)
	}

	var migrate = &apisv1alpha1.LocalVolumeGroupMigrate{}
	migrate.Name = fakeLocalVolumeGroupMigrateName
	migrate.Namespace = fakeNamespace
	migrate.Spec.LocalVolumeGroupName = fakeLocalVolumeGroupName
	migrate.Spec.SourceNodesNames = fakeNodenames
	migrate.Spec.TargetNodesNames = fakeNodenames

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args: args{
				migrate: migrate,
			},
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
				localNodes:                  map[string]apisv1alpha1.State{},
				logger:                      log.WithField("Module", "ControllerManager"),
			}
			if err := m.VolumeGroupMigrateInProgress(tt.args.migrate); (err != nil) != tt.wantErr {
				t.Errorf("VolumeGroupMigrateInProgress() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_VolumeGroupMigrateStart(t *testing.T) {
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
		localNodes                  map[string]apisv1alpha1.State
		logger                      *log.Entry
		lock                        sync.Mutex
	}
	type args struct {
		migrate *apisv1alpha1.LocalVolumeGroupMigrate
	}

	client, _ := CreateFakeClient()

	// Create LocalVolumeGroupMigrate
	lvgm := GenFakeLocalVolumeGroupMigrateObject()
	lvgm.Name = fakeLocalVolumeGroupMigrateName
	lvgm.Namespace = fakeNamespace
	err := client.Create(context.Background(), lvgm)
	if err != nil {
		t.Errorf("Create LocalVolumeGroupMigrate fail %v", err)
	}

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = fakeNamespace
	err = client.Create(context.Background(), lvg)
	if err != nil {
		t.Errorf("Create LocalVolumeGroup fail %v", err)
	}

	var migrate = &apisv1alpha1.LocalVolumeGroupMigrate{}
	migrate.Name = fakeLocalVolumeGroupMigrateName
	migrate.Namespace = fakeNamespace
	migrate.Spec.LocalVolumeGroupName = fakeLocalVolumeGroupName
	migrate.Spec.SourceNodesNames = fakeNodenames
	migrate.Spec.TargetNodesNames = fakeNodenames

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args: args{
				migrate: migrate,
			},
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
				localNodes:                  map[string]apisv1alpha1.State{},
				logger:                      log.WithField("Module", "ControllerManager"),
			}
			if err := m.VolumeGroupMigrateStart(tt.args.migrate); (err != nil) != tt.wantErr {
				t.Errorf("VolumeGroupMigrateStart() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_VolumeGroupMigrateSubmit(t *testing.T) {
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
		localNodes                  map[string]apisv1alpha1.State
		logger                      *log.Entry
		lock                        sync.Mutex
	}
	type args struct {
		migrate *apisv1alpha1.LocalVolumeGroupMigrate
	}

	client, _ := CreateFakeClient()

	// Create LocalVolumeGroupMigrate
	lvgm := GenFakeLocalVolumeGroupMigrateObject()
	lvgm.Name = fakeLocalVolumeGroupMigrateName
	lvgm.Namespace = fakeNamespace
	err := client.Create(context.Background(), lvgm)
	if err != nil {
		t.Errorf("Create LocalVolumeGroupMigrate fail %v", err)
	}

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = fakeNamespace
	err = client.Create(context.Background(), lvg)
	if err != nil {
		t.Errorf("Create LocalVolumeGroup fail %v", err)
	}

	var migrate = &apisv1alpha1.LocalVolumeGroupMigrate{}
	migrate.Name = fakeLocalVolumeGroupMigrateName
	migrate.Namespace = fakeNamespace
	migrate.Spec.LocalVolumeGroupName = fakeLocalVolumeGroupName
	migrate.Spec.SourceNodesNames = fakeNodenames
	migrate.Spec.TargetNodesNames = fakeNodenames

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args: args{
				migrate: migrate,
			},
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
				localNodes:                  map[string]apisv1alpha1.State{},
				logger:                      log.WithField("Module", "ControllerManager"),
			}
			if err := m.VolumeGroupMigrateSubmit(tt.args.migrate); (err != nil) != tt.wantErr {
				t.Errorf("VolumeGroupMigrateSubmit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_processVolumeGroupMigrate(t *testing.T) {
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
		localNodes                  map[string]apisv1alpha1.State
		logger                      *log.Entry
		lock                        sync.Mutex
	}
	type args struct {
		vgmNamespacedName string
	}

	client, _ := CreateFakeClient()

	// Create LocalVolumeGroupMigrate
	lvgm := GenFakeLocalVolumeGroupMigrateObject()
	lvgm.Name = fakeLocalVolumeGroupMigrateName
	lvgm.Namespace = fakeNamespace
	lvgm.Spec.LocalVolumeGroupName = fakeLocalVolumeGroupName
	lvgm.Spec.SourceNodesNames = fakeNodenames
	lvgm.Spec.TargetNodesNames = fakeNodenames
	err := client.Create(context.Background(), lvgm)
	if err != nil {
		t.Errorf("Create LocalVolumeGroupMigrate fail %v", err)
	}

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = fakeNamespace
	err = client.Create(context.Background(), lvg)
	if err != nil {
		t.Errorf("Create LocalVolumeGroup fail %v", err)
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
				vgmNamespacedName: fakeVgmNamespacedName,
			},
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
				localNodes:                  map[string]apisv1alpha1.State{},
				logger:                      log.WithField("Module", "ControllerManager"),
			}
			if err := m.processVolumeGroupMigrate(tt.args.vgmNamespacedName); (err != nil) != tt.wantErr {
				t.Errorf("processVolumeGroupMigrate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
