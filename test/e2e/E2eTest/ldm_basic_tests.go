package E2eTest

import (
	"context"
	clientset "github.com/hwameistor/hwameistor/pkg/apis/client/clientset/versioned/scheme"
	"github.com/hwameistor/hwameistor/test/e2e/utils"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"time"

	v1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/test/e2e/framework"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
)

var _ = ginkgo.Describe("test Local Disk Manager", ginkgo.Label("periodCheck"), ginkgo.Ordered, func() {
	var f *framework.Framework
	var client ctrlclient.Client
	ctx := context.TODO()
	ginkgo.Context("test Local Disk", func() {
		ginkgo.It("Configure the base environment", func() {
			result := utils.ConfigureEnvironment(ctx)
			gomega.Expect(result).To(gomega.BeNil())
			f = framework.NewDefaultFramework(clientset.AddToScheme)
			client = f.GetClient()

		})
		ginkgo.It("Check existed Local Disk", func() {
			time.Sleep(2 * time.Minute)
			localDiskList := &v1alpha1.LocalDiskList{}
			err := client.List(ctx, localDiskList)
			if err != nil {
				logrus.Error(err)
			}
			logrus.Printf("There are %d local volumes ", len(localDiskList.Items))
			gomega.Expect(len(localDiskList.Items)).To(gomega.Equal(9))
		})
		ginkgo.It("Manage new disks", func() {
			output := utils.RunInLinux("sh adddisk.sh")
			logrus.Info("add  disk :", output)
			err := wait.PollImmediate(3*time.Second, 3*time.Minute, func() (done bool, err error) {
				localDiskList := &v1alpha1.LocalDiskList{}
				err = client.List(ctx, localDiskList)
				if err != nil {
					logrus.Error("add disk failed")
					logrus.Error(err)
				}
				if len(localDiskList.Items) != 10 {
					return false, nil
				} else {
					logrus.Infof("There are %d local volumes ", len(localDiskList.Items))
					return true, nil
				}
			})
			if err != nil {
				logrus.Error("Manage new disks error", err)
			}
			gomega.Expect(err).To(gomega.BeNil())

		})

	})
	ginkgo.Context("test LocalDiskClaim", func() {
		ginkgo.It("Create new LocalDiskClaim", func() {
			err := utils.CreateLdc(ctx)
			gomega.Expect(err).To(gomega.BeNil())

		})
	})
	ginkgo.Context("Clean up the environment", func() {
		ginkgo.It("Clean helm & crd", func() {
			utils.UninstallHelm()
		})
	})
	ginkgo.AfterAll(func() {
		output := utils.RunInLinux("sh deletedisk.sh")
		logrus.Info("delete disk", output)
	})
})
