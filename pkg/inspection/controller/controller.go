package controller

import (
	"context"
	"os"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	privilegedTrue = true
)

type Controller struct {
	KubeCli *kubernetes.Clientset
}

func New(kubeCli *kubernetes.Clientset) *Controller {
	return &Controller{
		KubeCli: kubeCli,
	}
}

func (c *Controller) Run() {
	log.Info("controller running")
	image, exist := os.LookupEnv("IMAGE")
	if !exist {
		log.Info("env not exist")
		return
	}
	nodeList, err := c.KubeCli.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.WithError(err).Error("list nodes err")
		return
	}

	potTemplate := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			HostPID: true,
			Containers: []corev1.Container{
				{
					Name: "node-inspector",
					Args: []string{
						"--v=5",
					},
					ImagePullPolicy: corev1.PullIfNotPresent,
					Image: image,
					SecurityContext: &corev1.SecurityContext{
						Privileged: &privilegedTrue,
					},
				},
			},
		},
	}

	for _, node := range nodeList.Items {
		podName := "node-inspector-" + node.Name
		pod := potTemplate.DeepCopy()
		pod.Name = podName
		if _, err := c.KubeCli.CoreV1().Pods("default").Create(context.TODO(), pod, metav1.CreateOptions{}); err != nil {
			log.WithError(err).Error("create pod err")
		}
	}
}