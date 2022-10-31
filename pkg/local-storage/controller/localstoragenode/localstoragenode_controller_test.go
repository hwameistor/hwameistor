package localstoragenode

import (
	"context"
	"fmt"
	"testing"
	"time"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-storage/member"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var (
	fakeLocalStorageNodeName       = "local-storage-node-example"
	fakeLocalStorageNodeUID        = "local-storage-node-uid"
	fakeNamespace                  = "local-storage-node-test"
	fakeNodename                   = "10-6-118-10"
	fakeStorageIp                  = "10.6.118.11"
	fakeZone                       = "zone-test"
	fakeRegion                     = "region-test"
	fakeVgType                     = "LocalStorage_PoolHDD"
	fakeVgName                     = "vg-test"
	fakePoolClass                  = "HDD"
	fakePoolType                   = "REGULAR"
	fakeTotalCapacityBytes   int64 = 10 * 1024 * 1024 * 1024
	fakeFreeCapacityBytes    int64 = 8 * 1024 * 1024 * 1024
	fakeDiskCapacityBytes    int64 = 2 * 1024 * 1024 * 1024

	apiversion           = "hwameistor.io/v1alpha1"
	LocalStorageNodeKind = "LocalStorageNode"

	//fakeRecorder = record.NewFakeRecorder(100)
)

func TestNewLocalStorageNodeController(t *testing.T) {

	cli, s := CreateFakeClient()
	// Create a Reconcile for LocalStorageNode
	r := ReconcileLocalStorageNode{
		client:        cli,
		scheme:        s,
		storageMember: member.Member().ConfigureController(s),
	}

	// Create LocalStorageNode
	lsn := GenFakeLocalStorageNodeObject()
	err := r.client.Create(context.Background(), lsn)
	if err != nil {
		t.Errorf("Create LocalStorageNode fail %v", err)
	}
	defer r.DeleteFakeLocalStorageNode(t, lsn)

	// Get lsn
	err = r.client.Get(context.Background(), types.NamespacedName{Namespace: lsn.GetNamespace(), Name: lsn.GetName()}, lsn)
	if err != nil {
		t.Errorf("Get lsn fail %v", err)
	}
	fmt.Printf("lsn = %+v", lsn)
	fmt.Printf("r.storageMember = %+v", r.storageMember)

	// Mock LocalStorageNode request
	req := reconcile.Request{NamespacedName: types.NamespacedName{Namespace: lsn.GetNamespace(), Name: lsn.GetName()}}
	_, err = r.Reconcile(context.TODO(), req)
	if err != nil {
		t.Errorf("Reconcile fail %v", err)
	}

}

// DeleteFakeLocalStorageNode
func (r *ReconcileLocalStorageNode) DeleteFakeLocalStorageNode(t *testing.T, lsn *apisv1alpha1.LocalStorageNode) {
	if err := r.client.Delete(context.Background(), lsn); err != nil {
		t.Errorf("Delete LocalStorageNode %v fail %v", lsn.GetName(), err)
	}
}

// GenFakeLocalStorageNodeObject Create lsn request
func GenFakeLocalStorageNodeObject() *apisv1alpha1.LocalStorageNode {
	lsn := &apisv1alpha1.LocalStorageNode{}

	TypeMeta := metav1.TypeMeta{
		Kind:       LocalStorageNodeKind,
		APIVersion: apiversion,
	}

	ObjectMata := metav1.ObjectMeta{
		Name:              fakeLocalStorageNodeName,
		Namespace:         fakeNamespace,
		ResourceVersion:   "",
		UID:               types.UID(fakeLocalStorageNodeUID),
		CreationTimestamp: metav1.Time{Time: time.Now()},
	}

	Spec := apisv1alpha1.LocalStorageNodeSpec{
		HostName:  fakeNodename,
		StorageIP: fakeStorageIp,
		Topo: apisv1alpha1.Topology{
			Zone:   fakeZone,
			Region: fakeRegion,
		},
	}

	disks := make([]apisv1alpha1.LocalDevice, 0, 10)
	var localdisk1 apisv1alpha1.LocalDevice
	localdisk1.DevPath = "/dev/sdf"
	localdisk1.State = apisv1alpha1.DiskStateAvailable
	localdisk1.Class = fakePoolClass
	localdisk1.CapacityBytes = fakeDiskCapacityBytes
	disks = append(disks, localdisk1)

	volumes := make([]string, 0, 5)
	volumes = append(volumes, "volume-test1")

	pools := make(map[string]apisv1alpha1.LocalPool)
	pools[fakeVgType] = apisv1alpha1.LocalPool{
		Name:                     fakeVgName,
		Class:                    fakePoolClass,
		Type:                     fakePoolType,
		TotalCapacityBytes:       int64(fakeTotalCapacityBytes),
		UsedCapacityBytes:        int64(fakeTotalCapacityBytes) - int64(fakeFreeCapacityBytes),
		FreeCapacityBytes:        int64(fakeFreeCapacityBytes),
		VolumeCapacityBytesLimit: int64(fakeTotalCapacityBytes),
		TotalVolumeCount:         apisv1alpha1.LVMVolumeMaxCount,
		UsedVolumeCount:          int64(len(volumes)),
		FreeVolumeCount:          apisv1alpha1.LVMVolumeMaxCount - int64(len(volumes)),
		Disks:                    disks,
		Volumes:                  volumes,
	}

	lsn.ObjectMeta = ObjectMata
	lsn.TypeMeta = TypeMeta
	lsn.Spec = Spec
	lsn.Status.State = apisv1alpha1.NodeStateReady
	lsn.Status.Pools = pools
	return lsn
}

// CreateFakeClient Create LocalStorageNode resource
func CreateFakeClient() (client.Client, *runtime.Scheme) {
	lsn := GenFakeLocalStorageNodeObject()
	lsnList := &apisv1alpha1.LocalStorageNodeList{
		TypeMeta: metav1.TypeMeta{
			Kind:       LocalStorageNodeKind,
			APIVersion: apiversion,
		},
	}

	s := scheme.Scheme
	s.AddKnownTypes(apisv1alpha1.SchemeGroupVersion, lsn)
	s.AddKnownTypes(apisv1alpha1.SchemeGroupVersion, lsnList)
	return fake.NewFakeClientWithScheme(s), s
}
