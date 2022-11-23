package localdisk

import (
	"context"

	v1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/filter"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/tools/reference"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Handler struct {
	client.Client
	record.EventRecorder
	localDisk *v1alpha1.LocalDisk
	filter    filter.LocalDiskFilter
}

func NewLocalDiskHandler(client client.Client, recorder record.EventRecorder) *Handler {
	return &Handler{
		Client:        client,
		EventRecorder: recorder,
	}
}

func (ldHandler *Handler) GetLocalDisk(key client.ObjectKey) (*v1alpha1.LocalDisk, error) {
	ld := v1alpha1.LocalDisk{}
	if err := ldHandler.Get(context.Background(), key, &ld); err != nil {
		return nil, err
	}

	return &ld, nil
}

func (ldHandler *Handler) GetLocalDiskWithLabels(labels labels.Set) (*v1alpha1.LocalDiskList, error) {
	list := &v1alpha1.LocalDiskList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "LocalDisk",
			APIVersion: "v1alpha1",
		},
	}
	return list, ldHandler.List(context.TODO(), list, &client.ListOptions{LabelSelector: labels.AsSelector()})
}

func (ldHandler *Handler) ListLocalDisk() (*v1alpha1.LocalDiskList, error) {
	list := &v1alpha1.LocalDiskList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "LocalDisk",
			APIVersion: "v1alpha1",
		},
	}

	err := ldHandler.List(context.TODO(), list)
	return list, err
}

func (ldHandler *Handler) ListNodeLocalDisk(node string) (*v1alpha1.LocalDiskList, error) {
	list := &v1alpha1.LocalDiskList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "LocalDisk",
			APIVersion: "v1alpha1",
		},
	}
	nodeMatcher := client.MatchingFields{"spec.nodeName": node}
	err := ldHandler.List(context.TODO(), list, nodeMatcher)
	return list, err
}

func (ldHandler *Handler) For(ld *v1alpha1.LocalDisk) *Handler {
	ldHandler.localDisk = ld
	ldHandler.filter = filter.NewLocalDiskFilter(ld)
	return ldHandler
}

// UnClaimed Bounded
func (ldHandler *Handler) UnClaimed() bool {
	return ldHandler.filter.
		Init().
		Available().
		GetTotalResult()
}

// BoundTo assign disk to ldc
func (ldHandler *Handler) BoundTo(ldc *v1alpha1.LocalDiskClaim) error {
	// If this disk has already bound to the ldc, return directly
	if ldHandler.localDisk.Spec.ClaimRef != nil &&
		ldc.GetName() == ldHandler.localDisk.Spec.ClaimRef.Name {
		return nil
	}

	// Update the disk.spec.ClaimRef field to indicate that the disk is claimed
	ldcRef, _ := reference.GetReference(nil, ldc)
	ldHandler.localDisk.Spec.ClaimRef = ldcRef

	err := ldHandler.Update()
	if err == nil {
		// Record a Bound Event
		ldHandler.RecordEvent(v1.EventTypeNormal, v1alpha1.LocalDiskEventReasonBound,
			"Bounded by LocalDiskClaim: %v", ldc.GetName())
	}

	return err
}

func (ldHandler *Handler) SetupStatus(status v1alpha1.LocalDiskState) {
	ldHandler.localDisk.Status.State = status
}

func (ldHandler *Handler) SetupLabel(labels labels.Set) {
	if ldHandler.localDisk.ObjectMeta.Labels == nil {
		ldHandler.localDisk.ObjectMeta.Labels = make(map[string]string)
	}
	for k, v := range labels {
		ldHandler.localDisk.ObjectMeta.Labels[k] = v
	}
}

func (ldHandler *Handler) RemoveLabel(labels labels.Set) {
	for k := range labels {
		delete(ldHandler.localDisk.ObjectMeta.Labels, k)
	}
}

func (ldHandler *Handler) UpdateStatus() error {
	return ldHandler.Status().Update(context.TODO(), ldHandler.localDisk)
}

func (ldHandler *Handler) Update() error {
	return ldHandler.Client.Update(context.TODO(), ldHandler.localDisk)
}

func (ldHandler *Handler) ClaimRef() *v1.ObjectReference {
	return ldHandler.localDisk.Spec.ClaimRef
}

func (ldHandler *Handler) ReserveDisk() {
	ldHandler.localDisk.Spec.Reserved = true
}

func (ldHandler *Handler) FilterDisk(ldc *v1alpha1.LocalDiskClaim) bool {
	// Bounded disk
	if ldHandler.filter.HasBoundWith(ldc.GetName()) {
		return true
	}

	// Unbound disk
	return ldHandler.filter.
		Init().
		Available().
		HasNotReserved().
		NodeMatch(ldc.Spec.NodeName).
		Capacity(ldc.Spec.Description.Capacity).
		DiskType(ldc.Spec.Description.DiskType).
		DevType().
		NoPartition().
		GetTotalResult()
}

func (ldHandler *Handler) IsEmpty() bool {
	return !ldHandler.localDisk.Spec.HasPartition
}

func (ldHandler *Handler) RecordEvent(eventtype, reason, messageFmt string, args ...interface{}) {
	ldHandler.Eventf(ldHandler.localDisk, eventtype, reason, messageFmt, args)
}
