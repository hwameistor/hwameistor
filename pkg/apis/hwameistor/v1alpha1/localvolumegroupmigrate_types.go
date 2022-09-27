package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LocalVolumeGroupMigrateSpec defines the desired state of LocalVolumeGroupMigrate
type LocalVolumeGroupMigrateSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// *** custom section of the operations ***

	LocalVolumeGroupName string `json:"localVolumeGroupName"`

	// target NodeNames
	TargetNodesNames []string `json:"targetNodesNames"`

	// source NodeNames
	SourceNodesNames []string `json:"sourceNodesNames"`

	// *** common section of all the operations ***

	// +kubebuilder:default:=false
	Abort bool `json:"abort,omitempty"`
}

// LocalVolumeGroupMigrateStatus defines the observed state of LocalVolumeGroupMigrate
type LocalVolumeGroupMigrateStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// record the volume's replica number, it will be set internally
	ReplicaNumber int64 `json:"replicaNumber,omitempty"`
	// record the node where the specified replica is migrated to
	TargetNodesNames []string `json:"targetNodesNames,omitempty"`

	// State of the operation, e.g. submitted, started, completed, abort, ...
	State State `json:"state,omitempty"`
	// error message to describe some states
	Message string `json:"message,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalVolumeGroupMigrate is the Schema for the LocalVolumeGroupMigrates API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=LocalVolumeGroupMigrates,scope=Cluster,shortName=lvmigrate
// +kubebuilder:printcolumn:name="volume",type=string,JSONPath=`.spec.volumeName`,description="Name of the volume to be migrated"
// +kubebuilder:printcolumn:name="node",type=string,JSONPath=`.spec.nodeName`,description="Node name of the volume replica to be migrated"
// +kubebuilder:printcolumn:name="target",type=string,JSONPath=`.status.targetNodeName`,description="Node name of the new volume replica"
// +kubebuilder:printcolumn:name="state",type=string,JSONPath=`.status.state`,description="State of the migration"
// +kubebuilder:printcolumn:name="age",type=date,JSONPath=`.metadata.creationTimestamp`
type LocalVolumeGroupMigrate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LocalVolumeGroupMigrateSpec   `json:"spec,omitempty"`
	Status LocalVolumeGroupMigrateStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalVolumeGroupMigrateList contains a list of LocalVolumeGroupMigrate
type LocalVolumeGroupMigrateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LocalVolumeGroupMigrate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LocalVolumeGroupMigrate{}, &LocalVolumeGroupMigrateList{})
}
