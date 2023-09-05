package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +kubebuilder:validation:Enum=reserve
type Action string

const LocalDiskActionReserve Action = "reserve"

// LocalDiskActionSpec defines the desired state of LocalDiskAction
type LocalDiskActionRule struct {
	// Device capacity should less than this value
	// +optional
	MaxCapacity int64 `json:"maxCapacity,omitempty"`
	// Device capacity should larger than this value
	// +optional
	MinCapacity int64 `json:"minCapacity,omitempty"`
	// Matched by glob, e.g. /dev/rbd*
	// +optional
	DevicePath string `json:"devicePath,omitempty"`
}

// LocalDiskActionSpec defines the desired state of LocalDiskAction
type LocalDiskActionSpec struct {
	// +optional
	Rule LocalDiskActionRule `json:"rule,omitempty"`

	// +kubebuilder:validation:Required
	Action Action `json:"action"`
}

// LocalDiskActionStatus defines the observed state of LocalDiskAction
type LocalDiskActionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// latest matched local disks
	LatestMatchedLds []string `json:"latestMatchedLds,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalDiskAction is the Schema for the localdiskactions API
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=lda
// +kubebuilder:printcolumn:JSONPath=".spec.action",name=Action,type=string
// +kubebuilder:printcolumn:JSONPath=".spec.rule.maxCapacity",name=MaxCapacity,type=integer
// +kubebuilder:printcolumn:JSONPath=".spec.rule.minCapacity",name=MinCapacity,type=integer
// +kubebuilder:printcolumn:JSONPath=".spec.rule.devicePath",name=DevicePath,type=string
type LocalDiskAction struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LocalDiskActionSpec   `json:"spec,omitempty"`
	Status LocalDiskActionStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalDiskActionList contains a list of LocalDiskAction
type LocalDiskActionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LocalDiskAction `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LocalDiskAction{}, &LocalDiskActionList{})
}
