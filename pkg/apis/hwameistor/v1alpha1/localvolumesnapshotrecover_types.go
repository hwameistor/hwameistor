package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	SourceVolumeSnapshotAnnoKey           = "hwameistor.io/source-snapshot"
	VolumeSnapshotRecoverCompletedAnnoKey = "hwameistor.io/snapshot-recover-completed"
)

// LocalVolumeSnapshotRecoverSpec defines the desired state of LocalVolumeSnapshotRecover
type LocalVolumeSnapshotRecoverSpec struct {
	// TargetVolume is the name of the volume to recover to
	TargetVolume string `json:"targetVolume,omitempty"`

	// TargetVolume is the name of the target volume will place at
	TargetPoolName string `json:"targetPoolName,omitempty"`

	// SourceVolumeSnapshot represents which snapshot is used for volume to recover from
	// +kubebuilder:validation:Required
	SourceVolumeSnapshot string `json:"sourceVolumeSnapshot"`

	// RecoverType is the type about how to recover the volume, e.g. rollback, restore. By default restore.
	// +kubebuilder:default:=restore
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum:=rollback;restore
	RecoverType RecoverType `json:"recoverType"`

	// Abort can be used to abort the recover operation and clean up sub resources created by the recover operation automatically
	// +kubebuilder:default:=false
	Abort bool `json:"abort,omitempty"`
}

// LocalVolumeSnapshotRecoverStatus defines the observed state of LocalVolumeSnapshotRecover
type LocalVolumeSnapshotRecoverStatus struct {
	// VolumeReplicaSnapshotRecover is the replica snapshot to be recovered
	VolumeReplicaSnapshotRecover []string `json:"volumeReplicaSnapshotRecover,omitempty"`

	// State is the phase of recover volume snapshot, e.g. submitted, started, completed, abort, ...
	State State `json:"state,omitempty"`

	// Message error message to describe some states
	Message string `json:"message,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalVolumeSnapshotRecover is user's request for either recovering a local volume snapshot to a new volume, or merging into the old volume.
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=localvolumesnapshotrecovers,scope=Cluster,shortName=lvsrecover;lvsnaprecover
// +kubebuilder:printcolumn:name="targetvolume",type=string,JSONPath=`.spec.targetVolume`,description="Target for the recover"
// +kubebuilder:printcolumn:name="sourcesnapshot",type=string,JSONPath=`.spec.sourceVolumeSnapshot`,description="Source snapshot for the recover"
// +kubebuilder:printcolumn:name="state",type=string,JSONPath=`.status.state`,description="State of the recover"
// +kubebuilder:printcolumn:name="age",type=date,JSONPath=`.metadata.creationTimestamp`
type LocalVolumeSnapshotRecover struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LocalVolumeSnapshotRecoverSpec   `json:"spec,omitempty"`
	Status LocalVolumeSnapshotRecoverStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalVolumeSnapshotRecoverList contains a list of LocalVolumeSnapshotRecover
type LocalVolumeSnapshotRecoverList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LocalVolumeSnapshotRecover `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LocalVolumeSnapshotRecover{}, &LocalVolumeSnapshotRecoverList{})
}
