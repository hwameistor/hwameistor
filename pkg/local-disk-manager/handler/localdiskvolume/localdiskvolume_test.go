package localdiskvolume

import (
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"

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

func newEmptyVolumeHandler() *DiskVolumeHandler {
	return &DiskVolumeHandler{}
}
