// Package classification User API.
//
// The purpose of this service is to provide an application
// that is using plain go code to define an API
//
//	Host: localhost
//	Version: 0.0.1
//
// swagger:meta
package utils

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

// GetNodeName gets the node name from env, else
// returns an error
func GetNodeName() string {
	nodeName, ok := os.LookupEnv("NODENAME")
	if !ok {
		log.Errorf("Failed to get NODENAME from ENV")
		return ""
	}

	return nodeName
}

// GetNamespace get Namespace from env, else it returns error
func GetNamespace() string {
	ns, ok := os.LookupEnv("NAMESPACE")
	if !ok {
		log.Errorf("Failed to get NameSpace from ENV")
		return ""
	}

	return ns
}

func DivideOperate(num1, num2 int64) (float64, error) {
	value, err := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(num1)/float64(num2)), 64)
	if err != nil {
		return 0, err
	}
	return value, nil
}

// nolint
func DataPatination[T any](origin []T, page, pageSize int32) []T {
	if pageSize == -1 {
		return origin
	}

	if page < 1 {
		return make([]T, 0)
	}

	total := int32(len(origin))
	start := (page - 1) * pageSize
	end := page * pageSize

	if start > total {
		return make([]T, 0)
	}

	if end > total {
		end = total
	}

	return origin[start:end]
}

// ConvertNodeName e.g.(10.23.10.12 => 10-23-10-12)
func ConvertNodeName(node string) string {
	return strings.Replace(node, ".", "-", -1)
}
