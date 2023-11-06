package configer

import (
	"fmt"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
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
	var localVolumeReplica = &apisv1alpha1.LocalVolumeReplica{}
	localVolumeReplica.Spec.VolumeName = "volume1"
	localVolumeReplica.Spec.PoolName = "pool1"
	localVolumeReplica.Spec.NodeName = "node1"
	localVolumeReplica.Spec.RequiredCapacityBytes = 1240
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
	var localVolumeReplica = &apisv1alpha1.LocalVolumeReplica{}
	localVolumeReplica.Spec.VolumeName = "volume1"
	localVolumeReplica.Spec.PoolName = "pool1"
	localVolumeReplica.Spec.NodeName = "node1"
	localVolumeReplica.Spec.RequiredCapacityBytes = 1240
	localVolumeReplica.Name = "test1"

	var config apisv1alpha1.VolumeConfig
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
	var localVolumeReplica = &apisv1alpha1.LocalVolumeReplica{}
	localVolumeReplica.Spec.VolumeName = "volume1"
	localVolumeReplica.Spec.PoolName = "pool1"
	localVolumeReplica.Spec.NodeName = "node1"
	localVolumeReplica.Spec.RequiredCapacityBytes = 1240
	localVolumeReplica.Name = "test1"

	var config apisv1alpha1.VolumeConfig
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
	var localVolumeReplica = &apisv1alpha1.LocalVolumeReplica{}
	localVolumeReplica.Spec.VolumeName = "volume1"
	localVolumeReplica.Spec.PoolName = "pool1"
	localVolumeReplica.Spec.NodeName = "node1"
	localVolumeReplica.Spec.RequiredCapacityBytes = 1240
	localVolumeReplica.Name = "test1"

	var config apisv1alpha1.VolumeConfig
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
		Initialize(localVolumeReplica, config).
		Return(nil).
		Times(1)

	v := m.Initialize(localVolumeReplica, config)

	fmt.Printf("Test_drbdConfigure_Initialize v= %+v", v)
}

func Test_drbdConfigure_ConsistencyCheck(t *testing.T) {
	var lvrs []apisv1alpha1.LocalVolumeReplica
	var localVolumeReplica = apisv1alpha1.LocalVolumeReplica{}
	localVolumeReplica.Spec.VolumeName = "volume1"
	localVolumeReplica.Spec.PoolName = "pool1"
	localVolumeReplica.Spec.NodeName = "node1"
	localVolumeReplica.Spec.RequiredCapacityBytes = 1240
	localVolumeReplica.Name = "test1"
	lvrs = append(lvrs, localVolumeReplica)

	var config apisv1alpha1.VolumeConfig
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
		ConsistencyCheck(lvrs).
		Return().
		Times(1)

	m.ConsistencyCheck(lvrs)

	fmt.Printf("Test_drbdConfigure_ConsistencyCheck")
}

func Test_genConfigPath(t *testing.T) {
	type args struct {
		resourceName string
	}
	var wantRes = "/etc/drbd.d/test.res"
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{
			args: args{resourceName: "test"},
			want: wantRes,
		},
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
		{
			args: args{configName: "test.res"},
			want: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseConfigFileName(tt.args.configName); got != tt.want {
				t.Errorf("parseConfigFileName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_drbdConfigure_DeleteConfig(t *testing.T) {
	var localVolumeReplica = &apisv1alpha1.LocalVolumeReplica{}
	localVolumeReplica.Spec.VolumeName = "volume1"
	localVolumeReplica.Spec.PoolName = "pool1"
	localVolumeReplica.Spec.NodeName = "node1"
	localVolumeReplica.Spec.RequiredCapacityBytes = 1240
	localVolumeReplica.Name = "test1"

	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()

	m := NewMockConfiger(ctrl)
	m.
		EXPECT().
		DeleteConfig(localVolumeReplica).
		Return(nil).
		Times(1)

	err := m.DeleteConfig(localVolumeReplica)

	fmt.Printf("Test_drbdConfigure_DeleteConfig err = %v", err)
}

func Test_drbdConfigure_GetReplicaHAState(t *testing.T) {
	var localVolumeReplica = &apisv1alpha1.LocalVolumeReplica{}
	localVolumeReplica.Spec.VolumeName = "volume1"
	localVolumeReplica.Spec.PoolName = "pool1"
	localVolumeReplica.Spec.NodeName = "node1"
	localVolumeReplica.Spec.RequiredCapacityBytes = 1240
	localVolumeReplica.Name = "test1"
	var haState apisv1alpha1.HAState

	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()

	m := NewMockConfiger(ctrl)
	m.
		EXPECT().
		GetReplicaHAState(localVolumeReplica).
		Return(haState, nil).
		Times(1)

	haState, err := m.GetReplicaHAState(localVolumeReplica)

	fmt.Printf("Test_drbdConfigure_GetReplicaHAState haState = %v err = %v", haState, err)
}
func Test_handleDRBDEvent(t *testing.T) {

	testCases := []struct {
		Description  string
		HostName     string
		SysCfg       apisv1alpha1.SystemConfig
		ApiClient    client.Client
		SyncFunc     SyncReplicaStatus
		Events       []string
		ExpectResult map[string]*Resource
	}{
		{
			Description: "It is a drbdsetup events2 scsivol(resourceName) command result,caller:drbd-node3,peer-hostname:drbd-node1,peer-node-id:0",
			HostName:    "drbd-node3",
			SysCfg:      apisv1alpha1.SystemConfig{},
			ApiClient:   nil,
			SyncFunc:    nil,
			Events: strings.Split("exists resource name:scsivol role:Primary suspended:no\n"+
				"exists connection name:scsivol peer-node-id:0 conn-name:drbd-node1 connection:Connected role:Secondary\n"+
				"exists device name:scsivol volume:0 minor:0 disk:UpToDate client:no quorum:yes\n"+
				"exists peer-device name:scsivol peer-node-id:0 conn-name:drbd-node1 volume:0 replication:Established peer-disk:UpToDate peer-client:no resync-suspended:no", "\n"),
			ExpectResult: map[string]*Resource{"scsivol": &Resource{
				Name:        "scsivol",
				Role:        "Primary",
				Device:      struct{ State string }{State: "UpToDate"},
				Replication: "",
				PeerDevices: map[string]*PeerDevice{
					"drbd-node1": &PeerDevice{
						NodeID: 0, ConnectionName: "drbd-node1", Replication: "Established",
						DiskState: "UpToDate",
					},
				},
			},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Description, func(t *testing.T) {
			configer, err := NewDRBDConfiger(testCase.HostName, testCase.SysCfg, testCase.ApiClient, testCase.SyncFunc)
			if err != nil {
				t.Fatal("init DRBDConfiger failed")
			}
			for _, e := range testCase.Events {
				configer.handleDRBDEvent(e)
			}
			if !reflect.DeepEqual(configer.resourceCache, testCase.ExpectResult) {
				t.Fatal("resourceCache should be the same with the ExpectResult")
			}
		})
	}

}
