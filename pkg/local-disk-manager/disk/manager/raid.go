package manager

// IRaid
type IRaid interface {
	HasRaid() bool

	ParseRaidInfo() RaidInfo
}

// Raid
type RaidParser struct {
	// DiskIdentify Uniquely identify a disk
	DiskIdentify

	IRaid
}

// RaidInfo
type RaidInfo struct {
	// HasRaid
	HasRaid bool
}
