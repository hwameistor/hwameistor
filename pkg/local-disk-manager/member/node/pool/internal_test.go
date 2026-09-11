package pool

import (
	"testing"

	"k8s.io/kubernetes/pkg/volume/util/hostutil"

	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/member/types"
)

func TestExtendPoolReturnsExistingPoolDevice(t *testing.T) {
	const (
		poolName = "LocalDisk_PoolHDD"
		devLink  = "/dev/disk/by-path/pci-test"
		devName  = "pci-test"
	)
	poolDevicePath := types.ComposePoolDevicePath(poolName, devName)
	p := &diskPool{
		hu: hostutil.NewFakeHostUtil(map[string]hostutil.FileType{
			poolDevicePath: hostutil.FileTypeBlockDev,
		}),
	}

	exists, err := p.ExtendPool(poolName, []string{devLink}, "")
	if err != nil {
		t.Fatalf("ExtendPool() error = %v", err)
	}
	if !exists {
		t.Fatal("ExtendPool() should report the existing pool device")
	}
}
