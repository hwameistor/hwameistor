package localdiskclaim

import (
	"context"
	"fmt"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils"
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

	v1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	diskHandler "github.com/hwameistor/hwameistor/pkg/local-disk-manager/handler/localdisk"
)

const (
	// HwameiStorReclaim is used in annotation to check whether LocalDiskClaim is to reclaim or not
	HwameiStorReclaim = "hwameistor.io/reclaim"

	// HwameiStorLastClaimedDisks ius used in annotation to storage last claimed disks
	HwameiStorLastClaimedDisks = "hwameistor.io/last-claimed-disks"
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

	var finalAssignedDisks []string
	var originalDisks = func() (s []string) {
		for _, disk := range diskClaim.Spec.DiskRefs {
			s = append(s, disk.Name)
		}
		return
	}()

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

	// Check if the request needs more new disks
	if ldcHandler.NeedReclaim() {
		if err = ldcHandler.findAndSetLastClaimedDisksAnnotation(originalDisks, finalAssignedDisks); err != nil {
			return err
		}
	}

	// There must be more than one disk have been assigned or else return err
	if len(finalAssignedDisks) <= 0 {
		log.Infof("There is no available disk assigned to %v", diskClaim.GetName())
		return fmt.Errorf("there is no available disk assigned to %v", diskClaim.GetName())
	}

	log.Infof("Disk %v has been assigned to %v", finalAssignedDisks, diskClaim.GetName())
	return nil
}

// PatchBoundDiskRef update all disk bounded by the diskClaim to claim.spec.disks
func (ldcHandler *Handler) PatchBoundDiskRef() error {
	time.Sleep(time.Second)
	diskList, err := diskHandler.
		NewLocalDiskHandler(ldcHandler.Client, ldcHandler.EventRecorder).
		ListNodeLocalDisk(ldcHandler.diskClaim.Spec.NodeName)
	if err != nil {
		return err
	}

	oldDiskClaim := ldcHandler.diskClaim.DeepCopy()
	log.WithFields(log.Fields{"diskClaim": ldcHandler.diskClaim.GetName()}).
		Infof("Found %d localdisk(s) in cluster", len(diskList.Items))
	for _, disk := range diskList.Items {
		if disk.Spec.ClaimRef != nil &&
			disk.Spec.ClaimRef.Name == ldcHandler.diskClaim.GetName() {
			ldcHandler.AppendDiskRef(&disk)
		}
	}

	log.Infof("Found %d localdisk(s) bounded by claim %v",
		len(ldcHandler.diskClaim.Spec.DiskRefs), ldcHandler.diskClaim.GetName())
	for _, disk := range ldcHandler.diskClaim.Spec.DiskRefs {
		log.WithField("diskClaim", ldcHandler.diskClaim.GetName()).
			Infof("Bounded localdisk: %s", disk.Name)
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

func (ldcHandler *Handler) findAndSetLastClaimedDisksAnnotation(originalDisks, currentDisks []string) error {
	newDisks, found := utils.FoundNewStringElems(originalDisks, currentDisks)
	if !found {
		return fmt.Errorf("there is no disk(s) assigned to %s", ldcHandler.diskClaim.GetName())
	}

	oldDiskClaim := ldcHandler.diskClaim.DeepCopy()
	annotations := ldcHandler.diskClaim.GetAnnotations()
	// Set reclaim key to false
	annotations[HwameiStorReclaim] = "false"
	annotations[HwameiStorLastClaimedDisks] = func(disks []string) (s string) {
		for _, disk := range disks {
			s = disk + ","
		}
		return strings.TrimSuffix(s, ",")
	}(newDisks)

	ldcHandler.diskClaim.SetAnnotations(annotations)
	return ldcHandler.PatchClaimSpec(client.MergeFrom(oldDiskClaim))
}

func (ldcHandler *Handler) NeedReclaim() bool {
	annotation := ldcHandler.diskClaim.GetAnnotations()
	if annotation != nil {
		if val, ok := annotation[HwameiStorReclaim]; ok && strings.ToLower(val) == "true" {
			return true
		}
	}
	return false
}
