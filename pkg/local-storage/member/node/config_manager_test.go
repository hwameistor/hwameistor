package node

import (
	"reflect"
	"testing"

	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-storage/common"
	"github.com/hwameistor/hwameistor/pkg/local-storage/member/node/configer"
)

func TestNewConfigManager(t *testing.T) {
	type args struct {
		hostname  string
		config    v1alpha1.SystemConfig
		apiClient client.Client
	}
	tests := []struct {
		name    string
		args    args
		want    *configManager
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewConfigManager(tt.args.hostname, tt.args.config, tt.args.apiClient)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewConfigManager() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewConfigManager() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVolumeReplica(t *testing.T) {
	type fields struct {
		hostname               string
		systemConfig           v1alpha1.SystemConfig
		apiClient              client.Client
		configer               configer.Configer
		logger                 *log.Entry
		syncReplicaStatusQueue *common.TaskQueue
	}
	type args struct {
		replica *v1alpha1.LocalVolumeReplica
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
			m := &configManager{
				hostname:               tt.fields.hostname,
				systemConfig:           tt.fields.systemConfig,
				apiClient:              tt.fields.apiClient,
				configer:               tt.fields.configer,
				logger:                 tt.fields.logger,
				syncReplicaStatusQueue: tt.fields.syncReplicaStatusQueue,
			}
			if err := m.TestVolumeReplica(tt.args.replica); (err != nil) != tt.wantErr {
				t.Errorf("TestVolumeReplica() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_configManager_ConsistencyCheck(t *testing.T) {
	type fields struct {
		hostname               string
		systemConfig           v1alpha1.SystemConfig
		apiClient              client.Client
		configer               configer.Configer
		logger                 *log.Entry
		syncReplicaStatusQueue *common.TaskQueue
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &configManager{
				hostname:               tt.fields.hostname,
				systemConfig:           tt.fields.systemConfig,
				apiClient:              tt.fields.apiClient,
				configer:               tt.fields.configer,
				logger:                 tt.fields.logger,
				syncReplicaStatusQueue: tt.fields.syncReplicaStatusQueue,
			}
			m.ConsistencyCheck()
		})
	}
}

func Test_configManager_DeleteConfig(t *testing.T) {
	type fields struct {
		hostname               string
		systemConfig           v1alpha1.SystemConfig
		apiClient              client.Client
		configer               configer.Configer
		logger                 *log.Entry
		syncReplicaStatusQueue *common.TaskQueue
	}
	type args struct {
		replica *v1alpha1.LocalVolumeReplica
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
			m := &configManager{
				hostname:               tt.fields.hostname,
				systemConfig:           tt.fields.systemConfig,
				apiClient:              tt.fields.apiClient,
				configer:               tt.fields.configer,
				logger:                 tt.fields.logger,
				syncReplicaStatusQueue: tt.fields.syncReplicaStatusQueue,
			}
			if err := m.DeleteConfig(tt.args.replica); (err != nil) != tt.wantErr {
				t.Errorf("DeleteConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_configManager_EnsureConfig(t *testing.T) {
	type fields struct {
		hostname               string
		systemConfig           v1alpha1.SystemConfig
		apiClient              client.Client
		configer               configer.Configer
		logger                 *log.Entry
		syncReplicaStatusQueue *common.TaskQueue
	}
	type args struct {
		replica *v1alpha1.LocalVolumeReplica
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
			m := &configManager{
				hostname:               tt.fields.hostname,
				systemConfig:           tt.fields.systemConfig,
				apiClient:              tt.fields.apiClient,
				configer:               tt.fields.configer,
				logger:                 tt.fields.logger,
				syncReplicaStatusQueue: tt.fields.syncReplicaStatusQueue,
			}
			if err := m.EnsureConfig(tt.args.replica); (err != nil) != tt.wantErr {
				t.Errorf("EnsureConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_configManager_Run(t *testing.T) {
	type fields struct {
		hostname               string
		systemConfig           v1alpha1.SystemConfig
		apiClient              client.Client
		configer               configer.Configer
		logger                 *log.Entry
		syncReplicaStatusQueue *common.TaskQueue
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
			m := &configManager{
				hostname:               tt.fields.hostname,
				systemConfig:           tt.fields.systemConfig,
				apiClient:              tt.fields.apiClient,
				configer:               tt.fields.configer,
				logger:                 tt.fields.logger,
				syncReplicaStatusQueue: tt.fields.syncReplicaStatusQueue,
			}
			if err := m.Run(tt.args.stopCh); (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_configManager_ensureConfigForHA(t *testing.T) {
	type fields struct {
		hostname               string
		systemConfig           v1alpha1.SystemConfig
		apiClient              client.Client
		configer               configer.Configer
		logger                 *log.Entry
		syncReplicaStatusQueue *common.TaskQueue
	}
	type args struct {
		replica *v1alpha1.LocalVolumeReplica
		config  *v1alpha1.VolumeConfig
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
			m := &configManager{
				hostname:               tt.fields.hostname,
				systemConfig:           tt.fields.systemConfig,
				apiClient:              tt.fields.apiClient,
				configer:               tt.fields.configer,
				logger:                 tt.fields.logger,
				syncReplicaStatusQueue: tt.fields.syncReplicaStatusQueue,
			}
			if err := m.ensureConfigForHA(tt.args.replica, tt.args.config); (err != nil) != tt.wantErr {
				t.Errorf("ensureConfigForHA() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_configManager_ensureConfigForNonHA(t *testing.T) {
	type fields struct {
		hostname               string
		systemConfig           v1alpha1.SystemConfig
		apiClient              client.Client
		configer               configer.Configer
		logger                 *log.Entry
		syncReplicaStatusQueue *common.TaskQueue
	}
	type args struct {
		replica *v1alpha1.LocalVolumeReplica
		config  *v1alpha1.VolumeConfig
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
			m := &configManager{
				hostname:               tt.fields.hostname,
				systemConfig:           tt.fields.systemConfig,
				apiClient:              tt.fields.apiClient,
				configer:               tt.fields.configer,
				logger:                 tt.fields.logger,
				syncReplicaStatusQueue: tt.fields.syncReplicaStatusQueue,
			}
			if err := m.ensureConfigForNonHA(tt.args.replica, tt.args.config); (err != nil) != tt.wantErr {
				t.Errorf("ensureConfigForNonHA() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_configManager_ensureDirectory(t *testing.T) {
	type fields struct {
		hostname               string
		systemConfig           v1alpha1.SystemConfig
		apiClient              client.Client
		configer               configer.Configer
		logger                 *log.Entry
		syncReplicaStatusQueue *common.TaskQueue
	}
	type args struct {
		filepath string
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
			m := &configManager{
				hostname:               tt.fields.hostname,
				systemConfig:           tt.fields.systemConfig,
				apiClient:              tt.fields.apiClient,
				configer:               tt.fields.configer,
				logger:                 tt.fields.logger,
				syncReplicaStatusQueue: tt.fields.syncReplicaStatusQueue,
			}
			if err := m.ensureDirectory(tt.args.filepath); (err != nil) != tt.wantErr {
				t.Errorf("ensureDirectory() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_configManager_genDevicePath(t *testing.T) {
	type fields struct {
		hostname               string
		systemConfig           v1alpha1.SystemConfig
		apiClient              client.Client
		configer               configer.Configer
		logger                 *log.Entry
		syncReplicaStatusQueue *common.TaskQueue
	}
	type args struct {
		replica *v1alpha1.LocalVolumeReplica
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
			m := &configManager{
				hostname:               tt.fields.hostname,
				systemConfig:           tt.fields.systemConfig,
				apiClient:              tt.fields.apiClient,
				configer:               tt.fields.configer,
				logger:                 tt.fields.logger,
				syncReplicaStatusQueue: tt.fields.syncReplicaStatusQueue,
			}
			if got := m.genDevicePath(tt.args.replica); got != tt.want {
				t.Errorf("genDevicePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_configManager_genReplicaStateFromHAState(t *testing.T) {
	type fields struct {
		hostname               string
		systemConfig           v1alpha1.SystemConfig
		apiClient              client.Client
		configer               configer.Configer
		logger                 *log.Entry
		syncReplicaStatusQueue *common.TaskQueue
	}
	type args struct {
		haState v1alpha1.HAState
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   v1alpha1.State
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &configManager{
				hostname:               tt.fields.hostname,
				systemConfig:           tt.fields.systemConfig,
				apiClient:              tt.fields.apiClient,
				configer:               tt.fields.configer,
				logger:                 tt.fields.logger,
				syncReplicaStatusQueue: tt.fields.syncReplicaStatusQueue,
			}
			if got := m.genReplicaStateFromHAState(tt.args.haState); got != tt.want {
				t.Errorf("genReplicaStateFromHAState() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_configManager_getConfig(t *testing.T) {
	type fields struct {
		hostname               string
		systemConfig           v1alpha1.SystemConfig
		apiClient              client.Client
		configer               configer.Configer
		logger                 *log.Entry
		syncReplicaStatusQueue *common.TaskQueue
	}
	type args struct {
		replica *v1alpha1.LocalVolumeReplica
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *v1alpha1.VolumeConfig
		want1   bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &configManager{
				hostname:               tt.fields.hostname,
				systemConfig:           tt.fields.systemConfig,
				apiClient:              tt.fields.apiClient,
				configer:               tt.fields.configer,
				logger:                 tt.fields.logger,
				syncReplicaStatusQueue: tt.fields.syncReplicaStatusQueue,
			}
			got, got1, err := m.getConfig(tt.args.replica)
			if (err != nil) != tt.wantErr {
				t.Errorf("getConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getConfig() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getConfig() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_configManager_getCurrentNodeReplicas(t *testing.T) {
	type fields struct {
		hostname               string
		systemConfig           v1alpha1.SystemConfig
		apiClient              client.Client
		configer               configer.Configer
		logger                 *log.Entry
		syncReplicaStatusQueue *common.TaskQueue
	}
	tests := []struct {
		name         string
		fields       fields
		wantReplicas []v1alpha1.LocalVolumeReplica
		wantErr      bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &configManager{
				hostname:               tt.fields.hostname,
				systemConfig:           tt.fields.systemConfig,
				apiClient:              tt.fields.apiClient,
				configer:               tt.fields.configer,
				logger:                 tt.fields.logger,
				syncReplicaStatusQueue: tt.fields.syncReplicaStatusQueue,
			}
			gotReplicas, err := m.getCurrentNodeReplicas()
			if (err != nil) != tt.wantErr {
				t.Errorf("getCurrentNodeReplicas() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotReplicas, tt.wantReplicas) {
				t.Errorf("getCurrentNodeReplicas() gotReplicas = %v, want %v", gotReplicas, tt.wantReplicas)
			}
		})
	}
}

func Test_configManager_isThisNodePrimary(t *testing.T) {
	type fields struct {
		hostname               string
		systemConfig           v1alpha1.SystemConfig
		apiClient              client.Client
		configer               configer.Configer
		logger                 *log.Entry
		syncReplicaStatusQueue *common.TaskQueue
	}
	type args struct {
		config v1alpha1.VolumeConfig
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
			m := &configManager{
				hostname:               tt.fields.hostname,
				systemConfig:           tt.fields.systemConfig,
				apiClient:              tt.fields.apiClient,
				configer:               tt.fields.configer,
				logger:                 tt.fields.logger,
				syncReplicaStatusQueue: tt.fields.syncReplicaStatusQueue,
			}
			if got := m.isThisNodePrimary(tt.args.config); got != tt.want {
				t.Errorf("isThisNodePrimary() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_configManager_processReplicaStatusUpdate(t *testing.T) {
	type fields struct {
		hostname               string
		systemConfig           v1alpha1.SystemConfig
		apiClient              client.Client
		configer               configer.Configer
		logger                 *log.Entry
		syncReplicaStatusQueue *common.TaskQueue
	}
	type args struct {
		replicaName string
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
			m := &configManager{
				hostname:               tt.fields.hostname,
				systemConfig:           tt.fields.systemConfig,
				apiClient:              tt.fields.apiClient,
				configer:               tt.fields.configer,
				logger:                 tt.fields.logger,
				syncReplicaStatusQueue: tt.fields.syncReplicaStatusQueue,
			}
			if err := m.processReplicaStatusUpdate(tt.args.replicaName); (err != nil) != tt.wantErr {
				t.Errorf("processReplicaStatusUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_configManager_startReplicaStatusSyncWorker(t *testing.T) {
	type fields struct {
		hostname               string
		systemConfig           v1alpha1.SystemConfig
		apiClient              client.Client
		configer               configer.Configer
		logger                 *log.Entry
		syncReplicaStatusQueue *common.TaskQueue
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
			m := &configManager{
				hostname:               tt.fields.hostname,
				systemConfig:           tt.fields.systemConfig,
				apiClient:              tt.fields.apiClient,
				configer:               tt.fields.configer,
				logger:                 tt.fields.logger,
				syncReplicaStatusQueue: tt.fields.syncReplicaStatusQueue,
			}
			m.startReplicaStatusSyncWorker(tt.args.stopCh)
		})
	}
}

func Test_configManager_syncReplicaStatus(t *testing.T) {
	type fields struct {
		hostname               string
		systemConfig           v1alpha1.SystemConfig
		apiClient              client.Client
		configer               configer.Configer
		logger                 *log.Entry
		syncReplicaStatusQueue *common.TaskQueue
	}
	type args struct {
		replicaName string
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
			m := &configManager{
				hostname:               tt.fields.hostname,
				systemConfig:           tt.fields.systemConfig,
				apiClient:              tt.fields.apiClient,
				configer:               tt.fields.configer,
				logger:                 tt.fields.logger,
				syncReplicaStatusQueue: tt.fields.syncReplicaStatusQueue,
			}
			m.syncReplicaStatus(tt.args.replicaName)
		})
	}
}

func Test_configManager_updateConfig(t *testing.T) {
	type fields struct {
		hostname               string
		systemConfig           v1alpha1.SystemConfig
		apiClient              client.Client
		configer               configer.Configer
		logger                 *log.Entry
		syncReplicaStatusQueue *common.TaskQueue
	}
	type args struct {
		replica *v1alpha1.LocalVolumeReplica
		config  *v1alpha1.VolumeConfig
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
			m := &configManager{
				hostname:               tt.fields.hostname,
				systemConfig:           tt.fields.systemConfig,
				apiClient:              tt.fields.apiClient,
				configer:               tt.fields.configer,
				logger:                 tt.fields.logger,
				syncReplicaStatusQueue: tt.fields.syncReplicaStatusQueue,
			}
			if err := m.updateConfig(tt.args.replica, tt.args.config); (err != nil) != tt.wantErr {
				t.Errorf("updateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
