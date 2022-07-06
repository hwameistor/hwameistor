package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LocalVolumeExpandSpec defines the desired state of LocalVolumeExpand
type LocalVolumeExpandSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster

	// *** custom section of the operations ***

	VolumeName string `json:"volumeName,omitempty"`

	// +kubebuilder:validation:Minimum:=4194304
	RequiredCapacityBytes int64 `json:"requiredCapacityBytes,omitempty"`

	// *** common section of all the operations ***

	// +kubebuilder:default:=false
	Abort bool `json:"abort,omitempty"`
}

// LocalVolumeExpandStatus defines the observed state of LocalVolumeExpand
type LocalVolumeExpandStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster

	// *** custom section of the operations ***

	AllocatedCapacityBytes int64 `json:"allocatedCapacityBytes,omitempty"`

	// *** common section of all the operations ***

	// sub resources at different node.
	Subs []string `json:"subs,omitempty"`

	State State `json:"state,omitempty"`

	Message string `json:"message,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalVolumeExpand is the Schema for the localvolumeexpands API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=localvolumeexpands,scope=Cluster,shortName=lvexpand
// +kubebuilder:printcolumn:name="newCapacity",type=integer,JSONPath=`.spec.requiredCapacityBytes`,description="New capacity of the volume"
// +kubebuilder:printcolumn:name="abort",type=boolean,JSONPath=`.spec.abort`,description="Abort the operation"
// +kubebuilder:printcolumn:name="state",type=string,JSONPath=`.status.state`,description="State of the expansion"
// +kubebuilder:printcolumn:name="subs",type=string,JSONPath=`.status.subs`,description="Sub-operations on each volume replica expansion"
// +kubebuilder:printcolumn:name="message",type=string,JSONPath=`.status.message`,description="Event message of the expansion"
// +kubebuilder:printcolumn:name="age",type=date,JSONPath=`.metadata.creationTimestamp`
type LocalVolumeExpand struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LocalVolumeExpandSpec   `json:"spec,omitempty"`
	Status LocalVolumeExpandStatus `json:"status,omitempty"`
}

// AddSubs updates with subs info
func (v *LocalVolumeExpand) AddSubs(subNames ...string) {
	subMaps := make(map[string]bool)
	for _, sub := range subNames {
		subMaps[sub] = true
	}
	for _, sub := range v.Status.Subs {
		subMaps[sub] = true
	}

	v.Status.Subs = v.Status.Subs[:0]
	for sub := range subMaps {
		v.Status.Subs = append(v.Status.Subs, sub)
	}
}

// HasSub updates with sub-resource info
func (v *LocalVolumeExpand) HasSub(name string) bool {
	for _, subname := range v.Status.Subs {
		if subname == name {
			return true
		}
	}
	return false
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalVolumeExpandList contains a list of LocalVolumeExpand
type LocalVolumeExpandList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LocalVolumeExpand `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LocalVolumeExpand{}, &LocalVolumeExpandList{})
}
