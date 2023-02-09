package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LocalDiskClaimExpandSpec defines the desired state of LocalDiskClaimExpand
type LocalDiskClaimExpandSpec struct {
	// ClaimName represents which LocalDiskClaim needs to be extended
	// +kubebuilder:validation:Required
	ClaimName string `json:"claimName"`

	// +kubebuilder:default:=false
	Abort bool `json:"abort,omitempty"`
}

// LocalDiskClaimExpandStatus defines the observed state of LocalDiskClaimExpand
type LocalDiskClaimExpandStatus struct {
	// LastClaimedDisks records disks that claimed before creating this expand-job
	LastClaimedDisks []string `json:"lastClaimedDisks,omitempty"`

	// ExpandedDisks records disks that claimed after creating this expand-job
	ExpandedDisks []string `json:"expandedDisks,omitempty"`

	// State represents this expand-job state
	// +kubebuilder:validation:Enum:=Submitted;InProgress;Failed;Completed
	State State `json:"state,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalDiskClaimExpand is the Schema for the localdiskclaimexpands API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=localdiskclaimexpands,scope=Cluster
type LocalDiskClaimExpand struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LocalDiskClaimExpandSpec   `json:"spec,omitempty"`
	Status LocalDiskClaimExpandStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalDiskClaimExpandList contains a list of LocalDiskClaimExpand
type LocalDiskClaimExpandList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LocalDiskClaimExpand `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LocalDiskClaimExpand{}, &LocalDiskClaimExpandList{})
}
