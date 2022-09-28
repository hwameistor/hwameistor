package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LocalDiskNodeSpec defines the desired state of LocalDiskNode
type LocalDiskNodeSpec struct {
	// AttachNode represent where disk is attached
	AttachNode string `json:"attachNode"`
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
	// Disks key is the name of LocalDisk
	Disks map[string]Disk `json:"disks,omitempty"`

	// TotalDisk
	TotalDisk int64 `json:"totalDisk,omitempty"`

	// AllocatableDisk
	AllocatableDisk int64 `json:"allocatableDisk,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalDiskNode is the Schema for the localdisknodes API
// +kubebuilder:resource:path=localdisknodes,scope=Cluster,shortName=ldn
//+kubebuilder:printcolumn:JSONPath=".status.totalDisk",name=TotalDisk,type=integer
//+kubebuilder:printcolumn:JSONPath=".status.allocatableDisk",name=FreeDisk,type=integer
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
