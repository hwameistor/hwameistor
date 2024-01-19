package controller

import (
	"context"
	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-storage/common"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
	"testing"
)

func Test_manager_processVolumeSnapshot(t *testing.T) {
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
		volumeSnapshotTaskQueue     *common.TaskQueue
		localNodes                  map[string]v1alpha1.State
		logger                      *log.Entry
		lock                        sync.Mutex
	}
	type args struct {
		name string
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

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = fakeNamespace
	err = client.Create(context.Background(), lvg)
	if err != nil {
		t.Errorf("Create LocalVolumeGroup fail %v", err)
	}

	//Create LocalVolumeSnapshot
	lvs := GenFakeLocalVolumeSnapShotObject()
	lvs.Namespace = fakeNamespace
	err = client.Create(context.Background(), lvs)
	if err != nil {
		t.Errorf("Create LocalVolumeSnapshot fail %v", err)
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
				name: fakeLocalVolumeSnapshotName,
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
				volumeSnapshotTaskQueue:     common.NewTaskQueue("VolumeSnapshotTask", maxRetries),
				localNodes:                  map[string]v1alpha1.State{},
				logger:                      log.WithField("Module", "ControllerManager"),
			}
			if err := m.processVolumeSnapshot(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("processVolumeSnapshot() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

}

func Test_manager_volumeSnapshotSubmit(t *testing.T) {
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
		volumeSnapshotTaskQueue     *common.TaskQueue
		localNodes                  map[string]v1alpha1.State
		logger                      *log.Entry
		lock                        sync.Mutex
	}
	type args struct {
		snapShot *v1alpha1.LocalVolumeSnapshot
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

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = fakeNamespace
	err = client.Create(context.Background(), lvg)
	if err != nil {
		t.Errorf("Create LocalVolumeGroup fail %v", err)
	}

	// Create LocalVolumeConvert
	lvs := GenFakeLocalVolumeSnapShotObject()
	lvs.Namespace = fakeNamespace

	err = client.Create(context.Background(), lvs)
	if err != nil {
		t.Errorf("Create VolumeSnapshot fail %v", err)
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
				snapShot: lvs,
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
				volumeSnapshotTaskQueue:     common.NewTaskQueue("VolumeSnapshotTask", maxRetries),
				volumeGroupMigrateTaskQueue: common.NewTaskQueue("VolumeGroupMigrateTask", maxRetries),
				volumeConvertTaskQueue:      common.NewTaskQueue("VolumeConvertTask", maxRetries),
				volumeGroupConvertTaskQueue: common.NewTaskQueue("VolumeGroupConvertTask", maxRetries),
				localNodes:                  map[string]v1alpha1.State{},
				logger:                      log.WithField("Module", "ControllerManager"),
			}
			if err := m.volumeSnapshotSubmit(tt.args.snapShot); (err != nil) != tt.wantErr {
				t.Errorf("volumeMigrateSubmit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_volumeSnapshotCreate(t *testing.T) {
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
		volumeSnapshotTaskQueue     *common.TaskQueue
		localNodes                  map[string]v1alpha1.State
		logger                      *log.Entry
		lock                        sync.Mutex
	}
	type args struct {
		snapShot *v1alpha1.LocalVolumeSnapshot
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

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = fakeNamespace
	err = client.Create(context.Background(), lvg)
	if err != nil {
		t.Errorf("Create LocalVolumeGroup fail %v", err)
	}

	// Create LocalVolumeConvert
	lvs := GenFakeLocalVolumeSnapShotObject()
	lvs.Namespace = fakeNamespace

	err = client.Create(context.Background(), lvs)
	if err != nil {
		t.Errorf("Create VolumeSnapshot fail %v", err)
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
				snapShot: lvs,
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
				volumeSnapshotTaskQueue:     common.NewTaskQueue("VolumeSnapshotTask", maxRetries),
				volumeGroupMigrateTaskQueue: common.NewTaskQueue("VolumeGroupMigrateTask", maxRetries),
				volumeConvertTaskQueue:      common.NewTaskQueue("VolumeConvertTask", maxRetries),
				volumeGroupConvertTaskQueue: common.NewTaskQueue("VolumeGroupConvertTask", maxRetries),
				localNodes:                  map[string]v1alpha1.State{},
				logger:                      log.WithField("Module", "ControllerManager"),
			}
			if err := m.volumeSnapshotCreate(tt.args.snapShot); (err != nil) != tt.wantErr {
				t.Errorf("volumeMigrateSubmit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_volumeSnapshotDelete(t *testing.T) {
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
		volumeSnapshotTaskQueue     *common.TaskQueue
		localNodes                  map[string]v1alpha1.State
		logger                      *log.Entry
		lock                        sync.Mutex
	}
	type args struct {
		snapShot *v1alpha1.LocalVolumeSnapshot
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

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = fakeNamespace
	err = client.Create(context.Background(), lvg)
	if err != nil {
		t.Errorf("Create LocalVolumeGroup fail %v", err)
	}

	// Create LocalVolumeConvert
	lvs := GenFakeLocalVolumeSnapShotObject()
	lvs.Namespace = fakeNamespace

	err = client.Create(context.Background(), lvs)
	if err != nil {
		t.Errorf("Create VolumeSnapshot fail %v", err)
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
				snapShot: lvs,
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
				volumeSnapshotTaskQueue:     common.NewTaskQueue("VolumeSnapshotTask", maxRetries),
				volumeConvertTaskQueue:      common.NewTaskQueue("VolumeConvertTask", maxRetries),
				volumeGroupConvertTaskQueue: common.NewTaskQueue("VolumeGroupConvertTask", maxRetries),
				localNodes:                  map[string]v1alpha1.State{},
				logger:                      log.WithField("Module", "ControllerManager"),
			}
			if err := m.volumeSnapshotDelete(tt.args.snapShot); (err != nil) != tt.wantErr {
				t.Errorf("volumeMigrateSubmit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_volumeSnapshotCleanup(t *testing.T) {
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
		volumeSnapshotTaskQueue     *common.TaskQueue
		localNodes                  map[string]v1alpha1.State
		logger                      *log.Entry
		lock                        sync.Mutex
	}
	type args struct {
		snapShot *v1alpha1.LocalVolumeSnapshot
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

	// Create LocalVolumeGroup
	lvg := GenFakeLocalVolumeGroupObject()
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = fakeNamespace
	err = client.Create(context.Background(), lvg)
	if err != nil {
		t.Errorf("Create LocalVolumeGroup fail %v", err)
	}

	// Create LocalVolumeConvert
	lvs := GenFakeLocalVolumeSnapShotObject()
	lvs.Namespace = fakeNamespace

	err = client.Create(context.Background(), lvs)
	if err != nil {
		t.Errorf("Create VolumeSnapshot fail %v", err)
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
				snapShot: lvs,
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
				volumeSnapshotTaskQueue:     common.NewTaskQueue("VolumeSnapshotTask", maxRetries),
				volumeGroupMigrateTaskQueue: common.NewTaskQueue("VolumeGroupMigrateTask", maxRetries),
				volumeConvertTaskQueue:      common.NewTaskQueue("VolumeConvertTask", maxRetries),
				volumeGroupConvertTaskQueue: common.NewTaskQueue("VolumeGroupConvertTask", maxRetries),
				localNodes:                  map[string]v1alpha1.State{},
				logger:                      log.WithField("Module", "ControllerManager"),
			}
			if err := m.volumeSnapshotCleanup(tt.args.snapShot); (err != nil) != tt.wantErr {
				t.Errorf("volumeMigrateSubmit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
