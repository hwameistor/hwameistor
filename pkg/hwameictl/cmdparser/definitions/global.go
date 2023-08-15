package definitions

import "time"

// Global settings, Read from hwameictl flags
var (
	Kubeconfig string
	Timeout    time.Duration
)
