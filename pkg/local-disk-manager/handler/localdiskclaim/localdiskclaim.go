package localdiskclaim

import (
	"context"
	"fmt"

	v1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	localdisk2 "github.com/hwameistor/hwameistor/pkg/local-disk-manager/handler/localdisk"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/tools/reference"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// LocalDiskClaimHandler
type LocalDiskClaimHandler struct {
	client.Client
	record.EventRecorder
	ldc v1alpha1.LocalDiskClaim
}

// NewLocalDiskClaimHandler
func NewLocalDiskClaimHandler(client client.Client, recorder record.EventRecorder) *LocalDiskClaimHandler {
	return &LocalDiskClaimHandler{
		Client:        client,
		EventRecorder: recorder,
	}
}

// ListLocalDiskClaim
func (ldcHandler *LocalDiskClaimHandler) ListLocalDiskClaim() (*v1alpha1.LocalDiskClaimList, error) {
	list := &v1alpha1.LocalDiskClaimList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "LocalDiskClaim",
			APIVersion: "v1alpha1",
		},
	}

	err := ldcHandler.List(context.TODO(), list)
	return list, err
}

// GetLocalDiskClaim
func (ldcHandler *LocalDiskClaimHandler) GetLocalDiskClaim(key client.ObjectKey) (*v1alpha1.LocalDiskClaim, error) {
	ldc := &v1alpha1.LocalDiskClaim{}
	if err := ldcHandler.Get(context.Background(), key, ldc); err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return ldc, nil
}

// ListLocalDiskClaim
func (ldcHandler *LocalDiskClaimHandler) ListUnboundLocalDiskClaim() (*v1alpha1.LocalDiskClaimList, error) {
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

// For
func (ldcHandler *LocalDiskClaimHandler) For(ldc v1alpha1.LocalDiskClaim) *LocalDiskClaimHandler {
	ldcHandler.ldc = ldc
	return ldcHandler
}

// AssignFreeDisk
func (ldcHandler *LocalDiskClaimHandler) AssignFreeDisk() error {
	ldHandler := localdisk2.NewLocalDiskHandler(ldcHandler.Client, ldcHandler.EventRecorder)
	ldc := *ldcHandler.ldc.DeepCopy()
	ldList, err := ldHandler.ListLocalDisk()
	if err != nil {
		return err
	}

	var assignedDisks []string
	for _, ld := range ldList.Items {
		ldHandler.For(&ld)
		if !ldHandler.FilterDisk(ldc) {
			continue
		}
		if err = ldHandler.BoundTo(ldc); err != nil {
			return err
		}
		if err = ldcHandler.BoundWith(ld); err != nil {
			return err
		}

		assignedDisks = append(assignedDisks, ld.GetName())
	}

	if len(assignedDisks) == 0 {
		log.Infof("There is no available disk assigned to %v", ldc.GetName())
		return fmt.Errorf("there is no available disk assigned to %v", ldc.GetName())
	}

	log.Infof("Disk %v has been assigned to %v", assignedDisks, ldc.GetName())
	return ldcHandler.UpdateClaimStatus()
}

// Bounded
func (ldcHandler *LocalDiskClaimHandler) UpdateSpec() error {
	return ldcHandler.Update(context.Background(), &ldcHandler.ldc)
}

// Bounded
func (ldcHandler *LocalDiskClaimHandler) Bounded() bool {
	return ldcHandler.ldc.Status.Status == v1alpha1.LocalDiskClaimStatusBound
}

// DiskRefs
func (ldcHandler *LocalDiskClaimHandler) DiskRefs() []*v1.ObjectReference {
	return ldcHandler.ldc.Spec.DiskRefs
}

// DiskRefs
func (ldcHandler *LocalDiskClaimHandler) Phase() v1alpha1.DiskClaimStatus {
	return ldcHandler.ldc.Status.Status
}

// BoundWith
func (ldcHandler *LocalDiskClaimHandler) BoundWith(ld v1alpha1.LocalDisk) error {
	ldRef, err := reference.GetReference(nil, &ld)
	if err != nil {
		return err
	}

	// check if this disk has already bound
	needBound := true
	for _, boundDisk := range ldcHandler.ldc.Spec.DiskRefs {
		if boundDisk.Name == ld.GetName() {
			needBound = false
		}
	}
	if needBound {
		ldcHandler.ldc.Spec.DiskRefs = append(ldcHandler.ldc.Spec.DiskRefs, ldRef)
	}

	ldcHandler.ldc.Status.Status = v1alpha1.LocalDiskClaimStatusBound

	ldcHandler.EventRecorder.Eventf(&ldcHandler.ldc, v1.EventTypeNormal, "BoundLocalDisk", "Bound disk %v", ld.Name)
	return nil
}

// SetupClaimStatus
func (ldcHandler *LocalDiskClaimHandler) SetupClaimStatus(status v1alpha1.DiskClaimStatus) {
	ldcHandler.ldc.Status.Status = status
}

// UpdateStatus
func (ldcHandler *LocalDiskClaimHandler) UpdateClaimStatus() error {
	return ldcHandler.Update(context.Background(), &ldcHandler.ldc)
}

// Refresh
func (ldcHandler *LocalDiskClaimHandler) Refresh() error {
	ldc, err := ldcHandler.GetLocalDiskClaim(client.ObjectKey{Name: ldcHandler.ldc.GetName(), Namespace: ldcHandler.ldc.GetNamespace()})
	if err != nil {
		return err
	}
	ldcHandler.For(*ldc.DeepCopy())
	return nil
}
