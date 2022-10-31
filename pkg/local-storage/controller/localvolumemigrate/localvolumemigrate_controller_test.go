package localvolumemigrate

import (
	"context"
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
	fakeLocalVolumeMigrateName = "local-volume-convert-example"
	fakeLocalVolumeMigrateUID  = "local-volume-convert-uid"
	fakeNodename               = "10-6-118-10"
	fakeNodenames              = []string{"10-6-118-10"}

	LocalVolumeMigrateKind = "LocalVolumeMigrate"
	//fakeRecorder           = record.NewFakeRecorder(100)

	fakeLocalVolumeName      = "local-volume-test1"
	fakeLocalVolumeGroupName = "local-volume-group-example"
	// fakeLocalVolumeGroupMigrateName       = "local-volume-group-convert-example"
	// fakeLocalVolumeGroupMigrateUID        = "local-volume-group-convert-uid"
	fakeLocalVolumeGroupUID = "local-volume-group-uid"
	fakeLocalVolumeUID      = "local-volume-uid"
	fakeNamespace           = "local-volume-group-test"
	// fakeSourceNodenames                   = []string{"10-6-118-10"}
	// fakeTargetNodenames                   = []string{"10-6-118-11"}
	fakePersistentPvcName = "pvc-test"
	fakeVolumes           = []apisv1alpha1.VolumeInfo{{LocalVolumeName: fakeLocalVolumeName, PersistentVolumeClaimName: fakePersistentPvcName}}
	fakeStorageIp         = "10.6.118.11"
	fakeZone              = "zone-test"
	fakeRegion            = "region-test"
	fakePods              = []string{"pod-test1"}
	fakeVgType            = "LocalStorage_PoolHDD"
	fakeVgName            = "vg-test"
	fakePoolClass         = "HDD"
	fakePoolType          = "REGULAR"
	// fakePoolName                          = "pool-test-1"
	fakeTotalCapacityBytes int64 = 10 * 1024 * 1024 * 1024
	fakeFreeCapacityBytes  int64 = 8 * 1024 * 1024 * 1024
	fakeDiskCapacityBytes  int64 = 2 * 1024 * 1024 * 1024

	apiversion                  = "hwameistor.io/v1alpha1"
	LocalVolumeGroupMigrateKind = "LocalVolumeGroupMigrate"
	LocalVolumeGroupKind        = "LocalVolumeGroup"
	LocalVolumeKind             = "LocalVolume"
	fakeAcesscibility           = apisv1alpha1.AccessibilityTopology{Nodes: []string{"test-node1"}}
)

func TestNewLocalVolumeMigrateController(t *testing.T) {

	cli, s := CreateFakeClient()
	// Create a Reconcile for LocalVolumeMigrate
	r := ReconcileLocalVolumeMigrate{
		client:        cli,
		scheme:        s,
		storageMember: member.Member().ConfigureController(s),
	}

	// Create LocalVolume
	lv := GenFakeLocalVolumeObject()
	lv.Name = fakeLocalVolumeName
	lv.Spec.VolumeGroup = fakeLocalVolumeGroupName
	err := r.client.Create(context.Background(), lv)
	if err != nil {
		t.Errorf("Create LocalVolume fail %v", err)
	}
	defer r.DeleteFakeLocalVolume(t, lv)

	lvg := GenFakeLocalVolumeGroupObject(lv.Spec.VolumeGroup)
	lvg.Name = fakeLocalVolumeGroupName
	lvg.Namespace = fakeNamespace
	lvg.Spec.Volumes = fakeVolumes
	err = r.client.Create(context.Background(), lvg)
	if err != nil {
		t.Errorf("TestNewLocalVolumeGroupMigrateController CreateFakeLocalVolumeGroupObject fail %v", err)
	}

	// Create LocalVolumeMigrate
	lvm := GenFakeLocalVolumeMigrateObject()
	lvm.Name = fakeLocalVolumeMigrateName
	lvm.Namespace = fakeNamespace
	lvm.Spec.VolumeName = fakeLocalVolumeName
	err = r.client.Create(context.Background(), lvm)
	if err != nil {
		t.Errorf("Create LocalVolumeMigrate fail %v", err)
	}

	// Get lvm
	err = r.client.Get(context.Background(), types.NamespacedName{Namespace: lvg.GetNamespace(), Name: lvm.GetName()}, lvm)
	if err != nil {
		t.Errorf("Get lvm fail %v", err)
	}

	// Mock LocalVolumeMigrate request
	req := reconcile.Request{NamespacedName: types.NamespacedName{Namespace: lvm.GetNamespace(), Name: lvm.GetName()}}
	_, err = r.Reconcile(context.TODO(), req)
	if err != nil {
		t.Logf("Reconcile fail %v", err)
	}

}

// DeleteFakeLocalVolumeMigrate
func (r *ReconcileLocalVolumeMigrate) DeleteFakeLocalVolumeMigrate(t *testing.T, lvm *apisv1alpha1.LocalVolumeMigrate) {
	if err := r.client.Delete(context.Background(), lvm); err != nil {
		t.Errorf("Delete LocalVolumeMigrate %v fail %v", lvm.GetName(), err)
	}
}

// DeleteFakeLocalVolume
func (r *ReconcileLocalVolumeMigrate) DeleteFakeLocalVolume(t *testing.T, lv *apisv1alpha1.LocalVolume) {
	if err := r.client.Delete(context.Background(), lv); err != nil {
		t.Errorf("Delete LocalVolume %v fail %v", lv.GetName(), err)
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
		VolumeName:           fakeLocalVolumeName,
		TargetNodesSuggested: fakeNodenames,
		Abort:                true,
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

	lv := GenFakeLocalVolumeObject()
	lvList := &apisv1alpha1.LocalVolumeList{
		TypeMeta: metav1.TypeMeta{
			Kind:       LocalVolumeKind,
			APIVersion: apiversion,
		},
	}

	lvg := GenFakeLocalVolumeGroupObject(fakeLocalVolumeGroupName)
	lvgList := &apisv1alpha1.LocalVolumeGroupList{
		TypeMeta: metav1.TypeMeta{
			Kind:       LocalVolumeGroupKind,
			APIVersion: apiversion,
		},
	}

	s := scheme.Scheme
	s.AddKnownTypes(apisv1alpha1.SchemeGroupVersion, lvm)
	s.AddKnownTypes(apisv1alpha1.SchemeGroupVersion, lvmList)
	s.AddKnownTypes(apisv1alpha1.SchemeGroupVersion, lvg)
	s.AddKnownTypes(apisv1alpha1.SchemeGroupVersion, lvgList)
	s.AddKnownTypes(apisv1alpha1.SchemeGroupVersion, lv)
	s.AddKnownTypes(apisv1alpha1.SchemeGroupVersion, lvList)
	return fake.NewFakeClientWithScheme(s), s
}

// GenFakeLocalVolumeGroupMigrateObject Create lvgm request
func GenFakeLocalVolumeGroupObject(lvgName string) *apisv1alpha1.LocalVolumeGroup {
	lvg := &apisv1alpha1.LocalVolumeGroup{}

	TypeMeta := metav1.TypeMeta{
		Kind:       LocalVolumeGroupKind,
		APIVersion: apiversion,
	}

	ObjectMata := metav1.ObjectMeta{
		Name:              lvgName,
		Namespace:         fakeNamespace,
		ResourceVersion:   "",
		UID:               types.UID(fakeLocalVolumeGroupUID),
		CreationTimestamp: metav1.Time{Time: time.Now()},
	}

	Spec := apisv1alpha1.LocalVolumeGroupSpec{
		Volumes:       fakeVolumes,
		Accessibility: fakeAcesscibility,
		Pods:          fakePods,
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

	lvg.ObjectMeta = ObjectMata
	lvg.TypeMeta = TypeMeta
	lvg.Spec = Spec
	return lvg
}

// GenFakeLocalVolumeObject Create lv request
func GenFakeLocalVolumeObject() *apisv1alpha1.LocalVolume {
	lv := &apisv1alpha1.LocalVolume{}

	TypeMeta := metav1.TypeMeta{
		Kind:       LocalVolumeKind,
		APIVersion: apiversion,
	}

	ObjectMata := metav1.ObjectMeta{
		Name:              fakeLocalVolumeName,
		Namespace:         fakeNamespace,
		ResourceVersion:   "",
		UID:               types.UID(fakeLocalVolumeUID),
		CreationTimestamp: metav1.Time{Time: time.Now()},
	}

	Spec := apisv1alpha1.LocalVolumeSpec{
		RequiredCapacityBytes: fakeDiskCapacityBytes,
		ReplicaNumber:         1,
		PoolName:              fakeVgType,
		Delete:                false,
		Convertible:           true,
		Accessibility: apisv1alpha1.AccessibilityTopology{
			Nodes:   fakeNodenames,
			Regions: []string{fakeRegion},
			Zones:   []string{fakeZone},
		},
		Config: &apisv1alpha1.VolumeConfig{
			Convertible:           true,
			Initialized:           true,
			ReadyToInitialize:     true,
			RequiredCapacityBytes: fakeDiskCapacityBytes,
			ResourceID:            5,
			Version:               11,
			VolumeName:            fakeLocalVolumeName,
			Replicas: []apisv1alpha1.VolumeReplica{
				{
					Hostname: fakeNodename,
					ID:       1,
					IP:       fakeStorageIp,
					Primary:  true,
				},
			},
		},
	}

	lv.ObjectMeta = ObjectMata
	lv.TypeMeta = TypeMeta
	lv.Spec = Spec
	lv.Status.State = apisv1alpha1.VolumeStateCreating
	lv.Status.AllocatedCapacityBytes = fakeTotalCapacityBytes - fakeFreeCapacityBytes
	lv.Status.PublishedNodeName = fakeNodename
	lv.Status.Replicas = []string{fakeLocalVolumeName}

	return lv
}
