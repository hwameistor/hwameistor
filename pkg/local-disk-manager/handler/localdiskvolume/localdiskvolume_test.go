package localdiskvolume

import (
	"context"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/hwameistor/hwameistor/pkg/apis/client/clientset/versioned/scheme"
	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

func TestLocalDiskVolumeHandler_AppendMountPoint(t *testing.T) {
	v := newEmptyVolumeHandler()
	v.Ldv = &v1alpha1.LocalDiskVolume{}

	mountPointCases := []struct {
		Description string
		MountPath   string
		VolumeCap   *csi.VolumeCapability
		WantExist   bool
	}{
		{
			Description: "Should return success, mount block",
			MountPath:   "a/b/c",
			VolumeCap: &csi.VolumeCapability{
				AccessType: &csi.VolumeCapability_Block{},
			},
			WantExist: true,
		},
		{
			Description: "Should return success, mount filesystem",
			MountPath:   "a/b/c/d",
			VolumeCap: &csi.VolumeCapability{
				AccessType: &csi.VolumeCapability_Mount{
					Mount: &csi.VolumeCapability_MountVolume{
						FsType: "xfs",
					}},
			},
			WantExist: true,
		},
	}

	for _, testcase := range mountPointCases {
		t.Run(testcase.Description, func(t *testing.T) {
			v.AppendMountPoint(testcase.MountPath, testcase.VolumeCap)
			if v.ExistMountPoint(testcase.MountPath) != testcase.WantExist {
				t.Fatalf("MountPoints %s Append fail, want %v actual %v",
					testcase.MountPath, testcase.WantExist, !testcase.WantExist)
			}
		})
	}
}

func TestLocalDiskVolumeHandler_RemoveMountPoint(t *testing.T) {
	v := newEmptyVolumeHandler()
	v.Ldv = &v1alpha1.LocalDiskVolume{}

	umountPointCases := []struct {
		Description string
		MountPath   string
		VolumeCap   *csi.VolumeCapability
		WantExist   bool
	}{
		{
			Description: "Should return success, umount block",
			MountPath:   "a/b/c",
			VolumeCap: &csi.VolumeCapability{
				AccessType: &csi.VolumeCapability_Block{},
			},
			WantExist: false,
		},
		{
			Description: "Should return success, umount filesystem",
			MountPath:   "a/b/c/d",
			VolumeCap: &csi.VolumeCapability{
				AccessType: &csi.VolumeCapability_Mount{
					Mount: &csi.VolumeCapability_MountVolume{
						FsType: "xfs",
					}},
			},
			WantExist: false,
		},
	}
	for _, testcase := range umountPointCases {
		v.AppendMountPoint(testcase.MountPath, testcase.VolumeCap)
	}

	for _, testcase := range umountPointCases {
		t.Run(testcase.Description, func(t *testing.T) {
			v.RemoveMountPoint(testcase.MountPath)
			if v.ExistMountPoint(testcase.MountPath) != testcase.WantExist {
				t.Fatalf("UnMountPoints %s Append fail, want %v actual %v",
					testcase.MountPath, testcase.WantExist, !testcase.WantExist)
			}
		})
	}
}

func TestLocalDiskVolumeHandler_GetLocalDiskVolume(t *testing.T) {
	v := newEmptyVolumeHandler()
	c, _ := CreateFakeClient()
	v.Client = c
	ldv := &v1alpha1.LocalDiskVolume{
		ObjectMeta: v1.ObjectMeta{
			Namespace: "testns",
			Name: "testldv",
		},
	}
	v.Ldv = ldv

	if err := v.Create(context.TODO(), ldv); err != nil {
		t.Error(err)
	}

	key := types.NamespacedName{
		Namespace: "testns",
		Name: "testldv",
	}
	gotten, err := v.GetLocalDiskVolume(key)
	if err != nil {
		t.Error(err)
	}
	if (gotten.Name != v.Ldv.Name) && (gotten.Namespace != v.Ldv.Namespace) {
		t.Fail()
	}
}

// CreateFakeClient Create localDisk and LocalDiskNode resource
func CreateFakeClient() (client.Client, *runtime.Scheme) {
	s := scheme.Scheme
	s.AddKnownTypes(v1.SchemeGroupVersion, &v1alpha1.LocalVolumeGroup{})
	return fake.NewClientBuilder().WithScheme(s).WithObjects(&v1alpha1.LocalDisk{}, &v1alpha1.LocalDiskNode{}).Build(), s
}

func newEmptyVolumeHandler() *DiskVolumeHandler {
	return &DiskVolumeHandler{}
}
