package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PhysicalDiskSpec defines the desired state of PhysicalDisk
type PhysicalDiskSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// Node name of where the disk is attached
	NodeName string `json:"nodeName,omitempty"`
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
	// DevicePath is the path in the OS
	DevicePath string `json:"devicePath,omitempty"`
	// Protocol is for data transport, such as ATA, SCSI, NVMe
	Protocol string `json:"protocol,omitempty"`
	// Type is the disk type, such as ata, scsi, nvme, megaraid,N, ...
	Type string `json:"type,omitempty"`
	// Capacity of the disk
	Capacity int64 `json:"capacity,omitempty"`
	// IsRAID identifies if the disk is a raid disk or not
	IsRAID bool `json:"isRaid,omitempty"`
	// RAIDMaster is the master of the RAID disk, it works for only RAID slave disk, e.g. /dev/bus/0
	RAIDMaster string `json:"raidMaster,omitempty"`
	// SmartSupport identified if the disk supports SMART or not
	SmartSupport bool `json:"smartSupport,omitempty"`
}

// PhysicalDiskStatus defines the observed state of PhysicalDisk
type PhysicalDiskStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// disk is online or offline. Considering the disk replacement, the replaced disk should be offline
	Online bool `json:"online,omitempty"`

	SmartCheck *SmartCheck `json:"smartCheck,omitempty"`
}

// SmartCheck defines the result of the disk smartctl
type SmartCheck struct {
	// details of the health check by smartctl
	SmartStatus *PhyDiskSmartStatus       `json:"smart_status,omitempty"`
	Temperature *PhyDiskTemperatureStatus `json:"temperature,omitempty"`
	PowerOnTime *PhyDiskPowerOnTimeStatus `json:"power_on_time,omitempty"`

	NVMeSmartHealthStatus *NVMeSmartHealthDetailsInfo `json:"nvme_smart_health_information_log,omitempty"`
	ATASmartHealthStatus  *ATASmartHealthDetailsInfo  `json:"ata_smart_attributes,omitempty"`
	SCSISmartHealthStatus *SCSISmartHealthDetailsInfo `json:"scsi_error_counter_log,omitempty"`

	// latest time for health check
	LastTime *metav1.Time `json:"lastTime,omitempty"`
}

// PhyDiskSmartStatus struct
type PhyDiskSmartStatus struct {
	Passed bool `json:"passed,omitempty"`
}

// PhyDiskTemperatureStatus struct
type PhyDiskTemperatureStatus struct {
	Current int64 `json:"current,omitempty"`
}

// PhyDiskPowerOnTimeStatus struct
type PhyDiskPowerOnTimeStatus struct {
	Hours   int64 `json:"hours,omitempty"`
	Minutes int64 `json:"minutes,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PhysicalDisk is the Schema for the physicaldisks API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=physicaldisks,scope=Cluster,shortName=pd
// +kubebuilder:printcolumn:name="node",type=string,JSONPath=`.spec.nodeName`,description="Node name where the volume replica is located at"
// +kubebuilder:printcolumn:name="serialNumber",type=string,JSONPath=`.spec.serialNumber`,description="Serial number of the disk"
// +kubebuilder:printcolumn:name="modelName",type=string,JSONPath=`.spec.modelName`,description="Model name of the disk"
// +kubebuilder:printcolumn:name="device",type=string,JSONPath=`.spec.devicePath`,description="Disk path in OS"
// +kubebuilder:printcolumn:name="type",type=string,JSONPath=`.spec.type`,description="Disk type"
// +kubebuilder:printcolumn:name="protocol",type=string,JSONPath=`.spec.protocol`,description="Disk access protocol"
// +kubebuilder:printcolumn:name="health",type=boolean,JSONPath=`.status.smartCheck.details.smart_status.passed`,description="Disk health reported by smartctl"
// +kubebuilder:printcolumn:name="checkTime",type=date,JSONPath=`.status.smartCheck.lastTime`,description="Last time to check disk health"
// +kubebuilder:printcolumn:name="online",type=boolean,JSONPath=`.status.online`,description="Disk online or offline"
// +kubebuilder:printcolumn:name="age",type=date,JSONPath=`.metadata.creationTimestamp`
type PhysicalDisk struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PhysicalDiskSpec   `json:"spec,omitempty"`
	Status PhysicalDiskStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PhysicalDiskList contains a list of PhysicalDisk
type PhysicalDiskList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PhysicalDisk `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PhysicalDisk{}, &PhysicalDiskList{})
}
