package localdiskclaim

import (
	"context"
	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"strings"
	"testing"
	"time"
)

var (
	fakeLocalDiskClaimName       = "local-disk-claim-example"
	fakeLocalDiskClaimUID        = "local-disk-claim-example-uid"
	fakeLocalDiskName            = "local-disk-example"
	localDiskUID                 = "local-disk-example-uid"
	fakeNamespace                = "local-disk-manager-test"
	fakeNodename                 = "10-6-118-10"
	diskTypeHDD                  = "HDD"
	diskTypeSSD                  = "SSD"
	fakedevPath                  = "/dev/fake-sda"
	devType                      = "disk"
	vendorVMware                 = "VMware"
	proSCSI                      = "scsi"
	apiversion                   = "hwameistor.io/v1alpha1"
	localDiskKind                = "LocalDisk"
	localDiskNodeKind            = "LocalDiskNode"
	localDiskClaimKind           = "LocalDiskClaim"
	cap100G                int64 = 100 * 1024 * 1024 * 1024
	cap10G                 int64 = 10 * 1024 * 1024 * 1024
	fakeRecorder                 = record.NewFakeRecorder(100)
)

func TestLocalDiskClaimHandler_AssignDisk(t *testing.T) {
	cli, _ := CreateFakeClient()

	claimHandler := NewLocalDiskClaimHandler(cli, fakeRecorder)

	testCases := []struct {
		Description string
		DiskClaim   *v1alpha1.LocalDiskClaim
		FreeDisk    *v1alpha1.LocalDisk
		WantAssign  bool

		setProperty        func(diskClaim *v1alpha1.LocalDiskClaim)
		createNewFreeDisk  func(cli client.Client, disk *v1alpha1.LocalDisk) error
		deleteDisk         func(cli client.Client, disk *v1alpha1.LocalDisk) error
		createNewDiskClaim func(cli client.Client, disk *v1alpha1.LocalDiskClaim) error
		deleteDiskClaim    func(cli client.Client, disk *v1alpha1.LocalDiskClaim) error
	}{
		{
			Description: "Should return no error, Disk satisfied",
			DiskClaim:   GenFakeLocalDiskClaimObject(),
			FreeDisk:    GenFakeLocalDiskObject(),
			WantAssign:  true,

			setProperty: func(diskClaim *v1alpha1.LocalDiskClaim) {
				// By default, FakeLocalDisk matches the FakeLocalDiskClaim.
				// Do nothing here.
				return
			},
			createNewFreeDisk: createLocalDisk,
			deleteDisk:        deleteLocalDisk,

			createNewDiskClaim: createLocalDiskClaim,
			deleteDiskClaim:    deleteLocalDiskClaim,
		},
		{
			Description: "Should return error, Disk don't satisfied for DiskCapacity",
			DiskClaim:   GenFakeLocalDiskClaimObject(),
			FreeDisk:    GenFakeLocalDiskObject(),
			WantAssign:  false,

			setProperty: func(diskClaim *v1alpha1.LocalDiskClaim) {
				diskClaim.Spec.Description.Capacity = cap100G
				return
			},
			createNewFreeDisk: createLocalDisk,
			deleteDisk:        deleteLocalDisk,

			createNewDiskClaim: createLocalDiskClaim,
			deleteDiskClaim:    deleteLocalDiskClaim,
		},
		{
			Description: "Should return false, Disk don't satisfied for DiskType",
			DiskClaim:   GenFakeLocalDiskClaimObject(),
			FreeDisk:    GenFakeLocalDiskObject(),
			WantAssign:  false,

			setProperty: func(diskClaim *v1alpha1.LocalDiskClaim) {
				diskClaim.Spec.Description.DiskType = diskTypeSSD
				return
			},
			createNewFreeDisk: createLocalDisk,
			deleteDisk:        deleteLocalDisk,

			createNewDiskClaim: createLocalDiskClaim,
			deleteDiskClaim:    deleteLocalDiskClaim,
		},
		{
			Description: "Should return false, Disk don't satisfied for DiskNode",
			DiskClaim:   GenFakeLocalDiskClaimObject(),
			FreeDisk:    GenFakeLocalDiskObject(),
			WantAssign:  false,

			setProperty: func(diskClaim *v1alpha1.LocalDiskClaim) {
				diskClaim.Spec.NodeName = "test"
				return
			},
			createNewFreeDisk: createLocalDisk,
			deleteDisk:        deleteLocalDisk,

			createNewDiskClaim: createLocalDiskClaim,
			deleteDiskClaim:    deleteLocalDiskClaim,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Description, func(t *testing.T) {
			// Update DiskClaim according to testCase request first
			testCase.setProperty(testCase.DiskClaim)
			if err := testCase.createNewDiskClaim(cli, testCase.DiskClaim); err != nil {
				t.Fatalf("Failed to create DiskClaim %v", err)
			}

			claimHandler.For(testCase.DiskClaim)

			// Create new free disk
			if err := testCase.createNewFreeDisk(cli, testCase.FreeDisk); err != nil {
				t.Fatalf("Failed to create free disk %v", err)
			}

			// Assign free disk for claim request
			if err := claimHandler.AssignFreeDisk(); (err == nil) != testCase.WantAssign {
				t.Fatalf("Want assign: %v, got assign: %v", testCase.WantAssign, err == nil)
			}

			// Delete localDisk
			if err := testCase.deleteDisk(cli, testCase.FreeDisk); err != nil {
				t.Fatalf("Failed to delete Disk %v", err)
			}

			// Delete LocalDiskClaim
			if err := testCase.deleteDiskClaim(cli, testCase.DiskClaim); err != nil {
				t.Fatalf("Failed to delete DiskClaim %v", err)
			}
		})

	}

}

func TestLocalDiskClaimHandler_ListUnboundDiskClaim(t *testing.T) {
	cli, _ := CreateFakeClient()

	claimHandler := NewLocalDiskClaimHandler(cli, fakeRecorder)

	testCases := []struct {
		Name        string
		Description string
		DiskClaim   *v1alpha1.LocalDiskClaim
		WantResult  int

		setProperty        func(diskClaim *v1alpha1.LocalDiskClaim)
		createNewDiskClaim func(cli client.Client, disk *v1alpha1.LocalDiskClaim) error
		deleteDiskClaim    func(cli client.Client, disk *v1alpha1.LocalDiskClaim) error
	}{
		{
			Name:        "OneUnboundDiskClaim",
			Description: "Should return 1 Unbound disk claim",
			WantResult:  1,
			DiskClaim:   GenFakeLocalDiskClaimObject(),

			setProperty: func(diskClaim *v1alpha1.LocalDiskClaim) {
				// By default, FakeLocalDiskClaim's status is empty
				// Do nothing here
				return
			},
			createNewDiskClaim: createLocalDiskClaim,
			deleteDiskClaim:    deleteLocalDiskClaim,
		},
		// *** NOTE: fakeClient doesn't support option.FieldSelector ***
		// so we can't use `status.status == ""` as a filter here to filter Unbound DiskClaims.

		//{
		//	Name:        "NoUnboundDiskClaim",
		//	Description: "Should return 0 Unbound disk claim",
		//	WantResult:  0,
		//	DiskClaim:   GenFakeLocalDiskClaimObject(),
		//
		//	setProperty: func(diskClaim *v1alpha1.LocalDiskClaim) {
		//		diskClaim.Status.Status = v1alpha1.LocalDiskClaimStatusBound
		//		return
		//	},
		//	createNewDiskClaim: createLocalDiskClaim,
		//	deleteDiskClaim:    deleteLocalDiskClaim,
		//},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {

			// Update DiskClaim first
			testCase.setProperty(testCase.DiskClaim)

			// Create LocalDiskClaim
			if err := testCase.createNewDiskClaim(cli, testCase.DiskClaim); err != nil {
				t.Fatalf("Failed to create LocalDiskClaim %v", err)
			}

			// List Unbound DiskClaim
			claimList, err := claimHandler.ListUnboundLocalDiskClaim()
			if err != nil {
				t.Fatalf("Failed to list DiskClaims %v", err)
			}

			if len(claimList.Items) != testCase.WantResult {
				t.Fatalf("Want Unbound DiskClaim %d, But got %d", testCase.WantResult, len(claimList.Items))
			}

			if err := testCase.deleteDiskClaim(cli, testCase.DiskClaim); err != nil {
				t.Fatalf("Failed to delete diskclaim %v", err)
			}
		})
	}
}

// CreateFakeClient Create localDisk and LocalDiskNode resource
func CreateFakeClient() (client.Client, *runtime.Scheme) {
	s := scheme.Scheme
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, &v1alpha1.LocalDisk{})
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, &v1alpha1.LocalDiskList{})
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, &v1alpha1.LocalDiskNode{})
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, &v1alpha1.LocalDiskNodeList{})
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, &v1alpha1.LocalDiskClaim{})
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, &v1alpha1.LocalDiskClaimList{})
	return fake.NewFakeClientWithScheme(s, &v1alpha1.LocalDisk{}, &v1alpha1.LocalDiskNode{}), s
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
		Name:              fakeNodename + strings.Replace(fakedevPath, "/", "-", -1),
		Namespace:         fakeNamespace,
		UID:               types.UID(localDiskUID),
		CreationTimestamp: metav1.Time{Time: time.Now()},
	}

	Spec := v1alpha1.LocalDiskSpec{
		NodeName:     fakeNodename,
		DevicePath:   fakedevPath,
		Capacity:     cap10G,
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
			Capacity: cap10G,
		},
	}

	ldc.ObjectMeta = ObjectMata
	ldc.TypeMeta = TypeMeta
	ldc.Spec = Spec
	ldc.Status.Status = v1alpha1.DiskClaimStatusEmpty
	return ldc
}

func createLocalDisk(cli client.Client, disk *v1alpha1.LocalDisk) error {
	return cli.Create(context.Background(), disk)
}

func deleteLocalDisk(cli client.Client, disk *v1alpha1.LocalDisk) error {
	return cli.Delete(context.Background(), disk)
}

func createLocalDiskClaim(cli client.Client, diskClaim *v1alpha1.LocalDiskClaim) error {
	return cli.Create(context.Background(), diskClaim)
}

func deleteLocalDiskClaim(cli client.Client, diskClaim *v1alpha1.LocalDiskClaim) error {
	return cli.Delete(context.Background(), diskClaim)
}
