package localdisknode

import (
	"fmt"

	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

// Builder for LocalDiskNode resource
type Builder struct {
	DiskNode *v1alpha1.LocalDiskNode
	errs     []error
}

func NewBuilder() *Builder {
	return &Builder{
		DiskNode: &v1alpha1.LocalDiskNode{},
	}
}

func (builder *Builder) WithName(name string) *Builder {
	if builder.errs != nil {
		return builder
	}

	builder.DiskNode.Name = name
	return builder
}

func (builder *Builder) SetupAttachNode(node string) *Builder {
	if builder.errs != nil {
		return builder
	}

	builder.DiskNode.Spec.NodeName = node
	return builder
}

func (builder *Builder) Build() (*v1alpha1.LocalDiskNode, error) {
	if builder.errs != nil {
		return nil, fmt.Errorf("%v", builder.errs)
	}

	return builder.DiskNode, nil
}
