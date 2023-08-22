package definitions

import "time"

const DefaultKubeConfigPath = "~/.kube/config"

// Global settings, Read from hwameictl flags
var (
	KubeConfigPath string
	Timeout        time.Duration
	Debug          bool
)
