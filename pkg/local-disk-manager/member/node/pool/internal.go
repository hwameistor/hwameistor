package pool

import (
	"fmt"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/member/types"
	"k8s.io/kubernetes/pkg/volume/util/hostutil"
	"os"
	"strings"
)

// diskPool implement interface pool.Manager
type diskPool struct {
	name string
	hu   hostutil.HostUtils
}

func (p *diskPool) Init() error {
	exist, err := p.PoolExist(p.name)
	if err != nil {
		return err
	}
	if exist {
		return nil
	}

	return p.CreatePool(p.name)
}

func (p *diskPool) CreatePool(poolName string) error {
	poolPath := types.GetLocalDiskPoolPath(poolName)
	exist, _ := p.hu.PathExists(poolPath)
	if exist {
		return nil
	}

	// create LocalDisk_Pool{HDD,SSD,NVMe}
	return os.MkdirAll(poolPath, 0755)
}

func (p *diskPool) PoolExist(poolName string) (bool, error) {
	return p.hu.PathExists(types.GetLocalDiskPoolPath(poolName))
}

func (p *diskPool) GetPool(poolName string) (*Pool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *diskPool) ExtendPool(poolName string, devPath string) (bool, error) {
	err := p.CreatePool(poolName)
	if err != nil {
		return false, err
	}

	devName := getDeviceName(devPath)
	if devName == "" {
		return false, fmt.Errorf("devName can't be empty(devPath: %s)", devPath)
	}

	poolDevicePath := types.ComposePoolDevicePath(poolName, devName)
	err = os.Symlink(devPath, poolDevicePath)
	if err != nil {
		return false, err
	}

	return true, nil
}

func New() Manager {
	return &diskPool{
		hu: hostutil.NewHostUtil(),
	}
}

// /dev/sdc -> sda
func getDeviceName(devPath string) string {
	ss := strings.Split(devPath, "/")
	return ss[len(ss)-1]
}
