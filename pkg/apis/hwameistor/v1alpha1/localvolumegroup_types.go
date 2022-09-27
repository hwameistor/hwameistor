package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LocalVolumeGroupSpec defines the desired state of LocalVolumeGroup
type LocalVolumeGroupSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// Volumes is the collection of the volumes in the group
	Volumes []VolumeInfo `json:"volumes,omitempty"`

	// Accessibility is the topology requirement of the volume. It describes how to locate and distribute the volume replicas
	Accessibility AccessibilityTopology `json:"accessibility,omitempty"`

	Pods []string `json:"pods,omitempty"`
}

type VolumeInfo struct {
	// LocalVolumeName is the name of the LocalVolume
	LocalVolumeName string `json:"localvolume,omitempty"`

	// PersistentVolumeClaimName is the name of the associated PVC
	PersistentVolumeClaimName string `json:"pvc,omitempty"`
}

// LocalVolumeGroupStatus defines the observed state of LocalVolumeGroup
type LocalVolumeGroupStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalVolumeGroup is the Schema for the localvolumegroups API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=localvolumegroups,scope=Cluster,shortName=lvg
// +kubebuilder:printcolumn:name="pod",type=string,JSONPath=`.spec.pods`,description="Name of associated pod"
// +kubebuilder:printcolumn:name="namespace",type=string,JSONPath=`.spec.namespace`,description="Namespace of associated pod"
// +kubebuilder:printcolumn:name="age",type=date,JSONPath=`.metadata.creationTimestamp`
type LocalVolumeGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LocalVolumeGroupSpec   `json:"spec,omitempty"`
	Status LocalVolumeGroupStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalVolumeGroupList contains a list of LocalVolumeGroup
type LocalVolumeGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LocalVolumeGroup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LocalVolumeGroup{}, &LocalVolumeGroupList{})
}
