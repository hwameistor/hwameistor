package datacopy

import (
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// var (
// 	juiceSyncImageName = ""
// )

type JuiceSync struct {
	namespace string
	apiClient k8sclient.Client
}

func (js *JuiceSync) Prepare(targetNodeName, sourceNodeName, lvName string) error {
	return nil
}

func (js *JuiceSync) StartSync(jobName, lvName, excludedRunningNodeName, runningNodeName string) error {
	return nil
}
