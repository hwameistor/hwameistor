package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LocalVolumeReplicaSnapshotSpec represents the actual localvolume snapshot object in lvm
type LocalVolumeReplicaSnapshotSpec struct {
	// NodeName specifies which node the snapshot will be placed
	// +kubebuilder:validation:Required
	NodeName string `json:"nodeName"`

	// SourceVolume specifies the source volume name of the snapshot
	// +kubebuilder:validation:Required
	SourceVolume string `json:"sourceVolume"`

	// SourceVolume specifies the source volume replica name of the snapshot
	// +kubebuilder:validation:Required
	SourceVolumeReplica string `json:"sourceVolumeReplica"`

	// PoolName specifies which volume group the snapshot and source volume is placed
	// valid options are LocalStorage_PoolHDD, LocalStorage_PoolSSD, LocalStorage_PoolNVMe
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum:=LocalStorage_PoolHDD;LocalStorage_PoolSSD;LocalStorage_PoolNVMe
	PoolName string `json:"poolName"`

	// SnapSize specifies the space reserved for the snapshot
	// +kubebuilder:validation:Required
	SnapSize int64 `json:"snapSize"`
}

// LocalVolumeReplicaSnapshotStatus defines the observed state of LocalVolumeReplicaSnapshot
type LocalVolumeReplicaSnapshotStatus struct {
	// AllocatedCapacityBytes is the real allocated capacity in bytes
	AllocatedCapacityBytes int64 `json:"allocatedCapacityBytes,omitempty"`

	// State is the phase of volume replica, e.g. Creating, Ready, NotReady, ToBeDeleted, Deleted
	State State `json:"state,omitempty"`

	// Message error message to describe some states
	Message string `json:"reason,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalVolumeReplicaSnapshot is the Schema for the localvolumereplicasnapshots API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=localvolumereplicasnapshots,scope=Cluster,shortName=lvrs
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="SourceVolumeReplica",type=string,JSONPath=`.spec.sourceVolumeReplica`,description="Name of the snapshot's source volume replica"
// +kubebuilder:printcolumn:name="Capacity",type=integer,JSONPath=`.status.allocatedCapacityBytes`,description="Allocated capacity of the snapshot"
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`,description="State of the snapshot"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:printcolumn:name="SourceVolume",type=string,JSONPath=`.spec.sourceVolume`,description="Name of the snapshot's source volume",priority=1
type LocalVolumeReplicaSnapshot struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LocalVolumeReplicaSnapshotSpec   `json:"spec,omitempty"`
	Status LocalVolumeReplicaSnapshotStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalVolumeReplicaSnapshotList contains a list of LocalVolumeReplicaSnapshot
type LocalVolumeReplicaSnapshotList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LocalVolumeReplicaSnapshot `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LocalVolumeReplicaSnapshot{}, &LocalVolumeReplicaSnapshotList{})
}
