package v1alpha1

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// State is state type of resources
type State string

const (
	// blow state is for mountpoint
	MountPointStateEmpty  State = ""
	MountPointToBeMounted State = "ToBeMounted"
	MountPointToBeUnMount State = "ToBeUnMount"
	MountPointMounted     State = "Mounted"
	MountPointNotReady    State = "NotReady"

	// blow state is for volume crd
	VolumeStateEmpty       State = ""
	VolumeStateCreated     State = "Created"
	VolumeStateReady       State = "Ready"
	VolumeStateNotReady    State = "NotReady"
	VolumeStateToBeUnmount State = "ToBeUnmount"
	VolumeStateToBeDeleted State = "ToBeDeleted"
	VolumeStateDeleted     State = "Deleted"
)

// LocalDiskVolumeSpec defines the desired state of LocalDiskVolume
type LocalDiskVolumeSpec struct {
	// DiskType represents the type of drive like SSD, HDD etc.,
	DiskType string `json:"diskType"`

	// RequiredCapacityBytes
	RequiredCapacityBytes int64 `json:"requiredCapacityBytes,omitempty"`

	// Accessibility is the topology requirement of the volume. It describes how to locate and distribute the volume replicas
	Accessibility AccessibilityTopology `json:"accessibility,omitempty"`

	// PersistentVolumeClaimName is the reference of the associated PVC
	PersistentVolumeClaimName string `json:"persistentVolumeClaimName,omitempty"`
}

// MountPoint
type MountPoint struct {
	// TargetPath
	TargetPath string `json:"targetPath,omitempty"`

	// VolumeCap
	VolumeCap VolumeCapability `json:"volumeCap,omitempty"`

	// FsTye
	FsTye string `json:"fsTye,omitempty"`

	// MountOptions
	MountOptions []string `json:"mountOptions,omitempty"`

	// Phase indicates the volume's next or current operation
	Phase State `json:"phase,omitempty"`
}

type VolumeAccessMode int32

const (
	// Can only be published once as read/write on a single node, at
	// any given time.
	VolumeCapability_AccessMode_SINGLE_NODE_WRITER = csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER
)

type VolumeAccessType string

const (
	VolumeCapability_AccessType_Block = "Block"
	VolumeCapability_AccessType_Mount = "Mount"
)

type VolumeCapability struct {
	AccessMode VolumeAccessMode `json:"accessMode,omitempty"`
	AccessType VolumeAccessType `json:"accessType,omitempty"`
}

// AccessibilityTopology of the volume
type AccessibilityTopology struct {
	// One of the volume replica must be located at this node
	Node string `json:"node,omitempty"`

	// zones where the volume replicas should be distributed across, it's Optional
	// +kubebuilder:default:={default}
	Zones []string `json:"zones,omitempty"`

	// regions where the volume replicas should be distributed across, it's Optional
	// +kubebuilder:default:={default}
	Regions []string `json:"regions,omitempty"`
}

// LocalDiskVolumeStatus defines the observed state of LocalDiskVolume
type LocalDiskVolumeStatus struct {
	// LocalDiskName is disk name which is used to create this volume
	LocalDiskName string `json:"localDiskName,omitempty"`

	// DevPath is the disk path in the OS
	DevPath string `json:"devPath"`

	// AllocatedCapacityBytes is the real allocated capacity in bytes
	AllocatedCapacityBytes int64 `json:"allocatedCapacityBytes,omitempty"`

	// MountPoints
	MountPoints []MountPoint `json:"mountPoints,omitempty"`

	// State is the phase of volume replica, e.g. Creating, Ready, NotReady, ToBeDeleted, Deleted
	State State `json:"state,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalDiskVolume is the Schema for the localdiskvolumes API
// +kubebuilder:resource:path=localdiskvolumes,scope=Cluster,shortName=ldv
//+kubebuilder:printcolumn:JSONPath=".spec.accessibility.node",name=Node,type=string
//+kubebuilder:printcolumn:JSONPath=".status.devPath",name=Disk,type=string
//+kubebuilder:printcolumn:JSONPath=".status.allocatedCapacityBytes",name=AllocatedCap,type=integer
//+kubebuilder:printcolumn:JSONPath=".spec.diskType",name=Type,type=string
//+kubebuilder:printcolumn:JSONPath=".status.state",name=Status,type=string
type LocalDiskVolume struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LocalDiskVolumeSpec   `json:"spec,omitempty"`
	Status LocalDiskVolumeStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalDiskVolumeList contains a list of LocalDiskVolume
type LocalDiskVolumeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LocalDiskVolume `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LocalDiskVolume{}, &LocalDiskVolumeList{})
}
