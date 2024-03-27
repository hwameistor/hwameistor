package utils

import (
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
)

// ReadSysFSFileAsInt64 reads a file from the sysfs filesystem and converts its content to int64.
// Suitable for numeric information such as device size, etc.
func ReadSysFSFileAsInt64(sysFilePath string) (int64, error) {
	b, err := ioutil.ReadFile(filepath.Clean(sysFilePath))
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(strings.TrimSuffix(string(b), "\n"), 10, 64)
}

// ReadSysFSFileAsString reads a file from the sysfs filesystem and returns its content as a string.
// Suitable for text information such as device state, wwid, etc.
func ReadSysFSFileAsString(sysFilePath string) (string, error) {
	b, err := ioutil.ReadFile(filepath.Clean(sysFilePath))
	if err != nil {
		return "", err
	}
	return string(b), nil
}
