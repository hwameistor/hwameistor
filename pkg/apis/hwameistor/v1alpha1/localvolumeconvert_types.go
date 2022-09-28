package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LocalVolumeConvertSpec defines the desired state of LocalVolumeConvert
type LocalVolumeConvertSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// *** custom section of the operations ***

	VolumeName string `json:"volumeName,omitempty"`

	// ReplicaNumber is the number of replicas which the volume will be converted to
	// currently, only support the case of converting a non-HA volume to HA
	// +kubebuilder:validation:Minimum:=2
	// +kubebuilder:validation:Maximum:=2
	ReplicaNumber int64 `json:"replicaNumber,omitempty"`

	// *** common section of all the operations ***

	// +kubebuilder:default:=false
	Abort bool `json:"abort,omitempty"`
}

// LocalVolumeConvertStatus defines the observed state of LocalVolumeConvert
type LocalVolumeConvertStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	State State `json:"state,omitempty"`

	Message string `json:"message,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalVolumeConvert is the Schema for the localvolumeconverts API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=localvolumeconverts,scope=Cluster,shortName=lvconvert
// +kubebuilder:printcolumn:name="volume",type=string,JSONPath=`.spec.volumeName`,description="Name of the volume to convert"
// +kubebuilder:printcolumn:name="replicas",type=integer,JSONPath=`.spec.replicaNumber`,description="Number of volume replica"
// +kubebuilder:printcolumn:name="state",type=string,JSONPath=`.status.state`,description="State of the expansion"
// +kubebuilder:printcolumn:name="message",type=string,JSONPath=`.status.message`,description="Event message of the expansion"
// +kubebuilder:printcolumn:name="abort",type=boolean,JSONPath=`.spec.abort`,description="Abort the operation"
// +kubebuilder:printcolumn:name="age",type=date,JSONPath=`.metadata.creationTimestamp`
type LocalVolumeConvert struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LocalVolumeConvertSpec   `json:"spec,omitempty"`
	Status LocalVolumeConvertStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalVolumeConvertList contains a list of LocalVolumeConvert
type LocalVolumeConvertList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LocalVolumeConvert `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LocalVolumeConvert{}, &LocalVolumeConvertList{})
}
