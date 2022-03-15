package e2e

import (
	_ "github.com/hwameistor/local-storage/test/e2e/SmokeTest"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"math/rand"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UnixNano())
	os.Exit(m.Run())
}

func TestE2E(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "hwameistor e2e test")
}
