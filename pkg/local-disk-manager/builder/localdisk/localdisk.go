package localdisk

import (
	"fmt"

	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/disk/manager"
)

// Builder for localDisk resource
type Builder struct {
	disk *v1alpha1.LocalDisk
	errs []error
}

func NewBuilder() *Builder {
	return &Builder{
		disk: &v1alpha1.LocalDisk{},
	}
}

func (builder *Builder) WithName(name string) *Builder {
	if builder.errs != nil {
		return builder
	}
	builder.disk.Name = name
	return builder
}

func (builder *Builder) SetupAttribute(attribute manager.Attribute) *Builder {
	if builder.errs != nil {
		return builder
	}
	builder.disk.Spec.Capacity = attribute.Capacity
	builder.disk.Spec.DevicePath = attribute.DevName
	builder.disk.Spec.DiskAttributes.Type = attribute.DriverType
	builder.disk.Spec.DiskAttributes.Vendor = attribute.Vendor
	builder.disk.Spec.DiskAttributes.ModelName = attribute.Model
	builder.disk.Spec.DiskAttributes.Protocol = attribute.Bus
	builder.disk.Spec.DiskAttributes.SerialNumber = attribute.Serial
	builder.disk.Spec.DiskAttributes.DevType = attribute.DevType
	builder.disk.Spec.Major = attribute.Major
	builder.disk.Spec.Minor = attribute.Minor
	builder.disk.Spec.DevLinks = attribute.DevLinks

	return builder
}

func (builder *Builder) SetupState() *Builder {
	if builder.errs != nil {
		return builder
	}
	// fixme: update this state by using by health check tool
	builder.disk.Spec.State = v1alpha1.LocalDiskActive

	return builder
}

func (builder *Builder) SetupRaidInfo(raid manager.RaidInfo) *Builder {
	if builder.errs != nil {
		return builder
	}

	// complete RAID INFO here
	return builder
}

func (builder *Builder) SetupSmartInfo(smart manager.SmartInfo) *Builder {
	if builder.errs != nil {
		return builder
	}

	if smart.OverallHealthPassed {
		builder.disk.Spec.SmartInfo.OverallHealth = v1alpha1.AssessPassed
	} else {
		builder.disk.Spec.SmartInfo.OverallHealth = v1alpha1.AssessFailed
	}

	return builder
}

func (builder *Builder) SetupUUID(uuid string) *Builder {
	if builder.errs != nil {
		return builder
	}

	builder.disk.Spec.UUID = uuid
	return builder
}

func (builder *Builder) SetupNodeName(node string) *Builder {
	if builder.errs != nil {
		return builder
	}

	builder.disk.Spec.NodeName = node
	return builder
}

func (builder *Builder) SetupPartitionInfo(originParts []manager.PartitionInfo) *Builder {
	if builder.errs != nil {
		return builder
	}
	for _, part := range originParts {
		builder.disk.Spec.HasPartition = true
		p := v1alpha1.PartitionInfo{}
		p.HasFileSystem = true
		p.Path = part.Path
		p.FileSystem.Type = part.Filesystem
		builder.disk.Spec.PartitionInfo = append(builder.disk.Spec.PartitionInfo, p)
	}
	return builder
}

func (builder *Builder) GenerateStatus() *Builder {
	if builder.errs != nil {
		return builder
	}
	if builder.disk.Spec.HasPartition {
		builder.disk.Status.State = v1alpha1.LocalDiskBound
	} else {
		builder.disk.Status.State = v1alpha1.LocalDiskAvailable
	}
	return builder
}

func (builder *Builder) Build() (v1alpha1.LocalDisk, error) {
	if builder.errs != nil {
		return v1alpha1.LocalDisk{}, fmt.Errorf("%v", builder.errs)
	}

	return *builder.disk, nil
}
