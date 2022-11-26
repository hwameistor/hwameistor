package raid

// Manager define interfaces about how to get RAID info
type Manager interface {
	GetControllerCount() (count int, err error)
}

type CommandStatus struct {
	CLIVersion      string `json:"CLI Version"`
	OperatingSystem string `json:"Operating system"`
	Controller      int    `json:"Controller"`
	Status          string `json:"Status"`
	Description     string `json:"Description"`
}

type ResponseData struct {
	ControllerCount int `json:"Controller Count"`
}

type PhysicalDrive struct {
	// Storage Capacity
	Size string `json:"Size"`

	// Represents disk online or not
	State string `json:"State"`

	// Disk type e.g, HDD,SSD,NVMe
	Med string `json:"Med"`
}
