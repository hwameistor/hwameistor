package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// EventSpec defines the desired state of Event
type EventSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// HwameiStor resource type: Cluster, LocalStorageNode, LocalDiskNode, Pool,  LocalVolume, LocalDiskVolume, LocalDisk,
	// +kubebuilder:validation:Enum:=Cluster;StorageNode;DiskNode;Pool;Volume;DiskVolume;Disk
	ResourceType string `json:"resourceType"`

	// Name of the resource
	ResourceName string `json:"resourceName"`

	// Which node does the resource reside in
	// NodeName string `json:"nodeName"`

	Records []EventRecord `json:"records"`
}

type EventRecord struct {
	// The time when does the action happen
	Time metav1.Time `json:"time,omitempty"`

	// id is unique
	ID string `json:"id,omitempty"`

	// The action is the operation on the resource, such as Migrate a LocalVolume
	Action string `json:"action,omitempty"`
	// The content of the action which is a JSON string
	ActionContent string `json:"actionContent,omitempty"`

	// The state of the action
	State string `json:"state,omitempty"`
	// The content of the action state which is a JSON string
	StateContent string `json:"stateContent,omitempty"`
}

// EventStatus defines the observed state of Event
type EventStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Event is the Schema for the events API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=events,scope=Cluster,shortName=evt
// +kubebuilder:printcolumn:JSONPath=".spec.resourceType",name=type,type=string
// +kubebuilder:printcolumn:JSONPath=".spec.resourceName",name=resource,type=string
type Event struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EventSpec   `json:"spec,omitempty"`
	Status EventStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// EventList contains a list of Event
type EventList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Event `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Event{}, &EventList{})
}
