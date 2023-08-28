package storage

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

func Test_localRegistry_Disks(t *testing.T) {
	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()
	m := NewMockLocalRegistry(ctrl)
	m.
		EXPECT().
		Disks().
		Return(nil).
		Times(1)

	v := m.Disks()
	fmt.Printf("Test_localRegistry_Disks result v= %+v", v)
}

func Test_localRegistry_HasVolumeReplica(t *testing.T) {
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
	m := NewMockLocalRegistry(ctrl)
	m.
		EXPECT().
		HasVolumeReplica(localVolumeReplica).
		Return(false).
		Times(1)

	v := m.HasVolumeReplica(localVolumeReplica)
	fmt.Printf("Test_localRegistry_HasVolumeReplica result v= %+v", v)
}

func Test_localRegistry_Pools(t *testing.T) {
	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()
	m := NewMockLocalRegistry(ctrl)
	m.
		EXPECT().
		Pools().
		Return(nil).
		Times(1)

	v := m.Pools()
	fmt.Printf("Test_localRegistry_Pools result v= %+v", v)

}

func Test_localRegistry_UpdateNodeForVolumeReplica(t *testing.T) {
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
	m := NewMockLocalRegistry(ctrl)
	m.
		EXPECT().
		UpdateNodeForVolumeReplica(localVolumeReplica).
		Return().
		Times(1)

	m.UpdateNodeForVolumeReplica(localVolumeReplica)
	fmt.Printf("Test_localRegistry_UpdateNodeForVolumeReplica ends")
}

func Test_localRegistry_VolumeReplicas(t *testing.T) {
	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()
	m := NewMockLocalRegistry(ctrl)
	m.
		EXPECT().
		VolumeReplicas().
		Return(nil).
		Times(1)

	v := m.VolumeReplicas()
	fmt.Printf("Test_localRegistry_VolumeReplicas result v= %+v", v)

}

func Test_newLocalRegistry(t *testing.T) {
	//type args struct {
	//	lm *LocalManager
	//}
	//tests := []struct {
	//	name string
	//	args args
	//	want LocalRegistry
	//}{
	//	// TODO: Add test cases.
	//}
	//for _, tt := range tests {
	//	t.Run(tt.name, func(t *testing.T) {
	//		if got := newLocalRegistry(tt.args.lm); !reflect.DeepEqual(got, tt.want) {
	//			t.Errorf("newLocalRegistry() = %v, want %v", got, tt.want)
	//		}
	//	})
	//}
}

func Test_localRegistry_Init(t *testing.T) {
	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()
	m := NewMockLocalRegistry(ctrl)
	m.
		EXPECT().
		Init().
		Return().
		Times(1)

	m.Init()
}
