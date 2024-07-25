package scheduler

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"os"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	vgmock "github.com/hwameistor/hwameistor/pkg/local-storage/member/controller/volumegroup"
)

func Test_scheduler_Allocate(t *testing.T) {

	os.Setenv("KUBERNETES_MASTER", "test")
	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()

	var vol = &v1alpha1.LocalVolume{}
	vol.Name = "vol1"
	vol.Namespace = "test1"
	vol.Spec.RequiredCapacityBytes = 1240
	vol.Spec.PoolName = "pool1"
	vol.Spec.Accessibility.Nodes = []string{"node1"}

	var vc = &v1alpha1.VolumeConfig{}

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
	os.Setenv("KUBERNETES_MASTER", "test")

	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()

	var volList []*v1alpha1.LocalVolume
	var vol = &v1alpha1.LocalVolume{}
	vol.Name = "vol1"
	vol.Namespace = "test1"
	vol.Spec.RequiredCapacityBytes = 1240
	vol.Spec.PoolName = "pool1"
	vol.Spec.Accessibility.Nodes = []string{"node1"}
	volList = append(volList, vol)

	var lsns = []*v1alpha1.LocalStorageNode{}

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
	os.Setenv("KUBERNETES_MASTER", "test")

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
	os.Setenv("KUBERNETES_MASTER", "test")

	type args struct {
		lv1 *v1alpha1.LocalVolume
		lv2 *v1alpha1.LocalVolume
	}
	var lv1 *v1alpha1.LocalVolume
	var lv2 *v1alpha1.LocalVolume
	var lv12 = &v1alpha1.LocalVolume{}
	lv12.Name = "test12"
	lv12.Spec.PoolName = "pool12"
	var lv22 = &v1alpha1.LocalVolume{}
	lv22.Name = "test12"
	lv22.Spec.PoolName = "pool22"

	var lv13 = &v1alpha1.LocalVolume{}
	lv13.Name = "test13"
	lv13.Spec.PoolName = "pool13"
	lv13.Spec.ReplicaNumber = 1
	lv13.Spec.Convertible = true

	var lv23 = &v1alpha1.LocalVolume{}
	lv23.Name = "test13"
	lv23.Spec.PoolName = "pool13"
	lv23.Spec.ReplicaNumber = 2
	lv23.Spec.Convertible = true

	var lv14 = &v1alpha1.LocalVolume{}
	lv13.Name = "test13"
	lv13.Spec.PoolName = "pool13"
	lv13.Spec.ReplicaNumber = 1
	lv13.Spec.Convertible = true

	var lv24 = &v1alpha1.LocalVolume{}
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
			want: true,
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
	os.Setenv("KUBERNETES_MASTER", "test")

	type args struct {
		bigLv *v1alpha1.LocalVolume
		lv    *v1alpha1.LocalVolume
	}
	var bigLv *v1alpha1.LocalVolume
	var lv = &v1alpha1.LocalVolume{}
	lv.Name = "lv"

	var bigLv1 = &v1alpha1.LocalVolume{}
	bigLv1.Name = "bigLv1"
	var lv1 *v1alpha1.LocalVolume

	var bigLv2 = &v1alpha1.LocalVolume{}
	bigLv2.Name = "bigLv2"
	bigLv2.Spec.RequiredCapacityBytes = 1240

	var lv2 = &v1alpha1.LocalVolume{}
	lv2.Name = "lv2"
	lv2.Spec.RequiredCapacityBytes = 1240

	var bigLv22 = &v1alpha1.LocalVolume{}
	bigLv22 = bigLv2.DeepCopy()
	bigLv22.Spec.RequiredCapacityBytes = 2480

	tests := []struct {
		name string
		args args
		want *v1alpha1.LocalVolume
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
	os.Setenv("KUBERNETES_MASTER", "test")

	type args struct {
		strs1 []*v1alpha1.LocalStorageNode
		strs2 []*v1alpha1.LocalStorageNode
	}
	var strs1 []*v1alpha1.LocalStorageNode
	var strs2 []*v1alpha1.LocalStorageNode
	var lsn1 = &v1alpha1.LocalStorageNode{}
	lsn1.Name = "lsn1"
	strs1 = append(strs1, lsn1)

	var lsn2 = &v1alpha1.LocalStorageNode{}
	lsn2.Name = "lsn2"
	strs2 = append(strs2, lsn2)

	strs := []*v1alpha1.LocalStorageNode{}

	var strs11 []*v1alpha1.LocalStorageNode
	var strs21 []*v1alpha1.LocalStorageNode
	var lsn11 = &v1alpha1.LocalStorageNode{}
	lsn11.Name = "lsn1"
	strs11 = append(strs11, lsn11)

	var lsn21 = &v1alpha1.LocalStorageNode{}
	lsn21.Name = "lsn1"
	strs21 = append(strs21, lsn21)

	tests := []struct {
		name string
		args args
		want []*v1alpha1.LocalStorageNode
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

func TestIsTaintMatch(t *testing.T) {
	tests := []struct {
		name          string
		nodeTaints    []corev1.Taint
		tolerations   []corev1.Toleration
		expectedMatch bool
	}{
		{
			name:          "empty taints and tolerations",
			nodeTaints:    []corev1.Taint{},
			tolerations:   []corev1.Toleration{},
			expectedMatch: true,
		},
		{
			name: "taints present, tolerations empty",
			nodeTaints: []corev1.Taint{
				{
					Key:    "key1",
					Value:  "value1",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
			tolerations:   []corev1.Toleration{},
			expectedMatch: false,
		},
		{
			name:       "taints empty, tolerations present",
			nodeTaints: []corev1.Taint{},
			tolerations: []corev1.Toleration{
				{
					Key:      "key1",
					Operator: corev1.TolerationOpEqual,
					Value:    "value1",
					Effect:   corev1.TaintEffectNoSchedule,
				},
			},
			expectedMatch: true,
		},
		{
			name: "taints and tolerations present, matching",
			nodeTaints: []corev1.Taint{
				{
					Key:    "key1",
					Value:  "value1",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
			tolerations: []corev1.Toleration{
				{
					Key:      "key1",
					Operator: corev1.TolerationOpEqual,
					Value:    "value1",
					Effect:   corev1.TaintEffectNoSchedule,
				},
			},
			expectedMatch: true,
		},
		{
			name: "taints and tolerations present, not matching",
			nodeTaints: []corev1.Taint{
				{
					Key:    "key1",
					Value:  "value1",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
			tolerations: []corev1.Toleration{
				{
					Key:      "key2",
					Operator: corev1.TolerationOpEqual,
					Value:    "value2",
					Effect:   corev1.TaintEffectNoSchedule,
				},
			},
			expectedMatch: false,
		},
		{
			name: "taints present, toleration with Exists operator",
			nodeTaints: []corev1.Taint{
				{
					Key:    "key1",
					Value:  "value1",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
			tolerations: []corev1.Toleration{
				{
					Key:      "key1",
					Operator: corev1.TolerationOpExists,
					Effect:   corev1.TaintEffectNoSchedule,
				},
			},
			expectedMatch: true,
		},
		{
			name: "taints present, toleration with Exists operator does not match value",
			nodeTaints: []corev1.Taint{
				{
					Key:    "key1",
					Value:  "value1",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
			tolerations: []corev1.Toleration{
				{
					Key:      "key2",
					Operator: corev1.TolerationOpExists,
					Effect:   corev1.TaintEffectNoSchedule,
				},
			},
			expectedMatch: false,
		},
		{
			name: "single node with PreferNoSchedule taint, pod without tolerations",
			nodeTaints: []corev1.Taint{
				{
					Key:    "key1",
					Value:  "value1",
					Effect: corev1.TaintEffectPreferNoSchedule,
				},
			},
			tolerations: []corev1.Toleration{},
			// If there is only one node, the Pod should be scheduled to this node even if it has no toleration.
			expectedMatch: true,
		},

		{
			name: "single node with PreferNoSchedule taint, pod with mismatched tolerations",
			nodeTaints: []corev1.Taint{
				{
					Key:    "key1",
					Value:  "value1",
					Effect: corev1.TaintEffectPreferNoSchedule,
				},
			},
			tolerations: []corev1.Toleration{
				{
					Key:      "key2", // Key Mismatch
					Operator: corev1.TolerationOpEqual,
					Value:    "value2",
					Effect:   corev1.TaintEffectPreferNoSchedule,
				},
			},
			expectedMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &corev1.Node{
				Spec: corev1.NodeSpec{
					Taints: tt.nodeTaints,
				},
			}

			match := canTaintBeTolerated(node, tt.tolerations)
			if match != tt.expectedMatch {
				t.Errorf("canTaintBeTolerated() = %v, want %v", match, tt.expectedMatch)
			}
		})
	}
}

func TestMatchNodeSelectorTerm(t *testing.T) {
	tests := []struct {
		name           string
		term           corev1.NodeSelectorTerm
		nodeLabels     map[string]string
		expectedResult bool
	}{
		{
			name:           "Test with empty nodeLabels and empty term",
			term:           corev1.NodeSelectorTerm{},
			nodeLabels:     nil,
			expectedResult: true,
		},
		{
			name: "Test with matching nodeLabels",
			term: corev1.NodeSelectorTerm{
				MatchExpressions: []corev1.NodeSelectorRequirement{
					{
						Key:      "disktype",
						Operator: corev1.NodeSelectorOpIn,
						Values:   []string{"ssd", "hdd"},
					},
				},
			},
			nodeLabels: map[string]string{
				"disktype": "ssd",
			},
			expectedResult: true,
		},
		{
			name: "Test with non-matching nodeLabels",
			term: corev1.NodeSelectorTerm{
				MatchExpressions: []corev1.NodeSelectorRequirement{
					{
						Key:      "disktype",
						Operator: corev1.NodeSelectorOpIn,
						Values:   []string{"ssd", "hdd"},
					},
				},
			},
			nodeLabels: map[string]string{
				"disktype": "sd",
			},
			expectedResult: false,
		},
		{
			name: "Test with matching MatchFields",
			term: corev1.NodeSelectorTerm{
				MatchFields: []corev1.NodeSelectorRequirement{
					{
						Key:      "metadata.annotations.kubernetes.io/role",
						Operator: corev1.NodeSelectorOpIn,
						Values:   []string{"master"},
					},
				},
			},
			nodeLabels: map[string]string{
				"metadata.annotations.kubernetes.io/role": "master",
			},
			expectedResult: true,
		},
		{
			name: "Test with NodeSelectorOpDoesNotExist on MatchExpressions",
			term: corev1.NodeSelectorTerm{
				MatchExpressions: []corev1.NodeSelectorRequirement{
					{
						Key:      "disktype",
						Operator: corev1.NodeSelectorOpDoesNotExist,
					},
				},
			},
			nodeLabels: map[string]string{
				"otherKey": "someValue",
			},
			expectedResult: true,
		},
		{
			name: "Test with NodeSelectorOpDoesNotExist failing on MatchFields",
			term: corev1.NodeSelectorTerm{
				MatchFields: []corev1.NodeSelectorRequirement{
					{
						Key:      "metadata.annotations.kubernetes.io/role",
						Operator: corev1.NodeSelectorOpDoesNotExist,
					},
				},
			},
			nodeLabels: map[string]string{
				"metadata.annotations.kubernetes.io/role": "worker",
			},
			expectedResult: false,
		},
		{
			name: "Test with multiple MatchExpressions, some matching, some not",
			term: corev1.NodeSelectorTerm{
				MatchExpressions: []corev1.NodeSelectorRequirement{
					{
						Key:      "disktype",
						Operator: corev1.NodeSelectorOpIn,
						Values:   []string{"ssd"},
					},
					{
						Key:      "region",
						Operator: corev1.NodeSelectorOpNotIn,
						Values:   []string{"us-east-1"},
					},
				},
			},
			nodeLabels: map[string]string{
				"disktype": "ssd",
				"region":   "us-west-2",
			},
			expectedResult: true,
		},
		{
			name: "Test with multiple MatchExpressions, all not matching",
			term: corev1.NodeSelectorTerm{
				MatchExpressions: []corev1.NodeSelectorRequirement{
					{
						Key:      "disktype",
						Operator: corev1.NodeSelectorOpIn,
						Values:   []string{"ssd"},
					},
					{
						Key:      "region",
						Operator: corev1.NodeSelectorOpNotIn,
						Values:   []string{"us-west-2"},
					},
				},
			},
			nodeLabels: map[string]string{
				"disktype": "hdd",
				"region":   "us-west-2",
			},
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchNodeSelectorTerm(tt.term, tt.nodeLabels)
			if result != tt.expectedResult {
				t.Errorf("matchNodeSelectorTerm(%v, %v) = %v, expected %v", tt.term, tt.nodeLabels, result, tt.expectedResult)
			}
		})
	}
}

func TestMatchAffinityByPods(t *testing.T) {
	// Test cases
	tests := []struct {
		name           string
		nodeName       string
		pods           *corev1.PodList
		expectedResult bool
	}{
		{
			name:           "Matching nodeName",
			nodeName:       "node1",
			pods:           &corev1.PodList{Items: []corev1.Pod{{Spec: corev1.PodSpec{NodeName: "node1"}}}},
			expectedResult: true,
		},
		{
			name:           "Non-matching nodeName",
			nodeName:       "node2",
			pods:           &corev1.PodList{Items: []corev1.Pod{{Spec: corev1.PodSpec{NodeName: "node1"}}}},
			expectedResult: false,
		},
		{
			name:           "Empty PodList",
			nodeName:       "node1",
			pods:           &corev1.PodList{},
			expectedResult: false,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call matchAffinityByPods function
			result := MatchAffinityByPods(tt.nodeName, tt.pods)

			// Assert the result
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestMatchAntiAffinityByPods(t *testing.T) {
	// Test case 1: nodeName does not match any pod's nodeName
	pods := &corev1.PodList{
		Items: []corev1.Pod{
			{
				Spec: corev1.PodSpec{
					NodeName: "node1",
				},
			},
			{
				Spec: corev1.PodSpec{
					NodeName: "node2",
				},
			},
		},
	}
	result := MatchAntiAffinityByPods("node3", pods)
	if !result {
		t.Errorf("TestMatchAntiAffinityByPods failed for case 1")
	}

	// Test case 2: nodeName matches one of the pod's nodeName
	pods = &corev1.PodList{
		Items: []corev1.Pod{
			{
				Spec: corev1.PodSpec{
					NodeName: "node1",
				},
			},
			{
				Spec: corev1.PodSpec{
					NodeName: "node2",
				},
			},
		},
	}
	result = MatchAntiAffinityByPods("node1", pods)
	if result {
		t.Errorf("TestMatchAntiAffinityByPods failed for case 2")
	}
}
