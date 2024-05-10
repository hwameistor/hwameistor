package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// BaseModelSpec defines the desired state of BaseModel
type BaseModelSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	ModelFileName string `json:"modelName"`

	// +kubebuilder:default:=minio
	Proto string     `json:"proto"`
	MinIO *MinIOSpec `json:"minio,omitempty"`
	NFS   *NFSSpec   `json:"nfs,omitempty"`
	FTP   *FTPSpec   `json:"ftp,omitempty"`
	HTTP  *HTTPSpec  `json:"http,omitempty"`
}

// BaseModelStatus defines the observed state of BaseModel
type BaseModelStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BaseModel is the Schema for the basemodels API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=basemodels,scope=Namespaced,shortName=bm
type BaseModel struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BaseModelSpec   `json:"spec,omitempty"`
	Status BaseModelStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BaseModelList contains a list of BaseModel
type BaseModelList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BaseModel `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BaseModel{}, &BaseModelList{})
}
