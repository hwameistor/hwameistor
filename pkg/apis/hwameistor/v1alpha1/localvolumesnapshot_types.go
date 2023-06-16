package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VolumeSnapshotSpec describes the common attributes of a volume snapshot.

// LocalVolumeSnapshotSpec describes the common attributes of a localvolume snapshot.
type LocalVolumeSnapshotSpec struct {
	// SourceVolume specifies the source volume of the snapshot
	// +kubebuilder:validation:Required
	SourceVolume string `json:"sourceVolume"`

	// RequiredCapacityBytes specifies the space reserved for the snapshot
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum:=4194304
	RequiredCapacityBytes int64 `json:"requiredCapacityBytes"`
}

// LocalVolumeSnapshotStatus defines the observed state of LocalVolumeSnapshot
type LocalVolumeSnapshotStatus struct {
	// AllocatedCapacityBytes is the real allocated capacity in bytes
	// In case of HA volume with multiple replicas, the value is equal to the one of a replica's snapshot size
	AllocatedCapacityBytes int64 `json:"allocatedCapacityBytes,omitempty"`

	// ReplicaSnapshots represents the actual snapshots of replica
	ReplicaSnapshots []string `json:"snapshots,omitempty"`

	// State is the phase of volume replica, e.g. Creating, Ready, NotReady, ToBeDeleted, Deleted
	State State `json:"state,omitempty"`

	// Message error message to describe some states
	Message string `json:"message,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalVolumeSnapshot is a user's request for either creating a point-in-time
// snapshot of a persistent localvolume, or binding to a pre-existing snapshot.
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=localvolumesnapshots,scope=Cluster,shortName=lvs
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="SourceVolume",type=string,JSONPath=`.spec.sourceVolume`,description="Name of the snapshot's source volume"
// +kubebuilder:printcolumn:name="state",type=string,JSONPath=`.status.state`,description="State of the snapshot"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type LocalVolumeSnapshot struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LocalVolumeSnapshotSpec   `json:"spec,omitempty"`
	Status LocalVolumeSnapshotStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalVolumeSnapshotList contains a list of LocalVolumeSnapshot
type LocalVolumeSnapshotList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LocalVolumeSnapshot `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LocalVolumeSnapshot{}, &LocalVolumeSnapshotList{})
}
