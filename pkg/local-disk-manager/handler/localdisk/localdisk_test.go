package localdisk

import (
	"context"
	"fmt"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	v1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

var (
	fakeLocalDiskClaimName       = "local-disk-claim-example"
	fakeLocalDiskClaimUID        = "local-disk-claim-example-uid"
	fakeLocalDiskName            = "local-disk-example"
	localDiskUID                 = "local-disk-example-uid"
	fakeNamespace                = "local-disk-manager-test"
	fakeNodename                 = "10-6-118-10"
	diskTypeHDD                  = "HDD"
	fakedevPath                  = "/dev/fake-sda"
	devType                      = "disk"
	vendorVMware                 = "VMware"
	proSCSI                      = "scsi"
	apiversion                   = "hwameistor.io/v1alpha1"
	localDiskKind                = "localDisk"
	localDiskNodeKind            = "LocalDiskNode"
	localDiskClaimKind           = "LocalDiskClaim"
	cap100G                int64 = 100 * 1024 * 1024 * 1024
	cap10G                 int64 = 10 * 1024 * 1024 * 1024
	fakeRecorder                 = record.NewFakeRecorder(100)
	fakeLabelKey                 = "label1"
	fakeLabelValue               = "value"
)

func TestLocalDiskHandler_BoundTo(t *testing.T) {
	cli, _ := CreateFakeClient()
	handler := NewLocalDiskHandler(cli, fakeRecorder)

	createlocaldisk := func(cli client.Client, localdisk *v1alpha1.LocalDisk) error {
		return cli.Create(context.Background(), localdisk)
	}

	cleanlocaldisk := func(cli client.Client, localdisk *v1alpha1.LocalDisk) error {
		return cli.Delete(context.Background(), localdisk)
	}

	createResource := func(cli client.Client, resource interface{}) error {
		switch resource.(type) {
		case *v1alpha1.LocalDisk:
			return createlocaldisk(cli, resource.(*v1alpha1.LocalDisk))
		default:
			return fmt.Errorf("unknown resource type")
		}
	}

	cleanResource := func(cli client.Client, resource interface{}) error {
		switch resource.(type) {
		case *v1alpha1.LocalDisk:
			return cleanlocaldisk(cli, resource.(*v1alpha1.LocalDisk))
		default:
			return fmt.Errorf("unknown resource type")
		}
	}

	testCases := []struct {
		description string
		preAction   func(cli client.Client, resource interface{}) error
		postAction  func(cli client.Client, resource interface{}) error
		ld          *v1alpha1.LocalDisk
		ldc         *v1alpha1.LocalDiskClaim
		wantBound   bool
	}{
		{
			description: "Claim by one disk",
			preAction:   createResource,
			postAction:  cleanResource,
			ld:          GenFakeLocalDiskObject(),
			ldc:         GenFakeLocalDiskClaimObject(),
			wantBound:   true,
		},
		{
			description: "Claim by one disk don't satisfy the requirement",
			preAction:   createResource,
			postAction:  cleanResource,
			ld:          GenFakeLocalDiskObject(),
			ldc:         GenFakeLocalDiskClaimObject(),
			wantBound:   true,
		},
	}

	for _, testcase := range testCases {
		t.Run(testcase.description, func(t *testing.T) {
			defer testcase.postAction(cli, testcase.ld)
			err := testcase.preAction(cli, testcase.ld)
			if err != nil {
				t.Errorf("failed to create localdisk %v", err)
			}

			handler.For(testcase.ld)
			if err = handler.BoundTo(testcase.ldc); err != nil {
				t.Errorf("failed to bound localdiskclaim")
			}

			// refresh
			err = cli.Get(context.Background(), client.ObjectKey{Namespace: testcase.ld.GetNamespace(),
				Name: testcase.ld.GetName()}, testcase.ld)
			if err != nil {
				t.Errorf("failed to refresh localdisk")
				return
			}

			if testcase.wantBound && testcase.ld.Spec.ClaimRef != nil &&
				testcase.ld.Spec.ClaimRef.Name != testcase.ldc.Name {
				t.Errorf("Expect localdisk state is Bound but actual got %s", testcase.ld.Status.State)
			}
		})
	}
}

func TestHandler_GetLocalDiskFuncs(t *testing.T) {
	cli, _ := CreateFakeClient()
	handler := NewLocalDiskHandler(cli, fakeRecorder)

	handler.For(GenFakeLocalDiskObject())
	err := cli.Create(context.Background(), handler.localDisk)
	if err != nil {
		t.Errorf("Fail to create LocalDisk")
		return
	}

	// handler.GetLocalDisk
	ld1, err := handler.GetLocalDisk(types.NamespacedName{
		Name:      fakeNodename + fakedevPath,
		Namespace: fakeNamespace,
	})
	if err != nil || ld1.Name != fakeNodename+fakedevPath {
		t.Errorf("Fail to get LocalDisk")
	}

	_, err = handler.GetLocalDisk(types.NamespacedName{
		Name:      "invalid",
		Namespace: fakeNamespace,
	})
	if err == nil {
		t.Errorf("Fail to get invalid LocalDisk")
	}

	// handler.GetLocalDiskWithLabels
	labels := map[string]string{fakeLabelKey: fakeLabelValue}
	handler.SetupLabel(labels)
	if cli.Update(context.TODO(), handler.localDisk) != nil {
		t.Errorf("Fail to update localDisk")
	}
	ldList1, err := handler.GetLocalDiskWithLabels(labels)
	if err != nil || len(ldList1.Items) == 0 {
		t.Errorf("Fail to get LocalDiskList")
	}
	handler.RemoveLabel(labels)

	// handler.ListLocalDisk
	ldList2, err := handler.ListLocalDisk()
	if err != nil || len(ldList2.Items) == 0 {
		t.Errorf("Fail to list LocalDisk")
	}

	// handler.ListNodeLocalDisk
	ldList3, err := handler.ListNodeLocalDisk(fakeNodename)
	if err != nil || len(ldList3.Items) == 0 {
		t.Errorf("Fail to list node LocalDisk")
	}

	// handler.Unclaimed
	if handler.UnClaimed() == false {
		t.Errorf("Fail to get claimed")
	}

	// handler.SetupStatus
	handler.SetupStatus(v1alpha1.LocalDiskAvailable)
	if handler.UpdateStatus() != nil {
		t.Errorf("Fail to update status")
	}

	// handler.PatchDiskSpec
	oldLocalDisk := handler.localDisk.DeepCopy()
	handler.SetPartition(true)
	handler.SetOwner("testOwner")
	handler.ReserveDisk()
	if handler.PatchDiskSpec(client.MergeFrom(oldLocalDisk)) != nil {
		t.Errorf("Fail to patch LocalDisk spec")
	}

	if handler.ClaimRef() != nil {
		t.Errorf("Claim ref should be nil")
	}

	if handler.PatchDiskOwner("testOwner2") != nil {
		t.Errorf("Fail to patch disk owner")
	}

	if handler.IsEmpty() == true {
		t.Errorf("LocalDisk partision should be empty")
	}

	if handler.FilterDisk(GenFakeLocalDiskClaimObject()) != false {
		t.Errorf("Filter disk should be false")
	}
}

// CreateFakeClient Create localDisk and LocalDiskNode resource
func CreateFakeClient() (client.Client, *runtime.Scheme) {
	s := scheme.Scheme
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, &v1alpha1.LocalDisk{})
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, &v1alpha1.LocalDiskList{})
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, &v1alpha1.LocalDiskNode{})
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, &v1alpha1.LocalDiskNodeList{})
	return fake.NewClientBuilder().WithScheme(s).WithObjects(&v1alpha1.LocalDisk{}, &v1alpha1.LocalDiskNode{}).Build(), s
}

// GenFakeLocalDiskObject Create disk
// By default, disk can be claimed by the sample claim
func GenFakeLocalDiskObject() *v1alpha1.LocalDisk {
	ld := &v1alpha1.LocalDisk{}

	TypeMeta := metav1.TypeMeta{
		Kind:       localDiskKind,
		APIVersion: apiversion,
	}

	ObjectMata := metav1.ObjectMeta{
		Name:              fakeNodename + fakedevPath,
		Namespace:         fakeNamespace,
		UID:               types.UID(localDiskUID),
		CreationTimestamp: metav1.Time{Time: time.Now()},
	}

	Spec := v1alpha1.LocalDiskSpec{
		NodeName:     fakeNodename,
		DevicePath:   fakedevPath,
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

	Status := v1alpha1.LocalDiskStatus{State: v1alpha1.LocalDiskAvailable}

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
