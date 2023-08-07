package localvolumegroup

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// SystemMode of HA module
type SystemMode string

const (
	volumeGroupFinalizer = "hwameistor.io/localvolumegroup-protection"
)

var (
	fakeLocalVolumeGroupName = "local-volume-group-example"
	fakeLocalVolumeGroupUID  = "local-volume-group-convert-uid"
	fakeNamespace            = "local-volume-group-test"
	fakeNodename             = "10-6-118-10"
	fakePersistentPvcName    = "pvc-test"

	fakePods                             = []string{"pod-test1"}
	fakeAccessibility                    = apisv1alpha1.AccessibilityTopology{Nodes: []string{"test-node1"}}
	fakeLocalVolumeName                  = "local-volume-test1"
	fakeVolumes                          = []apisv1alpha1.VolumeInfo{{LocalVolumeName: fakeLocalVolumeName, PersistentVolumeClaimName: fakePersistentPvcName}}
	apiVersion                           = "hwameistor.io/v1alpha1"
	LocalVolumeGroupKind                 = "LocalVolumeGroup"
	fakeRecorder                         = record.NewFakeRecorder(100)
	SystemModeDRBD            SystemMode = "drbd"
	defaultDRBDStartPort                 = 43001
	defaultHAVolumeTotalCount            = 1000
)

// DeleteFakeLocalVolumeGroup
func (r *ReconcileLocalVolumeGroup) DeleteFakeLocalVolumeGroup(t *testing.T, lvg *apisv1alpha1.LocalVolumeGroup) {
	if err := r.client.Delete(context.Background(), lvg); err != nil {
		t.Errorf("Delete LocalVolumeGroup %v fail %v", lvg.GetName(), err)
	}
}

// GenFakeLocalVolumeGroupObject Create lvg request
func GenFakeLocalVolumeGroupObject() *apisv1alpha1.LocalVolumeGroup {
	lvg := &apisv1alpha1.LocalVolumeGroup{}

	TypeMeta := metav1.TypeMeta{
		Kind:       LocalVolumeGroupKind,
		APIVersion: apiVersion,
	}

	ObjectMata := metav1.ObjectMeta{
		Name:              fakeLocalVolumeGroupName,
		Namespace:         fakeNamespace,
		ResourceVersion:   "",
		UID:               types.UID(fakeLocalVolumeGroupUID),
		CreationTimestamp: metav1.Time{Time: time.Now()},
	}

	Spec := apisv1alpha1.LocalVolumeGroupSpec{
		Volumes:       fakeVolumes,
		Accessibility: fakeAccessibility,
		Pods:          fakePods,
	}

	lvg.ObjectMeta = ObjectMata
	lvg.TypeMeta = TypeMeta
	lvg.Spec = Spec
	lvg.Finalizers = append(lvg.Finalizers, volumeGroupFinalizer)
	return lvg
}

// CreateFakeClient Create LocalVolumeGroup resource
func CreateFakeClient() (client.Client, *runtime.Scheme) {
	lvg := GenFakeLocalVolumeGroupObject()
	lvgList := &apisv1alpha1.LocalVolumeGroupList{
		TypeMeta: metav1.TypeMeta{
			Kind:       LocalVolumeGroupKind,
			APIVersion: apiVersion,
		},
	}

	s := scheme.Scheme
	s.AddKnownTypes(apisv1alpha1.SchemeGroupVersion, lvg)
	s.AddKnownTypes(apisv1alpha1.SchemeGroupVersion, lvgList)
	return fake.NewClientBuilder().WithScheme(s).Build(), s
}

func validateSystemConfig() error {
	var errMsgs []string
	switch apisv1alpha1.SystemMode(SystemModeDRBD) {
	case apisv1alpha1.SystemModeDRBD:
	default:
		errMsgs = append(errMsgs, fmt.Sprintf("system mode %s not supported", SystemModeDRBD))
	}

	if len(errMsgs) != 0 {
		return fmt.Errorf(strings.Join(errMsgs, "; "))
	}
	return nil
}

func getSystemConfig() (apisv1alpha1.SystemConfig, error) {
	if err := validateSystemConfig(); err != nil {
		return apisv1alpha1.SystemConfig{}, err
	}

	config := apisv1alpha1.SystemConfig{
		Mode:             apisv1alpha1.SystemMode(SystemModeDRBD),
		MaxHAVolumeCount: defaultHAVolumeTotalCount,
	}

	switch config.Mode {
	case apisv1alpha1.SystemModeDRBD:
		{
			config.DRBD = &apisv1alpha1.DRBDSystemConfig{
				StartPort: defaultDRBDStartPort,
			}
		}
	}
	return config, nil
}
