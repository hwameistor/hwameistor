package pool

import "github.com/hwameistor/hwameistor/pkg/local-disk-manager/member/types"

// Pool contains disks, volume, pool info
type Pool struct {
	Name string
}

type Manager interface {
	// Init Storage Pool
	Init() error

	// CreatePool create StoragePool
	CreatePool(poolType types.DevType) error

	// PoolExist returns true if exist
	PoolExist(poolName string) (bool, error)

	// GetPool returns pool info, including capacity, volume counts, disk counts...
	GetPool(poolName string) (*Pool, error)

	// ExtendPool extend StoragePool with new disk
	ExtendPool(poolName string, devLinks []string, serial string) (bool, error)
}
