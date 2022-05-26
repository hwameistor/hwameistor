package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LocalVolumeSpec defines the desired state of LocalVolume
type LocalVolumeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster

	// +kubebuilder:validation:Minimum:=4194304
	RequiredCapacityBytes int64 `json:"requiredCapacityBytes,omitempty"`

	// PoolName is the name of the storage pool, e.g. LocalStorage_PoolHDD, LocalStorage_PoolSSD, etc..
	PoolName string `json:"poolName,omitempty"`

	// replica number: 1 - non-HA, 2 - HA, 3 - migration (temp)
	// +kubebuilder:validation:Minimum:=1
	// +kubebuilder:validation:Maximum:=3
	ReplicaNumber int64 `json:"replicaNumber,omitempty"`

	// Convertible is to indicate if the non-HA volume can be transitted to HA volume or not
	// +kubebuilder:default:=false
	Convertible bool `json:"convertible,omitempty"`

	// Accessibility is the topology requirement of the volume. It describes how to locate and distribute the volume replicas
	Accessibility AccessibilityTopology `json:"accessibility,omitempty"`

	// PersistentVolumeClaimNamespace is the namespace of the associated PVC
	PersistentVolumeClaimNamespace string `json:"pvcNamespace,omitempty"`

	// PersistentVolumeClaimName is the name of the associated PVC
	PersistentVolumeClaimName string `json:"pvcName,omitempty"`

	// VolumeGroup is the group name of the local volumes. It is designed for the scheduling and allocating.
	VolumeGroup string `json:"volumegroup,omitempty"`

	// Config is the configration for the volume replicas
	// It will be managed by the controller, and watched by all the nodes
	// Important: node will manage volume replica according this config
	Config *VolumeConfig `json:"config,omitempty"`

	// Delete is to indicate where the replica should be deleted or not.
	// It's different from the regular resource delete interface in Kubernetes.
	// The purpose is to protect it from any mistakes
	// +kubebuilder:default:=false
	Delete bool `json:"delete,omitempty"`
}

// AccessibilityTopology of the volume
type AccessibilityTopology struct {
	// Nodes is the collection of storage nodes the volume replicas must locate at
	Nodes []string `json:"nodes,omitempty"`

	// zones where the volume replicas should be distributed across, it's Optional
	// +kubebuilder:default:={default}
	Zones []string `json:"zones,omitempty"`

	// regions where the volume replicas should be distributed across, it's Optional
	// +kubebuilder:default:={default}
	Regions []string `json:"regions,omitempty"`
}

// VolumeConfig is the configration of the volume, including the replicas
type VolumeConfig struct {
	// Version of config, start from 0, plus 1 every time config update
	Version               int    `json:"version"`
	VolumeName            string `json:"volumeName"`
	RequiredCapacityBytes int64  `json:"requiredCapacityBytes,omitempty"`

	// Convertible is to indicate if the non-HA volume can be transitted to HA volume or not
	Convertible bool `json:"convertible,omitempty"`

	// ResourceID is for HA volume, set to '-1' for non-HA volume
	ResourceID        int             `json:"resourceID"`
	ReadyToInitialize bool            `json:"readyToInitialize"`
	Initialized       bool            `json:"initialized"`
	Replicas          []VolumeReplica `json:"replicas"`
}

// DeepEqual check if the two configs are equal completely or not
func (vc *VolumeConfig) DeepEqual(peer *VolumeConfig) bool {
	if peer == nil {
		return false
	}
	if vc.VolumeName != peer.VolumeName {
		return false
	}
	if vc.RequiredCapacityBytes != peer.RequiredCapacityBytes {
		return false
	}
	if vc.ResourceID != peer.ResourceID {
		return false
	}
	if vc.Convertible != peer.Convertible {
		return false
	}
	if len(vc.Replicas) != len(peer.Replicas) {
		return false
	}

	peerReplicasPos := map[string]int{}
	for i, replica := range peer.Replicas {
		peerReplicasPos[replica.Hostname] = i
	}
	for i, replica := range vc.Replicas {
		if pos, exists := peerReplicasPos[replica.Hostname]; exists {
			if !vc.Replicas[i].DeepEqual(&peer.Replicas[pos]) {
				return false
			}
		} else {
			return false
		}
	}
	return true
}

// VolumeReplica contains informations of replica peer
type VolumeReplica struct {
	ID       int    `json:"id"`
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
	Primary  bool   `json:"primary"`
}

// DeepEqual check if the two volumereplicas are equal completely or not
func (vr *VolumeReplica) DeepEqual(peer *VolumeReplica) bool {
	if peer == nil {
		return false
	}
	if vr.ID != peer.ID {
		return false
	}
	if vr.Hostname != peer.Hostname {
		return false
	}
	if vr.IP != peer.IP {
		return false
	}
	if vr.Primary != peer.Primary {
		return false
	}

	return true
}

// LocalVolumeStatus defines the observed state of LocalVolume
type LocalVolumeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster

	// AllocatedCapacityBytes is the real allocated capacity in bytes of the volume replicas.
	// In case of HA volume with multiple replicas, the value is equal to the one of a replica's size
	AllocatedCapacityBytes int64 `json:"allocatedCapacityBytes,omitempty"`

	// Volume is a logical concept and composed by one or many replicas which will be located at different node.
	Replicas []string `json:"replicas,omitempty"`

	// State is the phase of volume replica, e.g. Creating, Ready, NotReady, ToBeDeleted, Deleted
	State State `json:"state,omitempty"`

	// PublishedNodeName is the node where the volume is published and used by pod
	PublishedNodeName string `json:"publishedNode,omitempty"`

	// Synced is the sync state of the volume replica, which is important in HA volume
	// +kubebuilder:default:=false
	//Synced bool `json:"synced,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalVolume is the Schema for the volumes API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=localvolumes,scope=Cluster,shortName=lv
// +kubebuilder:printcolumn:name="pool",type=string,JSONPath=`.spec.poolName`,description="Name of storage pool"
// +kubebuilder:printcolumn:name="replicas",type=integer,JSONPath=`.spec.replicaNumber`,description="Number of volume replica"
// +kubebuilder:printcolumn:name="capacity",type=integer,JSONPath=`.spec.requiredCapacityBytes`,description="Required capacity of the volume"
// +kubebuilder:printcolumn:name="accessibility",type=string,JSONPath=`.spec.accessibility.node`,description="Accessibility of volume"
// +kubebuilder:printcolumn:name="state",type=string,JSONPath=`.status.state`,description="State of the volume"
// +kubebuilder:printcolumn:name="resource",type=integer,JSONPath=`.spec.config.resourceID`,description="Allocated resource ID for the volume"
// +kubebuilder:printcolumn:name="published",type=string,JSONPath=`.status.publishedNode`,description="Name of the node where the volume is in-use"
// +kubebuilder:printcolumn:name="age",type=date,JSONPath=`.metadata.creationTimestamp`
type LocalVolume struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LocalVolumeSpec   `json:"spec,omitempty"`
	Status LocalVolumeStatus `json:"status,omitempty"`
}

// SetReplicas add replicas into status
func (v *LocalVolume) SetReplicas(replicas []*LocalVolumeReplica) {
	v.Status.Replicas = []string{}
	for _, replica := range replicas {
		v.Status.Replicas = append(v.Status.Replicas, replica.Name)
	}
}

// IsHighAvailability return true if volume is HighAvailability
func (v *LocalVolume) IsHighAvailability() bool {
	return v.Spec.ReplicaNumber >= 2
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LocalVolumeList contains a list of LocalVolume
type LocalVolumeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LocalVolume `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LocalVolume{}, &LocalVolumeList{})
}
