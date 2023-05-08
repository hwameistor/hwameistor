package types

const MaxLimitVolume = 100

// Volume
type Volume struct {
	// Name
	Name string `json:"name"`

	// Ready
	Ready bool `json:"ready"`

	// Exist
	Exist bool `json:"exist"`

	// Capacity
	Capacity int64 `json:"capacity"`

	// VolumeContext
	VolumeContext map[string]string

	// AttachNode
	AttachNode string `json:"attachNode"`

	// AttachPath represents where the volume is attached(e.g. /dev/sdb)
	AttachPath string `json:"attachPath"`
}
