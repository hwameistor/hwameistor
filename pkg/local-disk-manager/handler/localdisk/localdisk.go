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

type LocalDiskHandler struct {
	client.Client
	record.EventRecorder
	localDisk *v1alpha1.LocalDisk
	filter    filter.LocalDiskFilter
}

func NewLocalDiskHandler(client client.Client, recorder record.EventRecorder) *LocalDiskHandler {
	return &LocalDiskHandler{
		Client:        client,
		EventRecorder: recorder,
	}
}

func (ldHandler *LocalDiskHandler) GetLocalDisk(key client.ObjectKey) (*v1alpha1.LocalDisk, error) {
	ld := v1alpha1.LocalDisk{}
	if err := ldHandler.Get(context.Background(), key, &ld); err != nil {
		return nil, err
	}

	return &ld, nil
}

func (ldHandler *LocalDiskHandler) GetLocalDiskWithLabels(labels labels.Set) (*v1alpha1.LocalDiskList, error) {
	list := &v1alpha1.LocalDiskList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "localDisk",
			APIVersion: "v1alpha1",
		},
	}
	return list, ldHandler.List(context.TODO(), list, &client.ListOptions{LabelSelector: labels.AsSelector()})
}

func (ldHandler *LocalDiskHandler) ListLocalDisk() (*v1alpha1.LocalDiskList, error) {
	list := &v1alpha1.LocalDiskList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "localDisk",
			APIVersion: "v1alpha1",
		},
	}

	err := ldHandler.List(context.TODO(), list)
	return list, err
}

func (ldHandler *LocalDiskHandler) ListNodeLocalDisk(node string) (*v1alpha1.LocalDiskList, error) {
	list := &v1alpha1.LocalDiskList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "localDisk",
			APIVersion: "v1alpha1",
		},
	}
	nodeMatcher := client.MatchingFields{"spec.nodeName": node}
	err := ldHandler.List(context.TODO(), list, nodeMatcher)
	return list, err
}

func (ldHandler *LocalDiskHandler) For(ld *v1alpha1.LocalDisk) *LocalDiskHandler {
	ldHandler.localDisk = ld
	ldHandler.filter = filter.NewLocalDiskFilter(ld)
	return ldHandler
}

// UnClaimed Bounded
func (ldHandler *LocalDiskHandler) UnClaimed() bool {
	return ldHandler.filter.
		Init().
		Available().
		GetTotalResult()
}

// BoundTo assign disk to ldc
func (ldHandler *LocalDiskHandler) BoundTo(ldc v1alpha1.LocalDiskClaim) error {
	ldcRef, err := reference.GetReference(nil, &ldc)
	if err != nil {
		return err
	}

	ldHandler.localDisk.Spec.ClaimRef = ldcRef
	ldHandler.localDisk.Status.State = v1alpha1.LocalDiskBound

	if err = ldHandler.UpdateStatus(); err != nil {
		return err
	}
	ldHandler.EventRecorder.Eventf(ldHandler.localDisk, v1.EventTypeNormal, "LocalDiskClaimed", "Claimed by %v/%v", ldc.Namespace, ldc.Name)
	return nil
}

func (ldHandler *LocalDiskHandler) SetupStatus(status v1alpha1.LocalDiskState) {
	ldHandler.localDisk.Status.State = status
}

func (ldHandler *LocalDiskHandler) SetupLabel(labels labels.Set) {
	if ldHandler.localDisk.ObjectMeta.Labels == nil {
		ldHandler.localDisk.ObjectMeta.Labels = make(map[string]string)
	}
	for k, v := range labels {
		ldHandler.localDisk.ObjectMeta.Labels[k] = v
	}
}

func (ldHandler *LocalDiskHandler) RemoveLabel(labels labels.Set) {
	for k := range labels {
		delete(ldHandler.localDisk.ObjectMeta.Labels, k)
	}
}

func (ldHandler *LocalDiskHandler) UpdateStatus() error {
	return ldHandler.Status().Update(context.TODO(), ldHandler.localDisk)
}

func (ldHandler *LocalDiskHandler) Update() error {
	return ldHandler.Client.Update(context.TODO(), ldHandler.localDisk)
}

func (ldHandler *LocalDiskHandler) ClaimRef() *v1.ObjectReference {
	return ldHandler.localDisk.Spec.ClaimRef
}

func (ldHandler *LocalDiskHandler) ReserveDisk() {
	ldHandler.localDisk.Spec.Reserved = true
}

func (ldHandler *LocalDiskHandler) FilterDisk(ldc v1alpha1.LocalDiskClaim) bool {
	// Bounded disk
	if ldHandler.filter.HasBoundWith(ldc.GetName()) {
		return true
	}

	// Unbound disk
	return ldHandler.filter.
		Init().
		Available().
		NodeMatch(ldc.Spec.NodeName).
		Capacity(ldc.Spec.Description.Capacity).
		DiskType(ldc.Spec.Description.DiskType).
		Unique(ldc.Spec.DiskRefs).
		DevType().
		NoPartition().
		GetTotalResult()
}

func (ldHandler *LocalDiskHandler) IsEmpty() bool {
	return !ldHandler.localDisk.Spec.HasPartition
}
