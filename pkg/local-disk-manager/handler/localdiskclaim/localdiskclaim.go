package localdiskclaim

import (
	"context"
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/tools/reference"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	diskHandler "github.com/hwameistor/hwameistor/pkg/local-disk-manager/handler/localdisk"
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
	diskList, err := localDiskHandler.ListNodeLocalDisk(diskClaim.Spec.NodeName)
	localDiskFailedMessages := make(map[string][]string)
	if err != nil {
		return err
	}

	var finalAssignedDisks []string

	// Find suitable disks
	for _, disk := range diskList.Items {
		localDiskHandler.For(&disk)
		// Disks that already assigned to this diskClaim will also be filtered in
		if !localDiskHandler.FilterDisk(diskClaim) {
			// append the disk name to failed message map when localdisk is assigned failed
			filterFailMessages := localDiskHandler.GetFilterFailMessages()
			for reason := range filterFailMessages {
				if _, ok := localDiskFailedMessages[reason]; !ok {
					localDiskFailedMessages[reason] = []string{}
				}
				localDiskFailedMessages[reason] = append(localDiskFailedMessages[reason], disk.Name)
			}
			continue
		}

		// Update disk.spec to indicate the disk has been Assigned to this diskClaim
		if err = localDiskHandler.BoundTo(diskClaim); err != nil {
			return err
		}

		finalAssignedDisks = append(finalAssignedDisks, disk.GetName())
	}

	// NOTE: Once found disk(s) already bound to this claim, return true directly
	if len(finalAssignedDisks) <= 0 {
		var fullFailMessages []string
		for reason, diskNames := range localDiskFailedMessages {
			fullMsg := strings.Join(diskNames, ",") + " are " + reason
			fullFailMessages = append(fullFailMessages, fullMsg)
		}
		ldcHandler.EventRecorder.Event(diskClaim, v1.EventTypeWarning, v1alpha1.LocalDiskClaimEventReasonAssignFail, strings.Join(fullFailMessages, ";"))

		log.Infof("There is no available disk assigned to %v", diskClaim.GetName())
		return fmt.Errorf("there is no available disk assigned to %v", diskClaim.GetName())
	}

	log.Infof("Disk %v has been assigned to %v", finalAssignedDisks, diskClaim.GetName())
	return nil
}

// PatchBoundDiskRef update all disk bounded by the diskClaim to claim.spec.disks
func (ldcHandler *Handler) PatchBoundDiskRef() error {
	time.Sleep(time.Second)
	logger := log.WithFields(log.Fields{"LocalDiskClaim": ldcHandler.diskClaim.GetName()})

	diskList, err := diskHandler.
		NewLocalDiskHandler(ldcHandler.Client, ldcHandler.EventRecorder).
		ListNodeLocalDisk(ldcHandler.diskClaim.Spec.NodeName)
	if err != nil {
		return err
	}

	oldDiskClaim := ldcHandler.diskClaim.DeepCopy()
	logger.Infof("Found %d localdisk(s) in cluster", len(diskList.Items))
	for _, disk := range diskList.Items {
		if disk.Spec.ClaimRef != nil &&
			// Since the claim can be applied repeatedly with a same name, thus compare UID here
			disk.Spec.ClaimRef.UID == ldcHandler.diskClaim.UID {
			ldcHandler.AppendDiskRef(&disk)
		}
	}

	logger.Infof("Found %d localdisk(s) bounded by claim %v",
		len(ldcHandler.diskClaim.Spec.DiskRefs), ldcHandler.diskClaim.GetName())
	for _, disk := range ldcHandler.diskClaim.Spec.DiskRefs {
		logger.Infof("Bounded localdisk: %s", disk.Name)
	}

	return ldcHandler.PatchClaimSpec(client.MergeFrom(oldDiskClaim))
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

func (ldcHandler *Handler) UpdateClaimStatusToBound() error {
	var err error

	// Check if any disk(s) have already bounded to the claim
	// Return error if not found any bound disk(s)
	if len(ldcHandler.DiskRefs()) <= 0 {
		err = fmt.Errorf("no disks bounded by the claim, need to reconcile the diskclaim %v again",
			ldcHandler.diskClaim.GetName())
		return err
	}

	ldcHandler.EventRecorder.Eventf(ldcHandler.diskClaim, v1.EventTypeNormal, v1alpha1.LocalDiskClaimEventReasonExtend,
		"Success to extend for localdiskclaim %v", ldcHandler.diskClaim.GetName())

	ldcHandler.SetupClaimStatus(v1alpha1.LocalDiskClaimStatusBound)
	return ldcHandler.UpdateClaimStatus()
}

func (ldcHandler *Handler) UpdateClaimSpec() error {
	return ldcHandler.Update(context.Background(), ldcHandler.diskClaim)
}

func (ldcHandler *Handler) PatchClaimSpec(patch client.Patch) error {
	return ldcHandler.Patch(context.Background(), ldcHandler.diskClaim, patch)
}

func (ldcHandler *Handler) Refresh() error {
	ldc, err := ldcHandler.GetLocalDiskClaim(client.ObjectKey{Name: ldcHandler.diskClaim.GetName(), Namespace: ldcHandler.diskClaim.GetNamespace()})
	if err != nil {
		return err
	}
	ldcHandler.For(ldc.DeepCopy())
	return nil
}

func (ldcHandler *Handler) ShowObjectInfo(msg string) {
	log.WithFields(log.Fields{
		"diskClaim":       ldcHandler.diskClaim.GetName(),
		"generation":      ldcHandler.diskClaim.GetGeneration(),
		"resourceVersion": ldcHandler.diskClaim.ResourceVersion,
		"Status":          ldcHandler.diskClaim.Status.Status,
		"diskRef":         ldcHandler.diskClaim.Spec.DiskRefs,
	}).Info(msg)
}

func (ldcHandler *Handler) DeleteLocalDiskClaim() error {
	return ldcHandler.Delete(context.Background(), ldcHandler.diskClaim)
}
