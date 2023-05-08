package utils

import (
	"bytes"
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"os"
)

// WriteFile writes data in the file at the given path
func WriteFile(filepath string, sCert *bytes.Buffer) error {
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(sCert.Bytes())
	if err != nil {
		return err
	}

	return nil
}

// GetStorageCondition extracts the provided condition from the given status and returns that.
// Returns nil and -1 if the condition is not present, and the index of the located condition.
func GetStorageCondition(conditions []apisv1alpha1.StorageNodeCondition,
	conditionType apisv1alpha1.StorageNodeConditionType) (int, *apisv1alpha1.StorageNodeCondition) {
	for i, condition := range conditions {
		if condition.Type == conditionType {
			return i, &conditions[i]
		}
	}

	return -1, nil
}
