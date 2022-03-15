package SmokeTest

import (
	"context"
	lsv1 "github.com/hwameistor/local-storage/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/local-storage/test/e2e/framework"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = ginkgo.Describe("volume", func() {
	f := framework.NewDefaultFramework(lsv1.AddToScheme)
	client := f.GetClient()

	ginkgo.It("Configure the base environment", func() {
		installHwameiStorByHelm()
		addLabels()
	})
	ginkgo.Context("test localstorage", func() {
		ginkgo.It("check status", func() {
			daemonset := &appsv1.DaemonSet{}
			daemonsetKey := k8sclient.ObjectKey{
				Name:      "hwameistor-local-storage",
				Namespace: "hwameistor",
			}

			err := client.Get(context.TODO(), daemonsetKey, daemonset)
			if err != nil {
				f.ExpectNoError(err)

			}

			gomega.Expect(daemonset.Status.DesiredNumberScheduled).To(gomega.Equal(daemonset.Status.NumberAvailable))
		})
	})
	ginkgo.Context("test hwameistor-csi-controller", func() {
		ginkgo.It("check status", func() {
			deployment := &appsv1.Deployment{}
			deploymentKey := k8sclient.ObjectKey{
				Name:      "hwameistor-csi-controller",
				Namespace: "hwameistor",
			}

			err := client.Get(context.TODO(), deploymentKey, deployment)
			if err != nil {
				f.ExpectNoError(err)
				logrus.Printf("%+v ", err)
			}
			gomega.Expect(deployment.Status.AvailableReplicas).To(gomega.Equal(int32(1)))
		})
	})
	ginkgo.Context("test hwameistor-scheduler", func() {
		ginkgo.It("check status", func() {
			deployment := &appsv1.Deployment{}
			deploymentKey := k8sclient.ObjectKey{
				Name:      "hwameistor-scheduler",
				Namespace: "hwameistor",
			}

			err := client.Get(context.TODO(), deploymentKey, deployment)
			if err != nil {
				f.ExpectNoError(err)
				logrus.Printf("%+v ", err)
			}
			gomega.Expect(deployment.Status.AvailableReplicas).To(gomega.Equal(int32(1)))
		})
	})
	ginkgo.It("Clean up the environment", func() {
		uninstallHelm()

	})
})
