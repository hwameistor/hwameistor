package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	hwameistorclient "github.com/hwameistor/hwameistor/pkg/apis/client/clientset/versioned"
	// "github.com/hwameistor/hwameistor/pkg/inspection"
	"github.com/hwameistor/hwameistor/pkg/inspection/controller"
	"github.com/hwameistor/hwameistor/pkg/inspection/node"
	log "github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	CONTROLLERMODE = "controller"
	NODEMODE = "node"
)

var (
	privilegedTrue = true
	hostPathDirectoryOrCreate = corev1.HostPathDirectoryOrCreate
	logLevel = flag.Int("v", 4 /*Log Info*/, "number for the log level verbosity")
	mode = flag.String("mode", NODEMODE, "run mode")
)

func main() {
	flag.Parse()
	setupLogging()

	restConfig := ctrl.GetConfigOrDie()
	cliset := kubernetes.NewForConfigOrDie(restConfig)

	if *mode == CONTROLLERMODE {
		controller.New(cliset).Run()
	} else if *mode == NODEMODE {
		node.NewNodeInspector().Inspect()
	} else {
		log.Errorf("wrong run mode: %v", mode)
		os.Exit(1)
	}
	
	// inspectCluster(cliset)
	// inspectHwameistor(restConfig)
	// createJob(cliset)
}

func setupLogging() {
	// parse log level(default level: info)
	var level log.Level
	if *logLevel >= int(log.TraceLevel) {
		level = log.TraceLevel
	} else if *logLevel <= int(log.PanicLevel) {
		level = log.PanicLevel
	} else {
		level = log.Level(*logLevel)
	}

	log.SetLevel(level)
	log.SetFormatter(&log.JSONFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			s := strings.Split(f.Function, ".")
			funcName := s[len(s)-1]
			fileName := path.Base(f.File)
			return funcName, fmt.Sprintf("%s:%d", fileName, f.Line)
		}})
	log.SetReportCaller(true)
}

func inspectCluster(cliset *kubernetes.Clientset) {
	podList, err := cliset.CoreV1().Pods("kube-system").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.WithError(err).Error("list pods in kube-system namespace err")
		return
	}
	for _, pod := range podList.Items {
		if strings.Contains(pod.Name, "kube-apiserver") {
			image := pod.Spec.Containers[0].Image
			subStrings := strings.Split(image, ":")
			imageTag := subStrings[len(subStrings) - 1]
			log.Infof("image tag of kube-apiserver is %v", imageTag)
		}
	}

	nodeList, err := cliset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.WithError(err).Error("list nodes err")
		return
	}
	for _, node := range nodeList.Items {
		nodeVersion := node.Status.NodeInfo.KubeletVersion
		log.Infof("node %v version is %v", node.Name, nodeVersion)
	}

	scList, err := cliset.StorageV1().StorageClasses().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.WithError(err).Error("list storageclass err")
		return
	}
	for _, sc := range scList.Items {
		if strings.Contains(sc.Provisioner, "hwamei") {
			log.Infof("storageclass related to hwameistor found: %v", sc.Name)
		}
	}
}

func inspectHwameistor(restConfig *rest.Config) {
	hwameiClient := hwameistorclient.NewForConfigOrDie(restConfig)
	lsnList, err := hwameiClient.HwameistorV1alpha1().LocalStorageNodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.WithError(err).Error("list localstoragenodes err")
		return
	}
	for _, lsn := range lsnList.Items {
		log.Infof("localstoragenode %v has %v pools", lsn.Name, len(lsn.Status.Pools))
		for _, pool := range lsn.Status.Pools {
			log.Infof("pool: %v, totalcapacity: %v", pool.Name, pool.TotalCapacityBytes)
		}
	}
}

func createJob(cliset *kubernetes.Clientset) {
	layout := "2006-01-02-15-04-05"
	now := time.Now().Format(layout)
	nodeInspectJob := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node-inspect-" + now,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					HostPID: true,
					Containers: []corev1.Container{
						{
							Name: "node-inspect",
							Image: "",
							// Command: ,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name: "log",
									MountPath: "/var/log/hwameistor-node-inspect",
								},
							},
							SecurityContext: &corev1.SecurityContext{
								Privileged: &privilegedTrue,
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "log",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/var/log/hwameistor-node-inspect",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						},
					},
				},
			},
		},
	}

	if _, err := cliset.BatchV1().Jobs("default").Create(context.TODO(), &nodeInspectJob, metav1.CreateOptions{}); err != nil {
		log.WithError(err).Error("create node inspect job err")
	}
}