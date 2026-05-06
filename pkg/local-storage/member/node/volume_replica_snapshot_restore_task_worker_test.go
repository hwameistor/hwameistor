package node

import (
	"testing"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGetSourceVolumeFromSnapshot(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := apisv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("AddToScheme() error = %v", err)
	}

	snapshot := &apisv1alpha1.LocalVolumeSnapshot{}
	snapshot.Name = "snapshot-a"
	snapshot.Spec.SourceVolume = "volume-a"

	volume := &apisv1alpha1.LocalVolume{}
	volume.Name = "volume-a"

	mgr := &manager{
		apiClient: fake.NewClientBuilder().
			WithScheme(scheme).
			WithObjects(snapshot, volume).
			Build(),
	}

	got, err := mgr.getSourceVolumeFromSnapshot(snapshot.Name)
	if err != nil {
		t.Fatalf("getSourceVolumeFromSnapshot() error = %v", err)
	}
	if got.Name != volume.Name {
		t.Fatalf("getSourceVolumeFromSnapshot() = %q, want %q", got.Name, volume.Name)
	}
}
