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
	"strings"
	"time"
)

var _ = ginkgo.Describe("apiserver test ", ginkgo.Label("api"), func() {
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
		ginkgo.It("check get cluster/drbd", func() {
			podList := &corev1.PodList{}
			err := client.List(ctx, podList)
			for _, pod := range podList.Items {
				b, _ := regexp.MatchString("hwameistor-apiserver", pod.Name)
				if b == true {
					logrus.Printf(pod.Status.HostIP)
					myUrl = pod.Status.HostIP
				}

			}
			time.Sleep(60 * time.Second)
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
			//gomega.Expect(drbd.Version).To(gomega.Equal("v9.0.32-1"))
			//gomega.Expect(drbd.Enable).To(gomega.Equal(false))
			//gomega.Expect(drbd.State).To(gomega.Equal(api.DrbdModuleStatusDisabled))

		})
		ginkgo.It("check post cluster/drbd", func() {

			req, err := json.Marshal(api.DrbdEnableSettingReqBody{Enable: true})
			if err != nil {
				logrus.Error(err)
			}
			gomega.Expect(err).To(gomega.BeNil())
			reqbody := strings.NewReader(string(req))
			resp, err := http.Post("http://"+myUrl+":31111/apis/hwameistor.io/v1alpha1/cluster/drbd", "application/json", reqbody)
			if err != nil {
				logrus.Error(err)
			}
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			drbd := &api.DrbdEnableSetting{}
			err = json.Unmarshal(body, drbd)
			if err != nil {
				logrus.Error("error:", err)
				return
			}
			logrus.Printf(string(body))
			logrus.Printf(drbd.Version)
			logrus.Println(drbd.Enable)

		})
		ginkgo.It("check cluster/nodes/", func() {

			resp, err := http.Get("http://" + myUrl + ":31111/apis/hwameistor.io/v1alpha1/cluster/nodes/k8s-node1")
			if err != nil {
				logrus.Error(err)
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			storageNode := &api.StorageNode{}
			err = json.Unmarshal(body, storageNode)
			if err != nil {
				fmt.Println("error:", err)
				return
			}
			logrus.Printf(string(storageNode.LocalStorageNode.Status.State))

		})

	})

})
