package localvolumereplica

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/cache"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-storage/member"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// SystemMode of HA module
type SystemMode string

var (
	fakeLocalVolumeReplicaName       = "local-volume-replica-example"
	fakeLocalVolumeReplicaUID        = "local-volume-replica-uid"
	fakeLocalVolumeName              = "local-volume-example"
	fakeLocalVolumeUID               = "local-volume-uid"
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
	LocalVolumeReplicaKind = "LocalVolumeReplica"
	fakeRecorder           = record.NewFakeRecorder(100)

	defaultDRBDStartPort                 = 43001
	defaultHAVolumeTotalCount            = 1000
	SystemModeDRBD            SystemMode = "drbd"
)

func TestNewLocalVolumeReplicaController(t *testing.T) {

	systemConfig, err := getSystemConfig()
	if err != nil {
		t.Errorf("invalid system config: %s", err)
	}
	var ca cache.Cache

	cli, s := CreateFakeClient()
	// Create a Reconcile for LocalVolumeReplica
	r := ReconcileLocalVolumeReplica{
		client:        cli,
		scheme:        s,
		storageMember: member.Member().ConfigureController(s).ConfigureBase(fakeNodename, fakeNamespace, systemConfig, cli, ca, fakeRecorder).ConfigureNode(s),
	}

	// Create LocalVolumeReplica
	lv := GenFakeLocalVolumeReplicaObject()
	err = r.client.Create(context.Background(), lv)
	if err != nil {
		t.Errorf("Create LocalVolumeReplica fail %v", err)
	}
	defer r.DeleteFakeLocalVolumeReplica(t, lv)

	// Get lv
	err = r.client.Get(context.Background(), types.NamespacedName{Namespace: lv.GetNamespace(), Name: lv.GetName()}, lv)
	if err != nil {
		t.Errorf("Get lv fail %v", err)
	}
	fmt.Printf("lv = %+v", lv)
	fmt.Printf("r.storageMember = %+v", r.storageMember)

	// Mock LocalVolumeReplica request
	req := reconcile.Request{NamespacedName: types.NamespacedName{Namespace: lv.GetNamespace(), Name: lv.GetName()}}
	_, err = r.Reconcile(context.TODO(), req)
	if err != nil {
		t.Errorf("Reconcile fail %v", err)
	}

}

// DeleteFakeLocalVolumeReplica
func (r *ReconcileLocalVolumeReplica) DeleteFakeLocalVolumeReplica(t *testing.T, lv *apisv1alpha1.LocalVolumeReplica) {
	if err := r.client.Delete(context.Background(), lv); err != nil {
		t.Errorf("Delete LocalVolumeReplica %v fail %v", lv.GetName(), err)
	}
}

// GenFakeLocalVolumeReplicaObject Create lv request
func GenFakeLocalVolumeReplicaObject() *apisv1alpha1.LocalVolumeReplica {
	lv := &apisv1alpha1.LocalVolumeReplica{}

	TypeMeta := metav1.TypeMeta{
		Kind:       LocalVolumeReplicaKind,
		APIVersion: apiversion,
	}

	ObjectMata := metav1.ObjectMeta{
		Name:              fakeLocalVolumeReplicaName,
		Namespace:         fakeNamespace,
		ResourceVersion:   "",
		UID:               types.UID(fakeLocalVolumeReplicaUID),
		CreationTimestamp: metav1.Time{Time: time.Now()},
	}

	Spec := apisv1alpha1.LocalVolumeReplicaSpec{
		RequiredCapacityBytes: fakeDiskCapacityBytes,
		VolumeName:            fakeLocalVolumeName,
		PoolName:              fakeVgType,
		Delete:                false,
		NodeName:              fakeNodename,
	}

	lv.ObjectMeta = ObjectMata
	lv.TypeMeta = TypeMeta
	lv.Spec = Spec
	lv.Status.State = apisv1alpha1.VolumeReplicaStateNotReady
	lv.Status.AllocatedCapacityBytes = fakeTotalCapacityBytes - fakeFreeCapacityBytes
	return lv
}

// CreateFakeClient Create LocalVolumeReplica resource
func CreateFakeClient() (client.Client, *runtime.Scheme) {
	lv := GenFakeLocalVolumeReplicaObject()
	lvList := &apisv1alpha1.LocalVolumeReplicaList{
		TypeMeta: metav1.TypeMeta{
			Kind:       LocalVolumeReplicaKind,
			APIVersion: apiversion,
		},
	}

	s := scheme.Scheme
	s.AddKnownTypes(apisv1alpha1.SchemeGroupVersion, lv)
	s.AddKnownTypes(apisv1alpha1.SchemeGroupVersion, lvList)
	return fake.NewFakeClientWithScheme(s), s
}

func validateSystemConfig() error {
	var errMsgs []string
	switch apisv1alpha1.SystemMode(SystemModeDRBD) {
	case apisv1alpha1.SystemModeDRBD:
	default:
		errMsgs = append(errMsgs, fmt.Sprintf("system mode %s not supported", SystemModeDRBD))
	}

	if len(errMsgs) != 0 {
		return fmt.Errorf(strings.Join(errMsgs, "; "))
	}
	return nil
}

func getSystemConfig() (apisv1alpha1.SystemConfig, error) {
	if err := validateSystemConfig(); err != nil {
		return apisv1alpha1.SystemConfig{}, err
	}

	config := apisv1alpha1.SystemConfig{
		Mode:             apisv1alpha1.SystemMode(SystemModeDRBD),
		MaxHAVolumeCount: defaultHAVolumeTotalCount,
	}

	switch config.Mode {
	case apisv1alpha1.SystemModeDRBD:
		{
			config.DRBD = &apisv1alpha1.DRBDSystemConfig{
				StartPort: defaultDRBDStartPort,
			}
		}
	}
	return config, nil
}
