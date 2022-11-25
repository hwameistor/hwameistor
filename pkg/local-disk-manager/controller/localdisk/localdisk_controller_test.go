package localdisk

import (
	"context"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/handler/localdisk"
	"testing"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"k8s.io/client-go/tools/reference"

	v1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	fakeLocalDiskClaimName       = "local-disk-claim-example"
	fakeLocalDiskClaimUID        = "local-disk-claim-example-uid"
	fakeLocalDiskName            = "local-disk-example"
	localDiskUID                 = "local-disk-example-uid"
	fakeNamespace                = "local-disk-manager-test"
	fakeNodename                 = "10-6-118-10"
	diskTypeHDD                  = "HDD"
	devPath                      = "/dev/fake-sda"
	devType                      = "disk"
	vendorVMware                 = "VMware"
	proSCSI                      = "scsi"
	apiversion                   = "hwameistor.io/v1alpha1"
	localDiskKind                = "LocalDisk"
	localDiskClaimKind           = "LocalDiskClaim"
	cap100G                int64 = 100 * 1024 * 1024 * 1024
	cap10G                 int64 = 10 * 1024 * 1024 * 1024
	fakeRecorder                 = record.NewFakeRecorder(100)
)

func TestReconcileLocalDisk_Reconcile(t *testing.T) {
	cli, s := CreateFakeClient()

	// reconcile object
	r := ReconcileLocalDisk{
		Client:      cli,
		Scheme:      s,
		Recorder:    fakeRecorder,
		diskHandler: localdisk.NewLocalDiskHandler(cli, fakeRecorder),
	}

	// set a LocalDiskClaim reference to a localDisk
	setClaimRef := func(ld *v1alpha1.LocalDisk, ldc *v1alpha1.LocalDiskClaim) {
		ld.Spec.ClaimRef, _ = reference.GetReference(nil, ldc)
	}
	cleanResource := func(c client.Client, ld *v1alpha1.LocalDisk, ldc *v1alpha1.LocalDiskClaim) error {
		return c.Delete(context.Background(), ld)
	}
	doNothing := func(ld *v1alpha1.LocalDisk, ldc *v1alpha1.LocalDiskClaim) {}

	testCases := []struct {
		description string
		ld          *v1alpha1.LocalDisk
		ldc         *v1alpha1.LocalDiskClaim
		pre         func(*v1alpha1.LocalDisk, *v1alpha1.LocalDiskClaim)
		wantState   v1alpha1.LocalDiskState
		post        func(client.Client, *v1alpha1.LocalDisk, *v1alpha1.LocalDiskClaim) error
	}{
		{
			description: "Claimed by a LocalDiskClaim",
			ld:          GenFakeLocalDiskObject(),
			ldc:         GenFakeLocalDiskClaimObject(),
			pre:         setClaimRef,
			wantState:   v1alpha1.LocalDiskBound,
			post:        cleanResource,
		},
		{
			description: "Available by any LocalDiskClaim",
			ld:          GenFakeLocalDiskObject(),
			ldc:         GenFakeLocalDiskClaimObject(),
			pre:         doNothing,
			wantState:   v1alpha1.LocalDiskAvailable,
			post:        cleanResource,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			defer testCase.post(r.Client, testCase.ld, testCase.ldc)

			// do some change on localDisk
			testCase.pre(testCase.ld, testCase.ldc)

			// create localDisk
			err := r.Create(context.Background(), testCase.ld)
			if err != nil {
				t.Error(err)
			}

			// create reconcile request for localDisk
			request := reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testCase.ld.GetNamespace(), Name: testCase.ld.GetName()}}

			// reconcile for LocalDisk
			if _, err = r.Reconcile(context.TODO(), request); err != nil {
				t.Error(err)
			}

			// refresh localDisk
			if err = r.Get(context.Background(), request.NamespacedName, testCase.ld); err != nil {
				t.Errorf("Failed to refresh localDisk %s for err %v", request.NamespacedName, err)
			}

			// check wanted state
			if testCase.wantState != testCase.ld.Status.State {
				t.Errorf("Expected LocalDiskClaim State %v but got State %v", testCase.wantState, testCase.ld.Status.State)
			}
		})
	}
}

// CreateFakeClient Create localDisk and LocalDiskClaim resource
func CreateFakeClient() (client.Client, *runtime.Scheme) {
	disk := GenFakeLocalDiskObject()
	diskList := &v1alpha1.LocalDiskList{
		TypeMeta: metav1.TypeMeta{
			Kind:       localDiskKind,
			APIVersion: apiversion,
		},
	}

	claim := GenFakeLocalDiskClaimObject()
	claimList := &v1alpha1.LocalDiskList{
		TypeMeta: metav1.TypeMeta{
			Kind:       localDiskClaimKind,
			APIVersion: apiversion,
		},
	}

	s := scheme.Scheme
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, disk)
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, diskList)
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, claim)
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, claimList)
	return fake.NewFakeClientWithScheme(s), s
}

// GenFakeLocalDiskObject Create disk
// By default, disk can be claimed by the sample calim
func GenFakeLocalDiskObject() *v1alpha1.LocalDisk {
	ld := &v1alpha1.LocalDisk{}

	TypeMeta := metav1.TypeMeta{
		Kind:       localDiskKind,
		APIVersion: apiversion,
	}

	ObjectMata := metav1.ObjectMeta{
		Name:              fakeNodename + devPath,
		Namespace:         fakeNamespace,
		UID:               types.UID(localDiskUID),
		CreationTimestamp: metav1.Time{Time: time.Now()},
	}

	Spec := v1alpha1.LocalDiskSpec{
		NodeName:     fakeNodename,
		DevicePath:   devPath,
		Capacity:     cap100G,
		HasPartition: false,
		HasRAID:      false,
		RAIDInfo:     v1alpha1.RAIDInfo{},
		HasSmartInfo: false,
		SmartInfo:    v1alpha1.SmartInfo{},
		DiskAttributes: v1alpha1.DiskAttributes{
			Type:     diskTypeHDD,
			DevType:  devType,
			Vendor:   vendorVMware,
			Protocol: proSCSI,
		},
		State: v1alpha1.LocalDiskActive,
	}

	Status := v1alpha1.LocalDiskStatus{State: v1alpha1.LocalDiskPending}

	ld.TypeMeta = TypeMeta
	ld.ObjectMeta = ObjectMata
	ld.Spec = Spec
	ld.Status = Status
	return ld
}

// GenFakeLocalDiskClaimObject Create claim request
// By default, claim can be bound to the sample disk
func GenFakeLocalDiskClaimObject() *v1alpha1.LocalDiskClaim {
	ldc := &v1alpha1.LocalDiskClaim{}

	TypeMeta := metav1.TypeMeta{
		Kind:       localDiskClaimKind,
		APIVersion: apiversion,
	}

	ObjectMata := metav1.ObjectMeta{
		Name:              fakeLocalDiskClaimName,
		Namespace:         fakeNamespace,
		ResourceVersion:   "",
		UID:               types.UID(fakeLocalDiskClaimUID),
		CreationTimestamp: metav1.Time{Time: time.Now()},
	}

	Spec := v1alpha1.LocalDiskClaimSpec{
		NodeName: fakeNodename,
		Description: v1alpha1.DiskClaimDescription{
			DiskType: diskTypeHDD,
			Capacity: cap100G,
		},
	}

	ldc.ObjectMeta = ObjectMata
	ldc.TypeMeta = TypeMeta
	ldc.Spec = Spec
	ldc.Status.Status = v1alpha1.LocalDiskClaimStatusPending
	return ldc
}
