package scheduler

import (
	"fmt"
	"github.com/golang/mock/gomock"
	apisv1alpha1 "github.com/hwameistor/local-storage/pkg/apis/hwameistor/v1alpha1"
	vgmock "github.com/hwameistor/local-storage/pkg/member/controller/volumegroup"
	"reflect"
	"testing"
)

func Test_scheduler_Allocate(t *testing.T) {

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

	var vc = &apisv1alpha1.VolumeConfig{}

	m := vgmock.NewMockVolumeScheduler(ctrl)
	m.
		EXPECT().
		Allocate(vol).
		Return(vc, nil).
		Times(1)

	v, err := m.Allocate(vol)

	fmt.Printf("Test_scheduler_Allocate v= %+v, err= %+v", v, err)

}

func Test_scheduler_GetNodeCandidates(t *testing.T) {

	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()

	var volList []*apisv1alpha1.LocalVolume
	var vol = &apisv1alpha1.LocalVolume{}
	vol.Name = "vol1"
	vol.Namespace = "test1"
	vol.Spec.RequiredCapacityBytes = 1240
	vol.Spec.PoolName = "pool1"
	vol.Spec.Accessibility.Nodes = []string{"node1"}
	volList = append(volList, vol)

	var lsns = []*apisv1alpha1.LocalStorageNode{}

	m := vgmock.NewMockVolumeScheduler(ctrl)
	m.
		EXPECT().
		GetNodeCandidates(volList).
		Return(lsns).
		Times(1)

	v := m.GetNodeCandidates(volList)
	fmt.Printf("Test_scheduler_GetNodeCandidates v= %+v", v)

}

func Test_scheduler_Init(t *testing.T) {

	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()

	m := vgmock.NewMockVolumeScheduler(ctrl)
	m.
		EXPECT().
		Init().
		Return().
		Times(1)

	m.Init()
}

func Test_isLocalVolumeSameClass(t *testing.T) {
	type args struct {
		lv1 *apisv1alpha1.LocalVolume
		lv2 *apisv1alpha1.LocalVolume
	}
	var lv1 *apisv1alpha1.LocalVolume
	var lv2 *apisv1alpha1.LocalVolume
	var lv12 = &apisv1alpha1.LocalVolume{}
	lv12.Name = "test12"
	lv12.Spec.PoolName = "pool12"
	var lv22 = &apisv1alpha1.LocalVolume{}
	lv22.Name = "test12"
	lv22.Spec.PoolName = "pool22"

	var lv13 = &apisv1alpha1.LocalVolume{}
	lv13.Name = "test13"
	lv13.Spec.PoolName = "pool13"
	lv13.Spec.ReplicaNumber = 1
	lv13.Spec.Convertible = true

	var lv23 = &apisv1alpha1.LocalVolume{}
	lv23.Name = "test13"
	lv23.Spec.PoolName = "pool13"
	lv23.Spec.ReplicaNumber = 2
	lv23.Spec.Convertible = true

	var lv14 = &apisv1alpha1.LocalVolume{}
	lv13.Name = "test13"
	lv13.Spec.PoolName = "pool13"
	lv13.Spec.ReplicaNumber = 1
	lv13.Spec.Convertible = true

	var lv24 = &apisv1alpha1.LocalVolume{}
	lv23.Name = "test13"
	lv23.Spec.PoolName = "pool13"
	lv23.Spec.ReplicaNumber = 1
	lv23.Spec.Convertible = false

	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
		{
			args: args{lv1: lv1, lv2: lv2},
			want: true,
		},
		{
			args: args{lv1: lv12, lv2: lv22},
			want: false,
		},
		{
			args: args{lv1: lv13, lv2: lv23},
			want: false,
		},
		{
			args: args{lv1: lv14, lv2: lv24},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isLocalVolumeSameClass(tt.args.lv1, tt.args.lv2); got != tt.want {
				t.Errorf("isLocalVolumeSameClass() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_appendLocalVolume(t *testing.T) {
	type args struct {
		bigLv *apisv1alpha1.LocalVolume
		lv    *apisv1alpha1.LocalVolume
	}
	var bigLv *apisv1alpha1.LocalVolume
	var lv = &apisv1alpha1.LocalVolume{}
	lv.Name = "lv"

	var bigLv1 = &apisv1alpha1.LocalVolume{}
	bigLv1.Name = "bigLv1"
	var lv1 *apisv1alpha1.LocalVolume

	var bigLv2 = &apisv1alpha1.LocalVolume{}
	bigLv2.Name = "bigLv2"
	bigLv2.Spec.RequiredCapacityBytes = 1240
	var lv2 = &apisv1alpha1.LocalVolume{}
	lv2.Name = "lv2"
	lv2.Spec.RequiredCapacityBytes = 1240

	var bigLv22 = &apisv1alpha1.LocalVolume{}
	bigLv22 = bigLv2.DeepCopy()
	bigLv2.Spec.RequiredCapacityBytes = 2480

	tests := []struct {
		name string
		args args
		want *apisv1alpha1.LocalVolume
	}{
		// TODO: Add test cases.
		{
			args: args{bigLv: bigLv, lv: lv},
			want: lv,
		},
		{
			args: args{bigLv: bigLv1, lv: lv1},
			want: bigLv1,
		},
		{
			args: args{bigLv: bigLv2, lv: lv2},
			want: bigLv22,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := appendLocalVolume(tt.args.bigLv, tt.args.lv); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("appendLocalVolume() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_unionSet(t *testing.T) {
	type args struct {
		strs1 []*apisv1alpha1.LocalStorageNode
		strs2 []*apisv1alpha1.LocalStorageNode
	}
	var strs1 []*apisv1alpha1.LocalStorageNode
	var strs2 []*apisv1alpha1.LocalStorageNode
	var lsn1 = &apisv1alpha1.LocalStorageNode{}
	lsn1.Name = "lsn1"
	strs1 = append(strs1, lsn1)

	var lsn2 = &apisv1alpha1.LocalStorageNode{}
	lsn2.Name = "lsn2"
	strs2 = append(strs2, lsn2)

	strs := []*apisv1alpha1.LocalStorageNode{}

	var strs11 []*apisv1alpha1.LocalStorageNode
	var strs21 []*apisv1alpha1.LocalStorageNode
	var lsn11 = &apisv1alpha1.LocalStorageNode{}
	lsn11.Name = "lsn1"
	strs11 = append(strs11, lsn11)

	var lsn21 = &apisv1alpha1.LocalStorageNode{}
	lsn21.Name = "lsn1"
	strs21 = append(strs21, lsn21)

	tests := []struct {
		name string
		args args
		want []*apisv1alpha1.LocalStorageNode
	}{
		// TODO: Add test cases.
		{
			args: args{strs1: strs1, strs2: strs2},
			want: strs,
		},
		{
			args: args{strs1: strs11, strs2: strs21},
			want: strs11,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := unionSet(tt.args.strs1, tt.args.strs2); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("unionSet() = %v, want %v", got, tt.want)
			}
		})
	}
}
