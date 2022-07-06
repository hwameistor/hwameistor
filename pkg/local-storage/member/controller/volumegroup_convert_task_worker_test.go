package controller

import (
	"context"
	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/local-storage/v1alpha1"
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/local-storage/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-storage/common"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
	"testing"
)

func Test_manager_VolumeGroupConvertCleanup(t *testing.T) {
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
		convert *apisv1alpha1.LocalVolumeGroupConvert
	}
	client, _ := CreateFakeClient()

	var convert = &apisv1alpha1.LocalVolumeGroupConvert{}
	convert.Name = fakeLocalVolumeGroupConvertName
	convert.Namespace = fakeNamespace
	convert.Spec.LocalVolumeGroupName = fakeLocalVolumeGroupName
	convert.Spec.ReplicaNumber = 2

	// Create LocalVolumeGroupConvert
	err := client.Create(context.Background(), convert)
	if err != nil {
		t.Errorf("Create LocalVolumeGroupConvert fail %v", err)
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
				convert: convert,
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
			if err := m.VolumeGroupConvertCleanup(tt.args.convert); (err != nil) != tt.wantErr {
				t.Errorf("VolumeGroupConvertCleanup() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_VolumeGroupConvertFailed(t *testing.T) {
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
		convert *apisv1alpha1.LocalVolumeGroupConvert
	}

	client, _ := CreateFakeClient()

	lvgc := GenFakeLocalVolumeGroupConvertObject()
	lvgc.Name = fakeLocalVolumeGroupConvertName
	lvgc.Namespace = fakeNamespace
	lvgc.Spec.LocalVolumeGroupName = fakeLocalVolumeGroupName
	lvgc.Spec.ReplicaNumber = 2
	err := client.Create(context.Background(), lvgc)
	if err != nil {
		t.Errorf("Create LocalVolumeGroupConvert fail %v", err)
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
				convert: lvgc,
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
			if err := m.VolumeGroupConvertFailed(tt.args.convert); (err != nil) != tt.wantErr {
				t.Errorf("VolumeGroupConvertFailed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_VolumeGroupConvertInProgress(t *testing.T) {
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
		convert *apisv1alpha1.LocalVolumeGroupConvert
	}

	client, _ := CreateFakeClient()

	lvgc := GenFakeLocalVolumeGroupConvertObject()
	lvgc.Name = fakeLocalVolumeGroupConvertName
	lvgc.Namespace = fakeNamespace
	lvgc.Spec.LocalVolumeGroupName = fakeLocalVolumeGroupName
	lvgc.Spec.ReplicaNumber = 2
	err := client.Create(context.Background(), lvgc)
	if err != nil {
		t.Errorf("Create LocalVolumeGroupConvert fail %v", err)
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
				convert: lvgc,
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
			if err := m.VolumeGroupConvertInProgress(tt.args.convert); (err != nil) != tt.wantErr {
				t.Errorf("VolumeGroupConvertInProgress() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_VolumeGroupConvertStart(t *testing.T) {
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
		convert *apisv1alpha1.LocalVolumeGroupConvert
	}
	client, _ := CreateFakeClient()

	lvgc := GenFakeLocalVolumeGroupConvertObject()
	lvgc.Name = fakeLocalVolumeGroupConvertName
	lvgc.Namespace = fakeNamespace
	lvgc.Spec.LocalVolumeGroupName = fakeLocalVolumeGroupName
	lvgc.Spec.ReplicaNumber = 2
	err := client.Create(context.Background(), lvgc)
	if err != nil {
		t.Errorf("Create LocalVolumeGroupConvert fail %v", err)
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
		{
			args: args{
				convert: lvgc,
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
			if err := m.VolumeGroupConvertStart(tt.args.convert); (err != nil) != tt.wantErr {
				t.Errorf("VolumeGroupConvertStart() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_VolumeGroupConvertSubmit(t *testing.T) {
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
		convert *apisv1alpha1.LocalVolumeGroupConvert
	}

	client, _ := CreateFakeClient()

	lvgc := GenFakeLocalVolumeGroupConvertObject()
	lvgc.Name = fakeLocalVolumeGroupConvertName
	lvgc.Namespace = fakeNamespace
	lvgc.Spec.LocalVolumeGroupName = fakeLocalVolumeGroupName
	lvgc.Spec.ReplicaNumber = 2
	err := client.Create(context.Background(), lvgc)
	if err != nil {
		t.Errorf("Create LocalVolumeGroupConvert fail %v", err)
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
				convert: lvgc,
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
			if err := m.VolumeGroupConvertSubmit(tt.args.convert); (err != nil) != tt.wantErr {
				t.Errorf("VolumeGroupConvertSubmit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_processVolumeGroupConvert(t *testing.T) {
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
		name string
	}

	client, _ := CreateFakeClient()

	lvgc := GenFakeLocalVolumeGroupConvertObject()
	lvgc.Name = fakeLocalVolumeGroupConvertName
	lvgc.Namespace = fakeNamespace
	lvgc.Spec.LocalVolumeGroupName = fakeLocalVolumeGroupName
	lvgc.Spec.ReplicaNumber = 2
	err := client.Create(context.Background(), lvgc)
	if err != nil {
		t.Errorf("Create LocalVolumeGroupConvert fail %v", err)
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
				name: fakeLocalVolumeGroupConvertName,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				name:                        "test",
				namespace:                   "test",
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
			if err := m.processVolumeGroupConvert(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("processVolumeGroupConvert() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
