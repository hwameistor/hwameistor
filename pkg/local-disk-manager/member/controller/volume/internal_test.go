package volume

import (
	"github.com/hwameistor/hwameistor/pkg/apis/client/clientset/versioned/fake"
	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/builder/localdiskvolume"
	types2 "github.com/hwameistor/hwameistor/pkg/local-disk-manager/member/types"
	"reflect"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"testing"
	"time"
)

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

func CreateFakeKubeClient() (*localdiskvolume.Kubeclient, error) {
	kubeclient := &localdiskvolume.Kubeclient{}
	clientset := fake.NewSimpleClientset()
	kubeclient.SetClient(clientset)
	return kubeclient, nil
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

func Test_GetVolumeInfo(t *testing.T) {
	testcases := []struct {
		Description    string
		DiskVolumeName string
		DiskVolume     *v1alpha1.LocalDiskVolume
		ExpectVolume   *types2.Volume
	}{
		// TODO: Add More test case
		{
			Description:    "It is a GetVolumeInfo test.",
			DiskVolumeName: fakeLocalDiskVolumeName,
			DiskVolume:     GenFakeLocalDiskVolumeObject(),
			ExpectVolume: &types2.Volume{
				Name:       fakeLocalDiskVolumeName,
				Exist:      true,
				Capacity:   fakeAllocatedCapacityBytes,
				AttachNode: fakeAccessibility.Nodes[0],
			},
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
				t.Fatal("create LocalDiskNode failed")
			}
			getFakeClient := func() (*localdiskvolume.Kubeclient, error) {
				return client, nil
			}
			ldvManager := localDiskVolumeManager{}
			ldvManager.GetClient = getFakeClient
			volumeInfo, err := ldvManager.GetVolumeInfo(testcase.DiskVolumeName)
			if err != nil {
				t.Fatal("get VolumeInfo failed")
			}
			if !reflect.DeepEqual(volumeInfo, testcase.ExpectVolume) {
				t.Log(volumeInfo)
				t.Log(testcase.ExpectVolume)
				t.Fatal("get GetVolumeInfo failed")
			}
		})
	}
}
