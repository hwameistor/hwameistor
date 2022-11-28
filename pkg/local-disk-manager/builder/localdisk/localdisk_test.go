package localdisk

import (
	"reflect"
	"testing"

	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/disk/manager"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewBuilder(t *testing.T) {
	tests := []struct{
		name string
		want *Builder
	}{
		{
			want: &Builder{
				disk: &v1alpha1.LocalDisk{},
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
	type args struct {
		diskName string
	}
	tests := []struct{
		name string
		args args
		want *Builder
	}{
		{
			want: &Builder{
				disk: &v1alpha1.LocalDisk{
					ObjectMeta: v1.ObjectMeta{
						Name: "testdiskname",
					},
				},
			},
			args: args{
				diskName: "testdiskname",
			},
		},
	}

	builder := NewBuilder()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := builder.WithName(tt.args.diskName); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetupAttribute(t *testing.T) {
	capacity := int64(1000)
	devPath := "/dev/sdb"
	diskType := "HDD"
	vendor := "testVendor"
	modelName := "testModelName"
	protocol := "scsi"
	serialNumber := "testSerialNumber"
	devType := "disk"

	tests := []struct {
		name string
		want *Builder
	}{
		{
			want: &Builder{
				disk: &v1alpha1.LocalDisk{
					Spec: v1alpha1.LocalDiskSpec{
						Capacity: capacity,
						DevicePath: devPath,
						DiskAttributes: v1alpha1.DiskAttributes{
							Type: diskType,
							Vendor: vendor,
							ModelName: modelName,
							Protocol: protocol,
							SerialNumber: serialNumber,
							DevType: devType,
						},
					},
				},
			},
		},
	}

	builder := NewBuilder()
	attr := manager.Attribute{
		Capacity: capacity,
		DevName: devPath,
		DriverType: diskType,
		Vendor: vendor,
		Model: modelName,
		Bus: protocol,
		Serial: serialNumber,
		DevType: devType,
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := builder.SetupAttribute(attr); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetupAttribute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetupState(t *testing.T) {
	tests := []struct {
		name string
		want *Builder
	}{
		{
			want: &Builder{
				disk: &v1alpha1.LocalDisk{
					Spec: v1alpha1.LocalDiskSpec{
						State: v1alpha1.LocalDiskActive,
					},
				},
			},
		},
	}

	builder := NewBuilder()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := builder.SetupState(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetupState() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetupRaidInfo(t *testing.T) {
	tests := []struct {
		name string
		want *Builder
	}{
		{
			want: &Builder{
				disk: &v1alpha1.LocalDisk{},
			},
		},
	}

	builder := NewBuilder()
	raidInfo := manager.RaidInfo{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := builder.SetupRaidInfo(raidInfo); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetupRaidInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetupUUID(t *testing.T) {
	uuid := "testuuid"

	tests := []struct {
		name string
		want *Builder
	}{
		{
			want: &Builder{
				disk: &v1alpha1.LocalDisk{
					Spec: v1alpha1.LocalDiskSpec{
						UUID: uuid,
					},
				},
			},
		},
	}

	builder := NewBuilder()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := builder.SetupUUID(uuid); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetupUUID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetupNodeName(t *testing.T) {
	nodeName := "testNodeName"

	tests := []struct {
		name string
		want *Builder
	}{
		{
			want: &Builder{
				disk: &v1alpha1.LocalDisk{
					Spec: v1alpha1.LocalDiskSpec{
						NodeName: nodeName,
					},
				},
			},
		},
	}

	builder := NewBuilder()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := builder.SetupNodeName(nodeName); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetupNodeName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetupPartitionInfo(t *testing.T) {
	partitionInfos := []manager.PartitionInfo{
		{
			Filesystem: "xfs",
		},
	}

	tests := []struct {
		name string
		want *Builder
	}{
		{
			want: &Builder{
				disk: &v1alpha1.LocalDisk{
					Spec: v1alpha1.LocalDiskSpec{
						HasPartition: true,
						PartitionInfo: []v1alpha1.PartitionInfo{
							{
								HasFileSystem: true,
								FileSystem: v1alpha1.FileSystemInfo{
									Type: "xfs",
								},
							},
						},
					},
				},
			},
		},
	}

	builder := NewBuilder()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := builder.SetupPartitionInfo(partitionInfos); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetupPartitionInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateStatus(t *testing.T) {
	tests := []struct {
		name string
		builder *Builder
		want v1alpha1.LocalDiskStatus
	}{
		{
			builder: &Builder{
				disk: &v1alpha1.LocalDisk{
					Spec: v1alpha1.LocalDiskSpec{
						HasPartition: true,
					},
				},
			},
			want: v1alpha1.LocalDiskStatus{
				State: v1alpha1.LocalDiskBound,
			},
		},
		{
			builder: &Builder{
				disk: &v1alpha1.LocalDisk{
					Spec: v1alpha1.LocalDiskSpec{
						HasPartition: false,
					},
				},
			},
			want: v1alpha1.LocalDiskStatus{
				State: v1alpha1.LocalDiskAvailable,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.builder.GenerateStatus().disk.Status; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuild(t *testing.T) {
	tests := []struct {
		name string
		want v1alpha1.LocalDisk
	}{
		{
			want: v1alpha1.LocalDisk{},
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