package controller

import (
	"sync"
	"testing"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-storage/common"
)

func Test_manager_processVolumeExpand(t *testing.T) {
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
	tests := []struct {
		name    string
		fields  fields
		args    args
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
			if err := m.processVolumeExpand(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("processVolumeExpand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_startVolumeExpandTaskWorker(t *testing.T) {
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
	tests := []struct {
		name   string
		fields fields
		args   args
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
			m.startVolumeExpandTaskWorker(tt.args.stopCh)
		})
	}
}

func Test_manager_volumeExpandAbort(t *testing.T) {
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
		expand *v1alpha1.LocalVolumeExpand
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
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
			if err := m.volumeExpandAbort(tt.args.expand); (err != nil) != tt.wantErr {
				t.Errorf("volumeExpandAbort() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_volumeExpandCleanup(t *testing.T) {
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
		expand *v1alpha1.LocalVolumeExpand
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
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
			if err := m.volumeExpandCleanup(tt.args.expand); (err != nil) != tt.wantErr {
				t.Errorf("volumeExpandCleanup() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_volumeExpandInProgress(t *testing.T) {
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
		expand *v1alpha1.LocalVolumeExpand
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
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
			if err := m.volumeExpandInProgress(tt.args.expand); (err != nil) != tt.wantErr {
				t.Errorf("volumeExpandInProgress() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_volumeExpandStart(t *testing.T) {
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
		expand *v1alpha1.LocalVolumeExpand
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
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
			if err := m.volumeExpandStart(tt.args.expand); (err != nil) != tt.wantErr {
				t.Errorf("volumeExpandStart() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_volumeExpandSubmit(t *testing.T) {
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
		expand *v1alpha1.LocalVolumeExpand
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
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
			if err := m.volumeExpandSubmit(tt.args.expand); (err != nil) != tt.wantErr {
				t.Errorf("volumeExpandSubmit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
