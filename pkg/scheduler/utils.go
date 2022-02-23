package scheduler

import (
	"fmt"
	"os"
	"strings"
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
