package pool

import (
	"fmt"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/member/types"
	log "github.com/sirupsen/logrus"
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
	for _, poolType := range types.DefaultDevTypes {
		if err := p.CreatePool(poolType); err != nil {
			return err
		}
	}
	return nil
}

func (p *diskPool) CreatePool(poolType types.DevType) error {
	for _, poolPath := range []string{types.GetLocalDiskPoolPath(poolType), types.GetPoolDiskPath(poolType), types.GetPoolVolumePath(poolType)} {
		exist, _ := p.hu.PathExists(poolPath)
		if !exist {
			if err := os.MkdirAll(poolPath, 0755); err != nil {
				return err
			}
			log.Debugf("Succeed to create %s(mode: 0755) directory", poolPath)
			continue
		}
		log.Debugf("Directory %s already exist", poolPath)
	}
	return nil
}

func (p *diskPool) PoolExist(poolName string) (bool, error) {
	return p.hu.PathExists(types.GetLocalDiskPoolPath(poolName))
}

func (p *diskPool) GetPool(poolName string) (*Pool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *diskPool) ExtendPool(poolName string, devPath string) (bool, error) {
	devName := getDeviceName(devPath)
	if devName == "" {
		return false, fmt.Errorf("devName can't be empty(devPath: %s)", devPath)
	}

	poolDevicePath := types.ComposePoolDevicePath(poolName, devName)
	err := os.Symlink(devPath, poolDevicePath)
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
