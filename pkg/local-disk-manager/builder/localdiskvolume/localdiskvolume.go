package localdiskvolume

import (
	"fmt"
	"strings"

	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

type Builder struct {
	volume *v1alpha1.LocalDiskVolume
	errs   []error
}

func NewBuilder() *Builder {
	return &Builder{
		volume: &v1alpha1.LocalDiskVolume{},
	}
}

func NewBuilderFrom(volume *v1alpha1.LocalDiskVolume) *Builder {
	return &Builder{
		volume: volume,
	}
}

func (builder *Builder) WithName(name string) *Builder {
	if err := builder.assertVolumeNotNil(); err != nil {
		return builder
	}

	builder.volume.SetName(name)
	return builder
}

func (builder *Builder) WithFinalizer(finalizer []string) *Builder {
	if err := builder.assertVolumeNotNil(); err != nil {
		return builder
	}

	builder.volume.ObjectMeta.Finalizers = finalizer
	return builder
}

func (builder *Builder) SetupAccessibility(topology v1alpha1.AccessibilityTopology) *Builder {
	if err := builder.assertVolumeNotNil(); err != nil {
		return builder
	}

	builder.volume.Spec.Accessibility = topology
	return builder
}

func (builder *Builder) SetupPVCNameSpaceName(pvc string) *Builder {
	if err := builder.assertVolumeNotNil(); err != nil {
		return builder
	}

	builder.volume.Spec.PersistentVolumeClaimName = pvc
	return builder
}

func (builder *Builder) SetupLocalDiskName(ld string) *Builder {
	if err := builder.assertVolumeNotNil(); err != nil {
		return builder
	}

	builder.volume.Status.LocalDiskName = ld
	return builder
}

func (builder *Builder) SetupDisk(devPath string) *Builder {
	if err := builder.assertVolumeNotNil(); err != nil {
		return builder
	}

	builder.volume.Status.DevPath = devPath
	return builder
}

func (builder *Builder) SetupVolumePath(volPath string) *Builder {
	if err := builder.assertVolumeNotNil(); err != nil {
		return builder
	}

	builder.volume.Status.VolumePath = volPath
	return builder
}

func (builder *Builder) SetupDevSymLinks(devLinks []string) *Builder {
	if err := builder.assertVolumeNotNil(); err != nil {
		return builder
	}

	devLinksMap := make(map[v1alpha1.DevLinkType][]string, 0)
	for _, devLink := range devLinks {
		switch {
		case strings.Contains(devLink, v1alpha1.LinkByPath):
			devLinksMap[v1alpha1.LinkByPath] = append(devLinksMap[v1alpha1.LinkByPath], devLink)
		case strings.Contains(devLink, v1alpha1.LinkByID):
			devLinksMap[v1alpha1.LinkByID] = append(devLinksMap[v1alpha1.LinkByID], devLink)
		case strings.Contains(devLink, v1alpha1.LinkByUUID):
			devLinksMap[v1alpha1.LinkByUUID] = append(devLinksMap[v1alpha1.LinkByUUID], devLink)
		default:
			continue
		}
	}

	builder.volume.Status.DevLinks = devLinksMap
	return builder
}

func (builder *Builder) SetupAllocateCap(caps int64) *Builder {
	if err := builder.assertVolumeNotNil(); err != nil {
		return builder
	}

	builder.volume.Status.AllocatedCapacityBytes = caps
	return builder
}

func (builder *Builder) SetupRequiredCapacityBytes(caps int64) *Builder {
	if err := builder.assertVolumeNotNil(); err != nil {
		return builder
	}

	builder.volume.Spec.RequiredCapacityBytes = caps
	return builder
}

func (builder *Builder) SetupDiskType(diskTpe string) *Builder {
	if err := builder.assertVolumeNotNil(); err != nil {
		return builder
	}

	builder.volume.Spec.DiskType = diskTpe
	return builder
}

func (builder *Builder) SetupStatus(status v1alpha1.State) *Builder {
	if err := builder.assertVolumeNotNil(); err != nil {
		return builder
	}

	builder.volume.Status.State = status
	return builder
}

func (builder *Builder) assertVolumeNotNil() error {
	if builder.volume == nil {
		err := fmt.Errorf("volume object is nil")
		builder.errs = append(builder.errs, err)
		return err
	}

	return nil
}

func (builder *Builder) Build() (*v1alpha1.LocalDiskVolume, error) {
	if builder.errs != nil {
		return nil, fmt.Errorf("%v", builder.errs)
	}

	return builder.volume, nil
}
