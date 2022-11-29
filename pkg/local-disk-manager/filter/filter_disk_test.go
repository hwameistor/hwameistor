package filter

import (
	"testing"
	"time"

	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils/sys"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
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
	devPath                      = "/dev/fake-sda"
	devType                      = "disk"
	vendorVMware                 = "VMware"
	proSCSI                      = "scsi"
	apiversion                   = "hwameistor.io/v1alpha1"
	localDiskKind                = "localDisk"
	localDiskClaimKind           = "LocalDiskClaim"
	cap100G                int64 = 100 * 1024 * 1024 * 1024
	cap10G                 int64 = 10 * 1024 * 1024 * 1024
	fakeRecorder                 = record.NewFakeRecorder(100)
)

func TestLocalDiskFilter(t *testing.T) {

	testCases := []struct {
		Description      string
		WantFilterResult bool

		WantCapacity        int64
		WantDiskType        string
		WantDiskUnclaimed   bool
		WantDiskNode        string
		WantDevType         string
		WantDiskNoPartition bool
		WantReserved        bool
		WantBoundWithClaim  string

		disk        *v1alpha1.LocalDisk
		setProperty func(disk *v1alpha1.LocalDisk)
	}{
		{
			Description:      "Should return true, Has Sufficient Capacity",
			WantFilterResult: true,
			WantCapacity:     cap10G,
			disk:             GenFakeLocalDiskObject(),
			setProperty: func(disk *v1alpha1.LocalDisk) {
				disk.Spec.Capacity = cap100G
			},
		},
		{
			Description:      "Should return false, Has InSufficient Capacity",
			WantFilterResult: false,
			WantCapacity:     cap100G,
			disk:             GenFakeLocalDiskObject(),
			setProperty: func(disk *v1alpha1.LocalDisk) {
				disk.Spec.Capacity = cap10G
			},
		},
		{
			Description:      "Should return true, Has Correct DiskType",
			WantFilterResult: true,
			WantDiskType:     diskTypeHDD,
			disk:             GenFakeLocalDiskObject(),
			setProperty: func(disk *v1alpha1.LocalDisk) {
				disk.Spec.DiskAttributes.Type = diskTypeHDD
			},
		},
		{
			Description:      "Should return false, Has InCorrect DiskType",
			WantFilterResult: false,
			WantDiskType:     diskTypeSSD,
			disk:             GenFakeLocalDiskObject(),
			setProperty: func(disk *v1alpha1.LocalDisk) {
				disk.Spec.DiskAttributes.Type = diskTypeHDD
			},
		},
		{
			Description:       "Should return true, Has Available Disk",
			WantFilterResult:  true,
			WantDiskUnclaimed: true,
			disk:              GenFakeLocalDiskObject(),
			setProperty: func(disk *v1alpha1.LocalDisk) {
				disk.Status.State = v1alpha1.LocalDiskAvailable
			},
		},
		{
			Description:       "Should return false, Has Bound Disk",
			WantFilterResult:  false,
			WantDiskUnclaimed: true,
			disk:              GenFakeLocalDiskObject(),
			setProperty: func(disk *v1alpha1.LocalDisk) {
				disk.Status.State = v1alpha1.LocalDiskBound
			},
		},
		{
			Description:      "Should return true, Has Correct DiskNode",
			WantFilterResult: true,
			WantDiskNode:     fakeNodename,
			disk:             GenFakeLocalDiskObject(),
			setProperty: func(disk *v1alpha1.LocalDisk) {
				disk.Spec.NodeName = fakeNodename
			},
		},
		{
			Description:      "Should return false, Has InCorrect DiskNode",
			WantFilterResult: false,
			WantDiskNode:     fakeNodename,
			disk:             GenFakeLocalDiskObject(),
			setProperty: func(disk *v1alpha1.LocalDisk) {
				disk.Spec.NodeName = "test-node"
			},
		},
		{
			Description:      "Should return true, Has Correct DevType",
			WantFilterResult: true,
			WantDevType:      sys.BlockDeviceTypeDisk,
			disk:             GenFakeLocalDiskObject(),
			setProperty: func(disk *v1alpha1.LocalDisk) {
				disk.Spec.DiskAttributes.DevType = sys.BlockDeviceTypeDisk
			},
		},
		{
			Description:      "Should return false, Has InCorrect DevType",
			WantFilterResult: false,
			WantDevType:      sys.BlockDeviceTypeDisk,
			disk:             GenFakeLocalDiskObject(),
			setProperty: func(disk *v1alpha1.LocalDisk) {
				disk.Spec.DiskAttributes.DevType = sys.BlockDeviceType
			},
		},
		{
			Description:         "Should return true, Has NoPartition Disk",
			WantFilterResult:    true,
			WantDiskNoPartition: true,
			disk:                GenFakeLocalDiskObject(),
			setProperty: func(disk *v1alpha1.LocalDisk) {
				disk.Spec.HasPartition = false
			},
		},
		{
			Description:         "Should return false, Has Partition Disk",
			WantFilterResult:    false,
			WantDiskNoPartition: true,
			disk:                GenFakeLocalDiskObject(),
			setProperty: func(disk *v1alpha1.LocalDisk) {
				disk.Spec.PartitionInfo = append(disk.Spec.PartitionInfo, v1alpha1.PartitionInfo{
					Path:          "",
					HasFileSystem: false,
					FileSystem:    v1alpha1.FileSystemInfo{},
				})
				disk.Spec.HasPartition = true
			},
		},
		{
			Description:         "Should return true, Has Reserved",
			WantFilterResult:    false,
			WantReserved:        true,
			disk:                GenFakeLocalDiskObject(),
			setProperty: func(disk *v1alpha1.LocalDisk) {
				disk.Spec.Reserved = true
			},
		},
		{
			Description:         "Should return false, Has Not Reserved",
			WantFilterResult:    true,
			WantReserved:        false,
			disk:                GenFakeLocalDiskObject(),
			setProperty: func(disk *v1alpha1.LocalDisk) {
				disk.Spec.Reserved = false
			},
		},
		{
			Description:      "Should return true, Has Correct ClaimRef Name",
			WantFilterResult: true,
			WantBoundWithClaim:  "ClaimFoo",
			disk:             GenFakeLocalDiskObject(),
			setProperty: func(disk *v1alpha1.LocalDisk) {
				disk.Spec.ClaimRef = &v1.ObjectReference{
					Name: "ClaimFoo",
				}
			},
		},
		{
			Description:      "Should return false, Has InCorrect ClaimRef Name",
			WantFilterResult: true,
			WantBoundWithClaim: "ClaimFoo",
			disk:             GenFakeLocalDiskObject(),
			setProperty: func(disk *v1alpha1.LocalDisk) {
				disk.Spec.ClaimRef = &v1.ObjectReference{
					Name: "ClaimBar",
				}
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Description, func(t *testing.T) {
			// set test property
			testCase.setProperty(testCase.disk)

			filter := NewLocalDiskFilter(testCase.disk)
			filter.Init()

			if testCase.WantCapacity > 0 {
				filter.Capacity(testCase.WantCapacity)
			}

			if testCase.WantDiskNode != "" {
				filter.NodeMatch(testCase.WantDiskNode)
			}

			if testCase.WantDiskNoPartition {
				filter.NoPartition()
			}

			if testCase.WantDevType != "" {
				filter.DevType()
			}

			if testCase.WantDiskUnclaimed {
				filter.Available()
			}

			if testCase.WantDiskType != "" {
				filter.DiskType(testCase.WantDiskType)
			}

			if testCase.WantReserved {
				filter.HasNotReserved()
			}

			if testCase.WantBoundWithClaim != "" {
				filter.HasBoundWith(testCase.WantBoundWithClaim)
			}

			if filter.GetTotalResult() != testCase.WantFilterResult {
				t.Fatalf("Filter disk fail,want result: %v, got: %v", testCase.WantFilterResult, !testCase.WantFilterResult)
			}
		})
	}
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

	Status := v1alpha1.LocalDiskStatus{State: v1alpha1.LocalDiskAvailable}

	ld.TypeMeta = TypeMeta
	ld.ObjectMeta = ObjectMata
	ld.Spec = Spec
	ld.Status = Status
	return ld
}
