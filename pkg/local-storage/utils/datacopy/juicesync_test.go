package datacopy

import (
	"context"
	"reflect"
	"testing"

	apis "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func testJuiceSync(t *testing.T) *JuiceSync {
	t.Helper()
	scheme := runtime.NewScheme()
	for _, add := range []func(*runtime.Scheme) error{corev1.AddToScheme, batchv1.AddToScheme, apis.AddToScheme} {
		if err := add(scheme); err != nil {
			t.Fatal(err)
		}
	}
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: GetConfigMapName(SyncConfigMapName, "volume"), Namespace: "test"},
		Data:       map[string]string{SyncConfigSourceNodeNameKey: "source", SyncConfigTargetNodeNameKey: "target", juiceFSVersionEnv: "ignored"},
	}
	source := &apis.LocalStorageNode{ObjectMeta: metav1.ObjectMeta{Name: "source"}}
	source.Spec.StorageIP = "192.0.2.1"
	target := &apis.LocalStorageNode{ObjectMeta: metav1.ObjectMeta{Name: "target"}}
	target.Spec.StorageIP = "192.0.2.2"
	return &JuiceSync{namespace: "test", apiClient: fake.NewClientBuilder().WithScheme(scheme).WithObjects(cm, source, target).Build()}
}

func TestJuiceSyncJobConfiguration(t *testing.T) {
	for _, tc := range []struct {
		name, version, image, wantVersion, script string
		check                                     bool
	}{
		{"default", "", "", "1.2.0", "sync_hwameistor_volumes_check.sh", true},
		{"explicit-baseline", "1.2.0", "example.com/sync:candidate", "1.2.0", "sync_hwameistor_volumes_check.sh", true},
		{"alternate", "1.4.1", "", "1.4.1", "sync_hwameistor_volumes_check.sh", true},
		{"check-disabled", "1.2.0", "", "1.2.0", "sync_hwameistor_volumes.sh", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(juiceFSVersionEnv, tc.version)
			t.Setenv("MIGRAGE_JUICESYNC_IMAGE", tc.image)
			js := testJuiceSync(t)
			if err := js.StartSync("copy", "volume", "source", "", tc.check); err != nil {
				t.Fatal(err)
			}
			job := &batchv1.Job{}
			if err := js.apiClient.Get(context.Background(), types.NamespacedName{Namespace: "test", Name: "copy"}, job); err != nil {
				t.Fatal(err)
			}
			container := job.Spec.Template.Spec.Containers[0]
			if !reflect.DeepEqual(container.Command, []string{"sh", "-c", tc.script}) {
				t.Fatalf("command: %v", container.Command)
			}
			image := tc.image
			if image == "" {
				image = juiceSyncImageName
			}
			if container.Image != image {
				t.Fatalf("image: %q, want %q", container.Image, image)
			}
			env := map[string]string{}
			for _, item := range container.Env {
				env[item.Name] = item.Value
			}
			if env[juiceFSVersionEnv] != tc.wantVersion || env[SyncConfigSourceNodeNameKey] != "192.0.2.1" || env[SyncConfigTargetNodeNameKey] != "192.0.2.2" {
				t.Fatalf("Job env did not preserve resolved version/IPs: %v", env)
			}
		})
	}
}

func TestJuiceSyncInvalidConfigurationDoesNotCreateJob(t *testing.T) {
	for _, version := range []string{"latest", "v1.2.0", "../../bin/sh"} {
		t.Run(version, func(t *testing.T) {
			t.Setenv(juiceFSVersionEnv, version)
			js := testJuiceSync(t)
			if err := js.StartSync("copy", "volume", "source", "", true); err == nil {
				t.Fatal("invalid version accepted")
			}
			jobs := &batchv1.JobList{}
			if err := js.apiClient.List(context.Background(), jobs); err != nil {
				t.Fatal(err)
			}
			if len(jobs.Items) != 0 {
				t.Fatal("created Job with invalid version")
			}
		})
	}
	t.Setenv(juiceFSVersionEnv, "1.2.0")
	if err := testJuiceSync(t).StartSync("copy", "missing-volume", "source", "", true); err == nil {
		t.Fatal("missing config map did not return an error")
	}
}
