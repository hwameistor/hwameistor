package node

import (
	"context"
	"fmt"
	"github.com/hwameistor/hwameistor/pkg/apis/client/clientset/versioned/fake"
	"github.com/hwameistor/hwameistor/pkg/apis/client/clientset/versioned/scheme"
	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/builder/localdisknode"
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
	fakeNamespace                      = "hwameistor"
	LocalDiskNodeKind                  = "LocalDiskNode"
	LocalDiskKind                      = "LocalDisk"
	LeaseKind                          = "Lease"
	LocalDiskClaimKind                 = "LocalDiskClaim"
	LocalDiskVolumeKind                = "LocalDiskVolume"
	apiversion                         = "hwameistor.io/v1alpha1"
	fakeRecorder                       = record.NewFakeRecorder(100)
	fakeTimestamp                      = time.Now()
	fakeKubeClient, _                  = CreateFakeKubeClient()
	fakeDevPath                        = "/dev/sda"
	fakeDisk                           = v1alpha1.LocalDevice{
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
		// ldvName list, empty slice is not allowed
		Volumes: nil,
	}
	fakeLocalDiskClaim     = GenFakeLocalDiskClaimObject()
	fakeLocalDiskClaimSpec = GenFakeLocalDiskClaimSpecObject()
	fakeLocalDiskNode      = GenFakeLocalDiskNodeObject()
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
		DiskRefs:    nil,
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
func GenFakeLocalDiskClaimSpecObject() v1alpha1.LocalDiskClaimSpec {
	spec := v1alpha1.LocalDiskClaimSpec{
		NodeName:    fakeLocalDiskNodeName,
		Description: v1alpha1.DiskClaimDescription{},
		DiskRefs: []*corev1.ObjectReference{
			&corev1.ObjectReference{
				APIVersion: apiversion,
				Kind:       LocalDiskKind,
				Namespace:  fakeNamespace,
				Name:       fakeLocalDiskName,
			},
		},
		Owner: v1alpha1.LocalDiskManager,
	}
	return spec
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
func GenFakeNodeManager() *nodeManager {
	m := nodeManager{}
	m.nodeName = fakeLocalDiskNodeName
	m.k8sClient = GenFakeClient()
	return &m
}
func Test_UpdatePoolExtendRecord(t *testing.T) {
	testcases := []struct {
		Description          string
		LocalPoolName        []string
		LocalDiskClaimRecord v1alpha1.LocalDiskClaimSpec
		DiskNode             *v1alpha1.LocalDiskNode
		ExpectDiskNodeStatus v1alpha1.LocalDiskNodeStatus
	}{
		// TODO: Add More test case
		{
			Description:          "It is an UpdatePoolExtendRecord test.",
			LocalPoolName:        []string{v1alpha1.PoolNameForHDD},
			LocalDiskClaimRecord: fakeLocalDiskClaimSpec,
			DiskNode:             fakeLocalDiskNode,
			ExpectDiskNodeStatus: v1alpha1.LocalDiskNodeStatus{
				TotalCapacity: fakeTotalCapacityBytes,
				FreeCapacity:  fakeFreeCapacityBytes,
				FreeDisk:      fakeFreeDiskCount,
				TotalDisk:     fakeTotalDiskCount,
				Pools:         map[string]v1alpha1.LocalPool{v1alpha1.PoolNameForHDD: fakeLocalPool},
				State:         "",
				PoolExtendRecords: map[string]v1alpha1.LocalDiskClaimSpecArray{
					v1alpha1.PoolNameForHDD: v1alpha1.LocalDiskClaimSpecArray{fakeLocalDiskClaim.Spec, fakeLocalDiskClaimSpec},
				},
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.Description, func(t *testing.T) {
			m := GenFakeNodeManager()
			err := m.k8sClient.Create(context.Background(), testcase.DiskNode)
			if err != nil {
				t.Fatal("failed to create a localDiskNode")
			}
			m.updatePoolExtendRecord(testcase.LocalPoolName, testcase.LocalDiskClaimRecord)
			ldn := v1alpha1.LocalDiskNode{}
			err = m.k8sClient.Get(context.Background(), client.ObjectKey{Name: fakeLocalDiskNode.Name}, &ldn)
			if err != nil {
				t.Fatal("failed to get a specified localDiskNode")
			}
			if !reflect.DeepEqual(ldn.Status, testcase.ExpectDiskNodeStatus) {
				t.Log(ldn.Status)
				t.Log(testcase.ExpectDiskNodeStatus)
				t.Fatal("get a localDiskNodeStatus not same as expectDiskNodeStatus")
			}
		})
	}

}
