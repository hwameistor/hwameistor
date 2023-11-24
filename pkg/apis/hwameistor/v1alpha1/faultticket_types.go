package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type (
	TicketType  string
	TicketPhase string

	FaultEffectScope string
	VolumeFaultType  string
)

const (
	DiskFaultTicket   TicketType = "Disk"
	NodeFaultTicket   TicketType = "Node"
	VolumeFaultTicket TicketType = "Volume"

	Node   FaultEffectScope = "Node"
	Pool   FaultEffectScope = "Pool"
	Volume FaultEffectScope = "Volume"
	App    FaultEffectScope = "App"

	Evaluating TicketPhase = "Evaluating"
	Recovering TicketPhase = "Recovering"
	Completed  TicketPhase = "Completed"

	FileSystemFault VolumeFaultType = "filesystem"
	BadBlockFault   VolumeFaultType = "badblock"
)

// FaultTicketSpec defines the desired state of FaultTicket
type FaultTicketSpec struct {
	// NodeName represents which node the fault happened at
	// +kubebuilder:validation:Required
	NodeName string `json:"nodeName"`

	// Type represents what caused this fault e.g., Disk, Volume, Node
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum:=Volume;Node;Disk
	Type TicketType `json:"type"`

	// Source represents where the fault comes from e.g., prometheus, disk analysis tools
	// +kubebuilder:validation:Required
	Source string `json:"source"`

	// Device represents which device the fault happened
	Device FaultDevice `json:"device,omitempty"`

	// Volume represents which volume the fault happened
	Volume FaultVolume `json:"volume,omitempty"`

	// Message represents the details of the fault which caused this fault actually
	Message string `json:"message,omitempty"`
}

// FaultDevice can be used to identify the device
type FaultDevice struct {
	// DevPath represents the path of the fault disk e.g., /dev/sda
	DevPath string `json:"devPath,omitempty"`

	// DevLinks are symbol links for this fault device
	DevLinks []string `json:"devLinks,omitempty"`

	// SerialNumber represents the serial number of the fault disk
	SerialNumber string `json:"serialNumber,omitempty"`

	// LocalDiskName represents the name of the LocalDisk
	LocalDiskName string `json:"localDiskName,omitempty"`
}

type FaultVolume struct {
	// Name represents the name of the fault volume, this can be pv name or lv name
	Name string `json:"volumeName,omitempty"`

	// Path represents the path of the fault volume on the host
	Path string `json:"volumePath,omitempty"`

	// FaultType represents the fault type of the fault volume
	FaultType VolumeFaultType `json:"volume,omitempty"`
}

// FaultTicketStatus defines the observed state of FaultTicket
type FaultTicketStatus struct {
	// Phase represents the	phase of the ticket
	// The state machine operates as follows
	// Empty -> Evaluating -> Recovering ->                         -> Completed
	// SubScope:                            Recovering -> Completed
	Phase TicketPhase `json:"phase,omitempty"`

	// Effects represent these scopes that effected by this fault and the handle results
	Effects []Effect `json:"effectScope,omitempty"`
}

// Effect represents the scope of the fault effects. The corresponding modules that care about this should handle these fault
// NOTES: which modules should handle this fault can be configured at somewhere
type Effect struct {
	// Scope represents the scope of the fault effects
	Scope FaultEffectScope `json:"scope,omitempty"`

	// RecoverInfo represents info about if and how to recover this fault
	RecoverInfo RecoverInfo `json:"recoverInfo,omitempty"`
}

// RecoverInfo represents information about if this fault is recoverable and the recover phase
type RecoverInfo struct {
	// Recoverable represents whether this fault is recoverable
	// +kubebuilder:validation:Required
	Recoverable bool `json:"recoverable"`

	// Phase represents the phase of the fault recovery
	// The phase Completed indicates this effect is recovered totally
	Phase TicketPhase `json:"phase,omitempty"`

	// Message represents the details of the recovery
	Message string `json:"message,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FaultTicket is the Schema for the faulttickets API
// +kubebuilder:resource:scope=Cluster,shortName=ticket
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".spec.nodeName",name=NodeName,type=string
// +kubebuilder:printcolumn:JSONPath=".spec.type",name=Type,type=string
// +kubebuilder:printcolumn:JSONPath=".spec.source",name=Source,type=string
// +kubebuilder:printcolumn:JSONPath=".status.phase",name=Phase,type=string
type FaultTicket struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FaultTicketSpec   `json:"spec,omitempty"`
	Status FaultTicketStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FaultTicketList contains a list of FaultTicket
type FaultTicketList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FaultTicket `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FaultTicket{}, &FaultTicketList{})
}
