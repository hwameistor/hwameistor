package manager

// IPartition
type IPartition interface {
	HasPartition() bool

	ParsePartitionInfo() []PartitionInfo
}

// Partition
type PartitionParser struct {
	// DiskIdentify Uniquely identify a disk
	//*DiskIdentify

	// IPartition
	IPartition
}

// PartitionInfo
type PartitionInfo struct {
	// Name
	Name string

	// Size
	Size uint64

	// Label
	Label string

	// Filesystem
	Filesystem string
}
