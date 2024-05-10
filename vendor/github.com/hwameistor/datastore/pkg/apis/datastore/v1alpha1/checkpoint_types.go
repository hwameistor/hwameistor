package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CheckpointSpec defines the desired state of Checkpoint
type CheckpointSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// +kubebuilder:default:=2
	RetainCount int64 `json:"retain"`

	Backup *CheckpointBackup `json:"backup,omitempty"`
}

// CheckpointStatus defines the observed state of Checkpoint
type CheckpointStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	Records []*CheckpointRecord `json:"records"`
}

type CheckpointBackup struct {
	// +kubebuilder:default:=minio
	Proto string     `json:"proto"`
	MinIO *MinIOSpec `json:"minio,omitempty"`
	//	NFS   *NFSSpec   `json:"nfs,omitempty"`
	// FTP   *FTPSpec   `json:"ftp,omitempty"`
	// HTTP  *HTTPSpec  `json:"http,omitempty"`
}

type CheckpointRecord struct {
	Name      string `json:"name"`
	NodeName  string `json:"node"`
	Checksum  string `json:"checksum"`
	DirOnHost string `json:"hostdir"`
	// +kubebuilder:default:=false
	BackedUp    bool         `json:"backedup"`
	Size        string       `json:"size"`
	CreateTime  *metav1.Time `json:"created"`
	ExpiredTime *metav1.Time `json:"expired,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Checkpoint is the Schema for the checkpoints API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=checkpoints,scope=Namespaced,shortName=ckpt
type Checkpoint struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CheckpointSpec   `json:"spec,omitempty"`
	Status CheckpointStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CheckpointList contains a list of Checkpoint
type CheckpointList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Checkpoint `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Checkpoint{}, &CheckpointList{})
}
