package manager

type ISmart interface {
	SupportSmart() (bool, error)
	// GetHealthStatus Show device SMART health status
	// true: passed false: not passed
	GetHealthStatus() (bool, error)
	ParseSmartInfo() SmartInfo
}

type SmartInfoParser struct {
	ISmart
}

type SmartInfo struct {
	// SupportSmart
	SupportSmart bool

	// OverallHealthPassed
	// true: passed false: not passed
	OverallHealthPassed bool
}
