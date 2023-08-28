package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LocalVolumeReplicaSnapshotSpec represents the actual localvolume snapshot object in lvm
type LocalVolumeReplicaSnapshotSpec struct {
	// NodeName specifies which node the snapshot will be placed
	// +kubebuilder:validation:Required
	NodeName string `json:"nodeName"`

	// VolumeSnapshotName represents the name of volume snapshot
	// +kubebuilder:validation:Required
	VolumeSnapshotName string `json:"volumeSnapshotName"`

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

	// RequiredCapacityBytes specifies the space reserved for the snapshot
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum:=4194304
	RequiredCapacityBytes int64 `json:"requiredCapacityBytes"`

	// Delete this snapshot if it is true
	// +kubebuilder:default:=false
	Delete bool `json:"delete,omitempty"`
}

// LocalVolumeReplicaSnapshotStatus defines the observed state of LocalVolumeReplicaSnapshot
type LocalVolumeReplicaSnapshotStatus struct {
	// AllocatedCapacityBytes is the real allocated capacity in bytes
	AllocatedCapacityBytes int64 `json:"allocatedCapacityBytes,omitempty"`

	// CreationTime is the host real snapshot creation time
	CreationTime *metav1.Time `json:"creationTime,omitempty"`

	// Attribute indicates attr on snapshot
	Attribute VolumeSnapshotAttr `json:"attr,omitempty"`

	// State is the phase of volume replica, e.g. Creating, Ready, NotReady, ToBeDeleted, Deleted
	State State `json:"state,omitempty"`

	// Message error message to describe some states
	Message string `json:"reason,omitempty"`
}

// VolumeSnapshotAttr defines attrs of volume, e.g. invalid, merging...
type VolumeSnapshotAttr struct {
	// Merging set true if snapshot is merging now
	Merging bool `json:"merging,omitempty"`

	// Invalid set true if snapshot is expiration
	Invalid bool `json:"invalid,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalVolumeReplicaSnapshot is the Schema for the localvolumereplicasnapshots API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=localvolumereplicasnapshots,scope=Cluster,shortName=lvrs
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Capacity",type=integer,JSONPath=`.status.allocatedCapacityBytes`,description="Allocated capacity of the snapshot"
// +kubebuilder:printcolumn:name="NodeName",type=string,JSONPath=`.spec.nodeName`,description="Node where snapshot is located"
// +kubebuilder:printcolumn:name="SourceVolume",type=string,JSONPath=`.spec.sourceVolume`,description="Name of the snapshot's source volume"
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`,description="State of the snapshot"
// +kubebuilder:printcolumn:name="Merging",type=string,JSONPath=`.status.attr.merging`,description="if the snapshot is merging"
// +kubebuilder:printcolumn:name="Invalid",type=string,JSONPath=`.status.attr.invalid`,description="if the snapshot is invalid"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.status.creationTime`
// +kubebuilder:printcolumn:name="SourceVolumeReplica",type=string,JSONPath=`.spec.sourceVolumeReplica`,description="Name of the snapshot's source volume replica",priority=1
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
