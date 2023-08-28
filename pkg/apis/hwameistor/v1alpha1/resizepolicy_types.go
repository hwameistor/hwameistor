package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ResizePolicySpec defines the desired state of ResizePolicy
type ResizePolicySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	WarningThreshold int8 `json:"warningThreshold"`

	ResizeThreshold int8 `json:"resizeThreshold"`

	NodePoolUsageLimit int8 `json:"nodePoolUsageLimit"`

	StorageClassSelector *metav1.LabelSelector `json:"storageClassSelector,omitempty"`

	NamespaceSelector *metav1.LabelSelector `json:"namespaceSelector,omitempty"`

	PVCSelector *metav1.LabelSelector `json:"pvcSelector,omitempty"`

}

// ResizePolicyStatus defines the observed state of ResizePolicy
type ResizePolicyStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ResizePolicy is the Schema for the resizepolicies API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=resizepolicies,scope=Cluster
type ResizePolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ResizePolicySpec   `json:"spec,omitempty"`
	Status ResizePolicyStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ResizePolicyList contains a list of ResizePolicy
type ResizePolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ResizePolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ResizePolicy{}, &ResizePolicyList{})
}
