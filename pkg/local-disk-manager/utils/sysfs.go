package utils

import (
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
)

// ReadSysFSFileAsInt64 reads a file and
// converts that content into int64
func ReadSysFSFileAsInt64(sysFilePath string) (int64, error) {
	b, err := ioutil.ReadFile(filepath.Clean(sysFilePath))
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(strings.TrimSuffix(string(b), "\n"), 10, 64)
}

func ReadSysFSFileAsString(sysFilePath string) (string, error) {
	b, err := ioutil.ReadFile(filepath.Clean(sysFilePath))
	if err != nil {
		return "", err
	}
	return string(b), nil
}
