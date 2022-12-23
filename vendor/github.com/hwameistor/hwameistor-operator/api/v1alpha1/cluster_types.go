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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ClusterSpec defines the desired state of Cluster
type ClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	TargetNamespace string `json:"targetNamespace"`
	// InstallDRBD bool `json:"installDRBD"`

	// LocalDiskManager represents settings about LocalDiskManager
	LocalDiskManager *LocalDiskManagerSpec `json:"localDiskManager,omitempty"`

	// LocalStorage represents settings about LocalStorage
	LocalStorage *LocalStorageSpec `json:"localStorage,omitempty"`

	// Scheduler represents settings about Scheduler
	Scheduler *SchedulerSpec `json:"scheduler,omitempty"`

	// Evictor represents settings about Evictor
	Evictor *EvictorSpec `json:"evictor,omitempty"`

	// AdmissionController represents settings about AdmissionController
	AdmissionController *AdmissionControllerSpec `json:"admissionController,omitempty"`

	ApiServer *ApiServerSpec `json:"apiServer,omitempty"`

	DRBD *DRBDSpec `json:"drbd,omitempty"`

	RBAC *RBACSpec `json:"rbac,omitempty"`
}

type ImageSpec struct {
	Registry string `json:"registry,omitempty"`
	Repository string `json:"repository,omitempty"`
	Tag string `json:"tag,omitempty"`
}

type ContainerCommonSpec struct {
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
	Image *ImageSpec `json:"image,omitempty"`
}

type PodCommonSpec struct {
	PriorityClassName string `json:"priorityClassName,omitempty"`
	NodeAffinity *corev1.NodeAffinity `json:"nodeAffinity,omitempty"`
	PodAffinty *corev1.PodAffinity `json:"podAffinity,omitempty"`
	Tolerations *[]corev1.Toleration `json:"tolerations,omitempty"`
}

type CSIControllerSpec struct {
	Common *PodCommonSpec `json:"common,omitempty"`
	Provisioner *ContainerCommonSpec `json:"provisioner,omitempty"`
	Attacher *ContainerCommonSpec `json:"attacher,omitempty"`
	Resizer *ContainerCommonSpec `json:"resizer,omitempty"`
}

type CSISpec struct {
	Enable bool `json:"enable,omitempty"`
	Registrar *ContainerCommonSpec `json:"registrar,omitempty"`
	Controller *CSIControllerSpec `json:"controller,omitempty"`
}

type LocalDiskManagerSpec struct {
	KubeletRootDir string `json:"kubeletRootDir,omitempty"`
	CSI *CSISpec `json:"csi,omitempty"`
	Common *PodCommonSpec `json:"common,omitempty"`
	Manager *ContainerCommonSpec `json:"manager,omitempty"`
}

type LocalStorageSpec struct {
	KubeletRootDir string `json:"kubeletRootDir,omitempty"`
	CSI *CSISpec `json:"csi,omitempty"`
	Member *MemberSpec `json:"member,omitempty"`
	Common *PodCommonSpec `json:"common,omitempty"`
}

type MemberSpec struct {
	DRBDStartPort int `json:"drbdStartPort,omitempty"`
	MaxHAVolumeCount int `json:"maxHAVolumeCount,omitempty"`
	RcloneImage *ImageSpec `json:"rcloneImage,omitempty"`
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
	Image *ImageSpec `json:"image,omitempty"`
}

type SchedulerSpec struct {
	Replicas int `json:"replicas"`
	Common *PodCommonSpec `json:"common,omitempty"`
	Scheduler *ContainerCommonSpec `json:"scheduler,omitempty"`
}

type EvictorSpec struct {
	Common *PodCommonSpec `json:"common,omitempty"`
	Evictor *ContainerCommonSpec `json:"evictor,omitempty"`
}

type AdmissionControllerSpec struct {
	Replicas int `json:"replicas"`
	FailurePolicy string `json:"failurePolicy,omitempty"`
	Common *PodCommonSpec `json:"common,omitempty"`
	Controller *ContainerCommonSpec `json:"controller,omitempty"`
}

type ApiServerSpec struct {
	Replicas int `json:"replicas"`
	Common *PodCommonSpec `json:"common,omitempty"`
	Server *ContainerCommonSpec `json:"server,omitempty"`
}

type DRBDSpec struct {
	Enable bool `json:"enable,omitempty"`
}

type RBACSpec struct {
	ServiceAccountName string `json:"serviceAccountName"`
}

// ClusterStatus defines the observed state of Cluster
type ClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Phase Phase `json:"phase,omitempty"`

	LocalDiskManager *LocalDiskManagerStatus `json:"localDiskManager"`
	LocalStorage *LocalStorageStatus `json:"localStorage"`
	Scheduler *SchedulerStatus `json:"scheduler"`
	Evictor *EvictorStatus `json:"evictor"`
	AdmissionController *AdmissionControllerStatus `json:"admissionController"`
	ApiServer *ApiServerStatus `json:"apiServer"`
}

type DeployStatus struct {
	Pods []PodStatus `json:"pods"`
	DesiredPodCount int `json:"desiredPodCount"`
	CurrentPodCount int `json:"currentPodCount"`
	ReadyPodCount int `json:"readyPodCount"`
	UpToDatePodCount int `json:"upToDatePodCount"`
	AvailablePodCount int `json:"availablePodCount"`
}

type PodStatus struct {
	Name string `json:"name"`
	Node string `json:"node"`
	Status string `json:"status"`
}

type LocalDiskManagerStatus struct {
	Instances *DeployStatus `json:"instances"`
	CSI *DeployStatus `json:"csi"`
}

type LocalStorageStatus struct {
	Instances *DeployStatus `json:"instances"`
	CSI DeployStatus `json:"csi"`
}

type SchedulerStatus struct {
	Instances *DeployStatus `json:"instances"`
}

type EvictorStatus struct {
	Instances *DeployStatus `json:"instances"`
}

type AdmissionControllerStatus struct {
	Instances *DeployStatus `json:"instances"`
}

type ApiServerStatus struct {
	Instances *DeployStatus `json:"instances"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Cluster is the Schema for the clusters API
// +kubebuilder:resource:path=clusters,scope=Cluster,shortName=hmcluster
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ClusterList contains a list of Cluster
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cluster{}, &ClusterList{})
}
