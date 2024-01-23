package cluster

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
)

var install = &cobra.Command{
	Use:     "install",
	Args:    cobra.ExactArgs(0),
	Short:   "install Hwameistor cluster.",
	Long:    "You can use 'hwameictl cluster install' to install  install the hwameistor cluster",
	Example: "hwameictl cluster install",
	RunE:    installRunE,
}

func installRunE(_ *cobra.Command, args []string) error {
	// helm repo add hwameistor-operator https://hwameistor.io/hwameistor-operator
	addCmd := exec.Command("helm", "repo", "add", "hwameistor-operator", "https://hwameistor.io/hwameistor-operator")
	addCmd.Stdout = os.Stdout
	addCmd.Stderr = os.Stderr

	err := addCmd.Run()
	if err != nil {
		return errors.Wrap(err, "failed to add Helm repository")
	}

	// helm repo update hwameistor-operator
	updateCmd := exec.Command("helm", "repo", "update", "hwameistor-operator")
	updateCmd.Stdout = os.Stdout
	updateCmd.Stderr = os.Stderr

	err = updateCmd.Run()
	if err != nil {
		return errors.Wrap(err, "Failed to update Helm repository")
	}

	// helm install hwameistor-operator hwameistor-operator/hwameistor-operator -n hwameistor --create-namespace
	installCmd := exec.Command("helm", "install", "hwameistor-operator", "hwameistor-operator/hwameistor-operator", "-n", "hwameistor", "--create-namespace")
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr

	err = installCmd.Run()
	if err != nil {
		return errors.Wrap(err, "Failed to install Helm hwameistor-operator chart")
	}

	fmt.Println("Hwameistor Operator installed successfully!")
	return nil
}
