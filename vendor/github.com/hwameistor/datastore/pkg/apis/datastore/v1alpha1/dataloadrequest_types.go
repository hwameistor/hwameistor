package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DataLoadRequestSpec defines the desired state of DataLoadRequest
type DataLoadRequestSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// indicate if the request is for all or not
	// +kubebuilder:default:=true
	IsGlobal bool `json:"isGlobal"`
	// name of the node who will loads the data, and it works only when isglobal is false
	Node string `json:"node,omitempty"`

	// name of the dataSet source
	DataSet string `json:"dataSet"`
	SubDir  string `json:"subDir,omitempty"`
}

// DataLoadRequestStatus defines the observed state of DataLoadRequest
type DataLoadRequestStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// when a node finishes the data loading, record it here
	ReadyNodes []string `json:"readyNodes"`
	// State of the operation, e.g. submitted, started, completed, abort, ...
	State State `json:"state,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DataLoadRequest is the Schema for the dataloadrequests API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=dataloadrequests,scope=Namespaced,shortName=dlr
// +kubebuilder:printcolumn:JSONPath=".spec.dataSet",name=DataSet,type=string
// +kubebuilder:printcolumn:JSONPath=".spec.subDir",name=SubDir,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date
type DataLoadRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DataLoadRequestSpec   `json:"spec,omitempty"`
	Status DataLoadRequestStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DataLoadRequestList contains a list of DataLoadRequest
type DataLoadRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DataLoadRequest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DataLoadRequest{}, &DataLoadRequestList{})
}
