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
	"github.com/hwameistor/hwameistor/pkg/apiserver/api"
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

type ByEventName []*api.EventAction

func (a ByEventName) Len() int {
	return len(a)
}

func (a ByEventName) Less(i, j int) bool {
	flag := false
	compare := strings.Compare(a[i].ResourceName, a[j].ResourceName)
	if compare < 0 {
		flag = true
	}
	return flag
}

func (a ByEventName) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

type ByEventType []*api.EventAction

func (a ByEventType) Len() int {
	return len(a)
}

func (a ByEventType) Less(i, j int) bool {
	flag := false
	compare := strings.Compare(a[i].ResourceType, a[j].ResourceType)
	if compare < 0 {
		flag = true
	}
	return flag
}

func (a ByEventType) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

type ByEventTime []*api.EventAction

func (a ByEventTime) Len() int {
	return len(a)
}

func (a ByEventTime) Less(i, j int) bool {
	return a[i].EventRecord.Time.Time.After(a[j].EventRecord.Time.Time)
}

func (a ByEventTime) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

// asc, desc

type ByVolumeNameAsc []*api.Volume

func (a ByVolumeNameAsc) Len() int {
	return len(a)
}

func (a ByVolumeNameAsc) Less(i, j int) bool {
	flag := false
	compare := strings.Compare(a[i].Name, a[j].Name)
	if compare < 0 {
		flag = true
	}
	return flag
}

func (a ByVolumeNameAsc) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

type ByVolumeNameDesc []*api.Volume

func (a ByVolumeNameDesc) Len() int {
	return len(a)
}

func (a ByVolumeNameDesc) Less(i, j int) bool {
	flag := false
	compare := strings.Compare(a[i].Name, a[j].Name)
	if compare > 0 {
		flag = true
	}
	return flag
}

func (a ByVolumeNameDesc) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

type ByVolumeNsAsc []*api.Volume

func (a ByVolumeNsAsc) Len() int {
	return len(a)
}

func (a ByVolumeNsAsc) Less(i, j int) bool {
	flag := false
	compare := strings.Compare(a[i].Namespace, a[j].Namespace)
	if compare < 0 {
		flag = true
	}
	return flag
}

func (a ByVolumeNsAsc) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

type ByVolumeNsDesc []*api.Volume

func (a ByVolumeNsDesc) Len() int {
	return len(a)
}

func (a ByVolumeNsDesc) Less(i, j int) bool {
	flag := false
	compare := strings.Compare(a[i].Namespace, a[j].Namespace)
	if compare > 0 {
		flag = true
	}
	return flag
}

func (a ByVolumeNsDesc) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

type ByVolumeTimeAsc []*api.Volume

func (a ByVolumeTimeAsc) Len() int {
	return len(a)
}

func (a ByVolumeTimeAsc) Less(i, j int) bool {
	return a[i].ObjectMeta.CreationTimestamp.Time.Before(a[j].ObjectMeta.CreationTimestamp.Time)
}

func (a ByVolumeTimeAsc) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

type ByVolumeTimeDesc []*api.Volume

func (a ByVolumeTimeDesc) Len() int {
	return len(a)
}

func (a ByVolumeTimeDesc) Less(i, j int) bool {
	return a[i].ObjectMeta.CreationTimestamp.Time.After(a[j].ObjectMeta.CreationTimestamp.Time)
}

func (a ByVolumeTimeDesc) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
