package E2eTest

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/sirupsen/logrus"

	"github.com/hwameistor/hwameistor/test/e2e/utils"
)

var _ = ginkgo.Describe("apiserver check test", ginkgo.Label("periodCheck"), func() {

	ginkgo.Context("check the fields of hwameistor-apiserver", func() {
		ginkgo.It("check serviceAccountName", func() {
			logrus.Infof("check serviceAccountName")
			output, _ := utils.RunInLinux("cat ../../helm/hwameistor/templates/api-server.yaml |grep serviceAccountName |wc -l")
			gomega.Expect(output).To(gomega.Equal("1\n"))
		})
	})

})
