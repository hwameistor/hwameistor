package e2e

import (
	"context"
	"github.com/hwameistor/hwameistor/test/e2e/utils"
	"github.com/sirupsen/logrus"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	_ "github.com/hwameistor/hwameistor/test/e2e/E2eTest"
	_ "github.com/hwameistor/hwameistor/test/e2e/adaptation_test"
)

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UnixNano())
	os.Exit(m.Run())
}

func TestE2E(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "hwameistor e2e test")
}

var _ = ginkgo.AfterSuite(func() {
	//判断ginkgo是否成功
	report := ginkgo.CurrentSpecReport()
	logrus.Info(report.State)
	logrus.Info(report.State)
	//输出所有default namespace下pod的events
	ctx := context.TODO()
	utils.GetAllPodEventsInDefaultNamespace(ctx)

})
