package pool

// Pool contains disks, volume, pool info
type Pool struct {
	Name string
}

type Manager interface {
	// Init Storage Pool
	Init() error

	// CreatePool create StoragePool
	CreatePool(poolName string) error

	// PoolExist returns true if exist
	PoolExist(poolName string) (bool, error)

	// GetPool returns pool info, including capacity, volume counts, disk counts...
	GetPool(poolName string) (*Pool, error)

	// ExtendPool extend StoragePool with new disk
	ExtendPool(poolName string, devPath string) (bool, error)
}
