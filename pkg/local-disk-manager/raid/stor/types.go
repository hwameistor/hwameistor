package stor

type OperateResult struct {
	Controllers []Controllers `json:"Controllers"`
}

type CommandStatus struct {
	CLIVersion      string `json:"CLI Version"`
	OperatingSystem string `json:"Operating system"`
	Controller      int    `json:"Controller"`
	Status          string `json:"Status"`
	Description     string `json:"Description"`
}
type Controllers struct {
	CommandStatus CommandStatus `json:"Command Status"`
}

type ControllerCounts struct {
	Controllers []ControllerCount `json:"Controllers"`
}

type CountResponseData struct {
	ControllerCount int `json:"Controller Count"`
}

type ControllerCount struct {
	CommandStatus CommandStatus     `json:"Command Status"`
	ResponseData  CountResponseData `json:"Response Data"`
}
