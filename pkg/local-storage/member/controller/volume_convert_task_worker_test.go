package controller

import (
	"context"
	"sync"
	"testing"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-storage/common"
)

func Test_manager_processVolumeConvert(t *testing.T) {
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
		name string
	}

	client, _ := CreateFakeClient()

	lvc := GenFakeLocalVolumeConvertObject()
	lvc.Name = fakeLocalVolumeConvertName
	lvc.Namespace = fakeNamespace
	lvc.Spec.VolumeName = fakeLocalVolumeName
	lvc.Spec.ReplicaNumber = 2
	err := client.Create(context.Background(), lvc)
	if err != nil {
		t.Errorf("Create LocalVolumeConvert fail %v", err)
	}

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = ""
	err = client.Create(context.Background(), lvg)
	if err != nil {
		t.Errorf("Create LocalVolumeGroup fail %v", err)
	}

	// Create LocalVolume
	lv := GenFakeLocalVolumeObject()
	lv.Name = fakeLocalVolumeName
	lv.Spec.VolumeGroup = fakeLocalVolumeGroupName
	lv.Spec.PersistentVolumeClaimName = fakePersistentVolumeClaimName
	lv.Spec.ReplicaNumber = 1
	lv.Spec.Convertible = true
	lv.Spec.PersistentVolumeClaimNamespace = fakeNamespace

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
				name: fakeLocalVolumeConvertName,
			},
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
				localNodes:                  map[string]v1alpha1.State{},
				logger:                      log.WithField("Module", "ControllerManager"),
			}
			if err := m.processVolumeConvert(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("processVolumeConvert() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_volumeConvertAbort(t *testing.T) {
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
		convert *v1alpha1.LocalVolumeConvert
	}

	client, _ := CreateFakeClient()

	lvc := GenFakeLocalVolumeConvertObject()
	lvc.Name = fakeLocalVolumeConvertName
	lvc.Namespace = fakeNamespace
	lvc.Spec.VolumeName = fakeLocalVolumeName
	lvc.Spec.ReplicaNumber = 2
	err := client.Create(context.Background(), lvc)
	if err != nil {
		t.Errorf("Create LocalVolumeConvert fail %v", err)
	}

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = ""
	err = client.Create(context.Background(), lvg)
	if err != nil {
		t.Errorf("Create LocalVolumeGroup fail %v", err)
	}

	// Create LocalVolume
	lv := GenFakeLocalVolumeObject()
	lv.Name = fakeLocalVolumeName
	lv.Spec.VolumeGroup = fakeLocalVolumeGroupName
	lv.Spec.PersistentVolumeClaimName = fakePersistentVolumeClaimName
	lv.Spec.ReplicaNumber = 1
	lv.Spec.Convertible = true
	lv.Spec.PersistentVolumeClaimNamespace = fakeNamespace

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
				convert: lvc,
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
				localNodes:                  map[string]v1alpha1.State{},
				logger:                      log.WithField("Module", "ControllerManager"),
			}
			if err := m.volumeConvertAbort(tt.args.convert); (err != nil) != tt.wantErr {
				t.Errorf("volumeConvertAbort() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_volumeConvertCleanup(t *testing.T) {
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
		convert *v1alpha1.LocalVolumeConvert
	}

	client, _ := CreateFakeClient()

	lvc := GenFakeLocalVolumeConvertObject()
	lvc.Name = fakeLocalVolumeConvertName
	lvc.Namespace = fakeNamespace
	lvc.Spec.VolumeName = fakeLocalVolumeName
	lvc.Spec.ReplicaNumber = 2
	err := client.Create(context.Background(), lvc)
	if err != nil {
		t.Errorf("Create LocalVolumeConvert fail %v", err)
	}

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = ""
	err = client.Create(context.Background(), lvg)
	if err != nil {
		t.Errorf("Create LocalVolumeGroup fail %v", err)
	}

	// Create LocalVolume
	lv := GenFakeLocalVolumeObject()
	lv.Name = fakeLocalVolumeName
	lv.Spec.VolumeGroup = fakeLocalVolumeGroupName
	lv.Spec.PersistentVolumeClaimName = fakePersistentVolumeClaimName
	lv.Spec.ReplicaNumber = 1
	lv.Spec.Convertible = true
	lv.Spec.PersistentVolumeClaimNamespace = fakeNamespace

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
				convert: lvc,
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
				localNodes:                  map[string]v1alpha1.State{},
				logger:                      log.WithField("Module", "ControllerManager"),
			}
			if err := m.volumeConvertCleanup(tt.args.convert); (err != nil) != tt.wantErr {
				t.Errorf("volumeConvertCleanup() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_volumeConvertInProgress(t *testing.T) {
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
		convert *v1alpha1.LocalVolumeConvert
	}

	client, _ := CreateFakeClient()

	lvc := GenFakeLocalVolumeConvertObject()
	lvc.Name = fakeLocalVolumeConvertName
	lvc.Namespace = fakeNamespace
	lvc.Spec.VolumeName = fakeLocalVolumeName
	lvc.Spec.ReplicaNumber = 2
	err := client.Create(context.Background(), lvc)
	if err != nil {
		t.Errorf("Create LocalVolumeConvert fail %v", err)
	}

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = ""
	err = client.Create(context.Background(), lvg)
	if err != nil {
		t.Errorf("Create LocalVolumeGroup fail %v", err)
	}

	// Create LocalVolume
	lv := GenFakeLocalVolumeObject()
	lv.Name = fakeLocalVolumeName
	lv.Spec.VolumeGroup = fakeLocalVolumeGroupName
	lv.Spec.PersistentVolumeClaimName = fakePersistentVolumeClaimName
	lv.Spec.ReplicaNumber = 1
	lv.Spec.Convertible = true
	lv.Spec.PersistentVolumeClaimNamespace = fakeNamespace

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
				convert: lvc,
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
				localNodes:                  map[string]v1alpha1.State{},
				logger:                      log.WithField("Module", "ControllerManager"),
			}
			if err := m.volumeConvertInProgress(tt.args.convert); (err != nil) != tt.wantErr {
				t.Errorf("volumeConvertInProgress() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_volumeConvertStart(t *testing.T) {
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
		convert *v1alpha1.LocalVolumeConvert
	}

	client, _ := CreateFakeClient()

	lvc := GenFakeLocalVolumeConvertObject()
	lvc.Name = fakeLocalVolumeConvertName
	lvc.Namespace = fakeNamespace
	lvc.Spec.VolumeName = fakeLocalVolumeName
	lvc.Spec.ReplicaNumber = 2
	err := client.Create(context.Background(), lvc)
	if err != nil {
		t.Errorf("Create LocalVolumeConvert fail %v", err)
	}

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = ""
	err = client.Create(context.Background(), lvg)
	if err != nil {
		t.Errorf("Create LocalVolumeGroup fail %v", err)
	}

	// Create LocalVolume
	lv := GenFakeLocalVolumeObject()
	lv.Name = fakeLocalVolumeName
	lv.Spec.VolumeGroup = fakeLocalVolumeGroupName
	lv.Spec.PersistentVolumeClaimName = fakePersistentVolumeClaimName
	lv.Spec.ReplicaNumber = 1
	lv.Spec.Convertible = true
	lv.Spec.PersistentVolumeClaimNamespace = fakeNamespace

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
				convert: lvc,
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
				localNodes:                  map[string]v1alpha1.State{},
				logger:                      log.WithField("Module", "ControllerManager"),
			}
			if err := m.volumeConvertStart(tt.args.convert); (err != nil) != tt.wantErr {
				t.Errorf("volumeConvertStart() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_volumeConvertSubmit(t *testing.T) {
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
		convert *v1alpha1.LocalVolumeConvert
	}

	client, _ := CreateFakeClient()

	lvc := GenFakeLocalVolumeConvertObject()
	lvc.Name = fakeLocalVolumeConvertName
	lvc.Namespace = fakeNamespace
	lvc.Spec.VolumeName = fakeLocalVolumeName
	lvc.Spec.ReplicaNumber = 2
	err := client.Create(context.Background(), lvc)
	if err != nil {
		t.Errorf("Create LocalVolumeConvert fail %v", err)
	}

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = ""
	err = client.Create(context.Background(), lvg)
	if err != nil {
		t.Errorf("Create LocalVolumeGroup fail %v", err)
	}

	// Create LocalVolume
	lv := GenFakeLocalVolumeObject()
	lv.Name = fakeLocalVolumeName
	lv.Spec.VolumeGroup = fakeLocalVolumeGroupName
	lv.Spec.PersistentVolumeClaimName = fakePersistentVolumeClaimName
	lv.Spec.ReplicaNumber = 1
	lv.Spec.Convertible = true
	lv.Spec.PersistentVolumeClaimNamespace = fakeNamespace

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
				convert: lvc,
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
				localNodes:                  map[string]v1alpha1.State{},
				logger:                      log.WithField("Module", "ControllerManager"),
			}
			if err := m.volumeConvertSubmit(tt.args.convert); (err != nil) != tt.wantErr {
				t.Errorf("volumeConvertSubmit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
