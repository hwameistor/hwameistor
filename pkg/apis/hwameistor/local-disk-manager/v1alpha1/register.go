// NOTE: Boilerplate only.  Ignore this file.

// Package v1alpha1 contains API Schema definitions for the hwameistor v1alpha1 API group
// +k8s:deepcopy-gen=package,register
// +groupName=hwameistor
package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: "hwameistor.io", Version: "v1alpha1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
)

// AddToSchemes may be used to add all resources defined in the project to a Scheme for Local Storage Member
var AddToSchemes runtime.SchemeBuilder

// AddToScheme adds all Resources to the Scheme for Local Storage Member
func AddToScheme(s *runtime.Scheme) error {
	return AddToSchemes.AddToScheme(s)
}

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, SchemeBuilder.AddToScheme)
}

// Kind takes an unqualified kind and returns back a Group qualified GroupKind
func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}
