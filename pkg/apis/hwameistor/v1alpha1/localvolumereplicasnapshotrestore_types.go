package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LocalVolumeReplicaSnapshotRestoreSpec defines the desired state of LocalVolumeReplicaSnapshotRestore
type LocalVolumeReplicaSnapshotRestoreSpec struct {
	// NodeName is the name of the node that snapshot will be restored at
	// +kubebuilder:validation:Required
	NodeName string `json:"nodeName"`

	// TargetVolume is the name of the volume to restore to
	// +kubebuilder:validation:Required
	TargetVolume string `json:"targetVolume"`

	// TargetVolume is the name of the target volume will place at
	// +kubebuilder:validation:Required
	TargetPoolName string `json:"targetPoolName"`

	// SourceVolumeSnapshot represents which snapshot is used for volume to restore from
	// +kubebuilder:validation:Required
	SourceVolumeSnapshot string `json:"sourceVolumeSnapshot"`

	// SourceVolumeReplicaSnapshot represents which replica snapshot is used for volume to restore from
	// +kubebuilder:validation:Required
	SourceVolumeReplicaSnapshot string `json:"sourceVolumeReplicaSnapshot"`

	// RestoreType is the type about how to restore the volume, e.g. rollback, create. By default create.
	// +kubebuilder:default:=create
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum:=rollback;create
	RestoreType RestoreType `json:"restoreType"`

	// +kubebuilder:validation:Required
	VolumeSnapshotRestore string `json:"volumeSnapshotRestore"`

	// Abort can be used to abort the restore operation and clean up sub resources created by the restore operation automatically
	// +kubebuilder:default:=false
	Abort bool `json:"abort,omitempty"`
}

// LocalVolumeReplicaSnapshotRestoreStatus defines the observed state of LocalVolumeReplicaSnapshotRestore
type LocalVolumeReplicaSnapshotRestoreStatus struct {
	// State is the phase of restore volume snapshot, e.g. submitted, started, completed, abort, ...
	State State `json:"state,omitempty"`

	// Message error message to describe some states
	Message string `json:"message,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalVolumeReplicaSnapshotRestore is the Schema for the localvolumereplicasnapshotrestores API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=localvolumereplicasnapshotrestores,scope=Cluster,shortName=lvrsrestore;lvrsnaprestore
// +kubebuilder:printcolumn:name="nodeName",type=string,JSONPath=`.spec.nodeName`,description="Node to restore"
// +kubebuilder:printcolumn:name="targetvolume",type=string,JSONPath=`.spec.targetVolume`,description="Target for the restore"
// +kubebuilder:printcolumn:name="sourcesnapshot",type=string,JSONPath=`.spec.sourceVolumeSnapshot`,description="Source snapshot for the restore"
// +kubebuilder:printcolumn:name="state",type=string,JSONPath=`.status.state`,description="State of the restore"
// +kubebuilder:printcolumn:name="age",type=date,JSONPath=`.metadata.creationTimestamp`
type LocalVolumeReplicaSnapshotRestore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LocalVolumeReplicaSnapshotRestoreSpec   `json:"spec,omitempty"`
	Status LocalVolumeReplicaSnapshotRestoreStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalVolumeReplicaSnapshotRestoreList contains a list of LocalVolumeReplicaSnapshotRestore
type LocalVolumeReplicaSnapshotRestoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LocalVolumeReplicaSnapshotRestore `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LocalVolumeReplicaSnapshotRestore{}, &LocalVolumeReplicaSnapshotRestoreList{})
}
