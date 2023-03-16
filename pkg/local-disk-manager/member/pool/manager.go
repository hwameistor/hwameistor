package pool

// Pool contains disks, volume, pool info
type Pool struct {
}

type Manager interface {
	// Init Storage Pool
	Init() error

	// CreatePool create StoragePool
	CreatePool(poolName string) error

	// PoolExist returns true if exist
	PoolExist(poolName string) bool

	// GetPool returns pool info, including capacity, volume counts, disk counts...
	GetPool(poolName string)
}
