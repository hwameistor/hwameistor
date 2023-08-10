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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ClusterSpec defines the desired state of Cluster
type ClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	TargetNamespace string `json:"targetNamespace,omitempty"`

	DiskReserveConfigurations []DiskReserveConfiguration `json:"diskReserveConfigurations,omitempty"`

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

	Exporter *ExporterSpec `json:"exporter,omitempty"`

	UI *UISpec `json:"ui,omitempty"`

	DRBD *DRBDSpec `json:"drbd,omitempty"`

	RBAC *RBACSpec `json:"rbac,omitempty"`

	StorageClass *StorageClassSpec `json:"storageClass,omitempty"`
}

type DiskReserveConfiguration struct {
	NodeName string `json:"nodeName"`
	Devices []string `json:"devices"`
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
	Replicas int32 `json:"replicas,omitempty"`
	Common *PodCommonSpec `json:"common,omitempty"`
	Provisioner *ContainerCommonSpec `json:"provisioner,omitempty"`
	Attacher *ContainerCommonSpec `json:"attacher,omitempty"`
	Monitor *ContainerCommonSpec `json:"monitor,omitempty"`
	Resizer *ContainerCommonSpec `json:"resizer,omitempty"`
}

type CSISpec struct {
	Registrar *ContainerCommonSpec `json:"registrar,omitempty"`
	Controller *CSIControllerSpec `json:"controller,omitempty"`
}

type LocalDiskManagerSpec struct {
	KubeletRootDir string `json:"kubeletRootDir,omitempty"`
	CSI *CSISpec `json:"csi,omitempty"`
	Common *PodCommonSpec `json:"common,omitempty"`
	Manager *ContainerCommonSpec `json:"manager,omitempty"`
	TolerationOnMaster bool `json:"tolerationOnMaster,omitempty"`
}

type LocalStorageSpec struct {
	Disable bool `json:"disable,omitempty"`
	KubeletRootDir string `json:"kubeletRootDir,omitempty"`
	CSI *CSISpec `json:"csi,omitempty"`
	Member *MemberSpec `json:"member,omitempty"`
	Common *PodCommonSpec `json:"common,omitempty"`
	TolerationOnMaster bool `json:"tolerationOnMaster,omitempty"`
}

type MemberSpec struct {
	DRBDStartPort int `json:"drbdStartPort,omitempty"`
	MaxHAVolumeCount int `json:"maxHAVolumeCount,omitempty"`
	RcloneImage *ImageSpec `json:"rcloneImage,omitempty"`
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
	Image *ImageSpec `json:"image,omitempty"`
}

type SchedulerSpec struct {
	Disable bool `json:"disable,omitempty"`
	Replicas int32 `json:"replicas"`
	Common *PodCommonSpec `json:"common,omitempty"`
	Scheduler *ContainerCommonSpec `json:"scheduler,omitempty"`
}

type EvictorSpec struct {
	Disable bool `json:"disable,omitempty"`
	Replicas int32 `json:"replicas"`
	Common *PodCommonSpec `json:"common,omitempty"`
	Evictor *ContainerCommonSpec `json:"evictor,omitempty"`
}

type AdmissionControllerSpec struct {
	Disable bool `json:"disable,omitempty"`
	Replicas int32 `json:"replicas,omitempty"`
	FailurePolicy string `json:"failurePolicy,omitempty"`
	Common *PodCommonSpec `json:"common,omitempty"`
	Controller *ContainerCommonSpec `json:"controller,omitempty"`
}

type Authentication struct {
	Enable bool `json:"enable,omitempty"`
	AccessId string `json:"accessId,omitempty"`
	SecretKey string `json:"secretKey,omitempty"`
}

type ApiServerSpec struct {
	Disable bool `json:"disable,omitempty"`
	Replicas int32 `json:"replicas,omitempty"`
	Common *PodCommonSpec `json:"common,omitempty"`
	Server *ContainerCommonSpec `json:"server,omitempty"`
	Authentication *Authentication `json:"authentication,omitempty"`
}

type ExporterSpec struct {
	Disable bool `json:"disable,omitempty"`
	Replicas int32 `json:"replicas,omitempty"`
	Common *PodCommonSpec `json:"common,omitempty"`
	Collector *ContainerCommonSpec `json:"collector,omitempty"`
}

type UISpec struct {
	Disable bool `json:"disable,omitempty"`
	Replicas int32 `json:"replicas"`
	Common *PodCommonSpec `json:"common,omitempty"`
	UI *ContainerCommonSpec `json:"ui,omitempty"`
}

type DRBDSpec struct {
	Disable bool `json:"disable,omitempty"`
	DeployOnMaster string `json:"deployOnMaster,omitempty"`
	ImageRegistry string `json:"imageRegistry,omitempty"`
	ImageRepoOwner string `json:"imageRepoOwner,omitempty"`
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`
	DRBDVersion string `json:"drbdVersion,omitempty"`
	Upgrade string `json:"upgrade,omitempty"`
	CheckHostName string `json:"checkHostName,omitempty"`
	UseAffinity string `json:"useAffinity,omitempty"`
	NodeAffinity *corev1.NodeAffinity `json:"nodeAffinity,omitempty"`
	ChartVersion string `json:"chartVersion,omitempty"`
}

type RBACSpec struct {
	ServiceAccountName string `json:"serviceAccountName"`
}

type StorageClassSpec struct {
	Disable bool `json:"disable,omitempty"`
	AllowVolumeExpansion bool `json:"allowVolumeExpansion,omitempty"`
	ReclaimPolicy corev1.PersistentVolumeReclaimPolicy `json:"reclaimPolicy,omitempty"`
	FSType string `json:"fsType,omitempty"`
}

// ClusterStatus defines the observed state of Cluster
type ClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	
	InstalledCRDS bool `json:"installedCRDS"`
	DRBDAdapterCreated bool `json:"drbdAdapterCreated"`
	DRBDAdapterCreatedJobNum int `json:"drbdAdapterCreatedJobNum"`
	DiskReserveState string `json:"diskReserveState,omitempty"`
	Phase string `json:"phase,omitempty"`
	ComponentStatus ComponentStatus `json:"componentStatus"`
}

type ComponentStatus struct {
	LocalDiskManager *LocalDiskManagerStatus `json:"localDiskManager,omitempty"`
	LocalStorage *LocalStorageStatus `json:"localStorage,omitempty"`
	Scheduler *SchedulerStatus `json:"scheduler,omitempty"`
	Evictor *EvictorStatus `json:"evictor,omitempty"`
	AdmissionController *AdmissionControllerStatus `json:"admissionController,omitempty"`
	ApiServer *ApiServerStatus `json:"apiServer,omitempty"`
	Exporter *ExporterStatus 	`json:"exporter,omitempty"`
}

type DeployStatus struct {
	Pods []PodStatus `json:"pods,omitempty"`
	DesiredPodCount int32 `json:"desiredPodCount"`
	AvailablePodCount int32 `json:"availablePodCount"`
	WorkloadType string `json:"workloadType"`
	WorkloadName string `json:"workloadName"`
}

type PodStatus struct {
	Name string `json:"name"`
	Node string `json:"node"`
	Status string `json:"status"`
}

type LocalDiskManagerStatus struct {
	Instances *DeployStatus `json:"instances,omitempty"`
	CSI *DeployStatus `json:"csi,omitempty"`
	Health string `json:"health,omitempty"`
}

type LocalStorageStatus struct {
	Instances *DeployStatus `json:"instances,omitempty"`
	CSI *DeployStatus `json:"csi,omitempty"`
	Health string `json:"health,omitempty"`
}

type SchedulerStatus struct {
	Instances *DeployStatus `json:"instances,omitempty"`
	Health string `json:"health,omitempty"`
}

type EvictorStatus struct {
	Instances *DeployStatus `json:"instances,omitempty"`
	Health string `json:"health,omitempty"`
}

type AdmissionControllerStatus struct {
	Instances *DeployStatus `json:"instances,omitempty"`
	Health string `json:"health,omitempty"`
}

type ApiServerStatus struct {
	Instances *DeployStatus `json:"instances,omitempty"`
	Health string `json:"health,omitempty"`
}

type ExporterStatus struct {
	Instances *DeployStatus `json:"instances,omitempty"`
	Health string `json:"health,omitempty"`
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
