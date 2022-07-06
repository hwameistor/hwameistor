package localdiskclaim

import (
	"context"
	"testing"
	"time"

	ldmv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/local-disk-manager/v1alpha1"
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

func TestLocalDiskClaimController_FilterByDiskCapacity(t *testing.T) {
	cli, s := CreateFakeClient()

	// Create a Reconcile for LocalDiskClaim
	r := ReconcileLocalDiskClaim{
		Client:   cli,
		Scheme:   s,
		Recorder: fakeRecorder,
	}

	testcases := []struct {
		description string
		ld          *ldmv1alpha1.LocalDisk
		ldc         *ldmv1alpha1.LocalDiskClaim
		setProperty func(claim *ldmv1alpha1.LocalDiskClaim, disk *ldmv1alpha1.LocalDisk)
		wantBound   bool
	}{
		// Disk cap100G is sufficient, should success
		{
			description: "Should return success, ldc state should be Bound",
			ld:          GenFakeLocalDiskObject(),
			ldc:         GenFakeLocalDiskClaimObject(),
			setProperty: func(claim *ldmv1alpha1.LocalDiskClaim, disk *ldmv1alpha1.LocalDisk) {
				// Modify disk cap100G to meet disk requirements
				disk.Spec.Capacity = cap100G
				claim.Spec.Description.Capacity = cap100G
			},
			wantBound: true,
		},

		// Disk cap10G is not enough, should fail
		{
			description: "Should return fail, ldc state should be Pending",
			ld:          GenFakeLocalDiskObject(),
			ldc:         GenFakeLocalDiskClaimObject(),
			setProperty: func(claim *ldmv1alpha1.LocalDiskClaim, disk *ldmv1alpha1.LocalDisk) {
				// Modify disk cap10G to do not meet disk requirements
				disk.Spec.Capacity = cap10G
				claim.Spec.Description.Capacity = cap100G
			},
			wantBound: false,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.description, func(t *testing.T) {
			// setProperty first
			testcase.setProperty(testcase.ldc, testcase.ld)

			// Reconcile
			r.ClaimLocalDisk(t, testcase.ld, testcase.ldc)

			// Check claim Status
			r.CheckLocalDiskClaimIsBound(t, testcase.ldc, testcase.wantBound)

			// Check disk bound relationship
			if testcase.wantBound {
				r.CheckDiskBound(t, testcase.ld, testcase.ldc)
			}
		})
	}
}

func TestReconcileLocalDiskClaim_Reconcile(t *testing.T) {
	cli, s := CreateFakeClient()
	// Create a Reconcile for LocalDiskClaim
	r := ReconcileLocalDiskClaim{
		Client:   cli,
		Scheme:   s,
		Recorder: fakeRecorder,
	}

	// Create LocalDisk
	disk := GenFakeLocalDiskObject()
	err := r.Create(context.Background(), disk)
	if err != nil {
		t.Errorf("Create LocalDisk fail %v", err)
	}
	defer r.DeleteFakeLocalDisk(t, disk)

	// Create LocalDiskClaim
	claim := GenFakeLocalDiskClaimObject()
	err = r.Create(context.Background(), claim)
	if err != nil {
		t.Errorf("Create LocalDiskClaim fail %v", err)
	}
	defer r.DeleteFakeLocalDiskClaim(t, claim)

	// Mock LocalDiskClaim request
	req := reconcile.Request{NamespacedName: types.NamespacedName{Namespace: claim.GetNamespace(), Name: claim.GetName()}}
	_, err = r.Reconcile(req)
	if err != nil {
		t.Errorf("Reconcile fail %v", err)
	}

	// Update claim
	err = r.Get(context.Background(), req.NamespacedName, claim)
	if err != nil {
		t.Errorf("Get disk claim fail %v", err)
	}

	// Checkout claim status, it should be bound
	r.CheckLocalDiskClaimIsBound(t, claim, true)
}

// CheckLocalDiskClaimIsBound
func (r *ReconcileLocalDiskClaim) CheckLocalDiskClaimIsBound(t *testing.T,
	claim *ldmv1alpha1.LocalDiskClaim, wantBound bool) {

	wantPhase := ldmv1alpha1.DiskClaimStatusEmpty
	if wantBound {
		wantPhase = ldmv1alpha1.LocalDiskClaimStatusBound
	} else {
		wantPhase = ldmv1alpha1.LocalDiskClaimStatusPending
	}

	if claim.Status.Status == wantPhase {
		t.Logf("LocalDiskClaim %v status is %v", claim.Name, claim.Status.Status)
	} else {
		t.Fatalf("LocalDiskClaim %v status: %v, want status: %v", claim.Name, claim.Status.Status, wantPhase)
	}
}

// ClaimLocalDisk Create disk and claim request, then try to reconcile the claim request
func (r *ReconcileLocalDiskClaim) ClaimLocalDisk(t *testing.T,
	disk *ldmv1alpha1.LocalDisk, claim *ldmv1alpha1.LocalDiskClaim) {

	// Create LocalDisk
	err := r.Create(context.Background(), disk)
	if err != nil {
		t.Errorf("Create LocalDisk fail %v", err)
	}
	defer r.DeleteFakeLocalDisk(t, disk)

	// Create LocalDiskClaim
	err = r.Create(context.Background(), claim)
	if err != nil {
		t.Errorf("Create LocalDiskClaim fail %v", err)
	}
	defer r.DeleteFakeLocalDiskClaim(t, claim)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Namespace: claim.GetNamespace(),
			Name:      claim.GetName(),
		},
	}

	// Reconcile request
	_, err = r.Reconcile(req)
	if err != nil {
		t.Errorf("Reconcile fail %v", err)
	}

	// Update status
	err = r.Client.Get(context.Background(), req.NamespacedName, claim)
	if err != nil {
		t.Errorf("Get LocalDiskClaim fail %v", err)
	}
}

// CheckDiskBound check disk is bound with designated claim
func (r *ReconcileLocalDiskClaim) CheckDiskBound(t *testing.T, disk *ldmv1alpha1.LocalDisk, claim *ldmv1alpha1.LocalDiskClaim) {
	// Check that DiskRef is the specified disk
	findDisk := false
	for _, boundDisk := range claim.Spec.DiskRefs {
		if boundDisk.Name == disk.Name {
			findDisk = true
			break
		}
	}
	if !findDisk {
		t.Fatalf("LocalDiskClaim %v has not bound disk: %v", claim.GetName(), claim.GetName())
	}

	t.Logf("LocalDisk %v has bound with LocalDiskClaim %v", claim.GetName(), claim.GetName())
}

// DeleteFakeLocalDisk
func (r *ReconcileLocalDiskClaim) DeleteFakeLocalDisk(t *testing.T, ld *ldmv1alpha1.LocalDisk) {
	if err := r.Delete(context.Background(), ld); err != nil {
		t.Errorf("Delete LocalDisk %v fail %v", ld.GetName(), err)
	}
}

// DeleteFakeLocalDiskClaim
func (r *ReconcileLocalDiskClaim) DeleteFakeLocalDiskClaim(t *testing.T, ldc *ldmv1alpha1.LocalDiskClaim) {
	if err := r.Delete(context.Background(), ldc); err != nil {
		t.Errorf("Delete LocalDiskClaim %v fail %v", ldc.GetName(), err)
	}
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
