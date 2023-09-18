package storage

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

func Test_lvmExecutor_CreateVolumeReplica(t *testing.T) {
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
	m := NewMockLocalVolumeExecutor(ctrl)
	m.
		EXPECT().
		CreateVolumeReplica(localVolumeReplica).
		Return(localVolumeReplica, nil).
		Times(1)

	lvr, err := m.CreateVolumeReplica(localVolumeReplica)
	fmt.Printf("Test_lvmExecutor_CreateVolumeReplica lvr = %v, err = %v", lvr, err)
}

func Test_lvmExecutor_GetReplicas(t *testing.T) {
	var localVolumeReplica = &apisv1alpha1.LocalVolumeReplica{}
	localVolumeReplica.Spec.VolumeName = "volume1"
	localVolumeReplica.Spec.PoolName = "pool1"
	localVolumeReplica.Spec.NodeName = "node1"
	localVolumeReplica.Spec.RequiredCapacityBytes = 1240
	localVolumeReplica.Name = "test1"
	var lvrmap = make(map[string]*apisv1alpha1.LocalVolumeReplica)
	lvrmap["test"] = localVolumeReplica

	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()
	m := NewMockLocalVolumeExecutor(ctrl)
	m.
		EXPECT().
		GetReplicas().
		Return(lvrmap, nil).
		Times(1)

	lvrmap, err := m.GetReplicas()
	fmt.Printf("Test_lvmExecutor_GetReplicas = %v, err = %v", lvrmap, err)
}

func Test_lvmExecutor_ExtendPoolsInfo(t *testing.T) {
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

func Test_lvmExecutor_ExpandVolumeReplica(t *testing.T) {
	var localVolumeReplica = &apisv1alpha1.LocalVolumeReplica{}
	localVolumeReplica.Spec.VolumeName = "volume1"
	localVolumeReplica.Spec.PoolName = "pool1"
	localVolumeReplica.Spec.NodeName = "node1"
	localVolumeReplica.Spec.RequiredCapacityBytes = 1240
	localVolumeReplica.Name = "test1"
	var newCapacity int64 = 102400

	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()
	m := NewMockLocalVolumeExecutor(ctrl)
	m.
		EXPECT().
		ExpandVolumeReplica(localVolumeReplica, newCapacity).
		Return(localVolumeReplica, nil).
		Times(1)

	lvr, err := m.ExpandVolumeReplica(localVolumeReplica, newCapacity)
	fmt.Printf("Test_lvmExecutor_ExpandVolumeReplica lvr = %v, err = %v", lvr, err)
}

func Test_lvmExecutor_DeleteVolumeReplica(t *testing.T) {
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
	m := NewMockLocalVolumeExecutor(ctrl)
	m.
		EXPECT().
		DeleteVolumeReplica(localVolumeReplica).
		Return(nil).
		Times(1)

	err := m.DeleteVolumeReplica(localVolumeReplica)
	fmt.Printf("Test_lvmExecutor_DeleteVolumeReplica err = %v", err)
}

func TestVolumeReplica(t *testing.T) {
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
	m := NewMockLocalVolumeExecutor(ctrl)
	m.
		EXPECT().
		TestVolumeReplica(localVolumeReplica).
		Return(localVolumeReplica, nil).
		Times(1)

	lvr, err := m.TestVolumeReplica(localVolumeReplica)
	fmt.Printf("TestVolumeReplica lvr = %v, err = %v", lvr, err)
}

func TestVolumeReplica2(t *testing.T) {
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
		TestVolumeReplica(localVolumeReplica).
		Return(localVolumeReplica, nil).
		Times(1)

	lvr, err := m.TestVolumeReplica(localVolumeReplica)
	fmt.Printf("TestVolumeReplica lvr = %v, err = %v", lvr, err)
}

func Test_lvmExecutor_ExtendPools(t *testing.T) {
	disks := make([]*apisv1alpha1.LocalDevice, 0, 10)
	var localdisk1 = &apisv1alpha1.LocalDevice{}
	localdisk1.DevPath = "/dev/sdf"
	localdisk1.State = apisv1alpha1.DiskStateAvailable
	localdisk1.Class = "HDD"
	localdisk1.CapacityBytes = 2 * 1024 * 1024 * 1024
	disks = append(disks, localdisk1)

	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()
	m := NewMockLocalPoolExecutor(ctrl)
	m.
		EXPECT().
		ExtendPools(disks).
		Return(nil).
		Times(1)

	_, err := m.ExtendPools(disks)
	fmt.Printf("Test_lvmExecutor_ExtendPools err = %v", err)
}

func Test_lvmExecutor_ConsistencyCheck(t *testing.T) {
	var localVolumeReplica = &apisv1alpha1.LocalVolumeReplica{}
	localVolumeReplica.Spec.VolumeName = "volume1"
	localVolumeReplica.Spec.PoolName = "pool1"
	localVolumeReplica.Spec.NodeName = "node1"
	localVolumeReplica.Spec.RequiredCapacityBytes = 1240
	localVolumeReplica.Name = "test1"
	var lvrmap = make(map[string]*apisv1alpha1.LocalVolumeReplica)
	lvrmap["test"] = localVolumeReplica

	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()
	m := NewMockLocalVolumeExecutor(ctrl)
	m.
		EXPECT().
		ConsistencyCheck(lvrmap).
		Return().
		Times(1)

	m.ConsistencyCheck(lvrmap)
	fmt.Printf("Test_lvmExecutor_ConsistencyCheck ends")
}
