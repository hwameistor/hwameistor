package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LocalStorageAlertSpec defines the desired state of LocalStorageAlert
type LocalStorageAlertSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Severity int `json:"severity"`

	Module string `json:"module"`

	Resource string `json:"resource"`

	Event string `json:"event"`

	Details string `json:"details"`
}

// LocalStorageAlertStatus defines the observed state of LocalStorageAlert
type LocalStorageAlertStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalStorageAlert is the Schema for the localstoragealerts API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=localstoragealerts,scope=Cluster,shortName=lsalert
// +kubebuilder:printcolumn:name="severity",type=integer,JSONPath=`.spec.severity`,description="Alert severity"
// +kubebuilder:printcolumn:name="module",type=string,JSONPath=`.spec.module`,description="Module of the alert"
// +kubebuilder:printcolumn:name="resource",type=string,JSONPath=`.spec.resource`,description="Resource name of the alert"
// +kubebuilder:printcolumn:name="event",type=string,JSONPath=`.spec.event`,description="Alert event"
type LocalStorageAlert struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LocalStorageAlertSpec   `json:"spec,omitempty"`
	Status LocalStorageAlertStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalStorageAlertList contains a list of LocalStorageAlert
type LocalStorageAlertList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LocalStorageAlert `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LocalStorageAlert{}, &LocalStorageAlertList{})
}
