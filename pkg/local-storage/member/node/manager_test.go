package node

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	memmock "github.com/hwameistor/hwameistor/pkg/local-storage/member/controller"
	"github.com/hwameistor/hwameistor/pkg/local-storage/member/node/storage"
)

func Test_manager_ReconcileVolumeReplica(t *testing.T) {

	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()

	var localVolumeReplica = &apisv1alpha1.LocalVolumeReplica{}
	localVolumeReplica.Spec.VolumeName = "volume1"
	localVolumeReplica.Spec.PoolName = "pool1"
	localVolumeReplica.Spec.NodeName = "node1"
	localVolumeReplica.Spec.RequiredCapacityBytes = 1240
	localVolumeReplica.Name = "test1"

	m := memmock.NewMockNodeManager(ctrl)
	m.
		EXPECT().
		ReconcileVolumeReplica(localVolumeReplica).
		Return().
		Times(1)

	m.ReconcileVolumeReplica(localVolumeReplica)

}

func Test_manager_Run(t *testing.T) {
	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()

	var stopCh <-chan struct{}
	m := memmock.NewMockNodeManager(ctrl)
	m.
		EXPECT().
		Run(stopCh).
		Return().
		Times(1)

	m.Run(stopCh)
}

func Test_manager_Storage(t *testing.T) {

	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()

	var lm = &storage.LocalManager{}

	m := memmock.NewMockNodeManager(ctrl)
	m.
		EXPECT().
		Storage().
		Return(lm).
		Times(1)

	v := m.Storage()

	fmt.Printf("Test_manager_Storage v= %+v", v)

}

func Test_manager_TakeVolumeReplicaTaskAssignment(t *testing.T) {

	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()

	var vol = &apisv1alpha1.LocalVolume{}
	vol.Name = "vol1"
	vol.Namespace = "test1"
	vol.Spec.RequiredCapacityBytes = 1240
	vol.Spec.PoolName = "pool1"
	vol.Spec.Accessibility.Nodes = []string{"node1"}

	m := memmock.NewMockNodeManager(ctrl)
	m.
		EXPECT().
		TakeVolumeReplicaTaskAssignment(vol).
		Return().
		Times(1)

	m.TakeVolumeReplicaTaskAssignment(vol)

}
