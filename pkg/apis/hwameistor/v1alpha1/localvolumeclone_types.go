package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LocalVolumeCloneSpec defines the desired state of LocalVolumeClone
type LocalVolumeCloneSpec struct {
	// SourceVolume specifies the source volume of the clone
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Immutable
	SourceVolume string `json:"sourceVolume"`

	// TargetVolume specifies the target volume of the clone
	// If the target volume is located at a remote node, the clone will happen cross-network
	// If the target volume does not exist, controller will create a new one that has the same attributes with source-volume
	// +kubebuilder:validation:Immutable
	TargetVolume string `json:"targetVolume,omitempty"`

	// TargetPoolName specifies which pool the target-volume will be placed at
	// The target pool can be different from the source-volume pool, for example,
	// a volume in PoolSSD can be cloned to PoolHDD.
	// +kubebuilder:validation:Immutable
	TargetPoolName string `json:"targetPoolName,omitempty"`

	// SourceVolumeSnapshot represents which snapshot is used for volume to restore from
	// it will be filled automatically by controller after snapshot created.
	// +kubebuilder:validation:Immutable
	SourceVolumeSnapshot string `json:"sourceVolumeSnapshot,omitempty"`

	// +kubebuilder:default:=false
	Abort bool `json:"abort,omitempty"`
}

// LocalVolumeCloneStatus defines the observed state of LocalVolumeClone
type LocalVolumeCloneStatus struct {
	// State is the phase of restore volume snapshot, e.g., submitted, started, completed, abort, ...
	State State `json:"state,omitempty"`

	// Message error message to describe some states
	Message string `json:"message,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalVolumeClone is the Schema for the localvolumeclones API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=localvolumeclones,scope=Cluster
// +kubebuilder:printcolumn:JSONPath=".spec.sourceVolume",name=sourcevolume,type=string
// +kubebuilder:printcolumn:JSONPath=".spec.targetVolume",name=targetvolume,type=string
// +kubebuilder:printcolumn:JSONPath=".spec.targetPoolName",name=targetpool,type=string
// +kubebuilder:printcolumn:JSONPath=".status.state",name=state,type=string
type LocalVolumeClone struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LocalVolumeCloneSpec   `json:"spec,omitempty"`
	Status LocalVolumeCloneStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalVolumeCloneList contains a list of LocalVolumeClone
type LocalVolumeCloneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LocalVolumeClone `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LocalVolumeClone{}, &LocalVolumeCloneList{})
}
