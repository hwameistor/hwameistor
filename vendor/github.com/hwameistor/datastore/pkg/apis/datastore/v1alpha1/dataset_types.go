package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DataSetSpec defines the desired state of DataSet
type DataSetSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// type of data source
	// +kubebuilder:validation:Enum:=minio;aws-s3;nfs;ftp;unknown
	// +kubebuilder:default:=unknown
	Type string `json:"type"`

	MinIO *MinIOSpec `json:"minio,omitempty"`

	NFS *NFSSpec `json:"nfs,omitempty"`

	FTP *FTPSpec `json:"ftp,omitempty"`

	// indicate if refresh data or not
	// +kubebuilder:default:=true
	Refresh bool `json:"refresh"`
}

// DataSetStatus defines the observed state of DataSet
type DataSetStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// is the data source connected or accessible?
	// +kubebuilder:default:=false
	Connected bool `json:"connected"`

	// how many time the data of backend have been refreshed
	// +kubebuilder:default:=0
	RefreshCount int `json:"refreshCount"`

	// last time to refresh the data from the backend
	LastRefreshTimestamp *metav1.Time `json:"lastRefreshTime,omitempty"`

	// any error message
	Error string `json:"error,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DataSet is the Schema for the DataSets API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=datasets,scope=Namespaced,shortName=dsrc
// +kubebuilder:printcolumn:JSONPath=".spec.type",name=Type,type=string
// +kubebuilder:printcolumn:JSONPath=".status.lastRefreshTime",name=LastRefreshTime,type=date
// +kubebuilder:printcolumn:JSONPath=".spec.refresh",name=Refresh,type=boolean,priority=1
// +kubebuilder:printcolumn:JSONPath=".status.refreshCount",name=RefreshCount,type=integer,priority=1
// +kubebuilder:printcolumn:JSONPath=".status.connected",name=Connected,type=boolean
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date
// +kubebuilder:printcolumn:JSONPath=".status.error",name=Error,type=string
type DataSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DataSetSpec   `json:"spec,omitempty"`
	Status DataSetStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DataSetList contains a list of DataSet
type DataSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DataSet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DataSet{}, &DataSetList{})
}
