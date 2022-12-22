package v1alpha1

//Phase is the phase of an operation
type Phase string

const (
	ClusterPhaseEmpty     Phase = ""
	ClusterPhaseToInstall Phase = "Toinstall"
	ClusterPhaseInstalled Phase = "Installed"
)