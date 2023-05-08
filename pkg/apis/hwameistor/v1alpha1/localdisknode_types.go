package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LocalDiskNodeSpec defines the desired state of LocalDiskNode
type LocalDiskNodeSpec struct {
	// NodeName represent where disk is attached
	NodeName string `json:"nodeName"`
}

type Disk struct {
	// DevPath
	DevPath string `json:"devPath"`

	// Capacity
	Capacity int64 `json:"capacity,omitempty"`

	// DiskType SSD/HDD/NVME...
	DiskType string `json:"diskType"`

	// Status
	Status string `json:"status"`
}

// LocalDiskNodeStatus defines the observed state of LocalDiskNode
type LocalDiskNodeStatus struct {
	// There may have multiple storage pools in a node.
	// e.g. HDD_POOL, SSD_POOL, NVMe_POOL
	// Pools: poolName -> LocalPool
	Pools map[string]LocalPool `json:"pools,omitempty"`

	// State of the Local Storage Node/Member: New, Active, Inactive, Failed
	State State `json:"state,omitempty"`

	// Represents the latest available observations of a localstoragenode's current state.
	// +optional
	Conditions []StorageNodeCondition `json:"conditions,omitempty"`

	// PoolExtendRecords record why disks are joined in the pool
	// +optional
	PoolExtendRecords map[string]LocalDiskClaimSpecArray `json:"poolExtendRecords,omitempty"`

	// TotalDisk
	TotalDisk int64 `json:"totalDisk,omitempty"`

	// FreeDisk
	FreeDisk int64 `json:"freeDisk,omitempty"`

	// TotalCapacity indicates the capacity of all the disks
	TotalCapacity int64 `json:"totalCapacity,omitempty"`

	// FreeCapacity indicates the free capacity of all the disks
	FreeCapacity int64 `json:"freeCapacity,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalDiskNode is the Schema for the localdisknodes API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=localdisknodes,scope=Cluster,shortName=ldn
// +kubebuilder:printcolumn:JSONPath=".status.freeCapacity",name=FreeCapacity,type=integer
// +kubebuilder:printcolumn:JSONPath=".status.totalCapacity",name=TotalCapacity,type=integer
// +kubebuilder:printcolumn:JSONPath=".status.totalDisk",name=TotalDisk,type=integer
// +kubebuilder:printcolumn:name="status",type=string,JSONPath=`.status.state`,description="State of the LocalDisk Node"
// +kubebuilder:printcolumn:name="age",type=date,JSONPath=`.metadata.creationTimestamp`
type LocalDiskNode struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LocalDiskNodeSpec   `json:"spec,omitempty"`
	Status LocalDiskNodeStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalDiskNodeList contains a list of LocalDiskNode
type LocalDiskNodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LocalDiskNode `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LocalDiskNode{}, &LocalDiskNodeList{})
}
