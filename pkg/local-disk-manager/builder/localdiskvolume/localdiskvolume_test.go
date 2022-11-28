package localdiskvolume

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
				volume: &v1alpha1.LocalDiskVolume{},
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

func TestNewBuilderFrom(t *testing.T) {
	volume := &v1alpha1.LocalDiskVolume{}

	tests := []struct {
		name string
		want *Builder
	}{
		{
			want: &Builder{
				volume: &v1alpha1.LocalDiskVolume{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBuilderFrom(volume); !reflect.DeepEqual(got, tt.want) {
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
				volume: &v1alpha1.LocalDiskVolume{
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

func TestWithFinalizer(t *testing.T) {
	finalizers := []string {
		"testFinalizer",
	}

	tests := []struct {
		name string
		want *Builder
	}{
		{
			want: &Builder{
				volume: &v1alpha1.LocalDiskVolume{
					ObjectMeta: v1.ObjectMeta{
						Finalizers: finalizers,
					},
				},
			},
		},
	}

	builder := NewBuilder()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := builder.WithFinalizer(finalizers); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithFinalizer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetupAccessibility(t *testing.T) {
	topology := v1alpha1.AccessibilityTopology{
		Nodes: []string{
			"testNode",
		},
	}

	tests := []struct {
		name string
		want *Builder
	}{
		{
			want: &Builder{
				volume: &v1alpha1.LocalDiskVolume{
					Spec: v1alpha1.LocalDiskVolumeSpec{
						Accessibility: topology,
					},
				},
			},
		},
	}

	builder := NewBuilder()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := builder.SetupAccessibility(topology); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetupAccessbility() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetupPVCNameSpaceName(t *testing.T) {
	pvc := "testPvc"

	tests := []struct {
		name string
		want *Builder
	}{
		{
			want: &Builder{
				volume: &v1alpha1.LocalDiskVolume{
					Spec: v1alpha1.LocalDiskVolumeSpec{
						PersistentVolumeClaimName: pvc,
					},
				},
			},
		},
	}

	builder := NewBuilder()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := builder.SetupPVCNameSpaceName(pvc); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetupPVCNameSpaceName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetupLocalDiskName(t *testing.T) {
	ld := "testLocalDisk"

	tests := []struct {
		name string
		want *Builder
	}{
		{
			want: &Builder{
				volume: &v1alpha1.LocalDiskVolume{
					Status: v1alpha1.LocalDiskVolumeStatus{
						LocalDiskName: ld,
					},
				},
			},
		},
	}

	builder := NewBuilder()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := builder.SetupLocalDiskName(ld); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetupLocalDiskName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetupDisk(t *testing.T) {
	devPath := "/dev/sdb"

	tests := []struct {
		name string
		want *Builder
	}{
		{
			want: &Builder{
				volume: &v1alpha1.LocalDiskVolume{
					Status: v1alpha1.LocalDiskVolumeStatus{
						DevPath: devPath,
					},
				},
			},
		},
	}

	builder := NewBuilder()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := builder.SetupDisk(devPath); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetupDisk() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetupAllocateCap(t *testing.T) {
	caps := int64(1000)

	tests := []struct {
		name string
		want *Builder
	}{
		{
			want: &Builder{
				volume: &v1alpha1.LocalDiskVolume{
					Status: v1alpha1.LocalDiskVolumeStatus{
						AllocatedCapacityBytes: caps,
					},
				},
			},
		},
	}

	builder := NewBuilder()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := builder.SetupAllocateCap(caps); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetupAllocateCap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetupRequiredCapacityBytes(t *testing.T) {
	caps := int64(1000)

	tests := []struct {
		name string
		want *Builder
	}{
		{
			want: &Builder{
				volume: &v1alpha1.LocalDiskVolume{
					Spec: v1alpha1.LocalDiskVolumeSpec{
						RequiredCapacityBytes: caps,
					},
				},
			},
		},
	}

	builder := NewBuilder()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := builder.SetupRequiredCapacityBytes(caps); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetupRequiredCapacityBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetupDiskType(t *testing.T) {
	diskType := "HDD"

	tests := []struct {
		name string
		want *Builder
	}{
		{
			want: &Builder{
				volume: &v1alpha1.LocalDiskVolume{
					Spec: v1alpha1.LocalDiskVolumeSpec{
						DiskType: diskType,
					},
				},
			},
		},
	}

	builder := NewBuilder()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := builder.SetupDiskType(diskType); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetupDiskType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetupStatus(t *testing.T) {
	tests := []struct {
		name string
		want *Builder
	}{
		{
			want: &Builder{
				volume: &v1alpha1.LocalDiskVolume{
					Status: v1alpha1.LocalDiskVolumeStatus{
						State: v1alpha1.VolumeReplicaStateReady,
					},
				},
			},
		},
	}

	builder := NewBuilder()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := builder.SetupStatus(v1alpha1.VolumeReplicaStateReady); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetupStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAssertVolumeNotNil(t *testing.T) {
	tests := []struct {
		name string
		builder Builder
		want error
	}{
		{
			builder: Builder{
				volume: &v1alpha1.LocalDiskVolume{},
			},
			want: nil,
		},
	}

	builder := NewBuilder()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := builder.assertVolumeNotNil(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("assertVolumeNotNil() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuild(t *testing.T) {
	tests := []struct {
		name string
		want *v1alpha1.LocalDiskVolume
	}{
		{
			want: &v1alpha1.LocalDiskVolume{},
		},
	}

	builder := NewBuilder()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := builder.Build(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Build() = %v, want %v", got, tt.want)
			}
		})
	}
}