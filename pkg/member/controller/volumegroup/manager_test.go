package volumegroup

import (
	"fmt"
	"github.com/golang/mock/gomock"
	apisv1alpha1 "github.com/hwameistor/local-storage/pkg/apis/hwameistor/v1alpha1"
	"testing"
)

func TestNewManager(t *testing.T) {

}

func Test_manager_GetLocalVolumeGroupByLocalVolume(t *testing.T) {
	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()

	var ns = "test_ns"
	var lvName = "test_lv_name"
	var lvg = &apisv1alpha1.LocalVolumeGroup{}

	m := NewMockVolumeGroupManager(ctrl)
	m.
		EXPECT().
		GetLocalVolumeGroupByLocalVolume(ns, lvName).
		Return(lvg, nil).
		Times(1)

	lvg, err := m.GetLocalVolumeGroupByLocalVolume(ns, lvName)
	fmt.Printf("Test_manager_GetLocalVolumeGroupByLocalVolume err = %+v", err)
	if err != nil {
		t.Fatal()
	}
	fmt.Printf("Test_manager_GetLocalVolumeGroupByLocalVolume lvg= %+v", lvg)
}

func Test_manager_GetLocalVolumeGroupByName(t *testing.T) {
	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()

	var ns = "test_ns"
	var lvgName = "test_lvg_name"
	var lvg = &apisv1alpha1.LocalVolumeGroup{}

	m := NewMockVolumeGroupManager(ctrl)
	m.
		EXPECT().
		GetLocalVolumeGroupByName(ns, lvgName).
		Return(lvg, nil).
		Times(1)

	lvg, err := m.GetLocalVolumeGroupByName(ns, lvgName)
	fmt.Printf("Test_manager_GetLocalVolumeGroupByName err = %+v", err)
	if err != nil {
		t.Fatal()
	}
	fmt.Printf("Test_manager_GetLocalVolumeGroupByName lvg= %+v", lvg)
}

func Test_manager_GetLocalVolumeGroupByPVC(t *testing.T) {
	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()

	var pvc_ns = "test_ns"
	var pvc_name = "test_pvc_name"
	var lvg = &apisv1alpha1.LocalVolumeGroup{}

	m := NewMockVolumeGroupManager(ctrl)
	m.
		EXPECT().
		GetLocalVolumeGroupByPVC(pvc_ns, pvc_name).
		Return(lvg, nil).
		Times(1)

	lvg, err := m.GetLocalVolumeGroupByPVC(pvc_ns, pvc_name)
	fmt.Printf("Test_manager_GetLocalVolumeGroupByPVC err = %+v", err)
	if err != nil {
		t.Fatal()
	}
	fmt.Printf("Test_manager_GetLocalVolumeGroupByPVC lvg= %+v", lvg)
}

func Test_manager_Init(t *testing.T) {
	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()
	var stopCh <-chan struct{}

	m := NewMockVolumeGroupManager(ctrl)
	m.
		EXPECT().
		Init(stopCh).
		Return().
		Times(1)

	m.Init(stopCh)
}

func Test_manager_ReconcileVolumeGroup(t *testing.T) {

	var lvg = &apisv1alpha1.LocalVolumeGroup{}
	lvg.Name = "test_lvg_name"
	lvg.Namespace = "test_lvg_ns"
	lvg.Spec.Accessibility.Nodes = []string{"node1"}
	lvg.Spec.Pods = []string{"pod1"}
	lvg.Spec.Volumes = []apisv1alpha1.VolumeInfo{{LocalVolumeName: "local-volume-test1", PersistentVolumeClaimName: "pvc-test1"}}

	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()

	m := NewMockVolumeGroupManager(ctrl)
	m.
		EXPECT().
		ReconcileVolumeGroup(lvg).
		Return().
		Times(1)

	m.ReconcileVolumeGroup(lvg)
}

func Test_namespacedName(t *testing.T) {
	type args struct {
		namespace string
		name      string
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
			if got := namespacedName(tt.args.namespace, tt.args.name); got != tt.want {
				t.Errorf("namespacedName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseNamespacedName(t *testing.T) {
	type args struct {
		nn string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := parseNamespacedName(tt.args.nn)
			if got != tt.want {
				t.Errorf("parseNamespacedName() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("parseNamespacedName() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
