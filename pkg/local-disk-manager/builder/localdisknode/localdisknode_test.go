package localdisknode

import (
	"reflect"
	"testing"

	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewBuilder(t *testing.T) {
	tests := []struct {
		name string
		want *Builder
	}{
		{
			want: &Builder{
				DiskNode: &v1alpha1.LocalDiskNode{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBuilder(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBuilder() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithName(t *testing.T) {
	name := "testName"

	tests := []struct {
		name string
		want *Builder
	}{
		{
			want: &Builder{
				DiskNode: &v1alpha1.LocalDiskNode{
					ObjectMeta: v1.ObjectMeta{
						Name: name,
					},
				},
			},
		},
	}

	builder := NewBuilder()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := builder.WithName(name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetupAttachNode(t *testing.T) {
	attachNode := "testAttachNode"

	tests := []struct {
		name string
		want *Builder
	}{
		{
			want: &Builder{
				DiskNode: &v1alpha1.LocalDiskNode{
					Spec: v1alpha1.LocalDiskNodeSpec{
						AttachNode: attachNode,
					},
				},
			},
		},
	}

	builder := NewBuilder()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := builder.SetupAttachNode(attachNode); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetupAttachNode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuild(t *testing.T) {
	tests := []struct {
		name string
		builder *Builder
		want *v1alpha1.LocalDiskNode
	}{
		{
			builder: &Builder{
				DiskNode: &v1alpha1.LocalDiskNode{},
			},
			want: &v1alpha1.LocalDiskNode{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := tt.builder.Build(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Build() = %v, want %v", got, tt.want)
			}
		})
	}
}