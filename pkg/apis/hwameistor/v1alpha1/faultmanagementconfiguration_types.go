package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FaultManagementConfiguration is the Schema for the faultmanagementconfigurations API
type FaultManagementConfiguration struct {
	metav1.TypeMeta `json:",inline"`

	// EvaluationProfile defines a set of evaluators
	EvaluatorProfiles []EvaluatorProfile `json:"evaluatorProfiles"`

	// RecoveryProfile
	RecoveryProfiles []RecoveryProfile `json:"recoveryProfiles"`

	Plugins []PluginConfig `json:"plugins"`
}

// EvaluatorProfile is a Evaluation profile
type EvaluatorProfile struct {
	// EvaluatorName is the name of a specific Evaluator.
	// This name can't be duplicated with other Evaluators
	EvaluatorName string `json:"evaluatorName"`

	// Plugins is a list of Plugins that can be used to evaluate a specify type fault
	Plugins *EvaluatorPlugins `json:"plugins"`
}

// RecoveryProfile is a Recovery profile
type RecoveryProfile struct {
	// RecoveryName is the name of a specific Recovery.
	// This name can't be duplicated with other Recoveries
	RecoveryName string `json:"recoveryName"`

	// Plugins is a list of Plugins that can be used to recover a specify type fault
	Plugins *RecoveryPlugins `json:"plugins"`
}

type EvaluatorPlugins struct {
	// DiskEvaluator for evaluating the fault disk
	DiskEvaluator *PluginSet `json:"disk"`

	// VolumeEvaluator for recovering the fault volume
	VolumeEvaluator *PluginSet `json:"volume"`

	// NodeEvaluator for recovering the fault node
	NodeEvaluator *PluginSet `json:"node"`
}

type RecoveryPlugins struct {
	// DiskRecovery for evaluating the fault disk
	DiskRecovery *PluginSet `json:"disk"`

	// VolumeRecovery for evaluating the fault
	VolumeRecovery *PluginSet `json:"volume"`

	// NodeRecovery
	NodeRecovery *PluginSet `json:"node"`
}

type PluginSet struct {
	// Enabled defines all plugins that are enabled to work
	Enabled []string `json:"enabled"`

	// Disabled defines all plugins that are disabled to work
	Disabled []string `json:"disabled"`
}

type PluginConfig struct {
	// Name defines the name of plugin being configured
	Name string `json:"name"`

	// Extender can be a shell or webhook service
	Extender *Extender `json:"extender"`
}

type Extender struct {
	Shell ExtenderShell `json:"shell"`
}

// ExtenderShell defined scripts that can replace faulty devices based on specific placeholders {device}
type ExtenderShell struct {
	Command   string    `json:"command"`
	Args      []string  `json:"args"`
	SucceedAt SucceedAt `json:"succeedAt"`
}

// SucceedAt defines when a shell can be identified as successful
type SucceedAt struct {
	// ReturnCode defines the successful return code, usually can be 0
	ReturnCode int `json:"returnCode"`
	// Output defines the successful output. The output must be the same as the actual output
	Output string `json:"output"`
}
