package storage

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

func Test_localVolumeReplicaManager_CreateVolumeReplica(t *testing.T) {
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
	m := NewMockLocalVolumeReplicaManager(ctrl)
	m.
		EXPECT().
		CreateVolumeReplica(localVolumeReplica).
		Return(nil, nil).
		Times(1)

	v, err := m.CreateVolumeReplica(localVolumeReplica)
	fmt.Printf("test err v= %+v", err)
	if err != nil {
		t.Fatal()
	}
	fmt.Printf("Test_localVolumeReplicaManager_CreateVolumeReplica result v= %+v", v)
}

func Test_localVolumeReplicaManager_DeleteVolumeReplica(t *testing.T) {
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
	m := NewMockLocalVolumeReplicaManager(ctrl)
	m.
		EXPECT().
		DeleteVolumeReplica(localVolumeReplica).
		Return(nil).
		Times(1)

	v := m.DeleteVolumeReplica(localVolumeReplica)
	fmt.Printf("Test_localVolumeReplicaManager_DeleteVolumeReplica result v= %+v", v)
}

func Test_localVolumeReplicaManager_ExpandVolumeReplica(t *testing.T) {
	var localVolumeReplica = &apisv1alpha1.LocalVolumeReplica{}
	localVolumeReplica.Spec.VolumeName = "volume1"
	localVolumeReplica.Spec.PoolName = "pool1"
	localVolumeReplica.Spec.NodeName = "node1"
	localVolumeReplica.Spec.RequiredCapacityBytes = 1240
	localVolumeReplica.Name = "test1"
	newCapacityBytes := 10240

	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()
	m := NewMockLocalVolumeReplicaManager(ctrl)
	m.
		EXPECT().
		ExpandVolumeReplica(localVolumeReplica, int64(newCapacityBytes)).
		Return(nil, nil).
		Times(1)

	v, err := m.ExpandVolumeReplica(localVolumeReplica, int64(newCapacityBytes))
	fmt.Printf("Test_localVolumeReplicaManager_ExpandVolumeReplica err v= %+v", err)
	if err != nil {
		t.Fatal()
	}
	fmt.Printf("Test_localVolumeReplicaManager_ExpandVolumeReplica result v= %+v", v)
}

func Test_localVolumeReplicaManager_GetVolumeReplica(t *testing.T) {
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
	m := NewMockLocalVolumeReplicaManager(ctrl)
	m.
		EXPECT().
		GetVolumeReplica(localVolumeReplica).
		Return(nil, nil).
		Times(1)

	v, err := m.GetVolumeReplica(localVolumeReplica)
	fmt.Printf("Test_localVolumeReplicaManager_GetVolumeReplica err v= %+v", err)
	if err != nil {
		t.Fatal()
	}
	fmt.Printf("Test_localVolumeReplicaManager_GetVolumeReplica result v= %+v", v)
}

func Test_localVolumeReplicaManager_ConsistencyCheck(t *testing.T) {
	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法åå
	defer ctrl.Finish()
	m := NewMockLocalVolumeReplicaManager(ctrl)
	m.
		EXPECT().
		ConsistencyCheck().
		Return().
		Times(1)

	m.ConsistencyCheck()
	fmt.Printf("Test_localVolumeReplicaManager_ConsistencyCheck ends")
}
