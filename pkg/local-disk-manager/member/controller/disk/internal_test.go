package disk

import (
	"fmt"
	"github.com/hwameistor/hwameistor/pkg/apis/client/clientset/versioned/fake"
	"github.com/hwameistor/hwameistor/pkg/apis/client/clientset/versioned/scheme"
	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/builder/localdisknode"
	localdisk2 "github.com/hwameistor/hwameistor/pkg/local-disk-manager/handler/localdisk"
	types2 "github.com/hwameistor/hwameistor/pkg/local-disk-manager/member/types"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fake2 "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"strings"
	"testing"
	"time"
)

var (
	fakeLocalDiskNodeName              = "local-disk-node-example"
	fakeNodename                       = "10-6-118-10"
	fakeFreeDiskCount                  = int64(1)
	fakeTotalDiskCount                 = int64(1)
	fakePoolClass                      = "HDD"
	fakePoolType                       = "REGULAR"
	fakeLocalDiskNodeUID               = "local-disk-node-uid"
	fakeLocalDiskClaimUID              = "local-disk-claim-uid"
	fakeLocalDiskUID                   = "local-disk-uid"
	fakeLocalDiskName                  = v1alpha1.LocalDiskObjectPrefix + utils.Hash("localdisk-example")
	fakeTotalCapacityBytes       int64 = 10 * 1024 * 1024 * 1024
	fakeUsedCapacityBytes        int64 = 0
	fakeVolumeCapacityBytesLimit int64 = 10 * 1024 * 1024 * 1024
	fakeFreeCapacityBytes        int64 = 10 * 1024 * 1024 * 1024
	fakeTotalVolumeCount         int64 = 1
	fakeUsedVolumeCount          int64 = 0
	fakeFreeVolumeCount          int64 = 1

	LocalDiskNodeKind   = "LocalDiskNode"
	LocalDiskKind       = "LocalDisk"
	LeaseKind           = "Lease"
	LocalDiskClaimKind  = "LocalDiskClaim"
	LocalDiskVolumeKind = "LocalDiskVolume"
	apiversion          = "hwameistor.io/v1alpha1"
	fakeRecorder        = record.NewFakeRecorder(100)
	fakeTimestamp       = time.Now()
	fakeKubeClient, _   = CreateFakeKubeClient()
	fakeDevPath         = "/dev/sda"
	fakeDisk            = v1alpha1.LocalDevice{
		DevPath:       fakeDevPath,
		Class:         types2.DevTypeHDD,
		CapacityBytes: fakeTotalCapacityBytes,
		State:         v1alpha1.DiskStateAvailable,
	}
	fakeLocalPool = v1alpha1.LocalPool{
		Name:                     "LocalDisk_PoolHDD-example",
		Class:                    v1alpha1.DiskClassNameHDD,
		Type:                     v1alpha1.PoolTypeRegular,
		TotalCapacityBytes:       fakeTotalCapacityBytes,
		UsedCapacityBytes:        fakeUsedCapacityBytes,
		VolumeCapacityBytesLimit: fakeVolumeCapacityBytesLimit,
		FreeCapacityBytes:        fakeFreeCapacityBytes,
		TotalVolumeCount:         fakeTotalVolumeCount,
		UsedVolumeCount:          fakeUsedVolumeCount,
		FreeVolumeCount:          fakeFreeVolumeCount,
		Disks:                    []v1alpha1.LocalDevice{fakeDisk},
		// ldvName list
		Volumes: []string{},
	}
	fakeLocalDiskClaim = GenFakeLocalDiskClaimObject()
)

func CreateFakeKubeClient() (*localdisknode.Kubeclient, error) {
	kubeclient := &localdisknode.Kubeclient{}
	clientset := fake.NewSimpleClientset()
	kubeclient.SetClient(clientset)
	return kubeclient, nil
}

func GenFakeClient() client.Client {
	ld := GenFakeLocalDiskObject()
	ldList := &v1alpha1.LocalDiskList{
		TypeMeta: v1.TypeMeta{
			Kind:       LocalDiskKind,
			APIVersion: apiversion,
		},
	}

	ldn := GenFakeLocalDiskNodeObject()
	ldnList := &v1alpha1.LocalDiskNodeList{
		TypeMeta: v1.TypeMeta{
			Kind:       LocalDiskNodeKind,
			APIVersion: apiversion,
		},
	}

	ldc := GenFakeLocalDiskClaimObject()
	ldcList := &v1alpha1.LocalDiskClaimList{
		TypeMeta: v1.TypeMeta{
			Kind:       LocalDiskClaimKind,
			APIVersion: apiversion,
		},
	}

	s := scheme.Scheme
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, ld)
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, ldList)
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, ldn)
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, ldnList)
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, ldc)
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, ldcList)

	return fake2.NewFakeClientWithScheme(s)

}

func GenFakeLocalDiskClaimObject() *v1alpha1.LocalDiskClaim {
	ldc := &v1alpha1.LocalDiskClaim{}

	typeMeta := v1.TypeMeta{
		Kind:       LocalDiskClaimKind,
		APIVersion: apiversion,
	}

	objectMeta := v1.ObjectMeta{
		Name:              strings.ToLower(fmt.Sprintf("%s-%s-claim-%s", fakeLocalDiskNodeName, types2.DevTypeHDD, "test-example")),
		ResourceVersion:   "",
		UID:               types.UID(fakeLocalDiskClaimUID),
		CreationTimestamp: v1.Time{fakeTimestamp},
	}

	spec := v1alpha1.LocalDiskClaimSpec{
		NodeName:    fakeLocalDiskNodeName,
		Description: v1alpha1.DiskClaimDescription{},
		DiskRefs:    []*corev1.ObjectReference{},
		Owner:       v1alpha1.LocalDiskManager,
	}

	status := v1alpha1.LocalDiskClaimStatus{
		Status: v1alpha1.DiskClaimStatusEmpty,
	}

	ldc.TypeMeta = typeMeta
	ldc.ObjectMeta = objectMeta
	ldc.Spec = spec
	ldc.Status = status

	return ldc
}
func GenFakeLocalDiskObject() *v1alpha1.LocalDisk {
	ld := &v1alpha1.LocalDisk{}

	typeMeta := v1.TypeMeta{
		Kind:       LocalDiskKind,
		APIVersion: apiversion,
	}

	objectMeta := v1.ObjectMeta{
		Name:              fakeLocalDiskName,
		ResourceVersion:   "",
		UID:               types.UID(fakeLocalDiskUID),
		CreationTimestamp: v1.Time{fakeTimestamp},
	}

	ld.TypeMeta = typeMeta
	ld.ObjectMeta = objectMeta

	return ld
}

func GenFakeLocalDiskNodeObject() *v1alpha1.LocalDiskNode {
	ldn := &v1alpha1.LocalDiskNode{}

	typeMeta := v1.TypeMeta{
		Kind:       LocalDiskNodeKind,
		APIVersion: apiversion,
	}

	objectMeta := v1.ObjectMeta{
		Name:              fakeLocalDiskNodeName,
		ResourceVersion:   "",
		UID:               types.UID(fakeLocalDiskNodeUID),
		CreationTimestamp: v1.Time{fakeTimestamp},
	}

	spec := v1alpha1.LocalDiskNodeSpec{
		NodeName: fakeLocalDiskNodeName,
	}

	status := v1alpha1.LocalDiskNodeStatus{
		TotalCapacity: fakeTotalCapacityBytes,
		FreeCapacity:  fakeFreeCapacityBytes,
		FreeDisk:      fakeFreeDiskCount,
		TotalDisk:     fakeTotalDiskCount,
		Pools:         map[string]v1alpha1.LocalPool{v1alpha1.PoolNameForHDD: fakeLocalPool},
		State:         "",
		PoolExtendRecords: map[string]v1alpha1.LocalDiskClaimSpecArray{
			v1alpha1.PoolNameForHDD: v1alpha1.LocalDiskClaimSpecArray{fakeLocalDiskClaim.Spec},
		},
	}

	ldn.Spec = spec
	ldn.TypeMeta = typeMeta
	ldn.ObjectMeta = objectMeta
	ldn.Status = status

	return ldn
}

func Test_GetNodeDisks(t *testing.T) {
	testcases := []struct {
		Description  string
		DiskNodeName string
		DiskNode     *v1alpha1.LocalDiskNode
		ExpectDisk   []types2.Disk
	}{
		// TODO: Add More test case
		{
			Description:  "It is a GetNodeDisks test.",
			DiskNodeName: fakeLocalDiskNodeName,
			DiskNode:     GenFakeLocalDiskNodeObject(),
			ExpectDisk: []types2.Disk{
				{
					AttachNode: fakeLocalDiskNodeName,
					Name:       fakeDevPath,
					DevPath:    fakeDevPath,
					Capacity:   fakeTotalCapacityBytes,
					DiskType:   types2.DevTypeHDD,
					Status:     types2.DiskStatus(v1alpha1.DiskStateAvailable),
				},
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.Description, func(t *testing.T) {
			client, err := CreateFakeKubeClient()
			if err != nil {
				t.Fatal("create a fake client failed")
			}
			_, err = client.Create(testcase.DiskNode)
			if err != nil {
				t.Fatal("create LocalDiskNode failed")
			}
			getFakeClient := func() (*localdisknode.Kubeclient, error) {
				return client, nil
			}
			ldnManager := localDiskNodesManager{}
			ldnManager.GetClient = getFakeClient
			fakeClient := GenFakeClient()
			ldnManager.DiskHandler = localdisk2.NewLocalDiskHandler(fakeClient, fakeRecorder)
			disks, err := ldnManager.GetNodeDisks(testcase.DiskNodeName)
			if !reflect.DeepEqual(disks, testcase.ExpectDisk) {
				t.Fatal("get a disks not same as expectDisk")
			}
		})
	}
}

func Test_GetNodeAvailableDisks(t *testing.T) {
	testcases := []struct {
		Description  string
		DiskNodeName string
		DiskNode     *v1alpha1.LocalDiskNode
		ExpectDisk   []types2.Disk
	}{
		// TODO: Add More test case
		{
			Description:  "It is a GetNodeAvailableDisks test.",
			DiskNodeName: fakeLocalDiskNodeName,
			DiskNode:     GenFakeLocalDiskNodeObject(),
			ExpectDisk: []types2.Disk{
				{
					AttachNode: fakeLocalDiskNodeName,
					Name:       fakeDevPath,
					DevPath:    fakeDevPath,
					Capacity:   fakeTotalCapacityBytes,
					DiskType:   types2.DevTypeHDD,
					Status:     types2.DiskStatus(v1alpha1.DiskStateAvailable),
				},
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.Description, func(t *testing.T) {
			client, err := CreateFakeKubeClient()
			if err != nil {
				t.Fatal("create a fake client failed")
			}
			_, err = client.Create(testcase.DiskNode)
			if err != nil {
				t.Fatal("create LocalDiskNode failed")
			}
			getFakeClient := func() (*localdisknode.Kubeclient, error) {
				return client, nil
			}
			ldnManager := localDiskNodesManager{}
			ldnManager.GetClient = getFakeClient
			fakeClient := GenFakeClient()
			ldnManager.DiskHandler = localdisk2.NewLocalDiskHandler(fakeClient, fakeRecorder)
			disks, err := ldnManager.GetNodeAvailableDisks(testcase.DiskNodeName)
			if !reflect.DeepEqual(disks, testcase.ExpectDisk) {
				t.Fatal("get a disks not same as expectDisk")
			}
		})
	}
}
