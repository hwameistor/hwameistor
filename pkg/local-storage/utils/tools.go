package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"k8s.io/client-go/tools/cache"
)

const (
	TimeFormatOnResourceName = "20060102-15"
)

// Route defines a rest api
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

var (
	// NoOpEventHandlerFuncs for informer
	NoOpEventHandlerFuncs = cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) {},
		UpdateFunc: func(oldObj, newObj interface{}) {},
		DeleteFunc: func(obj interface{}) {},
	}
)

// PrettyPrintJSON for debug
func PrettyPrintJSON(v interface{}) {
	prettyJSON, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		fmt.Printf("Failed to generate json: %s\n", err.Error())
	}
	fmt.Printf("%s\n", string(prettyJSON))
}

// Return K8S compatible resource name using substrs and settings
//
// e.g. sub1-sub2-20200101-01-f34rret4
func GenerateResourceName(substrs []string, addDatetime, addUuid bool, maxLen int) string {
	var tail []string
	if maxLen == 0 {
		maxLen = 253
	}
	remainingLen := maxLen
	if addDatetime {
		remainingLen -= 12
		tail = append(tail, time.Now().Format(TimeFormatOnResourceName))
	}
	if addUuid {
		remainingLen -= 8
		tail = append(tail, uuid.New().String()[:7])
	}

	substrsLen := 0
	for _, substr := range substrs {
		substrsLen += len(substr) + 1
	}

	for substrsLen > remainingLen {
		if len(substrs) == 0 {
			break
		}
		substrsLen -= len(substrs[len(substrs)-1]) + 1
		substrs = substrs[:len(substrs)-1]
	}

	substrs = append(substrs, tail...)
	return strings.ToLower(strings.Join(substrs, "-"))
}

func GetRFC3339LocalTime() string {
	return time.Now().Local().Format(time.RFC3339)
}
