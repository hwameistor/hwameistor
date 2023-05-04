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

	if t.Failed() {
		logrus.Info("e2e test failed")
		ctx := context.TODO()
		utils.GetAllPodEventsInDefaultNamespace(ctx)
		utils.GetAllPodLogsInHwameistorNamespace(ctx)
	} else {
		logrus.Info("e2e test success")
	}

}
