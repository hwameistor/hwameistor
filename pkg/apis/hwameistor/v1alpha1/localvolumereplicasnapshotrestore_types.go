package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LocalVolumeReplicaSnapshotRestoreSpec defines the desired state of LocalVolumeReplicaSnapshotRestore
type LocalVolumeReplicaSnapshotRestoreSpec struct {
	LocalVolumeSnapshotRestoreSpec `json:",inline"`

	// +kubebuilder:validation:Required
	VolumeSnapshotRestore string `json:"volumeSnapshotRestore"`

	// NodeName is the name of the node that snapshot will be restored
	// +kubebuilder:validation:Required
	NodeName string `json:"nodeName"`
}

// LocalVolumeReplicaSnapshotRestoreStatus defines the observed state of LocalVolumeReplicaSnapshotRestore
type LocalVolumeReplicaSnapshotRestoreStatus struct {
	LocalVolumeSnapshotRestoreStatus `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalVolumeReplicaSnapshotRestore is the Schema for the localvolumereplicasnapshotrestores API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=localvolumereplicasnapshotrestores,scope=Cluster,shortName=lvrsrestore;lvrsnaprestore
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
