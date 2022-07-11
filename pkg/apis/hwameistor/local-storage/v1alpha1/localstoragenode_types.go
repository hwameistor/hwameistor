package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

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

	Disks []LocalDisk `json:"disks,omitempty"`

	Volumes []string `json:"volumes,omitempty"`
}

// LocalDisk is disk struct
type LocalDisk struct {
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
// +kubebuilder:printcolumn:name="volumekind",type=string,JSONPath=`.spec.allowedVolumeKind`,description="volume kind"
// +kubebuilder:printcolumn:name="ramdiskQuota",type=integer,JSONPath=`.spec.allowedRAMDiskTotalCapacityBytes`,description="total storage space of ramdisk"
// +kubebuilder:printcolumn:name="zone",type=string,JSONPath=`.spec.topogoly.zone`,description="Zone of the node"
// +kubebuilder:printcolumn:name="region",type=string,JSONPath=`.spec.topogoly.region`,description="Region of the node"
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
