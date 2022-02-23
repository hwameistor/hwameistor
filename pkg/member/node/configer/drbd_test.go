package configer

import (
	"fmt"
	"github.com/golang/mock/gomock"
	"reflect"
	"sync"
	"testing"
	"text/template"

	localstoragev1alpha1 "github.com/hwameistor/local-storage/pkg/apis/localstorage/v1alpha1"
	"github.com/hwameistor/local-storage/pkg/exechelper"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Test_drbdConfigure_Run(t *testing.T) {
	var stopCh <-chan struct{}
	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()

	m := NewMockConfiger(ctrl)
	m.
		EXPECT().
		Run(stopCh).
		Return().
		Times(1)

	m.Run(stopCh)

	fmt.Printf("Test_drbdConfigure_Run ends")
}

func Test_drbdConfigure_HasConfig(t *testing.T) {
	var localVolumeReplica = &localstoragev1alpha1.LocalVolumeReplica{}
	localVolumeReplica.Spec.VolumeName = "volume1"
	localVolumeReplica.Spec.PoolName = "pool1"
	localVolumeReplica.Spec.NodeName = "node1"
	localVolumeReplica.Spec.RequiredCapacityBytes = 1240
	localVolumeReplica.Spec.Kind = localstoragev1alpha1.VolumeKindLVM
	localVolumeReplica.Name = "test1"

	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()

	m := NewMockConfiger(ctrl)
	m.
		EXPECT().
		HasConfig(localVolumeReplica).
		Return(false).
		Times(1)

	v := m.HasConfig(localVolumeReplica)

	fmt.Printf("Test_drbdConfigure_HasConfig v= %+v", v)
}

func Test_drbdConfigure_IsConfigUpdated(t *testing.T) {
	var localVolumeReplica = &localstoragev1alpha1.LocalVolumeReplica{}
	localVolumeReplica.Spec.VolumeName = "volume1"
	localVolumeReplica.Spec.PoolName = "pool1"
	localVolumeReplica.Spec.NodeName = "node1"
	localVolumeReplica.Spec.RequiredCapacityBytes = 1240
	localVolumeReplica.Spec.Kind = localstoragev1alpha1.VolumeKindLVM
	localVolumeReplica.Name = "test1"

	var config localstoragev1alpha1.VolumeConfig
	config.RequiredCapacityBytes = 1240
	config.VolumeName = "volume1"

	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()

	m := NewMockConfiger(ctrl)
	m.
		EXPECT().
		IsConfigUpdated(localVolumeReplica, config).
		Return(false).
		Times(1)

	v := m.IsConfigUpdated(localVolumeReplica, config)

	fmt.Printf("Test_drbdConfigure_IsConfigUpdated v= %+v", v)
}

func Test_drbdConfigure_ApplyConfig(t *testing.T) {
	var localVolumeReplica = &localstoragev1alpha1.LocalVolumeReplica{}
	localVolumeReplica.Spec.VolumeName = "volume1"
	localVolumeReplica.Spec.PoolName = "pool1"
	localVolumeReplica.Spec.NodeName = "node1"
	localVolumeReplica.Spec.RequiredCapacityBytes = 1240
	localVolumeReplica.Spec.Kind = localstoragev1alpha1.VolumeKindLVM
	localVolumeReplica.Name = "test1"

	var config localstoragev1alpha1.VolumeConfig
	config.RequiredCapacityBytes = 1240
	config.VolumeName = "volume1"

	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()

	m := NewMockConfiger(ctrl)
	m.
		EXPECT().
		ApplyConfig(localVolumeReplica, config).
		Return(nil).
		Times(1)

	v := m.ApplyConfig(localVolumeReplica, config)

	fmt.Printf("Test_drbdConfigure_ApplyConfig v= %+v", v)
}

func Test_drbdConfigure_Initialize(t *testing.T) {
	type fields struct {
		hostname               string
		apiClient              client.Client
		systemConfig           localstoragev1alpha1.SystemConfig
		statusSyncFunc         SyncReplicaStatus
		cmdExec                exechelper.Executor
		lock                   sync.Mutex
		once                   sync.Once
		localConfigs           map[string]localstoragev1alpha1.VolumeConfig
		resourceCache          map[string]*Resource
		resourceReplicaNameMap map[string]string
		template               *template.Template
		logger                 *log.Entry
		stopCh                 <-chan struct{}
	}
	type args struct {
		replica *localstoragev1alpha1.LocalVolumeReplica
		config  localstoragev1alpha1.VolumeConfig
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
			m := &drbdConfigure{
				hostname:               tt.fields.hostname,
				apiClient:              tt.fields.apiClient,
				systemConfig:           tt.fields.systemConfig,
				statusSyncFunc:         tt.fields.statusSyncFunc,
				cmdExec:                tt.fields.cmdExec,
				lock:                   tt.fields.lock,
				once:                   tt.fields.once,
				localConfigs:           tt.fields.localConfigs,
				resourceCache:          tt.fields.resourceCache,
				resourceReplicaNameMap: tt.fields.resourceReplicaNameMap,
				template:               tt.fields.template,
				logger:                 tt.fields.logger,
				stopCh:                 tt.fields.stopCh,
			}
			if err := m.Initialize(tt.args.replica, tt.args.config); (err != nil) != tt.wantErr {
				t.Errorf("drbdConfigure.Initialize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_drbdConfigure_ConsistencyCheck(t *testing.T) {
	type fields struct {
		hostname               string
		apiClient              client.Client
		systemConfig           localstoragev1alpha1.SystemConfig
		statusSyncFunc         SyncReplicaStatus
		cmdExec                exechelper.Executor
		lock                   sync.Mutex
		once                   sync.Once
		localConfigs           map[string]localstoragev1alpha1.VolumeConfig
		resourceCache          map[string]*Resource
		resourceReplicaNameMap map[string]string
		template               *template.Template
		logger                 *log.Entry
		stopCh                 <-chan struct{}
	}
	type args struct {
		replicas []localstoragev1alpha1.LocalVolumeReplica
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
			m := &drbdConfigure{
				hostname:               tt.fields.hostname,
				apiClient:              tt.fields.apiClient,
				systemConfig:           tt.fields.systemConfig,
				statusSyncFunc:         tt.fields.statusSyncFunc,
				cmdExec:                tt.fields.cmdExec,
				lock:                   tt.fields.lock,
				once:                   tt.fields.once,
				localConfigs:           tt.fields.localConfigs,
				resourceCache:          tt.fields.resourceCache,
				resourceReplicaNameMap: tt.fields.resourceReplicaNameMap,
				template:               tt.fields.template,
				logger:                 tt.fields.logger,
				stopCh:                 tt.fields.stopCh,
			}
			m.ConsistencyCheck(tt.args.replicas)
		})
	}
}

func Test_drbdConfigure_writeConfigFile(t *testing.T) {
	type fields struct {
		hostname               string
		apiClient              client.Client
		systemConfig           localstoragev1alpha1.SystemConfig
		statusSyncFunc         SyncReplicaStatus
		cmdExec                exechelper.Executor
		lock                   sync.Mutex
		once                   sync.Once
		localConfigs           map[string]localstoragev1alpha1.VolumeConfig
		resourceCache          map[string]*Resource
		resourceReplicaNameMap map[string]string
		template               *template.Template
		logger                 *log.Entry
		stopCh                 <-chan struct{}
	}
	type args struct {
		resourceName string
		conf         drbdConfig
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
			m := &drbdConfigure{
				hostname:               tt.fields.hostname,
				apiClient:              tt.fields.apiClient,
				systemConfig:           tt.fields.systemConfig,
				statusSyncFunc:         tt.fields.statusSyncFunc,
				cmdExec:                tt.fields.cmdExec,
				lock:                   tt.fields.lock,
				once:                   tt.fields.once,
				localConfigs:           tt.fields.localConfigs,
				resourceCache:          tt.fields.resourceCache,
				resourceReplicaNameMap: tt.fields.resourceReplicaNameMap,
				template:               tt.fields.template,
				logger:                 tt.fields.logger,
				stopCh:                 tt.fields.stopCh,
			}
			if err := m.writeConfigFile(tt.args.resourceName, tt.args.conf); (err != nil) != tt.wantErr {
				t.Errorf("drbdConfigure.writeConfigFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_genConfigPath(t *testing.T) {
	type args struct {
		resourceName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := genConfigPath(tt.args.resourceName); got != tt.want {
				t.Errorf("genConfigPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseConfigFileName(t *testing.T) {
	type args struct {
		configName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseConfigFileName(tt.args.configName); got != tt.want {
				t.Errorf("parseConfigFileName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_listNodeResourceNames(t *testing.T) {
	tests := []struct {
		name    string
		want    []string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := listNodeResourceNames()
			if (err != nil) != tt.wantErr {
				t.Errorf("listNodeResourceNames() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("listNodeResourceNames() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_removeResourceConfigFile(t *testing.T) {
	type args struct {
		resourceName string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := removeResourceConfigFile(tt.args.resourceName); (err != nil) != tt.wantErr {
				t.Errorf("removeResourceConfigFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_drbdConfigure_createMetadata(t *testing.T) {
	type fields struct {
		hostname               string
		apiClient              client.Client
		systemConfig           localstoragev1alpha1.SystemConfig
		statusSyncFunc         SyncReplicaStatus
		cmdExec                exechelper.Executor
		lock                   sync.Mutex
		once                   sync.Once
		localConfigs           map[string]localstoragev1alpha1.VolumeConfig
		resourceCache          map[string]*Resource
		resourceReplicaNameMap map[string]string
		template               *template.Template
		logger                 *log.Entry
		stopCh                 <-chan struct{}
	}
	type args struct {
		resourceName string
		peersCount   int
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
			m := &drbdConfigure{
				hostname:               tt.fields.hostname,
				apiClient:              tt.fields.apiClient,
				systemConfig:           tt.fields.systemConfig,
				statusSyncFunc:         tt.fields.statusSyncFunc,
				cmdExec:                tt.fields.cmdExec,
				lock:                   tt.fields.lock,
				once:                   tt.fields.once,
				localConfigs:           tt.fields.localConfigs,
				resourceCache:          tt.fields.resourceCache,
				resourceReplicaNameMap: tt.fields.resourceReplicaNameMap,
				template:               tt.fields.template,
				logger:                 tt.fields.logger,
				stopCh:                 tt.fields.stopCh,
			}
			if err := m.createMetadata(tt.args.resourceName, tt.args.peersCount); (err != nil) != tt.wantErr {
				t.Errorf("drbdConfigure.createMetadata() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_drbdConfigure_adjustResource(t *testing.T) {
	type fields struct {
		hostname               string
		apiClient              client.Client
		systemConfig           localstoragev1alpha1.SystemConfig
		statusSyncFunc         SyncReplicaStatus
		cmdExec                exechelper.Executor
		lock                   sync.Mutex
		once                   sync.Once
		localConfigs           map[string]localstoragev1alpha1.VolumeConfig
		resourceCache          map[string]*Resource
		resourceReplicaNameMap map[string]string
		template               *template.Template
		logger                 *log.Entry
		stopCh                 <-chan struct{}
	}
	type args struct {
		resourceName string
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
			m := &drbdConfigure{
				hostname:               tt.fields.hostname,
				apiClient:              tt.fields.apiClient,
				systemConfig:           tt.fields.systemConfig,
				statusSyncFunc:         tt.fields.statusSyncFunc,
				cmdExec:                tt.fields.cmdExec,
				lock:                   tt.fields.lock,
				once:                   tt.fields.once,
				localConfigs:           tt.fields.localConfigs,
				resourceCache:          tt.fields.resourceCache,
				resourceReplicaNameMap: tt.fields.resourceReplicaNameMap,
				template:               tt.fields.template,
				logger:                 tt.fields.logger,
				stopCh:                 tt.fields.stopCh,
			}
			if err := m.adjustResource(tt.args.resourceName); (err != nil) != tt.wantErr {
				t.Errorf("drbdConfigure.adjustResource() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_drbdConfigure_getResourceDiskState(t *testing.T) {
	type fields struct {
		hostname               string
		apiClient              client.Client
		systemConfig           localstoragev1alpha1.SystemConfig
		statusSyncFunc         SyncReplicaStatus
		cmdExec                exechelper.Executor
		lock                   sync.Mutex
		once                   sync.Once
		localConfigs           map[string]localstoragev1alpha1.VolumeConfig
		resourceCache          map[string]*Resource
		resourceReplicaNameMap map[string]string
		template               *template.Template
		logger                 *log.Entry
		stopCh                 <-chan struct{}
	}
	type args struct {
		resourceName string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantState string
		wantErr   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &drbdConfigure{
				hostname:               tt.fields.hostname,
				apiClient:              tt.fields.apiClient,
				systemConfig:           tt.fields.systemConfig,
				statusSyncFunc:         tt.fields.statusSyncFunc,
				cmdExec:                tt.fields.cmdExec,
				lock:                   tt.fields.lock,
				once:                   tt.fields.once,
				localConfigs:           tt.fields.localConfigs,
				resourceCache:          tt.fields.resourceCache,
				resourceReplicaNameMap: tt.fields.resourceReplicaNameMap,
				template:               tt.fields.template,
				logger:                 tt.fields.logger,
				stopCh:                 tt.fields.stopCh,
			}
			gotState, err := m.getResourceDiskState(tt.args.resourceName)
			if (err != nil) != tt.wantErr {
				t.Errorf("drbdConfigure.getResourceDiskState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotState != tt.wantState {
				t.Errorf("drbdConfigure.getResourceDiskState() = %v, want %v", gotState, tt.wantState)
			}
		})
	}
}

func Test_drbdConfigure_resizeResource(t *testing.T) {
	type fields struct {
		hostname               string
		apiClient              client.Client
		systemConfig           localstoragev1alpha1.SystemConfig
		statusSyncFunc         SyncReplicaStatus
		cmdExec                exechelper.Executor
		lock                   sync.Mutex
		once                   sync.Once
		localConfigs           map[string]localstoragev1alpha1.VolumeConfig
		resourceCache          map[string]*Resource
		resourceReplicaNameMap map[string]string
		template               *template.Template
		logger                 *log.Entry
		stopCh                 <-chan struct{}
	}
	type args struct {
		resourceName string
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
			m := &drbdConfigure{
				hostname:               tt.fields.hostname,
				apiClient:              tt.fields.apiClient,
				systemConfig:           tt.fields.systemConfig,
				statusSyncFunc:         tt.fields.statusSyncFunc,
				cmdExec:                tt.fields.cmdExec,
				lock:                   tt.fields.lock,
				once:                   tt.fields.once,
				localConfigs:           tt.fields.localConfigs,
				resourceCache:          tt.fields.resourceCache,
				resourceReplicaNameMap: tt.fields.resourceReplicaNameMap,
				template:               tt.fields.template,
				logger:                 tt.fields.logger,
				stopCh:                 tt.fields.stopCh,
			}
			if err := m.resizeResource(tt.args.resourceName); (err != nil) != tt.wantErr {
				t.Errorf("drbdConfigure.resizeResource() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_drbdConfigure_getResourceDevicePath(t *testing.T) {
	type fields struct {
		hostname               string
		apiClient              client.Client
		systemConfig           localstoragev1alpha1.SystemConfig
		statusSyncFunc         SyncReplicaStatus
		cmdExec                exechelper.Executor
		lock                   sync.Mutex
		once                   sync.Once
		localConfigs           map[string]localstoragev1alpha1.VolumeConfig
		resourceCache          map[string]*Resource
		resourceReplicaNameMap map[string]string
		template               *template.Template
		logger                 *log.Entry
		stopCh                 <-chan struct{}
	}
	type args struct {
		conf drbdConfig
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &drbdConfigure{
				hostname:               tt.fields.hostname,
				apiClient:              tt.fields.apiClient,
				systemConfig:           tt.fields.systemConfig,
				statusSyncFunc:         tt.fields.statusSyncFunc,
				cmdExec:                tt.fields.cmdExec,
				lock:                   tt.fields.lock,
				once:                   tt.fields.once,
				localConfigs:           tt.fields.localConfigs,
				resourceCache:          tt.fields.resourceCache,
				resourceReplicaNameMap: tt.fields.resourceReplicaNameMap,
				template:               tt.fields.template,
				logger:                 tt.fields.logger,
				stopCh:                 tt.fields.stopCh,
			}
			if got := m.getResourceDevicePath(tt.args.conf); got != tt.want {
				t.Errorf("drbdConfigure.getResourceDevicePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_drbdConfigure_DeleteConfig(t *testing.T) {
	type fields struct {
		hostname               string
		apiClient              client.Client
		systemConfig           localstoragev1alpha1.SystemConfig
		statusSyncFunc         SyncReplicaStatus
		cmdExec                exechelper.Executor
		lock                   sync.Mutex
		once                   sync.Once
		localConfigs           map[string]localstoragev1alpha1.VolumeConfig
		resourceCache          map[string]*Resource
		resourceReplicaNameMap map[string]string
		template               *template.Template
		logger                 *log.Entry
		stopCh                 <-chan struct{}
	}
	type args struct {
		replica *localstoragev1alpha1.LocalVolumeReplica
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
			m := &drbdConfigure{
				hostname:               tt.fields.hostname,
				apiClient:              tt.fields.apiClient,
				systemConfig:           tt.fields.systemConfig,
				statusSyncFunc:         tt.fields.statusSyncFunc,
				cmdExec:                tt.fields.cmdExec,
				lock:                   tt.fields.lock,
				once:                   tt.fields.once,
				localConfigs:           tt.fields.localConfigs,
				resourceCache:          tt.fields.resourceCache,
				resourceReplicaNameMap: tt.fields.resourceReplicaNameMap,
				template:               tt.fields.template,
				logger:                 tt.fields.logger,
				stopCh:                 tt.fields.stopCh,
			}
			if err := m.DeleteConfig(tt.args.replica); (err != nil) != tt.wantErr {
				t.Errorf("drbdConfigure.DeleteConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_drbdConfigure_downResource(t *testing.T) {
	type fields struct {
		hostname               string
		apiClient              client.Client
		systemConfig           localstoragev1alpha1.SystemConfig
		statusSyncFunc         SyncReplicaStatus
		cmdExec                exechelper.Executor
		lock                   sync.Mutex
		once                   sync.Once
		localConfigs           map[string]localstoragev1alpha1.VolumeConfig
		resourceCache          map[string]*Resource
		resourceReplicaNameMap map[string]string
		template               *template.Template
		logger                 *log.Entry
		stopCh                 <-chan struct{}
	}
	type args struct {
		resourceName string
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
			m := &drbdConfigure{
				hostname:               tt.fields.hostname,
				apiClient:              tt.fields.apiClient,
				systemConfig:           tt.fields.systemConfig,
				statusSyncFunc:         tt.fields.statusSyncFunc,
				cmdExec:                tt.fields.cmdExec,
				lock:                   tt.fields.lock,
				once:                   tt.fields.once,
				localConfigs:           tt.fields.localConfigs,
				resourceCache:          tt.fields.resourceCache,
				resourceReplicaNameMap: tt.fields.resourceReplicaNameMap,
				template:               tt.fields.template,
				logger:                 tt.fields.logger,
				stopCh:                 tt.fields.stopCh,
			}
			if err := m.downResource(tt.args.resourceName); (err != nil) != tt.wantErr {
				t.Errorf("drbdConfigure.downResource() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_drbdConfigure_wipeMetadata(t *testing.T) {
	type fields struct {
		hostname               string
		apiClient              client.Client
		systemConfig           localstoragev1alpha1.SystemConfig
		statusSyncFunc         SyncReplicaStatus
		cmdExec                exechelper.Executor
		lock                   sync.Mutex
		once                   sync.Once
		localConfigs           map[string]localstoragev1alpha1.VolumeConfig
		resourceCache          map[string]*Resource
		resourceReplicaNameMap map[string]string
		template               *template.Template
		logger                 *log.Entry
		stopCh                 <-chan struct{}
	}
	type args struct {
		resourceName string
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
			m := &drbdConfigure{
				hostname:               tt.fields.hostname,
				apiClient:              tt.fields.apiClient,
				systemConfig:           tt.fields.systemConfig,
				statusSyncFunc:         tt.fields.statusSyncFunc,
				cmdExec:                tt.fields.cmdExec,
				lock:                   tt.fields.lock,
				once:                   tt.fields.once,
				localConfigs:           tt.fields.localConfigs,
				resourceCache:          tt.fields.resourceCache,
				resourceReplicaNameMap: tt.fields.resourceReplicaNameMap,
				template:               tt.fields.template,
				logger:                 tt.fields.logger,
				stopCh:                 tt.fields.stopCh,
			}
			if err := m.wipeMetadata(tt.args.resourceName); (err != nil) != tt.wantErr {
				t.Errorf("drbdConfigure.wipeMetadata() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_drbdConfigure_EnsureDRBDResourceStateMonitorStated(t *testing.T) {
	type fields struct {
		hostname               string
		apiClient              client.Client
		systemConfig           localstoragev1alpha1.SystemConfig
		statusSyncFunc         SyncReplicaStatus
		cmdExec                exechelper.Executor
		lock                   sync.Mutex
		once                   sync.Once
		localConfigs           map[string]localstoragev1alpha1.VolumeConfig
		resourceCache          map[string]*Resource
		resourceReplicaNameMap map[string]string
		template               *template.Template
		logger                 *log.Entry
		stopCh                 <-chan struct{}
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &drbdConfigure{
				hostname:               tt.fields.hostname,
				apiClient:              tt.fields.apiClient,
				systemConfig:           tt.fields.systemConfig,
				statusSyncFunc:         tt.fields.statusSyncFunc,
				cmdExec:                tt.fields.cmdExec,
				lock:                   tt.fields.lock,
				once:                   tt.fields.once,
				localConfigs:           tt.fields.localConfigs,
				resourceCache:          tt.fields.resourceCache,
				resourceReplicaNameMap: tt.fields.resourceReplicaNameMap,
				template:               tt.fields.template,
				logger:                 tt.fields.logger,
				stopCh:                 tt.fields.stopCh,
			}
			m.EnsureDRBDResourceStateMonitorStated()
		})
	}
}

func Test_drbdConfigure_MonitorDRBDResourceState(t *testing.T) {
	type fields struct {
		hostname               string
		apiClient              client.Client
		systemConfig           localstoragev1alpha1.SystemConfig
		statusSyncFunc         SyncReplicaStatus
		cmdExec                exechelper.Executor
		lock                   sync.Mutex
		once                   sync.Once
		localConfigs           map[string]localstoragev1alpha1.VolumeConfig
		resourceCache          map[string]*Resource
		resourceReplicaNameMap map[string]string
		template               *template.Template
		logger                 *log.Entry
		stopCh                 <-chan struct{}
	}
	type args struct {
		stopCh <-chan struct{}
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
			m := &drbdConfigure{
				hostname:               tt.fields.hostname,
				apiClient:              tt.fields.apiClient,
				systemConfig:           tt.fields.systemConfig,
				statusSyncFunc:         tt.fields.statusSyncFunc,
				cmdExec:                tt.fields.cmdExec,
				lock:                   tt.fields.lock,
				once:                   tt.fields.once,
				localConfigs:           tt.fields.localConfigs,
				resourceCache:          tt.fields.resourceCache,
				resourceReplicaNameMap: tt.fields.resourceReplicaNameMap,
				template:               tt.fields.template,
				logger:                 tt.fields.logger,
				stopCh:                 tt.fields.stopCh,
			}
			if err := m.MonitorDRBDResourceState(tt.args.stopCh); (err != nil) != tt.wantErr {
				t.Errorf("drbdConfigure.MonitorDRBDResourceState() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_drbdConfigure_monitorDRBDResourceState(t *testing.T) {
	type fields struct {
		hostname               string
		apiClient              client.Client
		systemConfig           localstoragev1alpha1.SystemConfig
		statusSyncFunc         SyncReplicaStatus
		cmdExec                exechelper.Executor
		lock                   sync.Mutex
		once                   sync.Once
		localConfigs           map[string]localstoragev1alpha1.VolumeConfig
		resourceCache          map[string]*Resource
		resourceReplicaNameMap map[string]string
		template               *template.Template
		logger                 *log.Entry
		stopCh                 <-chan struct{}
	}
	type args struct {
		stopCh <-chan struct{}
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
			m := &drbdConfigure{
				hostname:               tt.fields.hostname,
				apiClient:              tt.fields.apiClient,
				systemConfig:           tt.fields.systemConfig,
				statusSyncFunc:         tt.fields.statusSyncFunc,
				cmdExec:                tt.fields.cmdExec,
				lock:                   tt.fields.lock,
				once:                   tt.fields.once,
				localConfigs:           tt.fields.localConfigs,
				resourceCache:          tt.fields.resourceCache,
				resourceReplicaNameMap: tt.fields.resourceReplicaNameMap,
				template:               tt.fields.template,
				logger:                 tt.fields.logger,
				stopCh:                 tt.fields.stopCh,
			}
			if err := m.monitorDRBDResourceState(tt.args.stopCh); (err != nil) != tt.wantErr {
				t.Errorf("drbdConfigure.monitorDRBDResourceState() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_drbdConfigure_handleDRBDEvent(t *testing.T) {
	type fields struct {
		hostname               string
		apiClient              client.Client
		systemConfig           localstoragev1alpha1.SystemConfig
		statusSyncFunc         SyncReplicaStatus
		cmdExec                exechelper.Executor
		lock                   sync.Mutex
		once                   sync.Once
		localConfigs           map[string]localstoragev1alpha1.VolumeConfig
		resourceCache          map[string]*Resource
		resourceReplicaNameMap map[string]string
		template               *template.Template
		logger                 *log.Entry
		stopCh                 <-chan struct{}
	}
	type args struct {
		event string
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
			m := &drbdConfigure{
				hostname:               tt.fields.hostname,
				apiClient:              tt.fields.apiClient,
				systemConfig:           tt.fields.systemConfig,
				statusSyncFunc:         tt.fields.statusSyncFunc,
				cmdExec:                tt.fields.cmdExec,
				lock:                   tt.fields.lock,
				once:                   tt.fields.once,
				localConfigs:           tt.fields.localConfigs,
				resourceCache:          tt.fields.resourceCache,
				resourceReplicaNameMap: tt.fields.resourceReplicaNameMap,
				template:               tt.fields.template,
				logger:                 tt.fields.logger,
				stopCh:                 tt.fields.stopCh,
			}
			m.handleDRBDEvent(tt.args.event)
		})
	}
}

func Test_drbdConfigure_getReplicaHAState(t *testing.T) {
	type fields struct {
		hostname               string
		apiClient              client.Client
		systemConfig           localstoragev1alpha1.SystemConfig
		statusSyncFunc         SyncReplicaStatus
		cmdExec                exechelper.Executor
		lock                   sync.Mutex
		once                   sync.Once
		localConfigs           map[string]localstoragev1alpha1.VolumeConfig
		resourceCache          map[string]*Resource
		resourceReplicaNameMap map[string]string
		template               *template.Template
		logger                 *log.Entry
		stopCh                 <-chan struct{}
	}
	type args struct {
		resource *Resource
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   localstoragev1alpha1.HAState
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &drbdConfigure{
				hostname:               tt.fields.hostname,
				apiClient:              tt.fields.apiClient,
				systemConfig:           tt.fields.systemConfig,
				statusSyncFunc:         tt.fields.statusSyncFunc,
				cmdExec:                tt.fields.cmdExec,
				lock:                   tt.fields.lock,
				once:                   tt.fields.once,
				localConfigs:           tt.fields.localConfigs,
				resourceCache:          tt.fields.resourceCache,
				resourceReplicaNameMap: tt.fields.resourceReplicaNameMap,
				template:               tt.fields.template,
				logger:                 tt.fields.logger,
				stopCh:                 tt.fields.stopCh,
			}
			if got := m.getReplicaHAState(tt.args.resource); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("drbdConfigure.getReplicaHAState() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_drbdConfigure_GetReplicaHAState(t *testing.T) {
	type fields struct {
		hostname               string
		apiClient              client.Client
		systemConfig           localstoragev1alpha1.SystemConfig
		statusSyncFunc         SyncReplicaStatus
		cmdExec                exechelper.Executor
		lock                   sync.Mutex
		once                   sync.Once
		localConfigs           map[string]localstoragev1alpha1.VolumeConfig
		resourceCache          map[string]*Resource
		resourceReplicaNameMap map[string]string
		template               *template.Template
		logger                 *log.Entry
		stopCh                 <-chan struct{}
	}
	type args struct {
		replica *localstoragev1alpha1.LocalVolumeReplica
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    localstoragev1alpha1.HAState
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &drbdConfigure{
				hostname:               tt.fields.hostname,
				apiClient:              tt.fields.apiClient,
				systemConfig:           tt.fields.systemConfig,
				statusSyncFunc:         tt.fields.statusSyncFunc,
				cmdExec:                tt.fields.cmdExec,
				lock:                   tt.fields.lock,
				once:                   tt.fields.once,
				localConfigs:           tt.fields.localConfigs,
				resourceCache:          tt.fields.resourceCache,
				resourceReplicaNameMap: tt.fields.resourceReplicaNameMap,
				template:               tt.fields.template,
				logger:                 tt.fields.logger,
				stopCh:                 tt.fields.stopCh,
			}
			got, err := m.GetReplicaHAState(tt.args.replica)
			if (err != nil) != tt.wantErr {
				t.Errorf("drbdConfigure.GetReplicaHAState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("drbdConfigure.GetReplicaHAState() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_drbdConfigure_hasMetadata(t *testing.T) {
	type fields struct {
		hostname               string
		apiClient              client.Client
		systemConfig           localstoragev1alpha1.SystemConfig
		statusSyncFunc         SyncReplicaStatus
		cmdExec                exechelper.Executor
		lock                   sync.Mutex
		once                   sync.Once
		localConfigs           map[string]localstoragev1alpha1.VolumeConfig
		resourceCache          map[string]*Resource
		resourceReplicaNameMap map[string]string
		template               *template.Template
		logger                 *log.Entry
		stopCh                 <-chan struct{}
	}
	type args struct {
		minor      int
		devicePath string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &drbdConfigure{
				hostname:               tt.fields.hostname,
				apiClient:              tt.fields.apiClient,
				systemConfig:           tt.fields.systemConfig,
				statusSyncFunc:         tt.fields.statusSyncFunc,
				cmdExec:                tt.fields.cmdExec,
				lock:                   tt.fields.lock,
				once:                   tt.fields.once,
				localConfigs:           tt.fields.localConfigs,
				resourceCache:          tt.fields.resourceCache,
				resourceReplicaNameMap: tt.fields.resourceReplicaNameMap,
				template:               tt.fields.template,
				logger:                 tt.fields.logger,
				stopCh:                 tt.fields.stopCh,
			}
			if got := m.hasMetadata(tt.args.minor, tt.args.devicePath); got != tt.want {
				t.Errorf("drbdConfigure.hasMetadata() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_drbdConfigure_config2DRBDConfig(t *testing.T) {
	type fields struct {
		hostname               string
		apiClient              client.Client
		systemConfig           localstoragev1alpha1.SystemConfig
		statusSyncFunc         SyncReplicaStatus
		cmdExec                exechelper.Executor
		lock                   sync.Mutex
		once                   sync.Once
		localConfigs           map[string]localstoragev1alpha1.VolumeConfig
		resourceCache          map[string]*Resource
		resourceReplicaNameMap map[string]string
		template               *template.Template
		logger                 *log.Entry
		stopCh                 <-chan struct{}
	}
	type args struct {
		replica *localstoragev1alpha1.LocalVolumeReplica
		config  localstoragev1alpha1.VolumeConfig
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   drbdConfig
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &drbdConfigure{
				hostname:               tt.fields.hostname,
				apiClient:              tt.fields.apiClient,
				systemConfig:           tt.fields.systemConfig,
				statusSyncFunc:         tt.fields.statusSyncFunc,
				cmdExec:                tt.fields.cmdExec,
				lock:                   tt.fields.lock,
				once:                   tt.fields.once,
				localConfigs:           tt.fields.localConfigs,
				resourceCache:          tt.fields.resourceCache,
				resourceReplicaNameMap: tt.fields.resourceReplicaNameMap,
				template:               tt.fields.template,
				logger:                 tt.fields.logger,
				stopCh:                 tt.fields.stopCh,
			}
			if got := m.config2DRBDConfig(tt.args.replica, tt.args.config); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("drbdConfigure.config2DRBDConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_drbdConfigure_isDeviceUpToDate(t *testing.T) {
	type fields struct {
		hostname               string
		apiClient              client.Client
		systemConfig           localstoragev1alpha1.SystemConfig
		statusSyncFunc         SyncReplicaStatus
		cmdExec                exechelper.Executor
		lock                   sync.Mutex
		once                   sync.Once
		localConfigs           map[string]localstoragev1alpha1.VolumeConfig
		resourceCache          map[string]*Resource
		resourceReplicaNameMap map[string]string
		template               *template.Template
		logger                 *log.Entry
		stopCh                 <-chan struct{}
	}
	type args struct {
		resourceName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &drbdConfigure{
				hostname:               tt.fields.hostname,
				apiClient:              tt.fields.apiClient,
				systemConfig:           tt.fields.systemConfig,
				statusSyncFunc:         tt.fields.statusSyncFunc,
				cmdExec:                tt.fields.cmdExec,
				lock:                   tt.fields.lock,
				once:                   tt.fields.once,
				localConfigs:           tt.fields.localConfigs,
				resourceCache:          tt.fields.resourceCache,
				resourceReplicaNameMap: tt.fields.resourceReplicaNameMap,
				template:               tt.fields.template,
				logger:                 tt.fields.logger,
				stopCh:                 tt.fields.stopCh,
			}
			got, err := m.isDeviceUpToDate(tt.args.resourceName)
			if (err != nil) != tt.wantErr {
				t.Errorf("drbdConfigure.isDeviceUpToDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("drbdConfigure.isDeviceUpToDate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_drbdConfigure_getDeviceState(t *testing.T) {
	type fields struct {
		hostname               string
		apiClient              client.Client
		systemConfig           localstoragev1alpha1.SystemConfig
		statusSyncFunc         SyncReplicaStatus
		cmdExec                exechelper.Executor
		lock                   sync.Mutex
		once                   sync.Once
		localConfigs           map[string]localstoragev1alpha1.VolumeConfig
		resourceCache          map[string]*Resource
		resourceReplicaNameMap map[string]string
		template               *template.Template
		logger                 *log.Entry
		stopCh                 <-chan struct{}
	}
	type args struct {
		resourceName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &drbdConfigure{
				hostname:               tt.fields.hostname,
				apiClient:              tt.fields.apiClient,
				systemConfig:           tt.fields.systemConfig,
				statusSyncFunc:         tt.fields.statusSyncFunc,
				cmdExec:                tt.fields.cmdExec,
				lock:                   tt.fields.lock,
				once:                   tt.fields.once,
				localConfigs:           tt.fields.localConfigs,
				resourceCache:          tt.fields.resourceCache,
				resourceReplicaNameMap: tt.fields.resourceReplicaNameMap,
				template:               tt.fields.template,
				logger:                 tt.fields.logger,
				stopCh:                 tt.fields.stopCh,
			}
			got, err := m.getDeviceState(tt.args.resourceName)
			if (err != nil) != tt.wantErr {
				t.Errorf("drbdConfigure.getDeviceState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("drbdConfigure.getDeviceState() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_drbdConfigure_setResourceUpToDate(t *testing.T) {
	type fields struct {
		hostname               string
		apiClient              client.Client
		systemConfig           localstoragev1alpha1.SystemConfig
		statusSyncFunc         SyncReplicaStatus
		cmdExec                exechelper.Executor
		lock                   sync.Mutex
		once                   sync.Once
		localConfigs           map[string]localstoragev1alpha1.VolumeConfig
		resourceCache          map[string]*Resource
		resourceReplicaNameMap map[string]string
		template               *template.Template
		logger                 *log.Entry
		stopCh                 <-chan struct{}
	}
	type args struct {
		resourceName string
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
			m := &drbdConfigure{
				hostname:               tt.fields.hostname,
				apiClient:              tt.fields.apiClient,
				systemConfig:           tt.fields.systemConfig,
				statusSyncFunc:         tt.fields.statusSyncFunc,
				cmdExec:                tt.fields.cmdExec,
				lock:                   tt.fields.lock,
				once:                   tt.fields.once,
				localConfigs:           tt.fields.localConfigs,
				resourceCache:          tt.fields.resourceCache,
				resourceReplicaNameMap: tt.fields.resourceReplicaNameMap,
				template:               tt.fields.template,
				logger:                 tt.fields.logger,
				stopCh:                 tt.fields.stopCh,
			}
			if err := m.setResourceUpToDate(tt.args.resourceName); (err != nil) != tt.wantErr {
				t.Errorf("drbdConfigure.setResourceUpToDate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_drbdConfigure_isAllResourcePeersConnected(t *testing.T) {
	type fields struct {
		hostname               string
		apiClient              client.Client
		systemConfig           localstoragev1alpha1.SystemConfig
		statusSyncFunc         SyncReplicaStatus
		cmdExec                exechelper.Executor
		lock                   sync.Mutex
		once                   sync.Once
		localConfigs           map[string]localstoragev1alpha1.VolumeConfig
		resourceCache          map[string]*Resource
		resourceReplicaNameMap map[string]string
		template               *template.Template
		logger                 *log.Entry
		stopCh                 <-chan struct{}
	}
	type args struct {
		resourceName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &drbdConfigure{
				hostname:               tt.fields.hostname,
				apiClient:              tt.fields.apiClient,
				systemConfig:           tt.fields.systemConfig,
				statusSyncFunc:         tt.fields.statusSyncFunc,
				cmdExec:                tt.fields.cmdExec,
				lock:                   tt.fields.lock,
				once:                   tt.fields.once,
				localConfigs:           tt.fields.localConfigs,
				resourceCache:          tt.fields.resourceCache,
				resourceReplicaNameMap: tt.fields.resourceReplicaNameMap,
				template:               tt.fields.template,
				logger:                 tt.fields.logger,
				stopCh:                 tt.fields.stopCh,
			}
			got, err := m.isAllResourcePeersConnected(tt.args.resourceName)
			if (err != nil) != tt.wantErr {
				t.Errorf("drbdConfigure.isAllResourcePeersConnected() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("drbdConfigure.isAllResourcePeersConnected() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_drbdConfigure_getResourceConnectionState(t *testing.T) {
	type fields struct {
		hostname               string
		apiClient              client.Client
		systemConfig           localstoragev1alpha1.SystemConfig
		statusSyncFunc         SyncReplicaStatus
		cmdExec                exechelper.Executor
		lock                   sync.Mutex
		once                   sync.Once
		localConfigs           map[string]localstoragev1alpha1.VolumeConfig
		resourceCache          map[string]*Resource
		resourceReplicaNameMap map[string]string
		template               *template.Template
		logger                 *log.Entry
		stopCh                 <-chan struct{}
	}
	type args struct {
		resourceName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &drbdConfigure{
				hostname:               tt.fields.hostname,
				apiClient:              tt.fields.apiClient,
				systemConfig:           tt.fields.systemConfig,
				statusSyncFunc:         tt.fields.statusSyncFunc,
				cmdExec:                tt.fields.cmdExec,
				lock:                   tt.fields.lock,
				once:                   tt.fields.once,
				localConfigs:           tt.fields.localConfigs,
				resourceCache:          tt.fields.resourceCache,
				resourceReplicaNameMap: tt.fields.resourceReplicaNameMap,
				template:               tt.fields.template,
				logger:                 tt.fields.logger,
				stopCh:                 tt.fields.stopCh,
			}
			got, err := m.getResourceConnectionState(tt.args.resourceName)
			if (err != nil) != tt.wantErr {
				t.Errorf("drbdConfigure.getResourceConnectionState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("drbdConfigure.getResourceConnectionState() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_drbdConfigure_newCurrentUUID(t *testing.T) {
	type fields struct {
		hostname               string
		apiClient              client.Client
		systemConfig           localstoragev1alpha1.SystemConfig
		statusSyncFunc         SyncReplicaStatus
		cmdExec                exechelper.Executor
		lock                   sync.Mutex
		once                   sync.Once
		localConfigs           map[string]localstoragev1alpha1.VolumeConfig
		resourceCache          map[string]*Resource
		resourceReplicaNameMap map[string]string
		template               *template.Template
		logger                 *log.Entry
		stopCh                 <-chan struct{}
	}
	type args struct {
		resourceName string
		clearBitmap  bool
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
			m := &drbdConfigure{
				hostname:               tt.fields.hostname,
				apiClient:              tt.fields.apiClient,
				systemConfig:           tt.fields.systemConfig,
				statusSyncFunc:         tt.fields.statusSyncFunc,
				cmdExec:                tt.fields.cmdExec,
				lock:                   tt.fields.lock,
				once:                   tt.fields.once,
				localConfigs:           tt.fields.localConfigs,
				resourceCache:          tt.fields.resourceCache,
				resourceReplicaNameMap: tt.fields.resourceReplicaNameMap,
				template:               tt.fields.template,
				logger:                 tt.fields.logger,
				stopCh:                 tt.fields.stopCh,
			}
			if err := m.newCurrentUUID(tt.args.resourceName, tt.args.clearBitmap); (err != nil) != tt.wantErr {
				t.Errorf("drbdConfigure.newCurrentUUID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_drbdConfigure_primaryResource(t *testing.T) {
	type fields struct {
		hostname               string
		apiClient              client.Client
		systemConfig           localstoragev1alpha1.SystemConfig
		statusSyncFunc         SyncReplicaStatus
		cmdExec                exechelper.Executor
		lock                   sync.Mutex
		once                   sync.Once
		localConfigs           map[string]localstoragev1alpha1.VolumeConfig
		resourceCache          map[string]*Resource
		resourceReplicaNameMap map[string]string
		template               *template.Template
		logger                 *log.Entry
		stopCh                 <-chan struct{}
	}
	type args struct {
		resourceName string
		force        bool
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
			m := &drbdConfigure{
				hostname:               tt.fields.hostname,
				apiClient:              tt.fields.apiClient,
				systemConfig:           tt.fields.systemConfig,
				statusSyncFunc:         tt.fields.statusSyncFunc,
				cmdExec:                tt.fields.cmdExec,
				lock:                   tt.fields.lock,
				once:                   tt.fields.once,
				localConfigs:           tt.fields.localConfigs,
				resourceCache:          tt.fields.resourceCache,
				resourceReplicaNameMap: tt.fields.resourceReplicaNameMap,
				template:               tt.fields.template,
				logger:                 tt.fields.logger,
				stopCh:                 tt.fields.stopCh,
			}
			if err := m.primaryResource(tt.args.resourceName, tt.args.force); (err != nil) != tt.wantErr {
				t.Errorf("drbdConfigure.primaryResource() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_drbdConfigure_secondaryResource(t *testing.T) {
	type fields struct {
		hostname               string
		apiClient              client.Client
		systemConfig           localstoragev1alpha1.SystemConfig
		statusSyncFunc         SyncReplicaStatus
		cmdExec                exechelper.Executor
		lock                   sync.Mutex
		once                   sync.Once
		localConfigs           map[string]localstoragev1alpha1.VolumeConfig
		resourceCache          map[string]*Resource
		resourceReplicaNameMap map[string]string
		template               *template.Template
		logger                 *log.Entry
		stopCh                 <-chan struct{}
	}
	type args struct {
		resourceName string
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
			m := &drbdConfigure{
				hostname:               tt.fields.hostname,
				apiClient:              tt.fields.apiClient,
				systemConfig:           tt.fields.systemConfig,
				statusSyncFunc:         tt.fields.statusSyncFunc,
				cmdExec:                tt.fields.cmdExec,
				lock:                   tt.fields.lock,
				once:                   tt.fields.once,
				localConfigs:           tt.fields.localConfigs,
				resourceCache:          tt.fields.resourceCache,
				resourceReplicaNameMap: tt.fields.resourceReplicaNameMap,
				template:               tt.fields.template,
				logger:                 tt.fields.logger,
				stopCh:                 tt.fields.stopCh,
			}
			if err := m.secondaryResource(tt.args.resourceName); (err != nil) != tt.wantErr {
				t.Errorf("drbdConfigure.secondaryResource() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_drbdConfigure_genResourceName(t *testing.T) {
	type fields struct {
		hostname               string
		apiClient              client.Client
		systemConfig           localstoragev1alpha1.SystemConfig
		statusSyncFunc         SyncReplicaStatus
		cmdExec                exechelper.Executor
		lock                   sync.Mutex
		once                   sync.Once
		localConfigs           map[string]localstoragev1alpha1.VolumeConfig
		resourceCache          map[string]*Resource
		resourceReplicaNameMap map[string]string
		template               *template.Template
		logger                 *log.Entry
		stopCh                 <-chan struct{}
	}
	type args struct {
		replica *localstoragev1alpha1.LocalVolumeReplica
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &drbdConfigure{
				hostname:               tt.fields.hostname,
				apiClient:              tt.fields.apiClient,
				systemConfig:           tt.fields.systemConfig,
				statusSyncFunc:         tt.fields.statusSyncFunc,
				cmdExec:                tt.fields.cmdExec,
				lock:                   tt.fields.lock,
				once:                   tt.fields.once,
				localConfigs:           tt.fields.localConfigs,
				resourceCache:          tt.fields.resourceCache,
				resourceReplicaNameMap: tt.fields.resourceReplicaNameMap,
				template:               tt.fields.template,
				logger:                 tt.fields.logger,
				stopCh:                 tt.fields.stopCh,
			}
			if got := m.genResourceName(tt.args.replica); got != tt.want {
				t.Errorf("drbdConfigure.genResourceName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_drbdConfigure_getReplicaName(t *testing.T) {
	type fields struct {
		hostname               string
		apiClient              client.Client
		systemConfig           localstoragev1alpha1.SystemConfig
		statusSyncFunc         SyncReplicaStatus
		cmdExec                exechelper.Executor
		lock                   sync.Mutex
		once                   sync.Once
		localConfigs           map[string]localstoragev1alpha1.VolumeConfig
		resourceCache          map[string]*Resource
		resourceReplicaNameMap map[string]string
		template               *template.Template
		logger                 *log.Entry
		stopCh                 <-chan struct{}
	}
	type args struct {
		resourceName string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &drbdConfigure{
				hostname:               tt.fields.hostname,
				apiClient:              tt.fields.apiClient,
				systemConfig:           tt.fields.systemConfig,
				statusSyncFunc:         tt.fields.statusSyncFunc,
				cmdExec:                tt.fields.cmdExec,
				lock:                   tt.fields.lock,
				once:                   tt.fields.once,
				localConfigs:           tt.fields.localConfigs,
				resourceCache:          tt.fields.resourceCache,
				resourceReplicaNameMap: tt.fields.resourceReplicaNameMap,
				template:               tt.fields.template,
				logger:                 tt.fields.logger,
				stopCh:                 tt.fields.stopCh,
			}
			if got := m.getReplicaName(tt.args.resourceName); got != tt.want {
				t.Errorf("drbdConfigure.getReplicaName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_drbdConfigure_isPrimary(t *testing.T) {
	type fields struct {
		hostname               string
		apiClient              client.Client
		systemConfig           localstoragev1alpha1.SystemConfig
		statusSyncFunc         SyncReplicaStatus
		cmdExec                exechelper.Executor
		lock                   sync.Mutex
		once                   sync.Once
		localConfigs           map[string]localstoragev1alpha1.VolumeConfig
		resourceCache          map[string]*Resource
		resourceReplicaNameMap map[string]string
		template               *template.Template
		logger                 *log.Entry
		stopCh                 <-chan struct{}
	}
	type args struct {
		config localstoragev1alpha1.VolumeConfig
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &drbdConfigure{
				hostname:               tt.fields.hostname,
				apiClient:              tt.fields.apiClient,
				systemConfig:           tt.fields.systemConfig,
				statusSyncFunc:         tt.fields.statusSyncFunc,
				cmdExec:                tt.fields.cmdExec,
				lock:                   tt.fields.lock,
				once:                   tt.fields.once,
				localConfigs:           tt.fields.localConfigs,
				resourceCache:          tt.fields.resourceCache,
				resourceReplicaNameMap: tt.fields.resourceReplicaNameMap,
				template:               tt.fields.template,
				logger:                 tt.fields.logger,
				stopCh:                 tt.fields.stopCh,
			}
			if got := m.isPrimary(tt.args.config); got != tt.want {
				t.Errorf("drbdConfigure.isPrimary() = %v, want %v", got, tt.want)
			}
		})
	}
}
