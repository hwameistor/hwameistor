package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	apisv1alpha1 "github.com/hwameistor/local-storage/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/local-storage/pkg/exechelper"

	"github.com/kubernetes-csi/csi-lib-utils/leaderelection"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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

//LogGRPC log grpc all info
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

// LogREST log rest api call info
func LogREST(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		inner.ServeHTTP(w, r)

		log.Infof(
			"%s  %s  %s  %s",
			r.Method,
			r.RequestURI,
			name,
			time.Since(start),
		)
	})
}

// PrettyPrintJSON for debug
func PrettyPrintJSON(v interface{}) {
	prettyJSON, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		fmt.Printf("Failed to generate json: %s\n", err.Error())
	}
	fmt.Printf("%s\n", string(prettyJSON))
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
func BuildStoragePoolName(poolClass string, poolType string) (string, error) {

	if poolClass == apisv1alpha1.DiskClassNameHDD && poolType == apisv1alpha1.PoolTypeRegular {
		return apisv1alpha1.PoolNameForHDD, nil
	}
	if poolClass == apisv1alpha1.DiskClassNameSSD && poolType == apisv1alpha1.PoolTypeRegular {
		return apisv1alpha1.PoolNameForSSD, nil
	}
	if poolClass == apisv1alpha1.DiskClassNameNVMe && poolType == apisv1alpha1.PoolTypeRegular {
		return apisv1alpha1.PoolNameForNVMe, nil
	}

	return "", fmt.Errorf("invalid pool info")
}

const (
	devDiskByPathDir   = "/dev/disk/by-path"
	prefixPCIDevice    = "pci-"
	pciNVMePlaceholder = "-nvme-"
)

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

// GetPCIDisks gets pci disks info on the local node
func GetPCIDisks(cmdExec exechelper.Executor) (map[string]*PCIDiskInfo, error) {
	pciDisks := make(map[string]*PCIDiskInfo)

	params := exechelper.ExecParams{
		CmdName: "ls",
		CmdArgs: []string{"-Al", devDiskByPathDir},
	}
	res := cmdExec.RunCommand(params)
	if res.ExitCode != 0 {
		return pciDisks, res.Error
	}

	for _, line := range strings.Split(res.OutBuf.String(), "\n") {
		items := regexp.MustCompile(" +").Split(strings.TrimPrefix(line, " "), -1)
		if len(items) < 9 {
			continue
		}
		pciName := strings.TrimSpace(items[8])
		if strings.HasPrefix(pciName, prefixPCIDevice) {
			var isNVMe bool
			if strings.Index(pciName, pciNVMePlaceholder) > 0 {
				isNVMe = true
			}
			devicePath := strings.Trim(items[len(items)-1], " ")
			diskNamePos := strings.LastIndex(devicePath, string(os.PathSeparator))
			if diskNamePos >= 0 {
				devicePath = devicePath[diskNamePos+1:]
			}
			rightIndex := len(devicePath)
			diskName := strings.TrimSpace(devicePath[:rightIndex])
			if len(diskName) > 0 {
				pciDisks[diskName] = &PCIDiskInfo{
					pciName:  pciName,
					diskName: diskName,
					isNVMe:   isNVMe,
				}
			}
		}
	}

	return pciDisks, nil
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
