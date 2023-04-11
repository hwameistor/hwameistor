package pool

import (
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/member/types"
	"k8s.io/kubernetes/pkg/volume/util/hostutil"
	"os"
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

func (p *diskPool) ExtendPool(poolName string, disk types.Disk) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func NewDiskPool(poolName string) Manager {
	return &diskPool{
		name: poolName,
		hu:   hostutil.NewHostUtil(),
	}
}
