package localvolumemigrate

import (
	"context"
	"fmt"
	"testing"
	"time"

	ldmv1alpha1 "github.com/hwameistor/local-disk-manager/pkg/apis/hwameistor/v1alpha1"
	apisv1alpha1 "github.com/hwameistor/local-storage/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/local-storage/pkg/member"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var (
	fakeLocalVolumeName              = "local-volume-example"
	fakeLocalVolumeMigrateName       = "local-volume-convert-example"
	fakeLocalVolumeMigrateUID        = "local-volume-convert-uid"
	fakeNamespace                    = "local-volume-test"
	fakeNodename                     = "10-6-118-10"
	fakeStorageIp                    = "10.6.118.11"
	fakeZone                         = "zone-test"
	fakeRegion                       = "region-test"
	fakeVgType                       = "LocalStorage_PoolHDD"
	fakeVgName                       = "vg-test"
	fakePoolClass                    = "HDD"
	fakePoolType                     = "REGULAR"
	fakeTotalCapacityBytes     int64 = 10 * 1024 * 1024 * 1024
	fakeFreeCapacityBytes      int64 = 8 * 1024 * 1024 * 1024
	fakeDiskCapacityBytes      int64 = 2 * 1024 * 1024 * 1024

	apiversion             = "hwameistor.io/v1alpha1"
	LocalVolumeMigrateKind = "LocalVolumeMigrate"
	fakeRecorder           = record.NewFakeRecorder(100)
)

func TestNewLocalVolumeMigrateController(t *testing.T) {

	cli, s := CreateFakeClient()
	// Create a Reconcile for LocalVolumeMigrate
	r := ReconcileLocalVolumeMigrate{
		client:        cli,
		scheme:        s,
		storageMember: member.Member().ConfigureController(s),
	}

	// Create LocalVolumeMigrate
	lvm := GenFakeLocalVolumeMigrateObject()
	err := r.client.Create(context.Background(), lvm)
	if err != nil {
		t.Errorf("Create LocalVolumeMigrate fail %v", err)
	}
	defer r.DeleteFakeLocalVolumeMigrate(t, lvm)

	// Get lvm
	err = r.client.Get(context.Background(), types.NamespacedName{Namespace: lvm.GetNamespace(), Name: lvm.GetName()}, lvm)
	if err != nil {
		t.Errorf("Get lvm fail %v", err)
	}
	fmt.Printf("lvm = %+v", lvm)
	fmt.Printf("r.storageMember = %+v", r.storageMember)

	// Mock LocalVolumeMigrate request
	req := reconcile.Request{NamespacedName: types.NamespacedName{Namespace: lvm.GetNamespace(), Name: lvm.GetName()}}
	_, err = r.Reconcile(req)
	if err != nil {
		t.Errorf("Reconcile fail %v", err)
	}

}

// DeleteFakeLocalVolumeMigrate
func (r *ReconcileLocalVolumeMigrate) DeleteFakeLocalVolumeMigrate(t *testing.T, lvm *apisv1alpha1.LocalVolumeMigrate) {
	if err := r.client.Delete(context.Background(), lvm); err != nil {
		t.Errorf("Delete LocalVolumeMigrate %v fail %v", lvm.GetName(), err)
	}
}

// GenFakeLocalVolumeMigrateObject Create lvm request
func GenFakeLocalVolumeMigrateObject() *apisv1alpha1.LocalVolumeMigrate {
	lvm := &apisv1alpha1.LocalVolumeMigrate{}

	TypeMeta := metav1.TypeMeta{
		Kind:       LocalVolumeMigrateKind,
		APIVersion: apiversion,
	}

	ObjectMata := metav1.ObjectMeta{
		Name:              fakeLocalVolumeMigrateName,
		Namespace:         fakeNamespace,
		ResourceVersion:   "",
		UID:               types.UID(fakeLocalVolumeMigrateUID),
		CreationTimestamp: metav1.Time{Time: time.Now()},
	}

	Spec := apisv1alpha1.LocalVolumeMigrateSpec{
		VolumeName: fakeLocalVolumeName,
		NodeName:   fakeNodename,
		Abort:      true,
	}

	disks := make([]apisv1alpha1.LocalDisk, 0, 10)
	var localdisk1 apisv1alpha1.LocalDisk
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

	lvm.ObjectMeta = ObjectMata
	lvm.TypeMeta = TypeMeta
	lvm.Spec = Spec
	lvm.Status.State = apisv1alpha1.VolumeStateCreating
	return lvm
}

// CreateFakeClient Create LocalVolumeMigrate resource
func CreateFakeClient() (client.Client, *runtime.Scheme) {
	lvm := GenFakeLocalVolumeMigrateObject()
	lvmList := &apisv1alpha1.LocalVolumeMigrateList{
		TypeMeta: metav1.TypeMeta{
			Kind:       LocalVolumeMigrateKind,
			APIVersion: apiversion,
		},
	}

	s := scheme.Scheme
	s.AddKnownTypes(ldmv1alpha1.SchemeGroupVersion, lvm)
	s.AddKnownTypes(ldmv1alpha1.SchemeGroupVersion, lvmList)
	return fake.NewFakeClientWithScheme(s), s
}
