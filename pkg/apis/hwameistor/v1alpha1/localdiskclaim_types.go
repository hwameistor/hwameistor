package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LocalDiskClaimSpec defines the desired state of LocalDiskClaim
type LocalDiskClaimSpec struct {
	// +kubebuilder:validation:Required
	// NodeName represents where disk has to be claimed.
	NodeName string `json:"nodeName"`

	// Description of the disk to be claimed
	// +optional
	Description DiskClaimDescription `json:"description,omitempty"`

	// DiskRefs represents which disks are assigned to the LocalDiskClaim
	// +optional
	DiskRefs []*v1.ObjectReference `json:"diskRefs,omitempty"`

	// Owner represents which system owns this claim(e.g. local-storage, local-disk-manager)
	// +kubebuilder:validation:Required
	Owner string `json:"owner"`
}

type LocalDiskClaimSpecArray []LocalDiskClaimSpec

// LocalDiskClaimStatus defines the observed state of LocalDiskClaim
type LocalDiskClaimStatus struct {
	// Status represents the current statue of the claim
	// +kubebuilder:validation:Enum:=Bound;Pending;Extending;Consumed;ToBeDeleted;Deleted
	Status DiskClaimStatus `json:"status,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalDiskClaim is the Schema for the localdiskclaims API
// +kubebuilder:validation:Required
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".spec.nodeName",name=NodeMatch,type=string
// +kubebuilder:printcolumn:JSONPath=".status.status",name=Phase,type=string
// +kubebuilder:printcolumn:JSONPath=".spec.owner",name=Owner,type=string
// +kubebuilder:resource:scope=Cluster,shortName=ldc
type LocalDiskClaim struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LocalDiskClaimSpec   `json:"spec,omitempty"`
	Status LocalDiskClaimStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalDiskClaimList contains a list of LocalDiskClaim
type LocalDiskClaimList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LocalDiskClaim `json:"items"`
}

// DiskClaimDescription defines the details of the disk that should be claimed
type DiskClaimDescription struct {
	// DiskType represents the type of drive like SSD, HDD etc.,
	// +optional
	DiskType string `json:"diskType,omitempty"`

	// Capacity of the disk in bytes
	Capacity int64 `json:"capacity,omitempty"`

	// Match the name of Localdisk
	// LocalDiskNames match all disks if LocalDiskNames are empty
	LocalDiskNames []string `json:"localDiskNames,omitempty"`

	// Match the device path of Localdisk
	// DevicePaths match all disks if DevicePaths are empty
	DevicePaths []string `json:"devicePaths,omitempty"`
}

// DiskClaimStatus is a typed string for phase field of BlockDeviceClaim.
type DiskClaimStatus string

// LocalDiskClaim CR, when created pass through phases before it got some Disks Assigned.
const (
	// LocalDiskClaimStatusEmpty represents that the LocalDiskClaim was just created.
	DiskClaimStatusEmpty DiskClaimStatus = ""

	// LocalDiskClaimStatusPending represents LocalDiskClaim has not been assigned devices yet. Rather
	// search is going on for matching disks.
	LocalDiskClaimStatusPending DiskClaimStatus = "Pending"

	// LocalDiskClaimStatusExtending represents LocalDiskClaim has been assigned devices yet, but need more disks.
	// Rather search is going on for matching disks.
	LocalDiskClaimStatusExtending DiskClaimStatus = "Extending"

	// LocalDiskClaimStatusBound represents LocalDiskClaim has been assigned backing disk and ready for use.
	LocalDiskClaimStatusBound DiskClaimStatus = "Bound"

	// LocalDiskClaimStatusConsumed represents disks backing this LocalDiskClaim is consumed by the consumer
	LocalDiskClaimStatusConsumed DiskClaimStatus = "Consumed"

	// LocalDiskClaimStatusToBeDeleted represents disks backing this LocalDiskClaim is consumed already and the claim
	// will be deleted after some clean job done
	LocalDiskClaimStatusToBeDeleted DiskClaimStatus = "ToBeDeleted"

	// LocalDiskClaimStatusDeleted  represents disks backing this LocalDiskClaim can be deleted
	LocalDiskClaimStatusDeleted DiskClaimStatus = "Deleted"
)

func init() {
	SchemeBuilder.Register(&LocalDiskClaim{}, &LocalDiskClaimList{})
}
