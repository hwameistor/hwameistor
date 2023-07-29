package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// EventSpec defines the desired state of Event
type EventSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	ResourceType string `json:"resourceType"`

	ResourceName string `json:"resourceName"`

	Records []EventRecord `json:"records"`
}

type EventRecord struct {
	Time metav1.Time `json:"time,omitempty"`

	Action        string `json:"action,omitempty"`
	ActionContent string `json:"actionContent,omitempty"`

	Result        string `json:"result,omitempty"`
	ResultContent string `json:"resultContent,omitempty"`
}

// EventStatus defines the observed state of Event
type EventStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Event is the Schema for the events API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=events,scope=Cluster
// +kubebuilder:printcolumn:JSONPath=".spec.resourceType",name=type,type=string
// +kubebuilder:printcolumn:JSONPath=".spec.resourceName",name=name,type=string
type Event struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EventSpec   `json:"spec,omitempty"`
	Status EventStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// EventList contains a list of Event
type EventList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Event `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Event{}, &EventList{})
}
