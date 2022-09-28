package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LocalVolumeReplicaSpec defines the desired state of LocalVolumeReplica
type LocalVolumeReplicaSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster

	// VolumeName is the name of the volume, e.g. pvc-fbf3ffc3-66db-4dae-9032-bda3c61b8f85
	VolumeName string `json:"volumeName,omitempty"`

	// PoolName is the name of the storage pool, e.g. LocalStorage_PoolHDD, LocalStorage_PoolSSD, etc..
	PoolName string `json:"poolName,omitempty"`

	// NodeName is the assigned node where the volume replica is located
	NodeName string `json:"nodeName,omitempty"`

	// +kubebuilder:validation:Minimum:=4194304
	RequiredCapacityBytes int64 `json:"requiredCapacityBytes,omitempty"`

	// Delete is to indicate where the replica should be deleted or not.
	// It's different from the regular resource delete interface in Kubernetes.
	// The purpose is to protect it from any mistakes
	// +kubebuilder:default:=false
	Delete bool `json:"delete,omitempty"`

	// Delete is to indicate this replica should be migrated or not.
	// If set to true, it will not be count into the total replicas number of the volume.
	// +kubebuilder:default:=false
	//Migrate bool `json:"migrate,omitempty"`
}

// LocalVolumeReplicaStatus defines the observed state of LocalVolumeReplica
type LocalVolumeReplicaStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster

	// StoragePath is a real path of the volume replica, like /dev/sdg.
	StoragePath string `json:"storagePath,omitempty"`

	// DevicePath is a link path of the StoragePath of the volume replica,
	// e.g. /dev/LocalStorage_PoolHDD/pvc-fbf3ffc3-66db-4dae-9032-bda3c61b8f85
	DevicePath string `json:"devPath,omitempty"`

	// Disks is a list of physical disks where the volume replica is spread cross, especially for striped LVM volume replica
	Disks []string `json:"disks,omitempty"`

	// AllocatedCapacityBytes is the real allocated capacity in bytes
	AllocatedCapacityBytes int64 `json:"allocatedCapacityBytes,omitempty"`

	// State is the phase of volume replica, e.g. Creating, Ready, NotReady, ToBeDeleted, Deleted
	State State `json:"state,omitempty"`

	// Synced is the sync state of the volume replica, which is important in HA volume
	// +kubebuilder:default:=false
	Synced bool `json:"synced,omitempty"`

	// HAState is state for ha replica, replica.Status.State == Ready only when HAState is Consistent of nil
	HAState *HAState `json:"haState,omitempty"`

	// InUse is one of volume replica's states, which indicates the replica is used by a Pod or not
	// +kubebuilder:default:=false
	InUse bool `json:"inuse,omitempty"`
}

// HAState is state for ha replica
type HAState struct {
	// Consistent, Inconsistent, replica is ready only when consistent
	State State `json:"state"`
	// Reason is why this state happened
	Reason string `json:"reason,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalVolumeReplica is the Schema for the volumereplicas API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=localvolumereplicas,scope=Cluster,shortName=lvr
// +kubebuilder:printcolumn:name="capacity",type=integer,JSONPath=`.spec.requiredCapacityBytes`,description="Required capacity of the volume replica"
// +kubebuilder:printcolumn:name="node",type=string,JSONPath=`.spec.nodeName`,description="Node name where the volume replica is located at"
// +kubebuilder:printcolumn:name="state",type=string,JSONPath=`.status.state`,description="State of the volume replica"
// +kubebuilder:printcolumn:name="synced",type=boolean,JSONPath=`.status.synced`,description="Sync status of the volume replica"
// +kubebuilder:printcolumn:name="device",type=string,JSONPath=`.status.devPath`,description="Device path of the volume replica"
// +kubebuilder:printcolumn:name="age",type=date,JSONPath=`.metadata.creationTimestamp`
type LocalVolumeReplica struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LocalVolumeReplicaSpec   `json:"spec,omitempty"`
	Status LocalVolumeReplicaStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalVolumeReplicaList contains a list of LocalVolumeReplica
type LocalVolumeReplicaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LocalVolumeReplica `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LocalVolumeReplica{}, &LocalVolumeReplicaList{})
}
