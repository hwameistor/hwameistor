package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	SourceVolumeSnapshotAnnoKey           = "hwameistor.io/source-snapshot"
	VolumeSnapshotRestoreCompletedAnnoKey = "hwameistor.io/snapshot-restore-completed"
)

// LocalVolumeSnapshotRestoreSpec defines the desired state of LocalVolumeSnapshotRestore
type LocalVolumeSnapshotRestoreSpec struct {
	// TargetVolume is the name of the volume to restore to
	TargetVolume string `json:"targetVolume,omitempty"`

	// TargetVolume is the name of the target volume will place at
	TargetPoolName string `json:"targetPoolName,omitempty"`

	// SourceVolumeSnapshot represents which snapshot is used for volume to restore from
	// +kubebuilder:validation:Required
	SourceVolumeSnapshot string `json:"sourceVolumeSnapshot"`

	// RestoreType is the type about how to restore the volume, e.g., rollback, create. By default, create.
	// +Kubebuilder:default:=create
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum:=rollback;create
	RestoreType RestoreType `json:"restoreType"`

	// Abort can be used to abort the restore operation and clean up sub resources created by the restore operation automatically
	// +kubebuilder:default:=false
	Abort bool `json:"abort,omitempty"`
}

// LocalVolumeSnapshotRestoreStatus defines the observed state of LocalVolumeSnapshotRestore
type LocalVolumeSnapshotRestoreStatus struct {
	// VolumeReplicaSnapshotRestore is the replica snapshot to be restored
	VolumeReplicaSnapshotRestore []string `json:"volumeReplicaSnapshotRestore,omitempty"`

	// State is the phase of restore volume snapshot, e.g. submitted, started, completed, abort, ...
	State State `json:"state,omitempty"`

	// Message error message to describe some states
	Message string `json:"message,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalVolumeSnapshotRestore is a user's request for either restoring a local volume snapshot to a new volume, or merging into the old volume.
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=localvolumesnapshotrestores,scope=Cluster,shortName=lvsrestore;lvsnaprestore
// +kubebuilder:printcolumn:name="targetvolume",type=string,JSONPath=`.spec.targetVolume`,description="Target for the restore"
// +kubebuilder:printcolumn:name="sourcesnapshot",type=string,JSONPath=`.spec.sourceVolumeSnapshot`,description="Source snapshot for the restore"
// +kubebuilder:printcolumn:name="state",type=string,JSONPath=`.status.state`,description="State of the restore"
// +kubebuilder:printcolumn:name="age",type=date,JSONPath=`.metadata.creationTimestamp`
type LocalVolumeSnapshotRestore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LocalVolumeSnapshotRestoreSpec   `json:"spec,omitempty"`
	Status LocalVolumeSnapshotRestoreStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalVolumeSnapshotRestoreList contains a list of LocalVolumeSnapshotRestore
type LocalVolumeSnapshotRestoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LocalVolumeSnapshotRestore `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LocalVolumeSnapshotRestore{}, &LocalVolumeSnapshotRestoreList{})
}
