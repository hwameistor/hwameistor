package localdiskactioncontroller

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/hwameistor/hwameistor/pkg/apis/client/clientset/versioned/fake"
	informers "github.com/hwameistor/hwameistor/pkg/apis/client/informers/externalversions"
	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/diff"
	core "k8s.io/client-go/testing"
)

var (
	alwaysReady        = func() bool { return true }
	noResyncPeriodFunc = func() time.Duration { return 0 }
	Mi                 = int64(1024 * 1024)
)

type fixture struct {
	t *testing.T

	client *fake.Clientset
	// Objects to put in the store.
	ldLister  []*v1alpha1.LocalDisk
	ldalister []*v1alpha1.LocalDiskAction
	// Actions expected to happen on the client.
	actions []core.Action
	// Objects from here preloaded into NewSimpleFake.
	objects []runtime.Object
}

func newFixture(t *testing.T) *fixture {
	f := &fixture{}
	f.t = t
	return f
}

func newLd(name string, devicePath string, capacity int64) *v1alpha1.LocalDisk {
	return &v1alpha1.LocalDisk{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.LocalDiskSpec{
			DevicePath: devicePath,
			Capacity:   capacity,
		},
	}
}

func newLda(name string, rule v1alpha1.LocalDiskActionRule) *v1alpha1.LocalDiskAction {
	return &v1alpha1.LocalDiskAction{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.LocalDiskActionSpec{
			Rule:   rule,
			Action: v1alpha1.LocalDiskActionReserve,
		},
	}
}

func (f *fixture) newController() (*LocalDiskActionController, informers.SharedInformerFactory) {
	f.client = fake.NewSimpleClientset(f.objects...)

	factory := informers.NewSharedInformerFactory(f.client, noResyncPeriodFunc())

	c := NewLocalDiskActionController(f.client,
		factory.Hwameistor().V1alpha1().LocalDisks(),
		factory.Hwameistor().V1alpha1().LocalDiskActions())

	c.localDisksSynced = alwaysReady
	c.localDiskActionsSynced = alwaysReady

	for _, ld := range f.ldLister {
		factory.Hwameistor().V1alpha1().LocalDisks().Informer().GetIndexer().Add(ld)
	}

	for _, lda := range f.ldalister {
		factory.Hwameistor().V1alpha1().LocalDiskActions().Informer().GetIndexer().Add(lda)
	}

	return c, factory
}

// filterInformerActions filters list and watch actions for testing resources.
// Since list and watch don't change resource state we can filter it to lower
// nose level in our tests.
func filterInformerActions(actions []core.Action) []core.Action {
	ret := []core.Action{}
	for _, action := range actions {
		if len(action.GetNamespace()) == 0 &&
			(action.Matches("list", "localdisks") ||
				action.Matches("watch", "localdisks") ||
				action.Matches("list", "localdiskactions") ||
				action.Matches("watch", "localdiskactions")) {
			continue
		}
		ret = append(ret, action)
	}

	return ret
}

func (f *fixture) runLDAHandler(ldaName string, startInformers bool, expectError bool) {
	ctx := context.TODO()
	c, factory := f.newController()
	if startInformers {
		factory.Start(ctx.Done())
	}

	err := c.syncLDAHandler(ctx, ldaName)
	if !expectError && err != nil {
		f.t.Errorf("error syncing foo: %v", err)
	} else if expectError && err == nil {
		f.t.Error("expected error syncing foo, got nil")
	}

	actions := filterInformerActions(f.client.Actions())
	for i, action := range actions {
		if len(f.actions) < i+1 {
			f.t.Errorf("%d unexpected actions: %+v", len(actions)-len(f.actions), actions[i:])
			break
		}

		expectedAction := f.actions[i]
		checkAction(expectedAction, action, f.t)
	}

	if len(f.actions) > len(actions) {
		f.t.Errorf("%d additional expected actions:%+v", len(f.actions)-len(actions), f.actions[len(actions):])
	}
}

func (f *fixture) runLDHandler(ldName string, startInformers bool, expectError bool) {
	ctx := context.TODO()
	c, factor := f.newController()
	if startInformers {
		factor.Start(ctx.Done())
	}

	err := c.syncLDHandler(ctx, ldName)
	if !expectError && err != nil {
		f.t.Errorf("error syncing foo: %v", err)
	} else if expectError && err == nil {
		f.t.Error("expected error syncing foo, got nil")
	}

	actions := filterInformerActions(f.client.Actions())
	for i, action := range actions {
		if len(f.actions) < i+1 {
			f.t.Errorf("%d unexpected actions: %+v", len(actions)-len(f.actions), actions[i:])
			break
		}

		expectedAction := f.actions[i]
		checkAction(expectedAction, action, f.t)
	}

	if len(f.actions) > len(actions) {
		f.t.Errorf("%d additional expected actions:%+v", len(f.actions)-len(actions), f.actions[len(actions):])
	}
}

func TestLDDoNothing(t *testing.T) {
	f := newFixture(t)

	ld := newLd("testLd", "/dev/sda", 1024*Mi)
	lda := newLda("testLda", v1alpha1.LocalDiskActionRule{})

	f.objects = append(f.objects, ld, lda)
	f.ldalister = append(f.ldalister, lda)
	f.ldLister = append(f.ldLister, ld)

	f.runLDHandler("testLd", true, false)
}

func TestLDADoNothing(t *testing.T) {
	f := newFixture(t)

	ld := newLd("testLd", "/dev/sda", 1024*Mi)
	lda := newLda("testLda", v1alpha1.LocalDiskActionRule{})

	f.objects = append(f.objects, ld, lda)
	f.ldalister = append(f.ldalister, lda)
	f.ldLister = append(f.ldLister, ld)

	f.runLDAHandler("testLda", true, false)
}

func TestLDFilterByMaxCapacity(t *testing.T) {
	f := newFixture(t)

	ld := newLd("testLd", "/dev/rbd0", 1024*Mi)
	lda := newLda("testLda", v1alpha1.LocalDiskActionRule{
		MaxCapacity: 1000 * Mi,
	})

	f.objects = append(f.objects, ld, lda)
	f.ldalister = append(f.ldalister, lda)
	f.ldLister = append(f.ldLister, ld)

	f.expectReserveLdAction(ld)
	f.expectLdaActionAddLd(lda, ld.Name)
	f.runLDHandler("testLd", true, false)
}

func TestLDFilterByMinCapacity(t *testing.T) {
	f := newFixture(t)

	ld := newLd("testLd", "/dev/rbd0", 1*Mi)
	lda := newLda("testLda", v1alpha1.LocalDiskActionRule{
		MinCapacity: 100 * Mi,
	})

	f.objects = append(f.objects, ld, lda)
	f.ldalister = append(f.ldalister, lda)
	f.ldLister = append(f.ldLister, ld)

	f.expectReserveLdAction(ld)
	f.expectLdaActionAddLd(lda, ld.Name)
	f.runLDHandler("testLd", true, false)
}

func TestLDFilterByDevicePath(t *testing.T) {
	f := newFixture(t)

	ld := newLd("testLd", "/dev/rbd0", 1024*Mi)
	lda := newLda("testLda", v1alpha1.LocalDiskActionRule{
		DevicePath: "/dev/rbd*",
	})

	f.objects = append(f.objects, ld, lda)
	f.ldalister = append(f.ldalister, lda)
	f.ldLister = append(f.ldLister, ld)

	f.expectReserveLdAction(ld)
	f.expectLdaActionAddLd(lda, ld.Name)
	f.runLDHandler("testLd", true, false)
}

func TestLDFilterByLdas(t *testing.T) {
	f := newFixture(t)

	ld := newLd("testLd", "/dev/rbd0", 1024*Mi)
	lda1 := newLda("testLda1", v1alpha1.LocalDiskActionRule{
		DevicePath: "/dev/sda*",
	})
	lda2 := newLda("testLda2", v1alpha1.LocalDiskActionRule{
		DevicePath: "/dev/nbd*",
	})
	lda3 := newLda("testLda3", v1alpha1.LocalDiskActionRule{
		DevicePath: "/dev/rbd*",
	})

	f.objects = append(f.objects, ld, lda1, lda2, lda3)
	f.ldalister = append(f.ldalister, lda1, lda2, lda3)
	f.ldLister = append(f.ldLister, ld)

	f.expectReserveLdAction(ld)
	f.expectLdaActionAddLd(lda3, ld.Name)
	f.runLDHandler("testLd", true, false)
}

func TestLDAFilterByLd(t *testing.T) {
	f := newFixture(t)

	lda := newLda("testLda", v1alpha1.LocalDiskActionRule{
		DevicePath: "/dev/rbd*",
	})
	ld1 := newLd("testLd1", "/dev/sda", 1024*Mi)
	ld2 := newLd("testLd2", "/dev/nbd1", 1024*Mi)
	ld3 := newLd("testLd3", "/dev/rbd0", 1024*Mi)

	f.objects = append(f.objects, lda, ld3)
	f.ldalister = append(f.ldalister, lda)
	f.ldLister = append(f.ldLister, ld1, ld2, ld3)

	f.expectReserveLdAction(ld3)
	f.expectLdaActionAddLd(lda, ld3.Name)
	f.runLDAHandler("testLda", true, false)
}

func (f *fixture) expectReserveLdAction(ld *v1alpha1.LocalDisk) {
	patchBytes := []byte("{\"spec\":{\"reserved\":true}}")
	f.actions = append(f.actions, core.NewPatchAction(schema.GroupVersionResource{Resource: "localdisks"}, ld.Namespace, ld.Name, types.MergePatchType, patchBytes))
}

func (f *fixture) expectLdaActionAddLd(lda *v1alpha1.LocalDiskAction, ldName string) {
	patchBytes := []byte(fmt.Sprintf("{\"status\":{\"latestMatchedLds\":[\"%s\"]}}", ldName))
	f.actions = append(f.actions, core.NewPatchSubresourceAction(schema.GroupVersionResource{Resource: "localdiskactions"}, lda.Namespace, lda.Name, types.MergePatchType, patchBytes, "status"))
}

// checkAction verifies that expected and actual actions are equal and both have
// same attached resources
func checkAction(expected, actual core.Action, t *testing.T) {
	if !(expected.Matches(actual.GetVerb(), actual.GetResource().Resource) && actual.GetSubresource() == expected.GetSubresource()) {
		t.Errorf("Expected\n\t%#v\ngot\n\t%#v", expected, actual)
		return
	}

	if reflect.TypeOf(actual) != reflect.TypeOf(expected) {
		t.Errorf("Action has wrong type. Expected: %t. Got: %t", expected, actual)
		return
	}

	switch a := actual.(type) {
	case core.CreateActionImpl:
		e, _ := expected.(core.CreateActionImpl)
		expObject := e.GetObject()
		object := a.GetObject()

		if !reflect.DeepEqual(expObject, object) {
			t.Errorf("Action %s %s has wrong object\nDiff:\n %s",
				a.GetVerb(), a.GetResource().Resource, diff.ObjectGoPrintSideBySide(expObject, object))
		}
	case core.UpdateActionImpl:
		e, _ := expected.(core.UpdateActionImpl)
		expObject := e.GetObject()
		object := a.GetObject()

		if !reflect.DeepEqual(expObject, object) {
			t.Errorf("Action %s %s has wrong object\nDiff:\n %s",
				a.GetVerb(), a.GetResource().Resource, diff.ObjectGoPrintSideBySide(expObject, object))
		}
	case core.PatchActionImpl:
		e, _ := expected.(core.PatchActionImpl)
		expPatch := e.GetPatch()
		patch := a.GetPatch()

		if !reflect.DeepEqual(expPatch, patch) {
			t.Errorf("Action %s %s has wrong patch\nDiff:\n %s",
				a.GetVerb(), a.GetResource().Resource, diff.ObjectGoPrintSideBySide(expPatch, patch))
		}
	default:
		t.Errorf("Uncaptured Action %s %s, you should explicitly add a case to capture it",
			actual.GetVerb(), actual.GetResource().Resource)
	}
}
