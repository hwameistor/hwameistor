package utils

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"
	"unicode"

	"github.com/kubernetes-csi/csi-lib-utils/leaderelection"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

const (
	leaderLeaseDuration      = 30 * time.Second
	leaderLeaseRenewDeadLine = 25 * time.Second
	leaderLeaseRetryDuration = 15 * time.Second
)

var unitMap = map[string]int64{
	"b":  1,
	"B":  1,
	"k":  1000,
	"K":  1024,
	"KB": 1024,
	"m":  1000 * 1000,
	"M":  1024 * 1024,
	"MB": 1024 * 1024,
	"g":  1000 * 1000 * 1000,
	"G":  1024 * 1024 * 1024,
	"GB": 1024 * 1024 * 1024,
	"t":  1000 * 1000 * 1000 * 1000,
	"T":  1024 * 1024 * 1024 * 1024,
	"TB": 1024 * 1024 * 1024 * 1024,
}

var unitArray = []string{"B", "KB", "MB", "GB", "TB"}

// LogGRPC log grpc all info
func LogGRPC(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	/*	logCtx := log.WithFields(log.Fields{"call": info.FullMethod, "request": req})
		resp, err := handler(ctx, req)
		if err != nil {
			logCtx.WithField("error", err).Debug("GRPC")
		} else {
			logCtx.WithField("response", resp).Debug("GRPC")
		}
		return resp, err
	*/
	return handler(ctx, req)
}

// BuildInClusterClientset builds a kubernetes in-cluster clientset
func BuildInClusterClientset() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.WithError(err).Error("Failed to build kubernetes config")
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

// RunWithLease run a process with acquired leader lease. It's a blocking function
func RunWithLease(ns string, identity string, lockName string, runFunc func(ctx context.Context)) error {
	clientset, err := BuildInClusterClientset()
	if err != nil {
		log.WithError(err).Error("Failed to build kubernetes clientset")
		return err
	}
	// Become the leader before proceeding. The lock will be released only after the Pod is terminated
	le := leaderelection.NewLeaderElectionWithLeases(clientset, lockName, runFunc)
	le.WithNamespace(ns)
	le.WithIdentity(identity)
	le.WithLeaseDuration(leaderLeaseDuration)
	le.WithRenewDeadline(leaderLeaseRenewDeadLine)
	le.WithRetryPeriod(leaderLeaseRetryDuration)

	return le.Run()
}

// RemoveStringItem removes a string from a slice
func RemoveStringItem(items []string, itemToDelete string) []string {
	for i, item := range items {
		if itemToDelete == item {
			return append(items[:i], items[i+1:]...)
		}
	}
	return items
}

// AddUniqueStringItem add a string from a slice without duplicate
func AddUniqueStringItem(items []string, itemToAdd string) []string {
	for _, item := range items {
		if itemToAdd == item {
			return items
		}
	}
	return append(items, itemToAdd)
}

// ParseBytes parse size from string into bytes
func ParseBytes(sizeStr string) (int64, error) {

	numStr := ""
	unitStr := ""
	for i := range sizeStr {
		if !unicode.IsDigit(rune(sizeStr[i])) {
			numStr = sizeStr[:i]
			unitStr = sizeStr[i:]
			break
		} else {
			numStr = sizeStr[:i+1]
		}
	}

	if numStr == "" {
		return -1, fmt.Errorf("wrong number: %s", sizeStr)
	}
	if unitStr == "" {
		unitStr = "B"
	}
	multiple, has := unitMap[unitStr]
	if !has {
		return -1, fmt.Errorf("wrong unit: %s", sizeStr)
	}

	size, err := strconv.ParseInt(numStr, 10, 32)
	if err != nil {
		return -1, err
	}
	return size * multiple, nil
}

// ConvertBytesToStr convert size into string
func ConvertBytesToStr(size int64) string {
	unitIndex := 0
	for size > 1024 {
		size /= 1024
		unitIndex++
	}
	return fmt.Sprintf("%d%s", size, unitArray[unitIndex])
}

// BuildStoragePoolName constructs storage pool name
func BuildStoragePoolName(poolClass string) (string, error) {
	if poolClass == apisv1alpha1.DiskClassNameHDD {
		return apisv1alpha1.PoolNameForHDD, nil
	}
	if poolClass == apisv1alpha1.DiskClassNameSSD {
		return apisv1alpha1.PoolNameForSSD, nil
	}
	if poolClass == apisv1alpha1.DiskClassNameNVMe {
		return apisv1alpha1.PoolNameForNVMe, nil
	}

	return "", fmt.Errorf("invalid pool info")
}

// const (
// 	devDiskByPathDir   = "/dev/disk/by-path"
// 	prefixPCIDevice    = "pci-"
// 	pciNVMePlaceholder = "-nvme-"
// )

// PCIDiskInfo struct
type PCIDiskInfo struct {
	isNVMe   bool
	pciName  string
	diskName string
}

// IsNVMe check if it's a nvme or not
func (d *PCIDiskInfo) IsNVMe() bool {
	return d.isNVMe
}

// SanitizeName sanitizes the provided string so it can be consumed by leader election library
// copy from github.com/kubernetes-csi/csi-lib-utils/leaderelection
func SanitizeName(name string) string {
	re := regexp.MustCompile("[^a-zA-Z0-9-]")
	name = re.ReplaceAllString(name, "-")
	if name[len(name)-1] == '-' {
		// name must not end with '-'
		name = name + "X"
	}
	return name
}

// GetNodeName gets the node name from env, else
// returns an error
func GetNodeName() string {
	nodeName, ok := os.LookupEnv("MY_NODENAME")
	if !ok {
		log.Errorf("Failed to get NODENAME from ENV")
		return ""
	}

	return nodeName
}

func GetPodName() string {
	podName, ok := os.LookupEnv("POD_NAME")
	if !ok {
		log.Errorf("Failed to get POD_NAME from ENV")
		return ""
	}

	return podName
}

// GetNamespace get Namespace from env, else it returns error
func GetNamespace() string {
	ns, ok := os.LookupEnv("POD_NAMESPACE")
	if !ok {
		log.Errorf("Failed to get NameSpace from ENV")
		return ""
	}

	return ns
}

func TouchFile(filepath string) error {
	if _, err := os.Stat(filepath); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		// file does not exist
		if _, e := os.Create(filepath); e != nil {
			return e
		}
	}
	return nil
}

func GetSnapshotRestoreNameByVolume(volumeName string) string {
	return fmt.Sprintf("snaprestore-%s", volumeName)
}
