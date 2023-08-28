package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LocalVolumeReplicaSnapshotRecoverSpec defines the desired state of LocalVolumeReplicaSnapshotRecover
type LocalVolumeReplicaSnapshotRecoverSpec struct {
	// NodeName is the name of the node that snapshot will be recovered at
	// +kubebuilder:validation:Required
	NodeName string `json:"nodeName"`

	// TargetVolume is the name of the volume to recover to
	// +kubebuilder:validation:Required
	TargetVolume string `json:"targetVolume"`

	// TargetVolume is the name of the target volume will place at
	// +kubebuilder:validation:Required
	TargetPoolName string `json:"targetPoolName"`

	// SourceVolumeSnapshot represents which snapshot is used for volume to recover from
	// +kubebuilder:validation:Required
	SourceVolumeSnapshot string `json:"sourceVolumeSnapshot"`

	// SourceVolumeReplicaSnapshot represents which replica snapshot is used for volume to recover from
	// +kubebuilder:validation:Required
	SourceVolumeReplicaSnapshot string `json:"sourceVolumeReplicaSnapshot"`

	// RecoverType is the type about how to recover the volume, e.g. rollback, restore. By default restore.
	// +kubebuilder:default:=restore
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum:=rollback;restore
	RecoverType RecoverType `json:"recoverType"`

	// +kubebuilder:validation:Required
	VolumeSnapshotRecover string `json:"volumeSnapshotRecover"`

	// Abort can be used to abort the recover operation and clean up sub resources created by the recover operation automatically
	// +kubebuilder:default:=false
	Abort bool `json:"abort,omitempty"`
}

// LocalVolumeReplicaSnapshotRecoverStatus defines the observed state of LocalVolumeReplicaSnapshotRecover
type LocalVolumeReplicaSnapshotRecoverStatus struct {
	// State is the phase of recover volume snapshot, e.g. submitted, started, completed, abort, ...
	State State `json:"state,omitempty"`

	// Message error message to describe some states
	Message string `json:"message,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalVolumeReplicaSnapshotRecover is the Schema for the localvolumereplicasnapshotrecovers API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=localvolumereplicasnapshotrecovers,scope=Cluster,shortName=lvrsrecover;lvrsnaprecover
// +kubebuilder:printcolumn:name="nodeName",type=string,JSONPath=`.spec.nodeName`,description="Node to recover"
// +kubebuilder:printcolumn:name="targetvolume",type=string,JSONPath=`.spec.targetVolume`,description="Target for the recover"
// +kubebuilder:printcolumn:name="sourcesnapshot",type=string,JSONPath=`.spec.sourceVolumeSnapshot`,description="Source snapshot for the recover"
// +kubebuilder:printcolumn:name="state",type=string,JSONPath=`.status.state`,description="State of the recover"
// +kubebuilder:printcolumn:name="age",type=date,JSONPath=`.metadata.creationTimestamp`
type LocalVolumeReplicaSnapshotRecover struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LocalVolumeReplicaSnapshotRecoverSpec   `json:"spec,omitempty"`
	Status LocalVolumeReplicaSnapshotRecoverStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalVolumeReplicaSnapshotRecoverList contains a list of LocalVolumeReplicaSnapshotRecover
type LocalVolumeReplicaSnapshotRecoverList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LocalVolumeReplicaSnapshotRecover `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LocalVolumeReplicaSnapshotRecover{}, &LocalVolumeReplicaSnapshotRecoverList{})
}
