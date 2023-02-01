package E2eTest

import (
	"context"
	"encoding/json"
	"fmt"
	clientset "github.com/hwameistor/hwameistor/pkg/apis/client/clientset/versioned/scheme"
	"github.com/hwameistor/hwameistor/pkg/apiserver/api"
	"github.com/hwameistor/hwameistor/test/e2e/framework"
	"github.com/hwameistor/hwameistor/test/e2e/utils"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"net/http"
	"regexp"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = ginkgo.Describe("apiserver check test ", ginkgo.Label("api"), func() {
	var f *framework.Framework
	var client ctrlclient.Client
	ctx := context.TODO()
	var myUrl string
	ginkgo.Context("check the fields of hwameistor-apiserver", func() {
		ginkgo.It("Configure the base environment", func() {
			result := utils.ConfigureEnvironment(ctx)
			gomega.Expect(result).To(gomega.BeNil())
			f = framework.NewDefaultFramework(clientset.AddToScheme)
			client = f.GetClient()
			utils.CreateLdc(ctx)

		})
		ginkgo.It("edit api-service", func() {
			service := &corev1.Service{}
			serviceKey := ctrlclient.ObjectKey{
				Name:      "hwameistor-apiserver",
				Namespace: "hwameistor",
			}
			err := client.Get(ctx, serviceKey, service)
			if err != nil {
				logrus.Printf("Failed to find service：%+v ", err)
				f.ExpectNoError(err)
			}
			service.Spec.Type = corev1.ServiceTypeNodePort
			servicePort := corev1.ServicePort{
				Protocol:   "TCP",
				Port:       80,
				TargetPort: intstr.FromString("http"),
				NodePort:   31111,
			}
			service.Spec.Ports[0] = servicePort

			err = client.Update(ctx, service)
		})
		ginkgo.It("check serviceAccountName", func() {
			podList := &corev1.PodList{}
			err := client.List(ctx, podList)
			for _, pod := range podList.Items {
				b, _ := regexp.MatchString("hwameistor-apiserver", pod.Name)
				if b == true {
					logrus.Printf(pod.Name)
					myUrl = pod.Status.HostIP
				}

			}
			//time.Sleep(60 * time.Second)
			resp, err := http.Get("http://" + myUrl + ":31111/apis/hwameistor.io/v1alpha1/cluster/drbd")
			if err != nil {
				logrus.Error(err)
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			drbd := &api.DrbdEnableSetting{}

			err = json.Unmarshal(body, drbd)
			if err != nil {
				fmt.Println("error:", err)
				return
			}
			logrus.Printf(drbd.Version)

			resp, err = http.Get("http://" + myUrl + ":31111/apis/hwameistor.io/v1alpha1/cluster/nodes/k8s-node1")
			if err != nil {
				logrus.Error(err)
			}
			defer resp.Body.Close()
			body, err = ioutil.ReadAll(resp.Body)
			storageNode := &api.StorageNode{}
			err = json.Unmarshal(body, storageNode)
			if err != nil {
				fmt.Println("error:", err)
				return
			}
			logrus.Printf(storageNode.LocalStorageNode.Name)

		})
	})

})
