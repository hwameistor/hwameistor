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

func Test_manager_processVolumeMigrate(t *testing.T) {
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
	lvm := GenFakeLocalVolumeMigrateObject()
	lvm.Name = fakeLocalVolumeMigrateName
	lvm.Namespace = fakeNamespace
	lvm.Spec.VolumeName = fakeLocalVolumeName
	lvm.Spec.TargetNodesSuggested = fakeNodenames
	err = client.Create(context.Background(), lvm)
	if err != nil {
		t.Errorf("Create LocalVolumeMigrate fail %v", err)
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
				name: fakeLocalVolumeMigrateName,
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
				localNodes:                  map[string]v1alpha1.State{},
				logger:                      log.WithField("Module", "ControllerManager"),
			}
			if err := m.processVolumeMigrate(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("processVolumeMigrate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_volumeMigrateAbort(t *testing.T) {
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
		migrate *v1alpha1.LocalVolumeMigrate
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
	lvm := GenFakeLocalVolumeMigrateObject()
	lvm.Name = fakeLocalVolumeMigrateName
	lvm.Namespace = fakeNamespace
	lvm.Spec.VolumeName = fakeLocalVolumeName
	lvm.Spec.TargetNodesSuggested = fakeNodenames
	err = client.Create(context.Background(), lvm)
	if err != nil {
		t.Errorf("Create LocalVolumeMigrate fail %v", err)
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
				migrate: lvm,
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
				localNodes:                  map[string]v1alpha1.State{},
				logger:                      log.WithField("Module", "ControllerManager"),
			}
			if err := m.volumeMigrateAbort(tt.args.migrate); (err != nil) != tt.wantErr {
				t.Errorf("volumeMigrateAbort() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_volumeMigrateCleanup(t *testing.T) {
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
		migrate *v1alpha1.LocalVolumeMigrate
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
	lvm := GenFakeLocalVolumeMigrateObject()
	lvm.Name = fakeLocalVolumeMigrateName
	lvm.Namespace = fakeNamespace
	lvm.Spec.VolumeName = fakeLocalVolumeName
	lvm.Spec.TargetNodesSuggested = fakeNodenames
	err = client.Create(context.Background(), lvm)
	if err != nil {
		t.Errorf("Create LocalVolumeMigrate fail %v", err)
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
				migrate: lvm,
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
				localNodes:                  map[string]v1alpha1.State{},
				logger:                      log.WithField("Module", "ControllerManager"),
			}
			if err := m.volumeMigrateCleanup(tt.args.migrate); (err != nil) != tt.wantErr {
				t.Errorf("volumeMigrateCleanup() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_volumeMigrateInProgress(t *testing.T) {
	type args struct {
		migrate *v1alpha1.LocalVolumeMigrate
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
	lsn := GenFakeLocalStorageNodeObject()
	lsn.ObjectMeta.Name = fakeLocalStorageNodename
	lsn.ObjectMeta.Namespace = ""
	err = client.Create(context.Background(), lsn)
	if err != nil {
		t.Errorf("Create LocalStorageNode fail %v", err)
	}
	// Create LocalVolumeConvert
	lvm := GenFakeLocalVolumeMigrateObject()
	lvm.Name = fakeLocalVolumeMigrateName
	lvm.Namespace = fakeNamespace
	lvm.Spec.VolumeName = fakeLocalVolumeName
	lvm.Spec.TargetNodesSuggested = fakeNodenames
	lvm.Status.TargetNode = fakeLocalStorageNodename
	err = client.Create(context.Background(), lvm)
	if err != nil {
		t.Errorf("Create LocalVolumeMigrate fail %v", err)
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args: args{
				migrate: lvm,
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
				localNodes:                  map[string]v1alpha1.State{},
				logger:                      log.WithField("Module", "ControllerManager"),
			}
			if err := m.volumeMigrateAddReplica(tt.args.migrate, lv); (err != nil) != tt.wantErr {
				t.Errorf("volumeMigrateInProgress() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_volumeMigrateStart(t *testing.T) {
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
		migrate *v1alpha1.LocalVolumeMigrate
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
	lvm := GenFakeLocalVolumeMigrateObject()
	lvm.Name = fakeLocalVolumeMigrateName
	lvm.Namespace = fakeNamespace
	lvm.Spec.VolumeName = fakeLocalVolumeName
	lvm.Spec.TargetNodesSuggested = fakeNodenames
	err = client.Create(context.Background(), lvm)
	if err != nil {
		t.Errorf("Create LocalVolumeMigrate fail %v", err)
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
				migrate: lvm,
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
				localNodes:                  map[string]v1alpha1.State{},
				logger:                      log.WithField("Module", "ControllerManager"),
			}
			if err := m.volumeMigrateStart(tt.args.migrate, lv, lvg); (err != nil) != tt.wantErr {
				t.Errorf("volumeMigrateStart() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_volumeMigrateSubmit(t *testing.T) {
	type args struct {
		migrate *v1alpha1.LocalVolumeMigrate
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
	lvg.Spec.Accessibility.Nodes = []string{fakeLocalVolumeName}
	lvg.Spec.Volumes = []v1alpha1.VolumeInfo{
		{
			LocalVolumeName: fakeLocalVolumeName,
		},
	}
	err = client.Create(context.Background(), lvg)
	if err != nil {
		t.Errorf("Create LocalVolumeGroup fail %v", err)
	}
	//  Create LocalVolumeReplicaObject
	lvr := GenFakeLocalVolumeReplicaObject()
	lvr.Status.State = v1alpha1.VolumeReplicaStateReady
	err = client.Create(context.Background(), lvr)
	// Create LocalVolumeMigrateObject
	lvm := GenFakeLocalVolumeMigrateObject()
	lvm.Name = fakeLocalVolumeMigrateName
	lvm.Namespace = fakeNamespace
	lvm.Spec.VolumeName = fakeLocalVolumeName
	lvm.Spec.TargetNodesSuggested = fakeNodenames
	lvm.Status.TargetNode = fakeLocalStorageNodename
	err = client.Create(context.Background(), lvm)
	if err != nil {
		t.Errorf("Create LocalVolumeMigrate fail %v", err)
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args: args{
				migrate: lvm,
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
				localNodes:                  map[string]v1alpha1.State{},
				logger:                      log.WithField("Module", "ControllerManager"),
			}
			if err := m.volumeMigrateSubmit(tt.args.migrate, lv, lvg); (err != nil) != tt.wantErr {
				t.Errorf("volumeMigrateSubmit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
