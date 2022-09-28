package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type ConditionStatus string

// These are valid condition statuses. "ConditionTrue" means a resource is in the condition.
// "ConditionFalse" means a resource is not in the condition. "ConditionUnknown" means kubernetes
// can't decide if a resource is in the condition or not. In the future, we could add other
// intermediate conditions, e.g. ConditionDegraded.
const (
	ConditionTrue    ConditionStatus = "True"
	ConditionFalse   ConditionStatus = "False"
	ConditionUnknown ConditionStatus = "Unknown"
)

type LocalStorageNodeConditionType string

// These are valid conditions of a localstoragenode.
const (
	// StorageAvailable Available means the localstoragenode is available, i.e. the free storage capacity is more than or equal 0
	StorageAvailable LocalStorageNodeConditionType = "Available"
	// StorageUnAvailable UnAvailable means the localstoragenode is unavailable, i.e. the free storage capacity is less than or equal 0
	StorageUnAvailable LocalStorageNodeConditionType = "UnAvailable"
	// StorageProgressing Progressing means the localstoragenode is progressing, i.e. extending storage capacity
	StorageProgressing LocalStorageNodeConditionType = "Progressing"
	// StorageExpandFailure is added in a localstoragenode when a disk fails to be joined the storage pool
	StorageExpandFailure LocalStorageNodeConditionType = "StorageExpandFailure"
	// StorageExpandSuccess is added in a localstoragenode when a disk succeeds to be joined the storage pool
	StorageExpandSuccess LocalStorageNodeConditionType = "StorageExpandSuccess"
)

// LocalStorageNodeSpec defines the desired state of LocalStorageNode
type LocalStorageNodeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster

	HostName string `json:"hostname,omitempty"`

	// IPv4 address is for HA replication traffic
	StorageIP string `json:"storageIP,omitempty"`

	Topo Topology `json:"topogoly,omitempty"`
}

// LocalStorageNodeStatus defines the observed state of LocalStorageNode
type LocalStorageNodeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster

	// There may have multiple storage pools in a node.
	// e.g. HDD_POOL, SSD_POOL, NVMe_POOL
	// Pools: poolName -> LocalPool
	Pools map[string]LocalPool `json:"pools,omitempty"`

	// State of the Local Storage Node/Member: New, Active, Inactive, Failed
	State State `json:"state,omitempty"`

	// Represents the latest available observations of a localstoragenode's current state.
	// +optional
	Conditions []LocalStorageNodeCondition `json:"conditions,omitempty"`
}

// LocalStorageNodeCondition describes the state of a localstoragenode at a certain point.
type LocalStorageNodeCondition struct {
	// Type of localstoragenode condition.
	Type LocalStorageNodeConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status ConditionStatus `json:"status"`
	// The last time this condition was updated.
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`
	// A human-readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
}

// NodeConfig defines local storage system configurations
type NodeConfig struct {
	Name      string    `json:"name,omitempty"`
	StorageIP string    `json:"ip,omitempty"`
	Topology  *Topology `json:"topology,omitempty"`
}

// Topology defines the topology info of Node
type Topology struct {

	// Zone is a collection of Local Storage Nodes
	// +kubebuilder:default:=default
	Zone string `json:"zone,omitempty"`

	// Region is a collection of Zones
	// +kubebuilder:default:=default
	Region string `json:"region,omitempty"`
}

// LocalPool is storage pool struct
type LocalPool struct {
	// Supported pool name: HDD_POOL, SSD_POOL, NVMe_POOL
	Name string `json:"name,omitempty"`

	// Supported class: HDD, SSD, NVMe
	// +kubebuilder:validation:Enum:=HDD;SSD;NVMe
	Class string `json:"class"`

	// Supported type: REGULAR
	// +kubebuilder:validation:Enum:=REGULAR
	// +kubebuilder:default:=REGULAR
	Type string `json:"type"`

	// VG path
	Path string `json:"path,omitempty"`

	TotalCapacityBytes int64 `json:"totalCapacityBytes"`

	UsedCapacityBytes int64 `json:"usedCapacityBytes"`

	VolumeCapacityBytesLimit int64 `json:"volumeCapacityBytesLimit"`

	FreeCapacityBytes int64 `json:"freeCapacityBytes"`

	TotalVolumeCount int64 `json:"totalVolumeCount"`

	UsedVolumeCount int64 `json:"usedVolumeCount"`

	FreeVolumeCount int64 `json:"freeVolumeCount"`

	Disks []LocalDevice `json:"disks,omitempty"`

	Volumes []string `json:"volumes,omitempty"`
}

// LocalDevice is disk struct
type LocalDevice struct {
	// e.g. /dev/sdb
	DevPath string `json:"devPath,omitempty"`

	// Supported: HDD, SSD, NVMe, RAM
	Class string `json:"type,omitempty"`

	// disk capacity
	CapacityBytes int64 `json:"capacityBytes,omitempty"`

	// Possible state: Available, Inuse, Offline
	State State `json:"state,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalStorageNode is the Schema for the localstoragenodes API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=localstoragenodes,scope=Cluster,shortName=lsn
// +kubebuilder:printcolumn:name="ip",type=string,JSONPath=`.spec.storageIP`,description="IPv4 address"
// +kubebuilder:printcolumn:name="status",type=string,JSONPath=`.status.state`,description="State of the Local Storage Node"
// +kubebuilder:printcolumn:name="age",type=date,JSONPath=`.metadata.creationTimestamp`
type LocalStorageNode struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LocalStorageNodeSpec   `json:"spec,omitempty"`
	Status LocalStorageNodeStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalStorageNodeList contains a list of LocalStorageNode
type LocalStorageNodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LocalStorageNode `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LocalStorageNode{}, &LocalStorageNodeList{})
}
