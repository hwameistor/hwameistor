package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	LocalStorage     = "local-storage"
	LocalDiskManager = "local-disk-manager"
	System           = "system"

	// LocalDiskObjectPrefix is used to prefix LocalDisk objects
	LocalDiskObjectPrefix = "localdisk-"
)

type DevLinkType = string

const (
	// LinkByPath is used to identify by-path symbolic link
	LinkByPath DevLinkType = "by-path"

	// LinkByID is used to identify by-id symbolic link
	LinkByID DevLinkType = "by-id"

	// LinkByUUID is used to identify by-uuid symbolic link
	LinkByUUID DevLinkType = "by-uuid"
)

// PartitionInfo contains partition information(e.g. FileSystem)
type PartitionInfo struct {
	// Path represents the partition path in the OS
	Path string `json:"path"`

	// HasFileSystem represents whether the filesystem is included
	HasFileSystem bool `json:"hasFileSystem"`

	// FileSystem contains mount point and filesystem type
	// +optional
	FileSystem FileSystemInfo `json:"filesystem,omitempty"`
}

// RAIDInfo contains infos of raid
type RAIDInfo struct {
	// RAIDMaster is the master of the RAID disk, it works for only RAID slave disk, e.g. /dev/bus/0
	RAIDMaster string `json:"raidMaster,omitempty"`
}

// DiskAttributes represent certain hardware/static attributes of the disk
type DiskAttributes struct {
	// Type is the disk type, such as ata, scsi, nvme, megaraid,N, ...
	Type string `json:"type,omitempty"`

	// DeviceType represents the type of device like
	// sparse, disk, partition, lvm, crypt
	DevType string `json:"devType,omitempty"`

	// Vendor is who provides the disk
	Vendor string `json:"vendor,omitempty"`

	// Product is a class of disks the vendor produces
	Product string `json:"product,omitempty"`

	// PCIVendorID is the ID of the PCI vendor, for NVMe disk only
	PCIVendorID string `json:"pciVendorID,omitempty"`

	// ModelName is the name of disk model
	ModelName string `json:"modelName,omitempty"`

	// SerialNumber is a unique number assigned to a disk
	SerialNumber string `json:"serialNumber,omitempty"`

	// FormFactor is the disk size, like 2.5 inches
	FormFactor string `json:"formFactor,omitempty"`

	// RotationRate is the rate of the disk rotation
	RotationRate int64 `json:"rotationRate,omitempty"`

	// Protocol is for data transport, such as ATA, SCSI, NVMe
	Protocol string `json:"protocol,omitempty"`
}

// FileSystemInfo defines the filesystem type and mountpoint of the disk if it exists
type FileSystemInfo struct {
	// Type represents the FileSystem type of the disk
	// +optional
	Type string `json:"fsType,omitempty"`

	// MountPoint represents the mountpoint of the disk
	// +optional
	Mountpoint string `json:"mountPoint,omitempty"`
}

// SmartInfo contains info collected by smartctl
type SmartInfo struct {
	// OverallHealth identifies if the disk is healthy or not
	OverallHealth SmartAssessResult `json:"overallHealth"`
}

// LocalDiskClaimState defines the observed state of LocalDisk
type LocalDiskClaimState string

const (
	// LocalDiskUnclaimed represents that the disk is not bound to any LDC,
	// and is available for claiming.
	LocalDiskUnclaimed LocalDiskClaimState = "Available"

	// LocalDiskReleased represents that the disk is released from the LDC,
	LocalDiskReleased LocalDiskClaimState = "Released"

	// LocalDiskClaimed represents that the disk is bound to a LDC
	LocalDiskClaimed LocalDiskClaimState = "Claimed"

	// LocalDiskInUse represents that the disk is in use but not claimed by a LDC
	LocalDiskInUse LocalDiskClaimState = "Inuse"
)

// LocalDiskState defines the observed state of the local disk
type LocalDiskState string

// NOTE: The follow-up state represent disk health status detected by
// system or health check tools(e.g., smartctl)
const (
	// LocalDiskActive is the state for the disk that is connected
	LocalDiskActive LocalDiskState = "Active"

	// LocalDiskInactive is the state for the disk that is disconnected
	LocalDiskInactive LocalDiskState = "Inactive"

	// LocalDiskUnknown is the state for the disk that cannot be determined
	// at this time(whether attached or detached)
	LocalDiskUnknown LocalDiskState = "Unknown"
)

// NOTE: The follow-up state represent LocalDisk instance status
const (
	// LocalDiskEmpty is temporary status, it can be updated to Available or Bound
	LocalDiskEmpty LocalDiskState = ""

	// LocalDiskPending is temporary status, it can be updated to Available or Bound
	LocalDiskPending LocalDiskState = "Pending"

	// LocalDiskAvailable represents the disk can be used which means:
	// 1) there is no filesystem or partitions exist
	// 2) the disk is not bound to any LocalDiskClaim object
	LocalDiskAvailable LocalDiskState = "Available"

	// LocalDiskBound represents the disk is used already.
	// There are follow-up use cases:
	// 1) used by system (e.g., rootfs)
	// 2) used by a LocalDiskClaim object
	// 3) there is already a filesystem or partition exist
	LocalDiskBound LocalDiskState = "Bound"
)

// SmartAssessResult defines the result of self-assessment test
type SmartAssessResult string

const (
	// AssessPassed indicates the disk is healthy
	AssessPassed SmartAssessResult = "Passed"

	// AssessFailed indicates the disk is unhealthy
	AssessFailed SmartAssessResult = "Failed"
)

// LocalDiskSpec defines the desired state of LocalDisk
type LocalDiskSpec struct {
	// NodeName represents the node where the disk is attached
	// +kubebuilder:validation:Required
	NodeName string `json:"nodeName"`

	// PreNodeName represents the node where the disk was attached
	PreNodeName string `json:"preNodeName,omitempty"`

	// UUID global unique identifier of the disk
	UUID string `json:"uuid,omitempty"`

	// DevicePath is the disk path in the OS
	DevicePath string `json:"devicePath,omitempty"`

	// PreDevicePath represents the last device path in the OS
	PreDevicePath string `json:"preDevicePath,omitempty"`

	// DevLinks are symbol links for this device
	DevLinks []string `json:"devLinks"`

	// Capacity of the disk in bytes
	Capacity int64 `json:"capacity,omitempty"`

	// HasPartition represents if the disk has partitions or not
	HasPartition bool `json:"partitioned,omitempty"`

	// PartitionInfo contains partition information
	// +optional
	PartitionInfo []PartitionInfo `json:"partitionInfo,omitempty"`

	// HasRAID identifies if the disk is a raid disk or not
	HasRAID bool `json:"isRaid,omitempty"`

	// RAIDInfo contains RAID information
	// +optional
	RAIDInfo RAIDInfo `json:"raidInfo,omitempty"`

	// HasSmartInfo identified if the disk supports SMART or not
	HasSmartInfo bool `json:"supportSmart,omitempty"`

	// SmartInfo contains infos collected by smartctl
	// +optional
	SmartInfo SmartInfo `json:"smartInfo,omitempty"`

	// DiskAttributes has hardware/static attributes of the disk
	DiskAttributes DiskAttributes `json:"diskAttributes,omitempty"`

	// State is the current state of the disk (Active/Inactive/Unknown)
	// +kubebuilder:validation:Enum:=Active;Inactive;Unknown
	State LocalDiskState `json:"state,omitempty"`

	// ClaimRef is the reference to the LDC which has claimed this LD
	// +optional
	ClaimRef *v1.ObjectReference `json:"claimRef,omitempty"`

	// Reserved represents the disk won't be used in hwameistor later, until it becomes unreserved
	// +optional
	Reserved bool `json:"reserved,omitempty"`

	// Owner represents which system owns this claim(e.g. local-storage, local-disk-manager)
	// +optional
	Owner string `json:"owner,omitempty"`

	// Major represents drive used by the device
	// +optional
	Major string `json:"major,omitempty"`

	// Minor is used to distinguish different devices
	// +optional
	Minor string `json:"minor,omitempty"`
}

// LocalDiskStatus defines the observed state of LocalDisk
type LocalDiskStatus struct {
	// State represents the claim state of the disk
	// +kubebuilder:validation:Enum:=Bound;Reserved;Available;Pending
	State LocalDiskState `json:"claimState,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalDisk is the Schema for the localdisks API
// +kubebuilder:resource:scope=Cluster,shortName=ld
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".spec.nodeName",name=NodeMatch,type=string
// +kubebuilder:printcolumn:JSONPath=".spec.capacity",name=Capacity,type=integer,priority=1
// +kubebuilder:printcolumn:JSONPath=".spec.preNodeName",name=PreNodeMatch,type=string,priority=1
// +kubebuilder:printcolumn:JSONPath=".spec.devicePath",name=DevicePath,type=string
// +kubebuilder:printcolumn:JSONPath=".spec.preDevicePath",name=PreDevicePath,type=string,priority=1
// +kubebuilder:printcolumn:JSONPath=".spec.owner",name=Owner,type=string
// +kubebuilder:printcolumn:JSONPath=".status.claimState",name=Phase,type=string
// +kubebuilder:printcolumn:JSONPath=".spec.smartInfo.overallHealth",name=Health,type=string,priority=1
// +kubebuilder:printcolumn:JSONPath=".spec.reserved",name=Reserved,type=boolean,priority=1
// +kubebuilder:printcolumn:JSONPath=".spec.state",name=State,type=string
// +kubebuilder:printcolumn:name="age",type=date,JSONPath=`.metadata.creationTimestamp`
type LocalDisk struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LocalDiskSpec   `json:"spec,omitempty"`
	Status LocalDiskStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalDiskList contains a list of LocalDisk
type LocalDiskList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LocalDisk `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LocalDisk{}, &LocalDiskList{})
}
