package localdisk

import (
	"context"
	"testing"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"k8s.io/client-go/tools/reference"

	ldmv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/local-disk-manager/v1alpha1"
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
		Client:   cli,
		Scheme:   s,
		Recorder: fakeRecorder,
	}

	// set a LocalDiskClaim reference to a LocalDisk
	setClaimRef := func(ld *ldmv1alpha1.LocalDisk, ldc *ldmv1alpha1.LocalDiskClaim) {
		ld.Spec.ClaimRef, _ = reference.GetReference(nil, ldc)
	}
	cleanResource := func(c client.Client, ld *ldmv1alpha1.LocalDisk, ldc *ldmv1alpha1.LocalDiskClaim) error {
		return c.Delete(context.Background(), ld)
	}
	doNothing := func(ld *ldmv1alpha1.LocalDisk, ldc *ldmv1alpha1.LocalDiskClaim) {}

	testCases := []struct {
		description string
		ld          *ldmv1alpha1.LocalDisk
		ldc         *ldmv1alpha1.LocalDiskClaim
		pre         func(*ldmv1alpha1.LocalDisk, *ldmv1alpha1.LocalDiskClaim)
		wantState   ldmv1alpha1.LocalDiskClaimState
		post        func(client.Client, *ldmv1alpha1.LocalDisk, *ldmv1alpha1.LocalDiskClaim) error
	}{
		{
			description: "Claimed by a LocalDiskClaim",
			ld:          GenFakeLocalDiskObject(),
			ldc:         GenFakeLocalDiskClaimObject(),
			pre:         setClaimRef,
			wantState:   ldmv1alpha1.LocalDiskClaimed,
			post:        cleanResource,
		},
		{
			description: "Unclaimed by any LocalDiskClaim",
			ld:          GenFakeLocalDiskObject(),
			ldc:         GenFakeLocalDiskClaimObject(),
			pre:         doNothing,
			wantState:   ldmv1alpha1.LocalDiskUnclaimed,
			post:        cleanResource,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			defer testCase.post(r.Client, testCase.ld, testCase.ldc)

			// do some change on LocalDisk
			testCase.pre(testCase.ld, testCase.ldc)

			// create LocalDisk
			err := r.Create(context.Background(), testCase.ld)
			if err != nil {
				t.Error(err)
			}

			// create reconcile request for LocalDisk
			request := reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testCase.ld.GetNamespace(), Name: testCase.ld.GetName()}}

			// reconcile for LocalDisk
			if _, err = r.Reconcile(request); err != nil {
				t.Error(err)
			}

			// refresh LocalDisk
			if err = r.Get(context.Background(), request.NamespacedName, testCase.ld); err != nil {
				t.Errorf("Failed to refresh LocalDisk %s for err %v", request.NamespacedName, err)
			}

			// check wanted state
			if testCase.wantState != testCase.ld.Status.State {
				t.Errorf("Expected LocalDiskClaim State %v but got State %v", testCase.wantState, testCase.ld.Status.State)
			}
		})
	}
}

// CreateFakeClient Create LocalDisk and LocalDiskClaim resource
func CreateFakeClient() (client.Client, *runtime.Scheme) {
	disk := GenFakeLocalDiskObject()
	diskList := &ldmv1alpha1.LocalDiskList{
		TypeMeta: metav1.TypeMeta{
			Kind:       localDiskKind,
			APIVersion: apiversion,
		},
	}

	claim := GenFakeLocalDiskClaimObject()
	claimList := &ldmv1alpha1.LocalDiskList{
		TypeMeta: metav1.TypeMeta{
			Kind:       localDiskClaimKind,
			APIVersion: apiversion,
		},
	}

	s := scheme.Scheme
	s.AddKnownTypes(ldmv1alpha1.SchemeGroupVersion, disk)
	s.AddKnownTypes(ldmv1alpha1.SchemeGroupVersion, diskList)
	s.AddKnownTypes(ldmv1alpha1.SchemeGroupVersion, claim)
	s.AddKnownTypes(ldmv1alpha1.SchemeGroupVersion, claimList)
	return fake.NewFakeClientWithScheme(s), s
}

// GenFakeLocalDiskObject Create disk
// By default, disk can be claimed by the sample calim
func GenFakeLocalDiskObject() *ldmv1alpha1.LocalDisk {
	ld := &ldmv1alpha1.LocalDisk{}

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

	Spec := ldmv1alpha1.LocalDiskSpec{
		NodeName:     fakeNodename,
		DevicePath:   devPath,
		Capacity:     cap100G,
		HasPartition: false,
		HasRAID:      false,
		RAIDInfo:     ldmv1alpha1.RAIDInfo{},
		HasSmartInfo: false,
		SmartInfo:    ldmv1alpha1.SmartInfo{},
		DiskAttributes: ldmv1alpha1.DiskAttributes{
			Type:     diskTypeHDD,
			DevType:  devType,
			Vendor:   vendorVMware,
			Protocol: proSCSI,
		},
		State: ldmv1alpha1.LocalDiskActive,
	}

	Status := ldmv1alpha1.LocalDiskStatus{State: ldmv1alpha1.LocalDiskUnclaimed}

	ld.TypeMeta = TypeMeta
	ld.ObjectMeta = ObjectMata
	ld.Spec = Spec
	ld.Status = Status
	return ld
}

// GenFakeLocalDiskClaimObject Create claim request
// By default, claim can be bound to the sample disk
func GenFakeLocalDiskClaimObject() *ldmv1alpha1.LocalDiskClaim {
	ldc := &ldmv1alpha1.LocalDiskClaim{}

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

	Spec := ldmv1alpha1.LocalDiskClaimSpec{
		NodeName: fakeNodename,
		Description: ldmv1alpha1.DiskClaimDescription{
			DiskType: diskTypeHDD,
			Capacity: cap100G,
		},
	}

	ldc.ObjectMeta = ObjectMata
	ldc.TypeMeta = TypeMeta
	ldc.Spec = Spec
	ldc.Status.Status = ldmv1alpha1.LocalDiskClaimStatusPending
	return ldc
}
