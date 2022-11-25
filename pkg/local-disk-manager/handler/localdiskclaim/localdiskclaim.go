package localdiskclaim

import (
	"context"
	"fmt"
	v1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	diskHandler "github.com/hwameistor/hwameistor/pkg/local-disk-manager/handler/localdisk"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/tools/reference"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Handler struct {
	client.Client
	record.EventRecorder
	diskClaim *v1alpha1.LocalDiskClaim
}

func NewLocalDiskClaimHandler(client client.Client, recorder record.EventRecorder) *Handler {
	return &Handler{
		Client:        client,
		EventRecorder: recorder,
	}
}

func (ldcHandler *Handler) ListLocalDiskClaim() (*v1alpha1.LocalDiskClaimList, error) {
	list := &v1alpha1.LocalDiskClaimList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "LocalDiskClaim",
			APIVersion: "v1alpha1",
		},
	}

	err := ldcHandler.List(context.TODO(), list)
	return list, err
}

func (ldcHandler *Handler) GetLocalDiskClaim(key client.ObjectKey) (*v1alpha1.LocalDiskClaim, error) {
	ldc := &v1alpha1.LocalDiskClaim{}
	if err := ldcHandler.Get(context.Background(), key, ldc); err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return ldc, nil
}

func (ldcHandler *Handler) ListUnboundLocalDiskClaim() (*v1alpha1.LocalDiskClaimList, error) {
	list := &v1alpha1.LocalDiskClaimList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "LocalDiskClaim",
			APIVersion: "v1alpha1",
		},
	}

	// NOTE: runtime selector is only support equal
	unboundSelector := fields.OneTermEqualSelector("status.status", "")

	err := ldcHandler.List(context.TODO(), list, &client.ListOptions{FieldSelector: unboundSelector})
	return list, err
}

func (ldcHandler *Handler) For(ldc *v1alpha1.LocalDiskClaim) *Handler {
	ldcHandler.diskClaim = ldc
	return ldcHandler
}

func (ldcHandler *Handler) AssignFreeDisk() error {
	localDiskHandler := diskHandler.NewLocalDiskHandler(ldcHandler.Client, ldcHandler.EventRecorder)
	diskClaim := ldcHandler.diskClaim.DeepCopy()
	diskList, err := localDiskHandler.ListLocalDisk()
	if err != nil {
		return err
	}

	var assignedDisks, finalAssignedDisks []string
	for _, disk := range diskClaim.Spec.DiskRefs {
		assignedDisks = append(assignedDisks, disk.Name)
	}

	// Find suitable disks
	for _, disk := range diskList.Items {
		localDiskHandler.For(&disk)

		// Disks that already assigned to this diskClaim will also be filtered in
		if !localDiskHandler.FilterDisk(diskClaim) {
			continue
		}

		// Update disk.spec to indicate the disk has been Assigned to this diskClaim
		if err = localDiskHandler.BoundTo(diskClaim); err != nil {
			return err
		}

		finalAssignedDisks = append(finalAssignedDisks, disk.GetName())
	}

	newAssignedDisks, foundNewDisks := utils.FoundNewStringElems(assignedDisks, finalAssignedDisks)
	if !foundNewDisks {
		log.Infof("There is no available disk assigned to %v", diskClaim.GetName())
		return fmt.Errorf("there is no available disk assigned to %v", diskClaim.GetName())
	}

	log.Infof("Disk %v has been assigned to %v", newAssignedDisks, diskClaim.GetName())
	return nil
}

// UpdateBoundDiskRef update all disk bounded by the diskClaim to claim.spec.disks
func (ldcHandler *Handler) UpdateBoundDiskRef() error {
	diskList, err := diskHandler.
		NewLocalDiskHandler(ldcHandler.Client, ldcHandler.EventRecorder).
		ListLocalDisk()
	if err != nil {
		return err
	}

	for _, disk := range diskList.Items {
		if disk.Spec.ClaimRef != nil &&
			disk.Spec.ClaimRef.Name == ldcHandler.diskClaim.GetName() {
			ldcHandler.AppendDiskRef(&disk)
		}
	}

	return ldcHandler.UpdateClaimSpec()
}

func (ldcHandler *Handler) Bounded() bool {
	return ldcHandler.diskClaim.Status.Status == v1alpha1.LocalDiskClaimStatusBound
}

func (ldcHandler *Handler) DiskRefs() []*v1.ObjectReference {
	return ldcHandler.diskClaim.Spec.DiskRefs
}

func (ldcHandler *Handler) Phase() v1alpha1.DiskClaimStatus {
	return ldcHandler.diskClaim.Status.Status
}

func (ldcHandler *Handler) AppendDiskRef(ld *v1alpha1.LocalDisk) {
	ldRef, _ := reference.GetReference(nil, ld)

	// check if this disk has already bound
	needBound := true
	for _, boundDisk := range ldcHandler.diskClaim.Spec.DiskRefs {
		if boundDisk.Name == ld.GetName() {
			needBound = false
		}
	}

	if needBound {
		ldcHandler.diskClaim.Spec.DiskRefs = append(ldcHandler.diskClaim.Spec.DiskRefs, ldRef)
	}
}

func (ldcHandler *Handler) SetupClaimStatus(status v1alpha1.DiskClaimStatus) {
	ldcHandler.diskClaim.Status.Status = status
}

func (ldcHandler *Handler) UpdateClaimStatus() error {
	return ldcHandler.Status().Update(context.Background(), ldcHandler.diskClaim)
}

func (ldcHandler *Handler) UpdateClaimSpec() error {
	return ldcHandler.Update(context.Background(), ldcHandler.diskClaim)
}

func (ldcHandler *Handler) Refresh() error {
	ldc, err := ldcHandler.GetLocalDiskClaim(client.ObjectKey{Name: ldcHandler.diskClaim.GetName(), Namespace: ldcHandler.diskClaim.GetNamespace()})
	if err != nil {
		return err
	}
	ldcHandler.For(ldc.DeepCopy())
	return nil
}
