/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LocalDiskClaimSpec defines the desired state of LocalDiskClaim
type LocalDiskClaimSpec struct {
	// NodeName represents where disk has to be claimed.
	NodeName string `json:"nodeName,omitempty"`

	// Description of the disk to be claimed
	// +optional
	Description DiskClaimDescription `json:"description,omitempty"`

	// DiskRefs represents which disks are assigned to the LocalDiskClaim
	DiskRefs []*v1.ObjectReference `json:"diskRefs,omitempty"`
}

// LocalDiskClaimStatus defines the observed state of LocalDiskClaim
type LocalDiskClaimStatus struct {
	// Status represents the current statue of the claim
	Status DiskClaimStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=ldc

// LocalDiskClaim is the Schema for the localdiskclaims API
type LocalDiskClaim struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LocalDiskClaimSpec   `json:"spec,omitempty"`
	Status LocalDiskClaimStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:scope=Namespaced,shortName=ldc

// LocalDiskClaimList contains a list of LocalDiskClaim
type LocalDiskClaimList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LocalDiskClaim `json:"items,omitempty"`
}

// DiskClaimDescription defines the details of the disk that should be claimed
type DiskClaimDescription struct {
	// DiskType represents the type of drive like SSD, HDD etc.,
	// +optional
	DiskType string `json:"diskType,omitempty"`
}

// DiskClaimStatus is a typed string for phase field of BlockDeviceClaim.
type DiskClaimStatus string

// LocalDiskClaim CR, when created pass through phases before it got some Disks Assigned.
const (
	// LocalDiskClaimStatusEmpty represents that the LocalDiskClaim was just created.
	DiskClaimStatusEmpty DiskClaimStatus = ""

	// LocalDiskClaimStatusPending represents LocalDiskClaim has not been assigned devices yet. Rather
	// search is going on for matching disks.
	LocalDiskClaimStatusPending DiskClaimStatus = "Pending"

	// LocalDiskClaimStatusDone represents LocalDiskClaim has been assigned backing disk and ready for use.
	LocalDiskClaimStatusDone DiskClaimStatus = "Bound"
)

func init() {
	SchemeBuilder.Register(&LocalDiskClaim{}, &LocalDiskClaimList{})
}
