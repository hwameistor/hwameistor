package localdiskvolume

import (
	"github.com/hwameistor/hwameistor/pkg/apis/client/clientset/versioned/fake"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"reflect"
	"testing"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
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
	finalizers := []string{
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
		name    string
		builder Builder
		want    error
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

var (
	fakeLocalDiskVolumeName          = "local-disk-volume-example"
	fakeNodename                     = "10-6-118-10"
	fakeLocalDiskName                = "localdisk-example"
	fakeTotalDiskCount               = int64(1)
	fakeDiskType                     = "HDD"
	fakePoolClass                    = "HDD"
	fakePoolType                     = "REGULAR"
	fakeLocalDiskVolumeUID           = "local-disk-volume-uid"
	fakeStorageClassName             = "sc-test"
	fakeRequiredCapacityBytes  int64 = 10 * 1024 * 1024 * 1024
	fakeAllocatedCapacityBytes int64 = 10 * 1024 * 1024 * 1024
	fakeCanWipe                      = true
	LocalDiskNodeKind                = "LocalDiskNode"
	LocalDiskKind                    = "LocalDisk"
	LeaseKind                        = "Lease"
	LocalDiskClaimKind               = "LocalDiskClaim"
	LocalDiskVolumeKind              = "LocalDiskVolume"
	fakeNamespace                    = "local-disk-volume-test"
	fakePersistentPvcName            = "pvc-test"
	apiversion                       = "hwameistor.io/v1alpha1"
	fakeRecorder                     = record.NewFakeRecorder(100)
	fakeTimestamp                    = time.Now()
	fakeAccessibility                = v1alpha1.AccessibilityTopology{Nodes: []string{fakeNodename}}
	fakeDevlinks                     = map[v1alpha1.DevLinkType][]string{}
	fakeMountPoint                   = []v1alpha1.MountPoint{
		{
			TargetPath: "/data",
		},
	}
	fakeVolumePath = "/etc/hwameistor/LocalDisk_PoolHDD/volume/" + fakeLocalDiskVolumeName
	fakeDevPath    = "/dev/sda"
)

func init() {
	fakeDevlinks[v1alpha1.LinkByID] = []string{"ata-test", "wwn-test"}
	fakeDevlinks[v1alpha1.LinkByPath] = []string{"pci-test"}
}
func CreateFakeKubeClient() (*Kubeclient, error) {
	kubeclient := Kubeclient{}
	clientset := fake.NewSimpleClientset()
	kubeclient.SetClient(clientset)
	return &kubeclient, nil
}
func GenFakePVCObject() *corev1.PersistentVolumeClaim {
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: v1.ObjectMeta{
			Name:      fakePersistentPvcName,
			Namespace: fakeNamespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			StorageClassName: &fakeStorageClassName,
		},
	}
	return pvc
}

func GenFakeLocalDiskVolumeObject() *v1alpha1.LocalDiskVolume {
	ldv := &v1alpha1.LocalDiskVolume{}

	typeMeta := v1.TypeMeta{
		Kind:       LocalDiskVolumeKind,
		APIVersion: apiversion,
	}

	objectMeta := v1.ObjectMeta{
		Name:              fakeLocalDiskVolumeName,
		ResourceVersion:   "",
		UID:               types.UID(fakeLocalDiskVolumeUID),
		CreationTimestamp: v1.Time{fakeTimestamp},
	}

	spec := v1alpha1.LocalDiskVolumeSpec{
		Accessibility:         fakeAccessibility,
		CanWipe:               fakeCanWipe,
		DiskType:              fakeDiskType,
		RequiredCapacityBytes: fakeRequiredCapacityBytes,
	}

	status := v1alpha1.LocalDiskVolumeStatus{
		LocalDiskName:          fakeLocalDiskName,
		DevLinks:               fakeDevlinks,
		MountPoints:            fakeMountPoint,
		VolumePath:             fakeVolumePath,
		DevPath:                fakeDevPath,
		AllocatedCapacityBytes: fakeAllocatedCapacityBytes,
		UsedCapacityBytes:      fakeRequiredCapacityBytes,
	}

	ldv.Spec = spec
	ldv.TypeMeta = typeMeta
	ldv.ObjectMeta = objectMeta
	ldv.Status = status

	return ldv
}

func Test_KubeClient_Create(t *testing.T) {
	testcases := []struct {
		Description      string
		DiskVolume       *v1alpha1.LocalDiskVolume
		ExpectDiskVolume *v1alpha1.LocalDiskVolume
	}{
		// TODO: Add More test case
		{
			Description:      "It is a create localDiskVolume test.",
			DiskVolume:       GenFakeLocalDiskVolumeObject(),
			ExpectDiskVolume: GenFakeLocalDiskVolumeObject(),
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.Description, func(t *testing.T) {
			client, err := CreateFakeKubeClient()
			if err != nil {
				t.Fatal("create a fake client failed")
			}
			_, err = client.Create(testcase.DiskVolume)
			if err != nil {
				t.Fatal("create a localDiskVolume on fake client failed")
			}
			ldn, err := client.Get(testcase.DiskVolume.Name)
			if err != nil {
				t.Fatal("get a specified localDiskVolume on fake client failed")
			}
			if !reflect.DeepEqual(ldn, testcase.ExpectDiskVolume) {
				t.Fatal("get a localDiskVolume not same as expectDiskVolume")
			}
		})
	}
}

func Test_KubeClient_Update(t *testing.T) {
	type args struct {
		CanWipe bool
	}
	testcases := []struct {
		Description   string
		DiskVolume    *v1alpha1.LocalDiskVolume
		modifyArgs    args
		ExpectCanWipe bool
	}{
		// TODO: Add More test case
		{
			Description: "It is an update localDiskVolume test.",
			DiskVolume:  GenFakeLocalDiskVolumeObject(),
			modifyArgs: args{
				CanWipe: false,
			},
			ExpectCanWipe: false,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.Description, func(t *testing.T) {
			client, err := CreateFakeKubeClient()
			if err != nil {
				t.Fatal("create a fake client failed")
			}
			// ensure the source ldn exists in storage
			_, err = client.Create(testcase.DiskVolume)
			if err != nil {
				t.Fatal("create a localDiskVolume on fake client failed")
			}
			ldv := testcase.DiskVolume.DeepCopy()
			ldv.Spec.CanWipe = testcase.modifyArgs.CanWipe
			_, err = client.Update(ldv)
			if err != nil {
				// the origin object must exist in cluster
				t.Fatal("update a localDiskVolume on fake client failed")
			}
			newLdv, err := client.Get(testcase.DiskVolume.Name)
			if err != nil {
				t.Fatal("get a specified localDiskVolume on fake client failed")
			}
			if !reflect.DeepEqual(testcase.ExpectCanWipe, newLdv.Spec.CanWipe) {
				t.Fatal("get a localDiskVolume not same as expectDiskVolume")
			}
		})
	}
}
