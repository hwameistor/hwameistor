package E2eTest

import (
	"context"
	ldapis "github.com/hwameistor/hwameistor/pkg/apis/generated/local-disk-manager/clientset/versioned/scheme"
	ldv1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/local-disk-manager/v1alpha1"
	"github.com/hwameistor/hwameistor/test/e2e/framework"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
	"time"
)

var _ = ginkgo.Describe("test Local Disk Manager", ginkgo.Label("periodCheck"), ginkgo.Ordered, func() {
	f := framework.NewDefaultFramework(ldapis.AddToScheme)
	client := f.GetClient()
	ctx := context.TODO()
	ginkgo.Context("test Local Disk", func() {
		ginkgo.It("Configure the base environment", func() {
			configureEnvironment(ctx)
		})
		ginkgo.It("Check existed Local Disk", func() {
			time.Sleep(2 * time.Minute)
			localDiskList := &ldv1.LocalDiskList{}
			err := client.List(ctx, localDiskList)
			if err != nil {
				logrus.Error(err)
			}
			logrus.Printf("There are %d local volumes ", len(localDiskList.Items))
			gomega.Expect(len(localDiskList.Items)).To(gomega.Equal(9))
		})
		ginkgo.It("Manage new disks", func() {
			output := runInLinux("sh adddisk.sh")
			logrus.Info("add  disk :", output)
			err := wait.PollImmediate(3*time.Second, 3*time.Minute, func() (done bool, err error) {
				localDiskList := &ldv1.LocalDiskList{}
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
			err := createLdc(ctx)
			gomega.Expect(err).To(gomega.BeNil())

		})
	})
	ginkgo.Context("Clean up the environment", func() {
		ginkgo.It("Clean helm & crd", func() {
			uninstallHelm()
		})
	})
	ginkgo.AfterAll(func() {
		output := runInLinux("sh deletedisk.sh")
		logrus.Info("delete disk", output)
	})
})
