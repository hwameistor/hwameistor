package localvolumegroupmigrate

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/wxnacy/wgo/arrays"
	"k8s.io/apimachinery/pkg/api/errors"
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
	fakeLocalVolumeName                   = "local-volume-test1"
	fakeLocalVolumeGroupName              = "local-volume-group-example"
	fakeLocalVolumeGroupMigrateName       = "local-volume-group-convert-example"
	fakeLocalVolumeGroupMigrateUID        = "local-volume-group-convert-uid"
	fakeLocalVolumeGroupUID               = "local-volume-group-uid"
	fakeLocalVolumeUID                    = "local-volume-uid"
	fakeNamespace                         = "local-volume-group-test"
	fakeSourceNodenames                   = []string{"10-6-118-10"}
	fakeTargetNodenames                   = []string{"10-6-118-11"}
	fakeVolumes                           = []apisv1alpha1.VolumeInfo{{LocalVolumeName: "local-volume-test1", PersistentVolumeClaimName: "pvc-test1"}}
	fakeStorageIp                         = "10.6.118.11"
	fakeZone                              = "zone-test"
	fakeRegion                            = "region-test"
	fakePods                              = []string{"pod-test1"}
	fakeVgType                            = "LocalStorage_PoolHDD"
	fakeVgName                            = "vg-test"
	fakePoolClass                         = "HDD"
	fakePoolType                          = "REGULAR"
	fakePoolName                          = "pool-test-1"
	fakeTotalCapacityBytes          int64 = 10 * 1024 * 1024 * 1024
	fakeFreeCapacityBytes           int64 = 8 * 1024 * 1024 * 1024
	fakeDiskCapacityBytes           int64 = 2 * 1024 * 1024 * 1024

	apiversion                  = "hwameistor.io/v1alpha1"
	LocalVolumeGroupMigrateKind = "LocalVolumeGroupMigrate"
	LocalVolumeGroupKind        = "LocalVolumeGroup"
	LocalVolumeKind             = "LocalVolume"
	fakeRecorder                = record.NewFakeRecorder(100)
	fakeAcesscibility           = apisv1alpha1.AccessibilityTopology{Nodes: []string{"test-node1"}}
)

func TestNewLocalVolumeGroupMigrateController(t *testing.T) {

	cli, s := CreateFakeClient()
	// Create a Reconcile for LocalVolumeGroupMigrate
	r := ReconcileLocalVolumeGroupMigrate{
		client:        cli,
		scheme:        s,
		storageMember: member.Member().ConfigureController(s),
	}

	// Create LocalVolumeGroupMigrate
	lvgm := GenFakeLocalVolumeGroupMigrateObject()
	err := r.client.Create(context.Background(), lvgm)
	if err != nil {
		t.Errorf("TestNewLocalVolumeGroupMigrateController Create LocalVolumeGroupMigrate fail %v", err)
	}
	defer r.DeleteFakeLocalVolumeGroupMigrate(t, lvgm)

	// Get lvgm
	err = r.client.Get(context.Background(), types.NamespacedName{Namespace: lvgm.GetNamespace(), Name: lvgm.GetName()}, lvgm)
	if err != nil {
		t.Errorf("TestNewLocalVolumeGroupMigrateController  Get lvgm fail %v", err)
	}

	lvg := GenFakeLocalVolumeGroupObject(lvgm.Spec.LocalVolumeGroupName)
	err = r.client.Create(context.Background(), lvg)
	if err != nil {
		t.Errorf("TestNewLocalVolumeGroupMigrateController CreateFakeLocalVolumeGroupObject fail %v", err)
	}
	for _, tmpvol := range lvg.Spec.Volumes {
		if tmpvol.LocalVolumeName == "" {
			continue
		}
		vol := GenFakeLocalVolumeObject()
		err = r.client.Create(context.Background(), vol)
		if err != nil {
			t.Errorf("TestNewLocalVolumeGroupMigrateController CreateFakeLocalVolumeObject fail %v", err)
		}
		for _, nodeName := range lvgm.Spec.TargetNodesNames {
			if arrays.ContainsString(vol.Spec.Accessibility.Nodes, nodeName) == -1 {
				vol.Spec.Accessibility.Nodes = append(vol.Spec.Accessibility.Nodes, nodeName)
			}
		}
		if err := r.client.Get(context.TODO(), types.NamespacedName{Name: tmpvol.LocalVolumeName}, vol); err != nil {
			if !errors.IsNotFound(err) {
				log.WithFields(log.Fields{"volName": tmpvol.LocalVolumeName, "error": err.Error()}).Error("Failed to query volume")
			}
		}
		if err := r.client.Update(context.TODO(), vol); err != nil {
			log.WithError(err).Error("TestNewLocalVolumeGroupMigrateController Reconcile : Failed to re-configure Volume")
			t.Errorf("TestNewLocalVolumeGroupMigrateController Update fail %v", err)
		}
	}

	// Mock LocalVolumeGroupMigrate request
	req := reconcile.Request{NamespacedName: types.NamespacedName{Namespace: lvgm.GetNamespace(), Name: lvgm.GetName()}}
	_, err = r.Reconcile(req)
	if err != nil {
		t.Errorf("Reconcile fail %v", err)
	}

}

// DeleteFakeLocalVolumeGroupMigrate
func (r *ReconcileLocalVolumeGroupMigrate) DeleteFakeLocalVolumeGroupMigrate(t *testing.T, lvgm *apisv1alpha1.LocalVolumeGroupMigrate) {
	if err := r.client.Delete(context.Background(), lvgm); err != nil {
		t.Errorf("Delete LocalVolumeGroupMigrate %v fail %v", lvgm.GetName(), err)
	}
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
		VolumeGroup:   fakeLocalVolumeGroupName,
		Accessibility: fakeAcesscibility,
		PoolName:      fakePoolName,
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

	lv.ObjectMeta = ObjectMata
	lv.TypeMeta = TypeMeta
	lv.Spec = Spec
	return lv
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

	lvg.ObjectMeta = ObjectMata
	lvg.TypeMeta = TypeMeta
	lvg.Spec = Spec
	return lvg
}

// GenFakeLocalVolumeGroupMigrateObject Create lvgm request
func GenFakeLocalVolumeGroupMigrateObject() *apisv1alpha1.LocalVolumeGroupMigrate {
	lvgm := &apisv1alpha1.LocalVolumeGroupMigrate{}

	TypeMeta := metav1.TypeMeta{
		Kind:       LocalVolumeGroupMigrateKind,
		APIVersion: apiversion,
	}

	ObjectMata := metav1.ObjectMeta{
		Name:              fakeLocalVolumeGroupMigrateName,
		Namespace:         fakeNamespace,
		ResourceVersion:   "",
		UID:               types.UID(fakeLocalVolumeGroupMigrateUID),
		CreationTimestamp: metav1.Time{Time: time.Now()},
	}

	Spec := apisv1alpha1.LocalVolumeGroupMigrateSpec{
		LocalVolumeGroupName: fakeLocalVolumeGroupName,
		SourceNodesNames:     fakeSourceNodenames,
		TargetNodesNames:     fakeTargetNodenames,
		Abort:                true,
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

	lvgm.ObjectMeta = ObjectMata
	lvgm.TypeMeta = TypeMeta
	lvgm.Spec = Spec
	lvgm.Status.State = apisv1alpha1.VolumeStateCreating
	return lvgm
}

// CreateFakeClient Create LocalVolumeGroupMigrate resource
func CreateFakeClient() (client.Client, *runtime.Scheme) {
	lvg := GenFakeLocalVolumeGroupObject(fakeLocalVolumeGroupName)
	lvgList := &apisv1alpha1.LocalVolumeGroupList{
		TypeMeta: metav1.TypeMeta{
			Kind:       LocalVolumeGroupKind,
			APIVersion: apiversion,
		},
	}
	lvgm := GenFakeLocalVolumeGroupMigrateObject()
	lvgmList := &apisv1alpha1.LocalVolumeGroupMigrateList{
		TypeMeta: metav1.TypeMeta{
			Kind:       LocalVolumeGroupMigrateKind,
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

	s := scheme.Scheme
	s.AddKnownTypes(ldmv1alpha1.SchemeGroupVersion, lvgm)
	s.AddKnownTypes(ldmv1alpha1.SchemeGroupVersion, lvgmList)
	s.AddKnownTypes(ldmv1alpha1.SchemeGroupVersion, lvg)
	s.AddKnownTypes(ldmv1alpha1.SchemeGroupVersion, lvgList)
	s.AddKnownTypes(ldmv1alpha1.SchemeGroupVersion, lv)
	s.AddKnownTypes(ldmv1alpha1.SchemeGroupVersion, lvList)
	return fake.NewFakeClientWithScheme(s), s
}
