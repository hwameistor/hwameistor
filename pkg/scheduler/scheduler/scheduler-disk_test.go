package scheduler

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	corev1 "k8s.io/api/core/v1"

	"github.com/hwameistor/hwameistor/pkg/scheduler/genscheduler"
)

func TestDiskVolumeScheduler_CSIDriverName(t *testing.T) {
	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()
	//
	m := genscheduler.NewMockVolumeScheduler(ctrl)
	m.
		EXPECT().CSIDriverName().
		Return("disk.hwameistor.io").
		Times(1)

	v := m.CSIDriverName()

	fmt.Printf("TestDiskVolumeScheduler_CSIDriverName v= %+v", v)
}

func TestDiskVolumeScheduler_Filter(t *testing.T) {
	var lvs []string
	lvs = append(lvs, "lv1")
	lvs = append(lvs, "lv2")
	var pendingPVCs []*corev1.PersistentVolumeClaim
	pvc1 := &corev1.PersistentVolumeClaim{}
	pvc1.Name = "pvc1"
	pvc1.Namespace = "default"
	pvc1.Spec.VolumeName = "lv1"
	pendingPVCs = append(pendingPVCs, pvc1)

	node1 := &corev1.Node{}
	node1.Name = "node1"
	node1.Namespace = "default"
	node1.Spec.Unschedulable = false

	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()
	//
	m := genscheduler.NewMockVolumeScheduler(ctrl)

	m.
		EXPECT().Filter(lvs, pendingPVCs, node1).
		Return(true, nil).
		Times(1)

	v, err := m.Filter(lvs, pendingPVCs, node1)
	if err != nil {
		t.Errorf("TestDiskVolumeScheduler_Filter Filter() err = %v ", err)
	}
	fmt.Printf("TestDiskVolumeScheduler_Filter v= %+v", v)
}
