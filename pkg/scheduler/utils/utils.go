package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

var (
	KubeConfigFilePath = "/etc/kubernetes/scheduler.conf"
)

func GetKubeconfigPath() (string, error) {
	fmt.Printf("%s\n", os.Args)
	for n, v := range os.Args {
		if v == "--kubeconfig" {
			return os.Args[n+1], nil
		}
		if strings.HasPrefix(v, "--kubeconfig=") {
			parts := strings.SplitN(v, "=", 2)
			return parts[1], nil
		}
	}

	return "", fmt.Errorf("flag --kubeconfig empty")
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

// PrettyPrintJSON for debug
func PrettyPrintJSON(v interface{}) {
	prettyJSON, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		fmt.Printf("Failed to generate json: %s\n", err.Error())
	}
	fmt.Printf("%s\n", string(prettyJSON))
}
